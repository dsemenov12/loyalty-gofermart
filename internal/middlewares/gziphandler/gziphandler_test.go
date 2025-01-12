package gziphandler

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipHandle(t *testing.T) {
	// Mock handler that just writes "Hello, World!"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Wrap the mock handler with GzipHandle
	gzipHandler := GzipHandle(handler)

	tests := []struct {
		name           string
		acceptEncoding string
		expectedBody   string
	}{
		{
			name:           "No gzip",
			acceptEncoding: "",
			expectedBody:   "Hello, World!",
		},
		{
			name:           "With gzip",
			acceptEncoding: "gzip",
			expectedBody:   "Hello, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			rr := httptest.NewRecorder()

			// Call the GzipHandle with the mock handler
			gzipHandler.ServeHTTP(rr, req)

			// Check for gzip encoding in the response header
			if tt.acceptEncoding == "gzip" && rr.Header().Get("Content-Encoding") != "gzip" {
				t.Errorf("expected gzip encoding, got %v", rr.Header().Get("Content-Encoding"))
			}

			// If we requested gzip encoding, we need to read the gzipped response
			var body []byte
			if tt.acceptEncoding == "gzip" {
				gr, err := gzip.NewReader(rr.Body)
				if err != nil {
					t.Fatal(err)
				}
				defer gr.Close()

				body, err = io.ReadAll(gr)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				body = rr.Body.Bytes()
			}

			if string(body) != tt.expectedBody {
				t.Errorf("expected body %v, got %v", tt.expectedBody, string(body))
			}
		})
	}
}