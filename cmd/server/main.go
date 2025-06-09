package main

import (
	"auth.alexmust/internal/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New())
	routes.IndexRoute(app, nil) // Pass DB if needed

	app.Get("/", func(c *fiber.Ctx) error {
		routes := app.GetRoutes()
		routeList := make([]string, 0)

		for _, route := range routes {
			routeList = append(routeList, route.Path)
		}

		return c.JSON(fiber.Map{
			"available_routes": routeList,
		})
	})

	app.Listen(":3000")
}
