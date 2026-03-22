package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &LoggingResponseWriter{ResponseWriter: w, body: bytes.NewBuffer([]byte{}), status: 200}

		// Read full request body
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
		}

		next.ServeHTTP(lrw, r)

		logEntry := map[string]interface{}{
			"requestDate":  start.Format(time.RFC3339),
			"clientIP":     r.RemoteAddr,
			"API":          r.Method + " " + r.URL.Path,
			"request":      string(reqBody),
			"response":     lrw.body.String(),
			"responseTime": time.Since(start).Milliseconds(),
			"status":       lrw.status,
		}

		logJSON, _ := json.Marshal(logEntry)
		log.Println(string(logJSON))
	})
}
