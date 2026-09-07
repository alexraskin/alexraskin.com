package main

import (
	"context"
	"embed"
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

func main() {
	port := flag.String("port", "8000", "port to listen on")
	devMode := flag.Bool("dev", false, "run in dev mode")
	cdnBase := flag.String("cdn", cdnBaseURL(), "base URL the review photos are served from")
	flag.Parse()

	var (
		tmplFunc    server.ExecuteTemplateFunc
		reviewsFunc server.ReviewsFunc
		assets      http.FileSystem
		assetHashes server.AssetHashes
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
		assetHashes = server.AssetHashes{}
		tmplFunc = func(wr io.Writer, name string, data any) error {
			tmpl, err := template.New("").Funcs(assetFuncs(assetHashes)).ParseGlob("templates/*.gohtml")
			if err != nil {
				return err
			}
			return tmpl.ExecuteTemplate(wr, name, data)
		}
		assets = http.Dir(".")
		reviewsFunc = func() ([]server.Review, error) {
			return server.LoadReviews(os.DirFS("."), *cdnBase)
		}
	} else {
		var err error
		assetHashes, err = server.HashAssets(Assets)
		if err != nil {
			logger.Error("failed to hash assets", slog.Any("error", err))
			os.Exit(-1)
		}

		tmpl, err := template.New("").Funcs(assetFuncs(assetHashes)).ParseFS(Templates, "templates/*.gohtml")
		if err != nil {
			logger.Error("failed to parse templates", slog.Any("error", err))
			os.Exit(-1)
		}
		tmplFunc = tmpl.ExecuteTemplate
		assets = http.FS(Assets)

		reviews, err := server.LoadReviews(Data, *cdnBase)
		if err != nil {
			logger.Error("failed to load reviews", slog.Any("error", err))
			os.Exit(-1)
		}
		reviewsFunc = func() ([]server.Review, error) { return reviews, nil }
	}

	httpClient := &http.Client{
		Timeout: 3 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := server.NewServer(
		version,
		ctx,
		*port,
		httpClient,
		assets,
		assetHashes,
		*cdnBase,
		tmplFunc,
		reviewsFunc,
		logger,
	)

	go srv.Start()

	logger.Debug("started web server", slog.Any("listen_addr", *port))

	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si

	logger.Debug("shutting down web server")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("err", err))
		srv.Close()
	}
}

// defaultCDNBase is where scripts/add-review.sh uploads to. Review photos are
// not embedded in the binary, so this is the only way the page can find them.
const defaultCDNBase = "https://cdn.alexraskin.com"

func cdnBaseURL() string {
	if base := os.Getenv("CDN_BASE_URL"); base != "" {
		return base
	}
	return defaultCDNBase
}

func assetFuncs(hashes server.AssetHashes) template.FuncMap {
	return template.FuncMap{
		"asset": hashes.URL,
	}
}
