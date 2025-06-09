package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"auth.alexmust/internal/models"
	"auth.alexmust/internal/oauth"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var (
	oauthGitHubConfig *oauth2.Config
)

func init() {
	_ = godotenv.Load()

	oauthGitHubConfig = oauth.GitHubConfig()
}

func GitHubLogin(c *fiber.Ctx) error {
	url := oauthGitHubConfig.AuthCodeURL("github-state")
	return c.Redirect(url)
}

func GitHubCallback(c *fiber.Ctx, db *gorm.DB) error {
	if c.Query("state") != "github-state" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid state")
	}
	code := c.Query("code")
	token, err := oauthGitHubConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Token exchange error: " + err.Error())
	}

	client := oauthGitHubConfig.Client(context.Background(), token)

	// Get user profile information
	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching user data")
	}
	defer userResp.Body.Close()

	var userData map[string]interface{}
	userBody, _ := ioutil.ReadAll(userResp.Body)
	_ = json.Unmarshal(userBody, &userData)

	// Get user emails
	emailsResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching emails")
	}
	defer emailsResp.Body.Close()

	var emails []map[string]interface{}
	emailsBody, _ := ioutil.ReadAll(emailsResp.Body)
	_ = json.Unmarshal(emailsBody, &emails)

	var email string
	for _, e := range emails {
		if primary, ok := e["primary"].(bool); ok && primary {
			email = e["email"].(string)
			break
		}
	}
	if email == "" {
		return c.Status(fiber.StatusInternalServerError).SendString("Primary email not found")
	}

	accessToken, _ := GenerateJWT(email)
	refreshToken, _ := GenerateRefreshToken(email)

	user := &models.User{
		Username:    userData["name"].(string),
		Email:       email,
		GithubToken: token.AccessToken,
	}

	// Проверяем, существует ли пользователь
	var existingUser models.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		// Обновляем существующего пользователя
		existingUser.GoogleToken = token.AccessToken
		existingUser.Avatar_url = userData["avatar_url"].(string)
		if err := db.Save(&existingUser).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to update user: " + err.Error())
		}
		user = &existingUser
	} else {
		if err := db.Create(user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to create user: " + err.Error())
		}
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"email":         email,
		"user_data":     userData,
	})
}
