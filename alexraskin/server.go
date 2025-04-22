package alexraskin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/yuin/goldmark"
)

type ExecuteTemplateFunc func(wr io.Writer, name string, data any) error

type Server struct {
	version    string
	ctx        context.Context
	port       string
	httpClient *http.Client
	server     *http.Server
	assets     http.FileSystem
	tmplFunc   ExecuteTemplateFunc
	md         goldmark.Markdown
	logger     *slog.Logger
}

func NewServer(version string, ctx context.Context, port string, httpClient *http.Client, assets http.FileSystem, tmplFunc ExecuteTemplateFunc, md goldmark.Markdown, logger *slog.Logger) *Server {

	s := &Server{
		version:    version,
		ctx:        ctx,
		port:       port,
		httpClient: httpClient,
		assets:     assets,
		tmplFunc:   tmplFunc,
		md:         md,
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

func FormatBuildVersion(version string, commit string, buildTime string) string {
	if len(commit) > 7 {
		commit = commit[:7]
	}

	buildTimeStr := "unknown"
	if buildTime != "unknown" {
		parsedTime, _ := time.Parse(time.RFC3339, buildTime)
		if !parsedTime.IsZero() {
			buildTimeStr = parsedTime.Format(time.ANSIC)
		}
	}
	return fmt.Sprintf("Go Version: %s\nVersion: %s\nCommit: %s\nBuild Time: %s\nOS/Arch: %s/%s\n", runtime.Version(), version, commit, buildTimeStr, runtime.GOOS, runtime.GOARCH)
}
