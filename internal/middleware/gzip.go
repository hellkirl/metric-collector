package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.gz == nil {
		var err error
		g.gz, err = gzip.NewWriterLevel(g.ResponseWriter, gzip.DefaultCompression)
		if err != nil {
			return 0, err
		}
	}

	if len(b) == 0 {
		return 0, nil
	}

	n, err := g.gz.Write(b)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (g *gzipResponseWriter) Close() error {
	if g.gz != nil {
		err := g.gz.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func GzipResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")

			gz := gzip.NewWriter(w)
			defer gz.Close()

			gzWriter := &gzipResponseWriter{
				ResponseWriter: w,
				gz:             gz,
			}

			next.ServeHTTP(gzWriter, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func GzipRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request body", http.StatusInternalServerError)
				return
			}
			defer gr.Close()

			r.Body = io.NopCloser(gr)
		}
		next.ServeHTTP(w, r)
	})
}
