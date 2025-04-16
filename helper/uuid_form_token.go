package helper

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func GetUserUUIDFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return os.Getenv("JWT_SECRET"), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Assuming UUID is stored in "user_id" field
		if uuid, ok := claims["user_id"].(string); ok {
			return uuid, nil
		}
		return "", fmt.Errorf("user_id not found in token claims")
	}

	return "", fmt.Errorf("invalid token")
}
