package loggerhandler

import (
	"time"
	"net/http"

	"go.uber.org/zap"
	"github.com/dsemenov12/loyalty-gofermart/internal/logger"
)

type (
    responseData struct {
        status int
        size int
    }
    loggingResponseWriter struct {
        http.ResponseWriter
        responseData *responseData
    }
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
    size, err := r.ResponseWriter.Write(b) 
    r.responseData.size += size
    return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
    r.ResponseWriter.WriteHeader(statusCode) 
    r.responseData.status = statusCode
}

func RequestLogger(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData {
            status: 0,
            size: 0,
        }
        lw := loggingResponseWriter {
            ResponseWriter: w,
            responseData: responseData,
        }

        handlerFunc(&lw, r)

		duration := time.Since(start)
		
		logger.Log.Info("got incoming HTTP request",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
			zap.Duration("duration", duration),
        )
		logger.Log.Info("got response",
            zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
        )
    })
}
