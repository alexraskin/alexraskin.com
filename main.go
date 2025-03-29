package main

import (
	"embed"
	"flag"
	"html/template"
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
	embeddedTemplates embed.FS

	//go:embed assets
	embeddedAssets embed.FS
)

func main() {
	port := flag.String("port", "8000", "port to listen on")
	flag.Parse()

	var (
		tmplFunc alexraskin.ExecuteTemplateFunc
	)

	slog.Info("Starting alexraskin.com...", slog.Any("version", version), slog.Any("commit", commit), slog.Any("buildTime", buildTime))

	funcs := template.FuncMap{
		"Content": func(content string) template.HTML {
			return template.HTML(content)
		},
	}

	tmpl, err := template.New("").Funcs(funcs).ParseFS(embeddedTemplates, "templates/*.gohtml")
	if err != nil {
		slog.Error("failed to parse templates", slog.Any("error", err))
		os.Exit(-1)
	}
	tmplFunc = tmpl.ExecuteTemplate

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
		embeddedTemplates,
		embeddedAssets,
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
