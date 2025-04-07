package helper

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/divyanshu050303/chat-app-backend/models"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(user models.UserModels) (accessToken string, refreshToken string, err error) {
	accessTokenClams := jwt.MapClaims{
		"userId": user.UserId,
		"exp":    time.Now().Add(time.Minute * 15).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClams)
	accessToken, err = token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}
	refreshTokenClams := jwt.MapClaims{
		"userId": user.UserId,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	}
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClams)
	refreshToken, err = token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func ValidateToken(token string) (claims jwt.MapClaims, err error) {
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")

	}
	tokenString, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	clamis, ok := tokenString.Claims.(jwt.MapClaims)
	if !ok || !tokenString.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return clamis, nil
}
