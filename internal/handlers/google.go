package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"time"

	"auth.alexmust/internal/models"
	"auth.alexmust/internal/oauth"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var oauthConfig *oauth2.Config

func init() {
	_ = godotenv.Load()
	oauthConfig = oauth.GoogleConfig()
}

func GoogleLogin(c *fiber.Ctx) error {
	state, err := generateRandomState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate state"})
	}
	url := oauthConfig.AuthCodeURL(state)
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HTTPOnly: true,
	})
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

func GoogleCallback(c *fiber.Ctx, db *gorm.DB) error {
	if c.Query("state") != c.Cookies("oauth_state") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid state parameter"})
	}
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Code not found"})
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token exchange failed: " + err.Error()})
	}

	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user info: " + err.Error()})
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read user info: " + err.Error()})
	}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user info"})
	}

	email, ok := userInfo["email"].(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Email not found"})
	}

	accessToken, err := GenerateJWT(email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "JWT generation failed"})
	}
	refreshToken, err := GenerateRefreshToken(email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Refresh token generation failed"})
	}

	user := &models.User{
		Username:    userInfo["name"].(string),
		Email:       email,
		GoogleToken: token.AccessToken,
		Avatar_url:  userInfo["picture"].(string),
	}

	var existingUser models.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		existingUser.GoogleToken = token.AccessToken
		existingUser.Avatar_url = userInfo["picture"].(string)
		if err := db.Save(&existingUser).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user: " + err.Error()})
		}
		user = &existingUser
	} else {
		if err := db.Create(user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user: " + err.Error()})
		}
	}

	// Set HTTP-only cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   false, // Set to true in production
		SameSite: "Lax",
	})

	// Redirect to frontend callback
	return c.Redirect("http://localhost:5173/callback", fiber.StatusTemporaryRedirect)
}

func generateRandomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
