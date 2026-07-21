package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/alexraskin/alexraskin.com/internal/ver"
)

type ExecuteTemplateFunc func(wr io.Writer, name string, data any) error

type Server struct {
	version    ver.Version
	ctx        context.Context
	port       string
	httpClient *http.Client
	server     *http.Server
	assets     http.FileSystem
	tmplFunc   ExecuteTemplateFunc
	logger     *slog.Logger
}

func NewServer(version ver.Version, ctx context.Context, port string, httpClient *http.Client, assets http.FileSystem, tmplFunc ExecuteTemplateFunc, logger *slog.Logger) *Server {

	s := &Server{
		version:    version,
		ctx:        ctx,
		port:       port,
		httpClient: httpClient,
		assets:     assets,
		tmplFunc:   tmplFunc,
		logger:     logger,
	}

	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: s.Routes(),
	}

	return s
}

func (s *Server) Start() {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("Error while listening", slog.Any("err", err))
		os.Exit(-1)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) Close() {
	if err := s.server.Close(); err != nil {
		s.logger.Error("Error while closing server", slog.Any("err", err))
	}
}
