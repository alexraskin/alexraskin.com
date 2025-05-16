package alexraskin

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

var statsStartTime = time.Now()

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
	r.Handle("/favicon.ico", s.serveFile(s.assets, "assets/icons/images/favicon.ico"))
	r.Get("/", s.index)
	r.Head("/", s.index)
	r.Get("/version", s.getVersion)
	r.Get("/stats", s.stats)

	r.Group(func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Route("/lastfm", func(r chi.Router) {
				r.Get("/", s.lastfm)
			})
		})
	})

	r.NotFound(s.notFound)

	return r
}

func (s *Server) getVersion(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(s.version))
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	readme, err := s.fetchGitHubProfile()
	if err != nil {
		s.logger.Error("failed to fetch profile", slog.Any("error", err))
		s.renderError(w, r, "Failed to fetch profile", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := s.md.Convert([]byte(readme), &buf); err != nil {
		s.logger.Error("failed to convert markdown", slog.Any("error", err))
		s.renderError(w, r, "Failed to convert markdown", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Home: HomeData{
			Content: template.HTML(buf.String()),
		},
	}

	err = s.tmplFunc(w, "index.gohtml", data)
	if err != nil {
		s.logger.Error("template execution failed", slog.Any("error", err))
		s.renderError(w, r, "Failed to render template", http.StatusInternalServerError)
	}
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.renderError(w, r, "Page not found", http.StatusNotFound)
}

func (s *Server) renderError(w http.ResponseWriter, r *http.Request, message string, status int) {
	requestID := middleware.GetReqID(r.Context())
	if requestID == "" {
		requestID = "unknown"
	}

	data := PageData{
		Error:     message,
		Status:    status,
		Path:      r.URL.Path,
		RequestID: requestID,
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

func (s *Server) lastfm(w http.ResponseWriter, r *http.Request) {
	track, err := s.fetchLastFMTrack()
	if err != nil {
		s.logger.Error("failed to fetch lastfm data", slog.Any("error", err))
		track = nil
	}

	data := struct {
		Track *LastFMTrack
	}{
		Track: track,
	}

	if err := s.tmplFunc(w, "lastfm.gohtml", data); err != nil {
		s.logger.Error("failed to execute lastfm template", slog.Any("error", err))
	}
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	page := struct {
		Go               string
		Uptime           string
		MemoryUsed       string
		TotalMemory      string
		GarbageCollected string
		Goroutines       int
	}{
		Go:               runtime.Version(),
		Uptime:           s.getDurationString(time.Since(statsStartTime)),
		MemoryUsed:       humanize.Bytes(stats.Alloc),
		TotalMemory:      humanize.Bytes(stats.Sys),
		GarbageCollected: humanize.Bytes(stats.TotalAlloc),
		Goroutines:       runtime.NumGoroutine(),
	}

	err := s.tmplFunc(w, "stats.gohtml", page)
	if err != nil {
		s.logger.Error("failed to execute stats template", slog.Any("error", err))
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

func (s *Server) getDurationString(duration time.Duration) string {
	return fmt.Sprintf(
		"%0.2d:%02d:%02d",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		int(duration.Seconds())%60,
	)
}
