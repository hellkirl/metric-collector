package middleware

import (
	"compress/gzip"
	"devops_analytics/internal/logger"
	"io"
	"net/http"
	"strings"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}

	gzipResponseWriter struct {
		http.ResponseWriter
		gz *gzip.Writer
	}

	gzipRequest struct {
		*http.Request
		gr *gzip.Reader
	}
)

func (r *loggingResponseWriter) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.responseData.status = status
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}

		start := time.Now()

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		logger.Log.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", lw.responseData.status,
			"size", lw.responseData.size,
			"duration", duration,
		)
	})
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

func (g *gzipRequest) Read(p []byte) (int, error) {
	if g.gr == nil {
		var err error
		g.gr, err = gzip.NewReader(g.Body)
		if err != nil {
			return 0, err
		}
	}
	return g.gr.Read(p)
}

func (g *gzipRequest) Close() error {
	if g.gr != nil {
		err := g.gr.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func Compress(next http.Handler) http.Handler {
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

func Decompress(next http.Handler) http.Handler {
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
