package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AuthJWT interface {
	BuildJWTString(userID string) (string, error)
	GetUserID(tokenString string) (string, error)
}

type Claims struct {
    jwt.RegisteredClaims
    UserID string
}

const TokenExp = time.Hour * 24
const SecretKey = "supersecretkey"

type userContextKey string
const UserIDKey userContextKey = "user_id"

func BuildJWTString(userID string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims {
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
        },
        UserID: userID,
    })

    tokenString, err := token.SignedString([]byte(SecretKey))
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

func GetUserID(tokenString string) (string, error) {
    claims := &Claims{}

    _, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
        return []byte(SecretKey), nil
    })
	if err != nil {
		return "", err
	}

    return claims.UserID, nil
}