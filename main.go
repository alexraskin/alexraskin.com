package main

import (
	"embed"
	"flag"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	slog.Info("Starting alexraskin.com...", slog.Any("version", version), slog.Any("commit", commit), slog.Any("buildTime", buildTime))

	if *devMode {
		slog.Info("running in dev mode")
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
			slog.Error("failed to parse templates", slog.Any("error", err))
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

	server := alexraskin.NewServer(
		alexraskin.FormatBuildVersion(version, commit, buildTime),
		*port,
		http.DefaultClient,
		assets,
		tmplFunc,
		md,
	)

	go server.Start()
	defer server.Close()

	slog.Info("started web server", slog.Any("listen_addr", *port))
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si
	slog.Info("shutting down web server")
}
