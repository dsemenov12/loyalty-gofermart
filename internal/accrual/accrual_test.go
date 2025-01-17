package accrual

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/dsemenov12/loyalty-gofermart/internal/models"
	"github.com/stretchr/testify/assert"
)

// Тестирование метода GetAccrualInfo
func TestClient_GetAccrualInfo(t *testing.T) {
	tests := []struct {
		name           string
		orderNumber    string
		responseCode   int
		responseBody   interface{}
		expectedResult *models.AccrualInfo
		expectedError  error
	}{
		{
			name:         "valid response",
			orderNumber:  "123456",
			responseCode: http.StatusOK,
			responseBody: models.AccrualInfo{OrderNumber: "123456", Status: "PROCESSED", Accrual: 100.0},
			expectedResult: &models.AccrualInfo{OrderNumber: "123456", Status: "PROCESSED", Accrual: 100.0},
			expectedError:  nil,
		},
		{
			name:         "order not found",
			orderNumber:  "654321",
			responseCode: http.StatusNoContent,
			responseBody: nil,
			expectedResult: nil,
			expectedError:  errors.New("order not found"),
		},
		{
			name:         "too many requests",
			orderNumber:  "789012",
			responseCode: http.StatusTooManyRequests,
			responseBody: nil,
			expectedResult: nil,
			expectedError:  errors.New("too many requests"),
		},
		{
			name:         "unexpected status code",
			orderNumber:  "999999",
			responseCode: http.StatusInternalServerError,
			responseBody: nil,
			expectedResult: nil,
			expectedError:  errors.New("unexpected response code: 500"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer ts.Close()

			client := NewClient(ts.URL)
			result, err := client.GetAccrualInfo(tt.orderNumber)

			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}