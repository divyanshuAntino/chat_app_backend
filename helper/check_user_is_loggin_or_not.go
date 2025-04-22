package helper

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func CheckUserIsLoggedInOrNot(c *fiber.Ctx) string {
	req := c.Request()
	tokenString := string(req.Header.Peek("Authorization"))
	fmt.Printf("Token string: '%s'\n", tokenString) // Debug print

	if tokenString == "" {

		return "Token Missing"
	}

	_, err := ValidateToken(tokenString)
	if err != nil {

		return "invalid token"
	}

	fmt.Println("No auth errors") // Debug
	return ""
}
