package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var compressibleTypes = []string{
	"application/json",
	"text/html",
}

var gzipTypes = []string{
	"application/gzip",
	"application/x-gzip",
}

type gzipReader struct {
	r  io.ReadCloser
	gr *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:  r,
		gr: gr,
	}, nil
}

func (gz *gzipReader) Read(p []byte) (n int, err error) {
	return gz.gr.Read(p)
}

func (gz *gzipReader) Close() error {
	if err := gz.gr.Close(); err != nil {
		return err
	}
	return gz.r.Close()
}

func WithGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var isGzipped bool

		contentEncoding := r.Header.Get("Content-Encoding")
		isGzipped = strings.Contains(contentEncoding, "gzip")

		if !isGzipped {
			contentType := r.Header.Get("Content-Type")
			for _, gzipType := range gzipTypes {
				if strings.Contains(contentType, gzipType) {
					isGzipped = true
					break
				}
			}
		}

		if isGzipped && r.Body != nil {
			gzReader, err := newGzipReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to read gzipped request body", http.StatusBadRequest)
				return
			}

			r.Body = gzReader

			contentType := r.Header.Get("Content-Type")
			for _, gzipType := range gzipTypes {
				if strings.Contains(contentType, gzipType) {
					r.Header.Set("Content-Type", "text/plain")
					break
				}
			}

			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
		}

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			gzipWriter := NewGzipResponseWriter(w)
			defer func() {
				if err := gzipWriter.Close(); err != nil {
					fmt.Printf("Error closing gzip writer: %v\n", err)
				}
			}()

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
