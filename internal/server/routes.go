package server

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func (s *Server) Routes() http.Handler {
	r := http.NewServeMux()
	files := http.FileServer(s.assets)
	r.Handle("GET /assets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=604800")
		files.ServeHTTP(w, r)
	}))
	for route, path := range map[string]string{
		"/robots.txt":  "/assets/robots.txt",
		"/sitemap.xml": "/assets/sitemap.xml",
		"/favicon.ico": "/assets/images/favicon.ico",
	} {
		r.HandleFunc("GET "+route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=604800")
			clone := r.Clone(r.Context())
			clone.URL.Path = path
			files.ServeHTTP(w, clone)
		})
	}
	r.HandleFunc("GET /{$}", s.index)
	r.HandleFunc("GET /franzbroetchen", s.franzbroetchen)
	r.HandleFunc("GET /version", s.getVersion)
	r.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			_, _ = io.WriteString(w, ".")
		}
	})
	r.HandleFunc("GET /", s.notFound)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		defer func() {
			if err := recover(); err != nil {
				if err == http.ErrAbortHandler {
					panic(err)
				}
				s.logger.Error("request panic", "error", err)
				w.Header().Set("Cache-Control", "no-store")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			s.logger.Info("request", "method", req.Method, "path", req.URL.Path, "peer", req.RemoteAddr, "duration", time.Since(start))
		}()
		w.Header().Set("Cache-Control", "no-cache")
		r.ServeHTTP(w, req)
	})
}

func (s *Server) getVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodHead {
		_, _ = w.Write([]byte(s.version.Format()))
	}
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	track, err := s.lastFMTrack(r.Context())
	if err != nil {
		s.logger.Error("failed to fetch lastfm data", slog.Any("error", err))
	}

	s.renderPage(w, r, "index.gohtml", PageData{Track: track})
}

func (s *Server) franzbroetchen(w http.ResponseWriter, r *http.Request) {
	reviews, err := s.reviewsFunc()
	if err != nil {
		s.logger.Error("failed to load reviews", slog.Any("error", err))
		s.renderError(w, r, "Failed to load reviews", http.StatusInternalServerError)
		return
	}

	s.renderPage(w, r, "franzbroetchen.gohtml", ReviewsPageData{Reviews: reviews, CDNBase: s.cdnBase})
}

func (s *Server) renderPage(w http.ResponseWriter, r *http.Request, name string, data any) {
	var page bytes.Buffer
	if err := s.tmplFunc(&page, name, data); err != nil {
		s.logger.Error("template execution failed", slog.Any("error", err))
		s.renderError(w, r, "Failed to render template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method != http.MethodHead {
		_, _ = w.Write(page.Bytes())
	}
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) renderError(w http.ResponseWriter, r *http.Request, message string, status int) {
	data := PageData{
		Error:  message,
		Status: status,
	}

	w.Header().Set("Cache-Control", "no-store")
	var page bytes.Buffer
	err := s.tmplFunc(&page, "error.gohtml", data)
	if err != nil {
		s.logger.Error("error template execution failed",
			slog.Any("error", err),
			slog.String("original_error", message),
		)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		if r.Method != http.MethodHead {
			_, _ = io.WriteString(w, "Internal Server Error\n")
		}
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_, _ = w.Write(page.Bytes())
	}
}
