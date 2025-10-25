package middleware

import (
	"net/http"
	"time"

	"bjoernblessin.de/go-utils/util/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *loggingResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		logger.Infof(
			"[%s] %s %s - Status: %d - Duration: %v - Size: %d bytes",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
			rw.written,
		)
	})
}
