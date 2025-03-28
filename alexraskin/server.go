package alexraskin

import (
	"embed"
	"errors"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type Server struct {
	port       string
	httpClient *http.Client
	server     *http.Server
	public     http.FileSystem
	assets     http.FileSystem
}

func NewServer(port string, httpClient *http.Client, rawPublic embed.FS, rawAssets embed.FS) *Server {
	publicFS, err := fs.Sub(rawPublic, "public")
	if err != nil {
		log.Fatal(err)
	}

	assetsFS, err := fs.Sub(rawAssets, "assets")
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		port:       port,
		httpClient: httpClient,
		public:     http.FS(publicFS),
		assets:     http.FS(assetsFS),
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
