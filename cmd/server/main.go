package main

import (
	"log"

	config "auth.alexmust/configs"
	"auth.alexmust/internal/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	config.ConnectDatabase()

	app := fiber.New()

	app.Use(cors.New())

	routes.IndexRoute(app, config.DB)

	app.Listen(":3000")
}
