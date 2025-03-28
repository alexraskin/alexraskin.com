package alexraskin

import (
	"io"
	"net/http"
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

	r.Use(httprate.Limit(
		10,
		time.Minute,
		httprate.WithLimitHandler(serveFile(s.public, "429.html")),
	))

	r.Mount("/assets", http.StripPrefix("/assets", http.FileServer(s.assets)))

	r.Handle("/robots.txt", serveFile(s.public, "robots.txt"))
	r.Handle("/sitemap.xml", serveFile(s.public, "sitemap.xml"))

	r.Get("/", serveFile(s.public, "index.html"))

	r.NotFound(serveFile(s.public, "404.html"))

	return r
}

func serveFile(fs http.FileSystem, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := fs.Open(path)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()
		_, _ = io.Copy(w, file)
	}
}
