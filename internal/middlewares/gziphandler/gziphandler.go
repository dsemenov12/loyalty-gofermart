package gziphandler

import (
	"net/http"
	"io"
	"strings"
	"compress/gzip"
)

type gzipWriter struct {
    http.ResponseWriter
    Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
    return w.Writer.Write(b)
} 

func GzipHandle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }

        gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
        if err != nil {
            io.WriteString(w, err.Error())
            return
        }
        defer gz.Close()

        sendsGzip := strings.Contains(r.Header.Get("Content-Encoding"), "gzip")
        if sendsGzip {
            cr, err := gzip.NewReader(r.Body)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
            r.Body = cr
            defer cr.Close()
        }

        w.Header().Set("Content-Encoding", "gzip")
        next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
    })
}