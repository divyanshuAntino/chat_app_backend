package helper

import "github.com/gofiber/fiber/v2"

func ApiResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"status":  statusCode,
		"message": message,
		"data":    data})
}
