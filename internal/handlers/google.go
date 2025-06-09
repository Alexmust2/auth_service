package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"auth.alexmust/internal/oauth"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	oauthConfig *oauth2.Config
)

func init() {
	_ = godotenv.Load()

	oauthConfig = oauth.GoogleConfig()
}

func GoogleLogin(c *fiber.Ctx) error {
	url := oauthConfig.AuthCodeURL("random-state-string")
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

func GoogleCallback(c *fiber.Ctx) error {
	if c.Query("state") != "random-state-string" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid state parameter")
	}

	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Code not found")
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Token exchange failed: " + err.Error())
	}

	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to get user info: " + err.Error())
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to parse user info")
	}

	email := userInfo["email"].(string)

	accessToken, err := GenerateJWT(email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("JWT generation failed")
	}
	refreshToken, err := GenerateRefreshToken(email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Refresh token generation failed")
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          userInfo,
	})
}
