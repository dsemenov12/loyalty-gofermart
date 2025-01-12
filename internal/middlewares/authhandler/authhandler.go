package authhandler

import (
	"context"
	"net/http"

	"github.com/dsemenov12/loyalty-gofermart/internal/auth"
)

func AuthHandle(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		jwtToken, _ := r.Cookie("JWT")
		if jwtToken == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} 

		userID, err := auth.GetUserID(jwtToken.Value)
		if err != nil || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(context.Background(), auth.UserIDKey, userID))

		handlerFunc(w, r)
	})
}