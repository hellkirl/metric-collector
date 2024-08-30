package middleware

import (
	"compress/gzip"
	"devops_analytics/internal/logger"
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

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.gz == nil {
		gz, err := gzip.NewWriterLevel(g.ResponseWriter, gzip.BestSpeed)
		if err != nil {
			return 0, err
		}
		g.gz = gz
	}

	if len(b) < 1500 {
		size, err := g.ResponseWriter.Write(b)
		return size, err
	}

	size, err := g.gz.Write(b)
	if err != nil {
		return size, err
	}

	return size, nil
}

func (g *gzipResponseWriter) Close() error {
	if g.gz != nil {
		return g.gz.Close()
	}
	return nil
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

func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Aceept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")

			next.ServeHTTP(&gzipResponseWriter{
				ResponseWriter: w,
			}, r)
		}
		next.ServeHTTP(w, r)
	})
}

func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			r.Body = http.MaxBytesReader(w, r.Body, 1048576)

			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress the request", http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}
