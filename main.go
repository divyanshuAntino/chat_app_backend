package main

import (
	"fmt"
	"log"
	"os"

	"github.com/divyanshu050303/chat-app-backend/database"
	"github.com/divyanshu050303/chat-app-backend/models"
	"github.com/divyanshu050303/chat-app-backend/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	config := &database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}
	fmt.Println("Database Config:", config)
	db, err := database.NewConnection(config)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		log.Fatal("Could Not load datebase")
	}
	fmt.Println("Database Connection Established")
	err = models.Migrate(db)

	if err != nil {
		log.Fatal("Could not migrate the databse")
	}
	app := fiber.New()
	routes.SetUpUserRoutes(app, db)

	// Start Socket.IO server
	// controller.OnSocketConnect(&fiber.Ctx{}, db)

	// Start Fiber server
	app.Listen(":8000")
}
