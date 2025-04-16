package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type responseBuffer struct {
	buf         *bytes.Buffer
	statusCode  int
	header      http.Header
	wroteHeader bool
}

func newResponseBuffer() *responseBuffer {
	return &responseBuffer{
		buf:    new(bytes.Buffer),
		header: make(http.Header),
	}
}

func (rb *responseBuffer) Header() http.Header {
	return rb.header
}

func (rb *responseBuffer) Write(p []byte) (int, error) {
	return rb.buf.Write(p)
}

func (rb *responseBuffer) WriteHeader(statusCode int) {
	if !rb.wroteHeader {
		rb.statusCode = statusCode
		rb.wroteHeader = true
	}
}

func WithGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Unable to decompress request", http.StatusBadRequest)
				return
			}
			defer gzReader.Close()
			r.Body = io.NopCloser(gzReader)
			r.Header.Del("Content-Encoding")
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		rwBuffer := newResponseBuffer()
		for k, v := range w.Header() {
			for _, vv := range v {
				rwBuffer.header.Add(k, vv)
			}
		}

		next.ServeHTTP(rwBuffer, r)

		contentType := rwBuffer.header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Content-Encoding", "gzip")

			w.Header().Del("Content-Length")
			for k, v := range rwBuffer.header {
				if strings.ToLower(k) == "content-length" {
					continue
				}
				w.Header()[k] = v
			}
			w.WriteHeader(rwBuffer.statusCode)
			gzWriter := gzip.NewWriter(w)
			defer gzWriter.Close()
			_, err := gzWriter.Write(rwBuffer.buf.Bytes())
			if err != nil {
			}
		} else {

			for k, v := range rwBuffer.header {
				w.Header()[k] = v
			}
			w.WriteHeader(rwBuffer.statusCode)
			_, _ = w.Write(rwBuffer.buf.Bytes())
		}
	})
}
