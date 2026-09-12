package server

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alexraskin/alexraskin.com/internal/ver"
)

type ExecuteTemplateFunc func(wr io.Writer, name string, data any) error

type Server struct {
	version         ver.Version
	httpClient      *http.Client
	assets          http.FileSystem
	cdnBase         string
	tmplFunc        ExecuteTemplateFunc
	reviewsFunc     ReviewsFunc
	logger          *slog.Logger
	trackMu         sync.Mutex
	track           *LastFMTrack
	trackExpires    time.Time
	trackRefreshing bool
}

type Config struct {
	Version         ver.Version
	HTTPClient      *http.Client
	Assets          http.FileSystem
	CDNBase         string
	ExecuteTemplate ExecuteTemplateFunc
	Reviews         ReviewsFunc
	Logger          *slog.Logger
}

func NewServer(cfg Config) *Server {
	return &Server{
		version:     cfg.Version,
		httpClient:  cfg.HTTPClient,
		assets:      cfg.Assets,
		cdnBase:     strings.TrimRight(cfg.CDNBase, "/"),
		tmplFunc:    cfg.ExecuteTemplate,
		reviewsFunc: cfg.Reviews,
		logger:      cfg.Logger,
	}
}
