package crypto

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultSigningKey = "secretKEY123"
const defaultExpirationHours = 256

// Возвращает установленное в системе TTL для куки и токенов.
//
// Можно установить TTL токена в часах в следующую переменную env: AUTH_EXPIRATION_HOURS.
// TTL по умолчанию 256 часов.
func GetExpirationHours() (expH time.Duration) {
	expHStr := os.Getenv("AUTH_EXPIRATION_HOURS")
	expHInt, err := strconv.Atoi(expHStr)
	if err != nil {
		expHInt = defaultExpirationHours
	}
	expH = time.Duration(expHInt) * time.Hour
	return
}

func getSigningKey() (key []byte) {
	content, err := os.ReadFile("./crypto/SIGNING_KEY.txt")
	if err != nil {
		key = []byte(defaultSigningKey)
	} else {
		key = []byte(content)
	}

	return
}

// Генерирует токен аутентификации.
//
// Можно установить TTL токена в часах в следующую переменную env: AUTH_EXPIRATION_HOURS.
// TTL по умолчанию 256 часов.
//
// Можно установить ключ для подписи в следующий файл: "./auth/SIGNING_KEY.txt".
// Или будет использован некоторый ключ по умолчанию.
func GenerateToken(userID string) (tokenString string, err error) {
	expirationTime := time.Now().Add(GetExpirationHours())

	claims := &jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	key := getSigningKey()

	tokenString, err = token.SignedString(key)
	if err != nil {
		err = fmt.Errorf("не удалось подписать токен: %v", err)
		return
	}

	return
}

// Проверяет токен и возвращает userId.
func ValidateToken(tokenString string) (userId string, err error) {

	claims := &jwt.RegisteredClaims{}
	key := getSigningKey()
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		return
	}

	if !token.Valid {
		err = errors.New("invalid token")
		return
	}

	userId = claims.Subject
	return
}
