package routes

import (
	"net/http"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

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

func Register(c echo.Context) error {
	if config.Conf.Auth.DisableAccountCreation {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "account creation is disabled",
		})
	}

	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}

	if IsUsernameReserved(req.Username) {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "username is reserved",
		})
	}

	usernameExists, err := db.DB.UsernameExists(req.Username)
	if err != nil {
		log.Error("error checking if username exists", "err", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "internal server error",
		})
	}
	if usernameExists {
		return c.JSON(http.StatusConflict, echo.Map{
			"message": "username already exists",
		})
	}

	if req.Email != "" {
		emailExists, err := db.DB.EmailExists(req.Email)
		if err != nil {
			log.Error("error checking if email exists", "err", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "internal server error",
			})
		}
		if emailExists {
			return c.JSON(http.StatusConflict, echo.Map{
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
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	token, err := utils.GenerateToken(user.ID, config.Conf.Auth.JWT.AccessTokenExpiration, config.Conf.Auth.JWT.SigningMethod, config.Conf.Auth.JWT.SigningKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
	})
}

func Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}
	user, err := db.DB.GetUserByUsername(req.Username)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"message": "user not found",
		})
	}
	if !utils.ComparePassword(user.PasswordHash, req.Password) {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"message": "incorrect password",
		})
	}

	token, err := utils.GenerateToken(user.ID, config.Conf.Auth.JWT.AccessTokenExpiration, config.Conf.Auth.JWT.SigningMethod, config.Conf.Auth.JWT.SigningKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
	})
}

func Logout(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	err = db.DB.BlacklistToken(user.Raw, expiration.Time)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	c.Set("user", nil)
	return c.NoContent(http.StatusOK)
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
