package routes

import (
	"auth.alexmust/internal/handlers"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func IndexRoute(app *fiber.App, db *gorm.DB) {
	app.Get("/login/google", handlers.GoogleLogin)
	app.Get("/callback/google", func(c *fiber.Ctx) error {
		return handlers.GoogleCallback(c, db)
	})
	app.Get("/login/github", handlers.GitHubLogin)
	app.Get("/callback/github", func(c *fiber.Ctx) error {
		return handlers.GitHubCallback(c, db)
	})

}
