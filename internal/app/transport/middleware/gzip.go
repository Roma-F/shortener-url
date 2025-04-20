package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

var compressibleTypes = []string{
	"application/json",
	"text/html",
}

func WithGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		contentEncoding := r.Header.Get("Content-Encoding")
		isGzipped := strings.Contains(contentEncoding, "gzip")

		if isGzipped {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to read gzipped request body", http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()

			r.Body = io.NopCloser(gzipReader)
			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
		}

		if supportsGzip {
			gzipWriter := NewGzipResponseWriter(w)
			defer gzipWriter.Close()

			next.ServeHTTP(gzipWriter, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type GzipResponseWriter struct {
	w             http.ResponseWriter
	gzipWriter    *gzip.Writer
	headerWritten bool
}

func NewGzipResponseWriter(w http.ResponseWriter) *GzipResponseWriter {
	gz := gzip.NewWriter(w)
	return &GzipResponseWriter{
		w:          w,
		gzipWriter: gz,
	}
}

func (gzw *GzipResponseWriter) Header() http.Header {
	return gzw.w.Header()
}

func (gzw *GzipResponseWriter) WriteHeader(statusCode int) {
	if gzw.headerWritten {
		return
	}

	contentType := gzw.w.Header().Get("Content-Type")
	shouldCompress := false

	for _, compressibleType := range compressibleTypes {
		if strings.HasPrefix(contentType, compressibleType) {
			shouldCompress = true
			break
		}
	}

	if shouldCompress {
		gzw.w.Header().Set("Content-Encoding", "gzip")
		gzw.w.Header().Del("Content-Length")
	}

	gzw.w.WriteHeader(statusCode)
	gzw.headerWritten = true
}

func (gzw *GzipResponseWriter) Write(b []byte) (int, error) {
	if !gzw.headerWritten {
		gzw.WriteHeader(http.StatusOK)
	}

	contentType := gzw.w.Header().Get("Content-Type")
	shouldCompress := false

	for _, compressibleType := range compressibleTypes {
		if strings.HasPrefix(contentType, compressibleType) {
			shouldCompress = true
			break
		}
	}

	if shouldCompress {
		return gzw.gzipWriter.Write(b)
	}

	return gzw.w.Write(b)
}

func (gzw *GzipResponseWriter) Close() error {
	contentType := gzw.w.Header().Get("Content-Type")
	shouldCompress := false

	for _, compressibleType := range compressibleTypes {
		if strings.HasPrefix(contentType, compressibleType) {
			shouldCompress = true
			break
		}
	}

	if shouldCompress {
		return gzw.gzipWriter.Close()
	}
	return nil
}

func (gzw *GzipResponseWriter) Flush() {
	if flusher, ok := gzw.w.(http.Flusher); ok {
		flusher.Flush()
	}
}
