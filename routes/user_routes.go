package routes

import (
	"github.com/divyanshu050303/chat-app-backend/controller"
	"github.com/divyanshu050303/chat-app-backend/repository"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetUpUserRoutes(app *fiber.App, db *gorm.DB) {
	userRepository := &repository.UserRepository{DB: db}
	userController := &controller.UserController{Repo: userRepository}
	api := app.Group("/api/user")
	api.Post("createUser", userController.Createuser)
	api.Post("/login", userController.LoginUser)
	api.Patch("/user-profile", userController.Createuser)
}
