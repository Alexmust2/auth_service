package handlers

import (
	"time"

	"auth.alexmust/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetMe(c *fiber.Ctx, db *gorm.DB) error {
	email, ok := c.Locals("email").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(fiber.Map{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"avatar_url": user.Avatar_url,
	})
}

func Logout(c *fiber.Ctx, db *gorm.DB) error {
	// Clear cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}
