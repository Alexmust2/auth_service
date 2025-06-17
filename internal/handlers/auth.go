package handlers

import (
	"auth.alexmust/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"time"
)

func RegisterUser(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user models.User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неправильные данные"})
		}

		if user.Email == "" || user.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Почта и пароль обязательны"})
		}

		var existingUser models.User
		if err := db.Unscoped().Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			if existingUser.DeletedAt.Valid {
				if err := db.Model(&existingUser).Update("deleted_at", nil).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка восстановления аккаунта"})
				}
				return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Пользователь восстановлен"})
			}
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Пользователь с такой почтой уже существует"})
		}

		if err := db.Create(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка создания пользователя"})
		}

		accessToken, err := GenerateJWT(user.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка получения токена"})
		}
		refreshToken, err := GenerateRefreshToken(user.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка получения refresh токена"})
		}

		// Set HTTP-only cookies
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Expires:  time.Now().Add(15 * time.Minute),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
		})
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
		})

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"user": fiber.Map{
				"id":         user.ID,
				"email":      user.Email,
				"username":   user.Username,
				"avatar_url": user.Avatar_url,
			},
		})
	}
}

func LoginUser(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var loginData struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&loginData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неправильные данные"})
		}

		if loginData.Email == "" || loginData.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Почта и пароль обязательны"})
		}

		var user models.User
		if err := db.Where("email = ?", loginData.Email).First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неправильная почта или пароль"})
		}

		if err := user.ComparePassword(loginData.Password); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неправильная почта или пароль"})
		}

		accessToken, err := GenerateJWT(user.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка получения токена"})
		}
		refreshToken, err := GenerateRefreshToken(user.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка получения refresh токена"})
		}

		// Set HTTP-only cookies
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Expires:  time.Now().Add(15 * time.Minute),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
		})
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
		})

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"user": fiber.Map{
				"id":         user.ID,
				"email":      user.Email,
				"username":   user.Username,
				"avatar_url": user.Avatar_url,
			},
		})
	}
}
