package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexraskin/alexraskin.com/internal/server"
	"github.com/alexraskin/alexraskin.com/internal/ver"
)

var (
	//go:embed templates/**
	Templates embed.FS

	//go:embed assets
	Assets embed.FS

	//go:embed data
	Data embed.FS
)

const cdnBase string = "https://cdn.alexraskin.com"

func main() {
	port := flag.String("port", "8000", "port to listen on")
	devMode := flag.Bool("dev", false, "run in dev mode")
	flag.Parse()

	var (
		tmplFunc    server.ExecuteTemplateFunc
		reviewsFunc server.ReviewsFunc
		assets      http.FileSystem
	)

	logger := slog.Default()
	if *devMode {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	version := ver.Load()

	logger.Debug("Starting alexraskin.com...",
		slog.String("version", version.Version),
		slog.String("commit", version.Revision),
		slog.String("buildTime", version.BuildTime),
	)

	if *devMode {
		logger.Debug("running in dev mode")
		tmplFunc = func(wr io.Writer, name string, data any) error {
			tmpl, err := template.New("").ParseGlob("templates/*.gohtml")
			if err != nil {
				return err
			}
			return tmpl.ExecuteTemplate(wr, name, data)
		}
		assets = http.Dir(".")
		reviewsFunc = func() ([]server.Review, error) {
			return server.LoadReviews(os.DirFS("."), cdnBase)
		}
	} else {
		tmpl, err := template.New("").ParseFS(Templates, "templates/*.gohtml")
		if err != nil {
			logger.Error("failed to parse templates", slog.Any("error", err))
			os.Exit(-1)
		}
		tmplFunc = tmpl.ExecuteTemplate
		assets = http.FS(Assets)

		reviews, err := server.LoadReviews(Data, cdnBase)
		if err != nil {
			logger.Error("failed to load reviews", slog.Any("error", err))
			os.Exit(-1)
		}
		reviewsFunc = func() ([]server.Review, error) { return reviews, nil }
	}

	httpClient := &http.Client{
		Timeout: 3 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	app := server.NewServer(server.Config{
		Version:         version,
		HTTPClient:      httpClient,
		Assets:          assets,
		CDNBase:         cdnBase,
		ExecuteTemplate: tmplFunc,
		Reviews:         reviewsFunc,
		Logger:          logger,
	})
	srv := &http.Server{
		Addr:              ":" + *port,
		Handler:           app.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.ListenAndServe() }()
	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", slog.Any("error", err))
			os.Exit(1)
		}
		return
	case <-ctx.Done():
	}
	stop()

	logger.Debug("shutting down web server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("err", err))
		_ = srv.Close()
	}
}
