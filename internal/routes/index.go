package routes

import (
	"auth.alexmust/internal/handlers"
	"auth.alexmust/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func IndexRoute(app *fiber.App, db *gorm.DB) {
	api := app.Group("/api/v1")
	api.Post("/login", handlers.LoginUser(db))
	api.Post("/register", handlers.RegisterUser(db))
	api.Get("/login/google", handlers.GoogleLogin)
	api.Get("/callback/google", func(c *fiber.Ctx) error {
		return handlers.GoogleCallback(c, db)
	})
	api.Get("/login/github", handlers.GitHubLogin)
	api.Get("/callback/github", func(c *fiber.Ctx) error {
		return handlers.GitHubCallback(c, db)
	})
	api.Get("/me", middleware.JWTProtected(), func(c *fiber.Ctx) error {
		return handlers.GetMe(c, db)
	})
	api.Post("/logout", middleware.JWTProtected(), func(c *fiber.Ctx) error {
		return handlers.Logout(c, db)
	})
	api.Post("/refresh", handlers.RefreshToken)
}
