package server

import (
	"bytes"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

const (
	assetsDir         = "assets"
	assetPrefix       = "/" + assetsDir + "/"
	assetVersionParam = "v"
)

// Go's built-in table has no woff2, and the alpine image ships no
// /etc/mime.types, so the font would otherwise be sniffed as octet-stream.
func init() {
	_ = mime.AddExtensionType(".woff2", "font/woff2")
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(s.cacheControl)
	r.Use(s.assetETag)

	r.Use(httprate.Limit(
		100,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByRealIP),
		httprate.WithLimitHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		})),
	))

	r.Mount("/assets", s.serveAssets(http.FileServer(s.assets)))
	r.Handle("/robots.txt", s.serveFile(s.assets, "assets/robots.txt"))
	r.Handle("/favicon.ico", s.serveFile(s.assets, "assets/images/favicon.ico"))
	r.Get("/", s.index)
	r.Head("/", s.index)
	r.Get("/franzbroetchen", s.franzbroetchen)
	r.Head("/franzbroetchen", s.franzbroetchen)
	r.Get("/version", s.getVersion)

	r.NotFound(s.notFound)

	return r
}

func (s *Server) getVersion(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(s.version.Format()))
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	// A missing track just drops the "listening to" line from the page.
	track, err := s.fetchLastFMTrack()
	if err != nil {
		s.logger.Error("failed to fetch lastfm data", slog.Any("error", err))
	}

	err = s.tmplFunc(w, "index.gohtml", PageData{Track: track})
	if err != nil {
		s.logger.Error("template execution failed", slog.Any("error", err))
		s.renderError(w, r, "Failed to render template", http.StatusInternalServerError)
	}
}

func (s *Server) franzbroetchen(w http.ResponseWriter, r *http.Request) {
	reviews, err := s.reviewsFunc()
	if err != nil {
		s.logger.Error("failed to load reviews", slog.Any("error", err))
		s.renderError(w, r, "Failed to load reviews", http.StatusInternalServerError)
		return
	}

	// Rendering into a buffer keeps the error page reachable when a template
	// fails, and gives the bytes an ETag: the page only changes when the review
	// file or an asset URL does, so a repeat visit costs a 304 instead of a
	// body. Cache-Control stays no-cache, which is revalidate-then-304.
	var page bytes.Buffer
	if err := s.tmplFunc(&page, "franzbroetchen.gohtml", ReviewsPageData{Reviews: reviews}); err != nil {
		s.logger.Error("template execution failed", slog.Any("error", err))
		s.renderError(w, r, "Failed to render template", http.StatusInternalServerError)
		return
	}

	if serveNotModified(w, r, hashBytes(page.Bytes())) {
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page.Bytes())
}

// Only a couple of pages, so anything unknown goes home instead of 404ing.
func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) renderError(w http.ResponseWriter, r *http.Request, message string, status int) {
	data := PageData{
		Error:  message,
		Status: status,
	}

	w.WriteHeader(status)
	err := s.tmplFunc(w, "error.gohtml", data)
	if err != nil {
		s.logger.Error("error template execution failed",
			slog.Any("error", err),
			slog.String("original_error", message),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) serveFile(fs http.FileSystem, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := fs.Open(path)
		if err != nil {
			s.logger.Error("file not found", slog.String("path", path), slog.Any("error", err))
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()
		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		// These are mounted at the root, so the asset middleware never sees
		// them under their embedded path.
		if asset, ok := s.assetHashes["/"+path]; ok && serveNotModified(w, r, asset.Hash) {
			return
		}

		_, _ = io.Copy(w, file)
	}
}

func (s *Server) cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, assetPrefix):
			// A versioned URL names its own content, so a change ships a new
			// URL rather than waiting out a TTL. Unversioned requests are
			// either stale markup or hand-typed, so they stay short-lived.
			if r.URL.Query().Get(assetVersionParam) != "" {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "public, max-age=3600")
			}
		case r.URL.Path == "/favicon.ico", r.URL.Path == "/robots.txt":
			// Fixed paths that can't carry a version, but revalidate cheaply
			// against their ETag.
			w.Header().Set("Cache-Control", "public, max-age=86400")
		default:
			// The page carries live last.fm data, so it always revalidates.
			// no-store would also cost the back/forward cache.
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}

// Assets are served from an embed.FS, whose files report a zero ModTime, so
// net/http emits no Last-Modified and never generates an ETag itself. Without
// this a revalidation has nothing to compare and refetches the whole body.
func (s *Server) assetETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asset, ok := s.assetHashes[r.URL.Path]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		if serveNotModified(w, r, asset.Hash) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

// serveNotModified sets the ETag and reports whether the request was answered
// with a 304.
func serveNotModified(w http.ResponseWriter, r *http.Request, hash string) bool {
	etag := `"` + hash + `"`
	w.Header().Set("ETag", etag)

	if !etagMatches(r.Header.Get("If-None-Match"), etag) {
		return false
	}

	w.WriteHeader(http.StatusNotModified)
	return true
}

func etagMatches(header, etag string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" || strings.TrimPrefix(candidate, "W/") == etag {
			return true
		}
	}
	return false
}

// serveAssets answers from the in-memory asset set, falling back to the file
// server that dev mode reads off disk.
func (s *Server) serveAssets(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.assetHashes.serve(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}
