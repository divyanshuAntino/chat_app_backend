package helper

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func GetUserUUIDFromToken(tokenString string) (string, error) {
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		if uuid, ok := claims["userId"].(string); ok {
			return uuid, nil
		}
		return "", fmt.Errorf("userId not found in token claims")
	}

	return "", fmt.Errorf("invalid token")
}
