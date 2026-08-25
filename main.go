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
)

func main() {
	port := flag.String("port", "8000", "port to listen on")
	devMode := flag.Bool("dev", false, "run in dev mode")
	flag.Parse()

	var (
		tmplFunc    server.ExecuteTemplateFunc
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
		// Assets come off disk here, so hashing the embedded copies would only
		// pin URLs to bytes that are no longer being served. An empty map
		// leaves asset paths unversioned.
		assetHashes = server.AssetHashes{}
		tmplFunc = func(wr io.Writer, name string, data any) error {
			tmpl, err := template.New("").Funcs(assetFuncs(assetHashes)).ParseGlob("templates/*.gohtml")
			if err != nil {
				return err
			}
			return tmpl.ExecuteTemplate(wr, name, data)
		}
		assets = http.Dir(".")
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
	}

	httpClient := &http.Client{
		// Keeps a slow last.fm upstream from stalling the page render.
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
		tmplFunc,
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

// assetFuncs exposes the content-hashed asset URLs to templates.
func assetFuncs(hashes server.AssetHashes) template.FuncMap {
	return template.FuncMap{
		"asset": hashes.URL,
	}
}
