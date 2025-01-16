package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
    "time"

	"github.com/dsemenov12/loyalty-gofermart/internal/models"
	"github.com/dsemenov12/loyalty-gofermart/internal/storage/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Тестирование метода UserRegister
func Test_app_UserRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockStorage(ctrl)

	m.EXPECT().GetUserByLogin("user").Return(&models.User{ID: 1, Login: "user"}, nil).AnyTimes()
	m.EXPECT().CreateUser("user", gomock.Any()).AnyTimes()
	m.EXPECT().CreateUser("user1", gomock.Any()).Return(errors.New("user already exists")).AnyTimes()

	// создадим экземпляр приложения и передадим ему «хранилище»
    app := NewApp(m)

	type want struct {
        code        int
        contentType string
    }
	tests := []struct {
		name string
		body string
		want want
	}{
		{
            name: "positive test #1",
			body: `{"login": "user","password": "123456"}`,
            want: want{
                code: http.StatusOK,
        		contentType: "application/json",
            },
        },
		{
            name: "error conflict test",
			body: `{"login": "user1","password": "123456"}`,
            want: want{
                code: http.StatusConflict,
        		contentType: "application/json",
            },
        },
		{
            name: "error test",
			body: `{"login": "user1"}`,
            want: want{
                code: http.StatusBadRequest,
        		contentType: "application/json",
            },
        },
		{
            name: "error empty test",
			body: ``,
            want: want{
                code: http.StatusBadRequest,
        		contentType: "application/json",
            },
        },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			app.UserRegister(response, request)

			res := response.Result()
			defer res.Body.Close()
            
            assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

// Тестирование метода UserLogin
func Test_app_UserLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockStorage(ctrl)

	m.EXPECT().GetUserByLogin("user").Return(&models.User{ID: 1, Login: "user", Password: "$2a$10$dJQ76.hRXamJDPf.wYT/suWxZU0K25tvubpcXy8lW8X6ERzzBGQX2"}, nil).AnyTimes()
	m.EXPECT().GetUserByLogin("user1").Return(nil, errors.New("user not found")).AnyTimes()

	// создадим экземпляр приложения и передадим ему «хранилище»
    app := NewApp(m)

	type want struct {
        code        int
        contentType string
    }
	tests := []struct {
		name string
		body string
		want want
	}{
		{
            name: "positive test",
			body: `{"login":"user","password":"123456"}`,
            want: want{
                code: http.StatusOK,
        		contentType: "application/json",
            },
        },
		{
            name: "error unauthorized test",
			body: `{"login":"user1","password":"123456"}`,
            want: want{
                code: http.StatusUnauthorized,
        		contentType: "application/json",
            },
        },
		{
            name: "error test",
			body: `{"login": "user1"}`,
            want: want{
                code: http.StatusBadRequest,
        		contentType: "application/json",
            },
        },
		{
            name: "error empty test",
			body: ``,
            want: want{
                code: http.StatusBadRequest,
        		contentType: "application/json",
            },
        },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			app.UserLogin(response, request)

			res := response.Result()
			defer res.Body.Close()
            
            assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

// Тестирование метода UserUploadOrder
func Test_app_UserUploadOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок хранилища
	m := mocks.NewMockStorage(ctrl)
	m.EXPECT().SaveOrder(gomock.Any(), "1234567890318").Return(true, nil).AnyTimes()

	// Создаем экземпляр приложения
	app := NewApp(m)

	tests := []struct {
		name string
		body string
		want int
	}{
		{
			name: "positive test",
			body: "1234567890318",
			want: http.StatusAccepted,
		},
		{
			name: "bad request test (invalid order)",
			body: "abcdef",
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "empty request",
			body: "",
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			app.UserUploadOrder(response, request)

			res := response.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}

// Тестирование метода UserGetOrders
func Test_app_UserGetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	app := NewApp(m)

	m.EXPECT().GetOrdersByUser(gomock.Any()).Return([]models.Order{}, nil).AnyTimes()

	tests := []struct {
		name string
		want int
	}{
        {
			name: "no content",
			want: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			response := httptest.NewRecorder()

			app.UserGetOrders(response, request)

			res := response.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}

// Тестирование метода GetUserBalance
func Test_app_GetUserBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	app := NewApp(m)

	m.EXPECT().GetBalance(gomock.Any()).Return(&models.Balance{Current: 100.0}, nil).AnyTimes()

	tests := []struct {
		name string
		want int
	}{
		{
			name: "positive test",
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			response := httptest.NewRecorder()

			app.GetUserBalance(response, request)

			res := response.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}

// Тестирование метода GetUserWithdrawals
func Test_app_GetUserWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок хранилища
	m := mocks.NewMockStorage(ctrl)

	mockWithdrawals := []models.Withdrawal{
		{Order: "123456789", Sum: 500, ProcessedAt: time.Now().Format(time.RFC3339)},
	}

	m.EXPECT().GetUserWithdrawals(gomock.Any()).Return(mockWithdrawals, nil).AnyTimes()

	app := NewApp(m)

	tests := []struct {
		name       string
		userID     int64
		wantCode   int
		wantResult string
	}{
		{
			name:     "positive test",
			userID:   1,
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			response := httptest.NewRecorder()

			ctx := context.WithValue(request.Context(), "userID", tt.userID)
			request = request.WithContext(ctx)

			app.GetUserWithdrawals(response, request)

			res := response.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantCode, res.StatusCode)
		})
	}
}

// Тестирование метода WithdrawUserBalance
func Test_app_WithdrawUserBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)

	m.EXPECT().WithdrawUserBalance(gomock.Any(), "123456789", float64(100)).Return(nil).AnyTimes()
	m.EXPECT().WithdrawUserBalance(gomock.Any(), "123456789", float64(100)).Return(errors.New("insufficient funds")).AnyTimes()

    m.EXPECT().GetBalance(gomock.Any()).Return(&models.Balance{Current: 100, Withdrawn: 0}, nil).AnyTimes()
	m.EXPECT().GetBalance(gomock.Any()).Return(&models.Balance{Current: 0, Withdrawn: 0}, nil).AnyTimes()

	app := NewApp(m)

	tests := []struct {
		name     string
		userID   int64
		body     string
		wantCode int
	}{
		{
			name:     "positive test",
			userID:   1,
			body:     `{"order": "123456789", "sum": 100}`,
			wantCode: http.StatusOK,
		},
		{
			name:     "empty body test",
			userID:   1,
			body:     ``,
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			ctx := context.WithValue(request.Context(), "userID", tt.userID)
			request = request.WithContext(ctx)

			app.WithdrawUserBalance(response, request)

			res := response.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantCode, res.StatusCode)
		})
	}
}

// Тестирование метода setCookieJWT
func Test_setCookieJWT(t *testing.T) {
	tests := []struct {
		name   string
		userID string
	}{
		{
			name:   "set valid cookie",
			userID: "1",
		},
		{
			name:   "set empty user ID",
			userID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			setCookieJWT(tt.userID, response)

			res := response.Result()
			defer res.Body.Close()

			cookie := res.Cookies()
			assert.NotEmpty(t, cookie, "Expected cookie to be set")
		})
	}
}