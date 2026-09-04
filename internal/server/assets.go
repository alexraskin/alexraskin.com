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

type Asset struct {
	Body        []byte
	Hash        string
	ContentType string
}

type AssetHashes map[string]*Asset

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

	for name, asset := range assets {
		if path.Ext(name) != ".css" {
			continue
		}
		asset.Body = assets.rewriteURLs(asset.Body)
		asset.Hash = hashBytes(asset.Body)
	}

	return assets, nil
}

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

func (a AssetHashes) URL(name string) string {
	asset, ok := a[name]
	if !ok {
		return name
	}
	return name + "?" + assetVersionParam + "=" + asset.Hash
}

func (a AssetHashes) serve(w http.ResponseWriter, r *http.Request) bool {
	asset, ok := a[r.URL.Path]
	if !ok {
		return false
	}

	if asset.ContentType != "" {
		w.Header().Set("Content-Type", asset.ContentType)
	}

	http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(asset.Body))

	return true
}

func hashBytes(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])[:12]
}
