package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"log"
	"time"
)

type TokenService struct {
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func (t *TokenService) GenerateJWT(userID int, secretKey string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Создаем новый JWT токен с указанными claims и методом подписи HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен с использованием секретного ключа и возвращаем подписанную строку токена
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}
	return signedToken, nil
}

func (t *TokenService) ParseJWT(tokenStr string, secretKey string) (*Claims, error) {
	claims := &Claims{}
	// ParseWithClaims разбирает токен и заполняет структуру claims.
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Println("Error parsing JWT:", err)
		return nil, errors.New("invalid token")
	}

	if !token.Valid {
		log.Println("Invalid JWT token")
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
