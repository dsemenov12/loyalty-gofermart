package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestBuildJWTString(t *testing.T) {
	// Тестируем корректный userID
	userID := "testUser123"
	tokenString, err := BuildJWTString(userID)

	// Проверяем, что ошибки при создании токена нет
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Разбираем токен, чтобы проверить его содержимое
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, token)

	// Проверяем, что UserID и время истечения срока действия правильные
	assert.Equal(t, userID, claims.UserID)
	assert.WithinDuration(t, time.Now().Add(TokenExp), claims.ExpiresAt.Time, time.Second)
}

func TestGetUserID(t *testing.T) {
	// Генерируем правильный JWT токен
	userID := "testUser123"
	tokenString, err := BuildJWTString(userID)
	assert.NoError(t, err)

	// Получаем userID из токена
	retrievedUserID, err := GetUserID(tokenString)
	assert.NoError(t, err)

	// Проверяем, что полученный userID совпадает с ожидаемым
	assert.Equal(t, userID, retrievedUserID)
}

func TestGetUserID_InvalidToken(t *testing.T) {
	// Некорректный токен
	invalidToken := "invalidTokenString"

	// Пытаемся получить userID из некорректного токена
	retrievedUserID, err := GetUserID(invalidToken)

	// Проверяем, что возникла ошибка и userID не был получен
	assert.Error(t, err)
	assert.Empty(t, retrievedUserID)
}

func TestBuildJWTString_EmptyUserID(t *testing.T) {
	// Тестируем создание токена с пустым userID
	tokenString, err := BuildJWTString("")
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Разбираем токен и проверяем его содержимое
	claims := &Claims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	assert.NoError(t, err)

	// Проверяем, что пустой userID обрабатывается правильно
	assert.Equal(t, "", claims.UserID)
}

func TestBuildJWTString_ExpiredToken(t *testing.T) {
	// Создаем токен с истекшим сроком действия
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour * 24)),
		},
		UserID: "expiredUser",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	assert.NoError(t, err)

	// Пытаемся получить userID из истекшего токена
	_, err = GetUserID(tokenString)
	assert.Error(t, err) // Токен должен быть истекшим
}