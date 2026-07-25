package server

import (
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

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(s.cacheControl)

	r.Use(httprate.Limit(
		100,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByRealIP),
		httprate.WithLimitHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		})),
	))

	r.Mount("/assets", http.FileServer(s.assets))
	r.Handle("/robots.txt", s.serveFile(s.assets, "assets/robots.txt"))
	r.Handle("/favicon.ico", s.serveFile(s.assets, "assets/images/favicon.ico"))
	r.Get("/", s.index)
	r.Head("/", s.index)
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

// One page only, so anything unknown goes home instead of 404ing.
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
		_, _ = io.Copy(w, file)
	}
}

func (s *Server) cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=86400")
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		next.ServeHTTP(w, r)
	})
}
