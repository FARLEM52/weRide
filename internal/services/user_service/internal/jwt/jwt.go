package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"

	"we_ride/internal/services/user_service/internal/models"
)

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

func NewToken(user models.User, secret string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID: user.UserID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}
	return tokenString, nil
}

func ValidateToken(tokenStr string, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
