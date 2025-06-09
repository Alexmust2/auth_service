package handlers

import (
	"auth.alexmust/internal/services"
	"github.com/gofiber/fiber/v2"
)

var userService = services.NewUserService()

func Register(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}
	if err := userService.Register(req.Email, req.Password); err != nil {
		return c.Status(fiber.StatusConflict).SendString(err.Error())
	}
	return c.SendStatus(fiber.StatusCreated)
}

func Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}
	user, err := userService.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}
	accessToken, _ := GenerateJWT(user.Email)
	refreshToken, _ := GenerateRefreshToken(user.Email)
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
