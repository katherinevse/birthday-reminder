package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"log"
	"time"
)

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int, secretKey string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func ParseJWT(tokenStr string, secretKey string) (*Claims, error) {
	claims := &Claims{}
	// ParseWithClaims разбирает токен и заполняет структуру claims.
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Println("Error parsing JWT:", err)
		return nil, errors.New("invalid token") // Возвращаем специальную ошибку
	}

	if !token.Valid {
		log.Println("Invalid JWT token")
		return nil, errors.New("invalid token") // Возвращаем специальную ошибку
	}

	return claims, nil
}
