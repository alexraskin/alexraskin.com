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

	"github.com/alexraskin/alexraskin.com/alexraskin"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	version   = "unknown"
	commit    = "unknown"
	buildTime = "unknown"
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
		tmplFunc alexraskin.ExecuteTemplateFunc
		assets   http.FileSystem
	)

	logger := slog.Default()
	if *devMode {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	logger.Debug("Starting alexraskin.com...", slog.Any("version", version), slog.Any("commit", commit), slog.Any("buildTime", buildTime))

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
	} else {
		tmpl, err := template.New("").ParseFS(Templates, "templates/*.gohtml")
		if err != nil {
			logger.Error("failed to parse templates", slog.Any("error", err))
			os.Exit(-1)
		}
		tmplFunc = tmpl.ExecuteTemplate
		assets = http.FS(Assets)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := alexraskin.NewServer(
		alexraskin.FormatBuildVersion(version, commit, buildTime),
		ctx,
		*port,
		httpClient,
		assets,
		tmplFunc,
		md,
		logger,
	)

	go server.Start()

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("err", err))
		server.Close()
	}

	logger.Debug("started web server", slog.Any("listen_addr", *port))
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si
	logger.Debug("shutting down web server")
}
