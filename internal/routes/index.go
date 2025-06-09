package routes

import (
	"auth.alexmust/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func IndexRoute(app *fiber.App, db interface{}) {
	app.Get("/login/google", handlers.GoogleLogin)
	app.Get("/callback/google", handlers.GoogleCallback)
	app.Get("/login/github", handlers.GitHubLogin)
	app.Get("/callback/github", handlers.GitHubCallback)

	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)
}
