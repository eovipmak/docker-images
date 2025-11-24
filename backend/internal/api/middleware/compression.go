package middleware

import (
	"compress/gzip"
	"io"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// gzipWriter wraps http.ResponseWriter to provide gzip compression
type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// gzipWriterPool reuses gzip writers to reduce allocations
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

// GzipCompression returns a middleware that compresses HTTP responses using GZIP
func GzipCompression() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if client accepts gzip encoding
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Skip compression for SSE endpoints
		if strings.Contains(c.Request.URL.Path, "/stream/") {
			c.Next()
			return
		}

		// Get a gzip writer from the pool
		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(c.Writer)

		// Create custom response writer
		gzWriter := &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		// Set response headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Replace the writer
		c.Writer = gzWriter

		// Process the request
		c.Next()

		// Flush and close the gzip writer
		gz.Close()

		// Return the writer to the pool
		gzipWriterPool.Put(gz)
	}
}
