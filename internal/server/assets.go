package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"time"
)

type AssetsFunc func() (*Assets, error)

// Assets holds the static files in memory, keyed by URL path, with a content
// hash used both as the fingerprint in asset URLs and as the ETag.
type Assets struct {
	files map[string]asset
}

type asset struct {
	data []byte
	hash string
}

var cssURL = regexp.MustCompile(`url\("(/assets/[^"]+)"\)`)

func LoadAssets(fsys fs.FS) (*Assets, error) {
	a := &Assets{files: map[string]asset{}}
	var stylesheets []string
	err := fs.WalkDir(fsys, "assets", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		a.files["/"+name] = asset{data: data, hash: hashBytes(data)}
		if path.Ext(name) == ".css" {
			stylesheets = append(stylesheets, "/"+name)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("assets: %w", err)
	}

	// Stylesheets are rewritten after everything else is hashed so the fonts they
	// load share URLs, and therefore cache entries, with the fingerprinted preloads.
	for _, name := range stylesheets {
		var missing error
		data := cssURL.ReplaceAllFunc(a.files[name].data, func(m []byte) []byte {
			url, err := a.Path(string(cssURL.FindSubmatch(m)[1]))
			if err != nil {
				missing = err
				return m
			}
			return []byte(`url("` + url + `")`)
		})
		if missing != nil {
			return nil, fmt.Errorf("assets: %s: %w", name, missing)
		}
		a.files[name] = asset{data: data, hash: hashBytes(data)}
	}

	return a, nil
}

// Path returns the URL for an asset with its content hash appended, so the URL
// changes whenever the file does and can be cached indefinitely.
func (a *Assets) Path(name string) (string, error) {
	f, ok := a.files[name]
	if !ok {
		return "", fmt.Errorf("asset %s not found", name)
	}
	return name + "?h=" + f.hash, nil
}

func (s *Server) serveAsset(w http.ResponseWriter, r *http.Request, name string) {
	assets, err := s.assetsFunc()
	if err != nil {
		s.logger.Error("failed to load assets", "error", err)
		s.renderError(w, r, "Failed to load assets", http.StatusInternalServerError)
		return
	}
	f, ok := assets.files[name]
	if !ok {
		http.NotFound(w, r)
		return
	}
	// A stale or missing hash still gets the current file, just without the
	// long-lived cache, so old HTML and hand-written URLs keep working.
	if r.URL.Query().Get("h") == f.hash {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=86400")
	}
	serveContent(w, r, name, f.data, f.hash)
}

// serveContent handles If-None-Match, HEAD and Content-Length. Content-Type is
// taken from the header if already set, otherwise inferred from name.
func serveContent(w http.ResponseWriter, r *http.Request, name string, data []byte, hash string) {
	w.Header().Set("ETag", `"`+hash+`"`)
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}
