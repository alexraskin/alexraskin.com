package main

import (
	"embed"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexraskin/alexraskin.com/alexraskin"
)

var (
	//go:embed public/*
	embeddedPublic embed.FS

	//go:embed assets/*
	embeddedAssets embed.FS
)

func main() {
	port := flag.String("port", "8000", "port to listen on")
	flag.Parse()

	server := alexraskin.NewServer(
		*port,
		http.DefaultClient,
		embeddedPublic,
		embeddedAssets,
	)

	go server.Start()
	defer server.Close()

	slog.Info("started alexraskin.com", slog.Any("listen_addr", *port))
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si
	slog.Info("shutting down alexraskin.com")
}
