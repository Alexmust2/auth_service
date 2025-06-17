package handlers

import (
	"context"
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

var oauthGitHubConfig *oauth2.Config

func init() {
	_ = godotenv.Load()
	oauthGitHubConfig = oauth.GitHubConfig()
}

func GitHubLogin(c *fiber.Ctx) error {
	state, err := generateRandomState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate state"})
	}
	url := oauthGitHubConfig.AuthCodeURL(state)
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HTTPOnly: true,
	})
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

func GitHubCallback(c *fiber.Ctx, db *gorm.DB) error {
	if c.Query("state") != c.Cookies("oauth_state") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid state"})
	}
	code := c.Query("code")
	token, err := oauthGitHubConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token exchange error: " + err.Error()})
	}

	client := oauthGitHubConfig.Client(context.Background(), token)
	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error fetching user data"})
	}
	defer userResp.Body.Close()

	var userData map[string]interface{}
	userBody, err := io.ReadAll(userResp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error reading user data"})
	}
	if err := json.Unmarshal(userBody, &userData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error parsing user data"})
	}

	emailsResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error fetching emails"})
	}
	defer emailsResp.Body.Close()

	var emails []map[string]interface{}
	emailsBody, err := io.ReadAll(emailsResp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error reading emails"})
	}
	if err := json.Unmarshal(emailsBody, &emails); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error parsing emails"})
	}

	var email string
	for _, e := range emails {
		if primary, ok := e["primary"].(bool); ok && primary {
			email = e["email"].(string)
			break
		}
	}
	if email == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Primary email not found"})
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
		Username:    userData["name"].(string),
		Email:       email,
		GithubToken: token.AccessToken,
		Avatar_url:  userData["avatar_url"].(string),
	}

	var existingUser models.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		existingUser.GithubToken = token.AccessToken
		existingUser.Avatar_url = userData["avatar_url"].(string)
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
		Secure:   false, // Set to true in production
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
