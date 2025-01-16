package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		a    *app
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.UserUploadOrder(tt.args.w, tt.args.r)
		})
	}
}

// Тестирование метода UserGetOrders
func Test_app_UserGetOrders(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		a    *app
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.UserGetOrders(tt.args.w, tt.args.r)
		})
	}
}

// Тестирование метода GetUserBalance
func Test_app_GetUserBalance(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		a    *app
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.GetUserBalance(tt.args.w, tt.args.r)
		})
	}
}

// Тестирование метода GetUserWithdrawals
func Test_app_GetUserWithdrawals(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		a    *app
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.GetUserWithdrawals(tt.args.w, tt.args.r)
		})
	}
}

func Test_app_WithdrawUserBalance(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		a    *app
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.WithdrawUserBalance(tt.args.w, tt.args.r)
		})
	}
}

// Тестирование метода setCookieJWT
func Test_setCookieJWT(t *testing.T) {
	type args struct {
		userID string
		w      http.ResponseWriter
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setCookieJWT(tt.args.userID, tt.args.w)
		})
	}
}

// Тестирование метода checkOrderStatus
func Test_app_checkOrderStatus(t *testing.T) {
	type args struct {
		ctx         context.Context
		orderNumber string
	}
	tests := []struct {
		name string
		a    *app
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.checkOrderStatus(tt.args.ctx, tt.args.orderNumber)
		})
	}
}