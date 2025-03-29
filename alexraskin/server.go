package alexraskin

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
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
	port       string
	httpClient *http.Client
	server     *http.Server
	templates  http.FileSystem
	assets     http.FileSystem
	tmplFunc   ExecuteTemplateFunc
	md         goldmark.Markdown
}

func NewServer(version string, port string, httpClient *http.Client, rawTemplates embed.FS, rawAssets embed.FS, tmplFunc ExecuteTemplateFunc, md goldmark.Markdown) *Server {
	templatesFS, err := fs.Sub(rawTemplates, "templates")
	if err != nil {
		log.Fatal(err)
	}

	assetsFS, err := fs.Sub(rawAssets, "assets")
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		version:    version,
		port:       port,
		httpClient: httpClient,
		templates:  http.FS(templatesFS),
		assets:     http.FS(assetsFS),
		tmplFunc:   tmplFunc,
		md:         md,
	}

	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: s.Routes(),
	}

	return s
}

func (s *Server) Start() {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Error while listening", slog.Any("err", err))
		os.Exit(-1)
	}
}

func (s *Server) Close() {
	if err := s.server.Close(); err != nil {
		slog.Error("Error while closing server", slog.Any("err", err))
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
