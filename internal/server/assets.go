package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"time"
)

// Asset is one embedded file, held in memory with the hash its URL is
// versioned by. The whole asset set is a handful of small files, so keeping the
// bytes costs little and lets CSS be rewritten once at startup.
type Asset struct {
	Body        []byte
	Hash        string
	ContentType string
}

// AssetHashes maps a request path ("/assets/css/style.css") to its asset.
// Versioned URLs are what make the year-long immutable TTL safe: the URL
// changes the moment the bytes do, so a stale copy is never reachable.
type AssetHashes map[string]*Asset

// HashAssets reads every embedded asset, rewrites the asset URLs inside CSS to
// their versioned form, and hashes the result.
func HashAssets(fsys fs.FS) (AssetHashes, error) {
	assets := AssetHashes{}

	err := fs.WalkDir(fsys, assetsDir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		body, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}

		assets["/"+p] = &Asset{
			Body:        body,
			Hash:        hashBytes(body),
			ContentType: mime.TypeByExtension(path.Ext(p)),
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// A font referenced from CSS has to arrive on the same URL the preload
	// asked for, or the browser fetches it twice. Rewriting happens after every
	// hash is known, then the CSS is rehashed over its final bytes.
	for name, asset := range assets {
		if path.Ext(name) != ".css" {
			continue
		}
		asset.Body = assets.rewriteURLs(asset.Body)
		asset.Hash = hashBytes(asset.Body)
	}

	return assets, nil
}

// rewriteURLs swaps bare asset paths for versioned ones. CSS only ever points
// at fonts here, and fonts are hashed independently, so there is no ordering
// cycle to worry about.
func (a AssetHashes) rewriteURLs(body []byte) []byte {
	for name, asset := range a {
		if path.Ext(name) == ".css" {
			continue
		}
		versioned := name + "?" + assetVersionParam + "=" + asset.Hash
		body = bytes.ReplaceAll(body, []byte(name), []byte(versioned))
	}
	return body
}

// URL appends the content hash as a query parameter. An unknown path is
// returned untouched, which is also the dev-mode behaviour: templates and
// assets are read from disk there, so versioning them would only pin URLs to
// bytes that are no longer being served.
func (a AssetHashes) URL(name string) string {
	asset, ok := a[name]
	if !ok {
		return name
	}
	return name + "?" + assetVersionParam + "=" + asset.Hash
}

// serve writes an in-memory asset, reporting whether it handled the request.
// Falling through leaves dev mode on the plain file server.
func (a AssetHashes) serve(w http.ResponseWriter, r *http.Request) bool {
	asset, ok := a[r.URL.Path]
	if !ok {
		return false
	}

	if asset.ContentType != "" {
		w.Header().Set("Content-Type", asset.ContentType)
	}

	// The zero modtime keeps ServeContent from emitting Last-Modified; the
	// ETag set upstream is the validator.
	http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(asset.Body))

	return true
}

func hashBytes(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])[:12]
}
