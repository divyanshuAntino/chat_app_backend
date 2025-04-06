package controller

import (
	"github.com/divyanshu050303/chat-app-backend/repository"
	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	Repo *repository.UserRepository
}

func (ctrl *UserController) Createuser(c *fiber.Ctx) error {
	return nil
}
func (ctrl *UserController) LoginUser(c *fiber.Ctx) error {
	return nil
}
