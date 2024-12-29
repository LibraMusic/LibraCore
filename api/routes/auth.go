package routes

import (
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/LibraMusic/LibraCore/config"
	"github.com/LibraMusic/LibraCore/db"
	"github.com/LibraMusic/LibraCore/types"
	"github.com/LibraMusic/LibraCore/utils"
)

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if config.Conf.Auth.DisableAccountCreation {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "account creation is disabled",
		})
	}

	if IsUsernameReserved(req.Username) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "username is reserved",
		})
	}

	usernameExists, err := db.DB.UsernameExists(req.Username)
	if err != nil {
		log.Error("error checking if username exists", "err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "internal server error",
		})
	}
	if usernameExists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "username already exists",
		})
	}

	if req.Email != "" {
		emailExists, err := db.DB.EmailExists(req.Email)
		if err != nil {
			log.Error("error checking if email exists", "err", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "internal server error",
			})
		}
		if emailExists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "email already exists",
			})
		}
	}

	user := types.User{
		ID:           utils.GenerateID(config.Conf.General.IDLength),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: utils.GeneratePassword(req.Password),
	}
	err = db.DB.CreateUser(user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	token, err := utils.GenerateToken(user.ID, config.Conf.Auth.JWTAccessTokenExpiration, config.Conf.Auth.JWTSigningMethod, config.Conf.Auth.JWTSigningKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	user, err := db.DB.GetUserByUsername(req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "user not found",
		})
	}
	if !utils.ComparePassword(user.PasswordHash, req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "incorrect password",
		})
	}

	token, err := utils.GenerateToken(user.ID, config.Conf.Auth.JWTAccessTokenExpiration, config.Conf.Auth.JWTSigningMethod, config.Conf.Auth.JWTSigningKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func Logout(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	err = db.DB.BlacklistToken(user.Raw, expiration.Time)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	c.Locals("user", nil)
	return c.SendStatus(fiber.StatusOK)
}

func GetReservedUsernames() []string {
	return []string{
		"default",
	}
}

func IsUsernameReserved(username string) bool {
	return slices.Contains(GetReservedUsernames(), username) || slices.ContainsFunc(config.Conf.General.ReservedUsernames, func(s string) bool {
		return strings.EqualFold(s, username)
	})
}
