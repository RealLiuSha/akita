package middleware

import (
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/itchenyi/akita"
	"net/http"
	"strings"
)

// binary file system
type bfs struct {
	fs http.FileSystem
}

func (b *bfs) Open(name string) (http.File, error) {
	return b.fs.Open(name)
}

func (b *bfs) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		if _, err := b.Open(p); err != nil {
			return false
		}
		return true
	}
	return false
}

// AssetFs Static returns a middleware handler that serves static files in the given directory.
func AssetFs(urlPrefix string, fs *assetfs.AssetFS) akita.MiddlewareFunc {
	// binary file system
	b := &bfs{fs}

	// file server
	s := http.FileServer(fs)
	if urlPrefix != "" {
		s = http.StripPrefix(urlPrefix, s)
	}

	return func(before akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			err := before(ctx)
			if err != nil {
				if ctx, ok := err.(*akita.HTTPError); !ok || ctx.Code != http.StatusNotFound {
					return err
				}
			}

			w, r := ctx.Response(), ctx.Request()
			if b.Exists(urlPrefix, r.URL.Path) {
				s.ServeHTTP(w, r)
				return nil
			}
			return err
		}
	}
}
