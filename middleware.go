package proxy

import (
	"net/http"
	"time"

	"github.com/mmatczuk/proxy/log"
)

// LoggingMiddleware is a HTTP middleware that logs HTTP requests.
type LoggingMiddleware struct {
	Inner  http.Handler
	Logger log.Logger
}

func (m LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Logger.Log(
		"msg", "request",
		"method", r.Method,
		"path", r.URL.Path,
	)

	start := time.Now()
	sw := &statusAwareWriter{ResponseWriter: w}
	m.Inner.ServeHTTP(sw, r)

	m.Logger.Log(
		"msg", "response",
		"duration", time.Since(start),
		"method", r.Method,
		"path", r.URL.Path,
		"status", sw.status,
	)
}

// statusAwareWriter is a http.ResponseWriter that provides information on
// status code.
type statusAwareWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusAwareWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func (w *statusAwareWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
