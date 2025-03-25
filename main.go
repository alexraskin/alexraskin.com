package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func main() {

	port := flag.String("port", "8000", "port to listen on")
	flag.Parse()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cacheControl)
	r.Use(middleware.Heartbeat("/ping"))

	r.Use(httprate.Limit(
		10,
		time.Minute,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "public/429.html")
		}),
	))

	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/index.html")
	})

	r.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})

	r.Get("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/sitemap.xml")
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/404.html")
	})

	log.Printf("Server starting on http://localhost:%s", *port)
	if err := http.ListenAndServe(":"+*port, r); err != nil {
		log.Fatal(err)
	}
}

func cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		next.ServeHTTP(w, r)
	})
}
