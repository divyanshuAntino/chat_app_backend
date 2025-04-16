package helper

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func CheckUserIsLoggedInOrNot(c *fiber.Ctx) error {
	req := c.Request()
	tokenString := string(req.Header.Peek("Authorization"))

	if tokenString == "" {
		ApiResponse(c, http.StatusUnauthorized, "token is missing", nil)

		return nil
	}
	fmt.Println(tokenString)

	_, err := ValidateToken(tokenString)
	if err != nil {
		ApiResponse(c, http.StatusUnauthorized, "invalid tokens", nil)

		return nil
	}
	return nil
}
