package routes

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"

	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/types"
	"github.com/libramusic/libracore/utils"
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

	ctx := c.Request().Context()

	usernameExists, err := db.DB.UsernameExists(ctx, req.Username)
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
		emailExists, err := db.DB.EmailExists(ctx, req.Email)
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

	user := types.DatabaseUser{
		ID:           utils.GenerateID(config.Conf.General.IDLength),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: utils.GeneratePassword(req.Password),
	}
	err = db.DB.CreateUser(ctx, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	token, err := utils.GenerateToken(
		user.ID,
		config.Conf.Auth.JWT.AccessTokenExpiration,
		config.Conf.Auth.JWT.SigningMethod,
		config.Conf.Auth.JWT.SigningKey,
	)
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

	ctx := c.Request().Context()

	user, err := db.DB.GetUserByUsername(ctx, req.Username)
	if errors.Is(err, db.ErrNotFound) {
		return c.JSON(http.StatusNotFound, echo.Map{
			"message": "user not found",
		})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	if !utils.ComparePassword(user.PasswordHash, req.Password) {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"message": "incorrect password",
		})
	}

	token, err := utils.GenerateToken(
		user.ID,
		config.Conf.Auth.JWT.AccessTokenExpiration,
		config.Conf.Auth.JWT.SigningMethod,
		config.Conf.Auth.JWT.SigningKey,
	)
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

	ctx := c.Request().Context()

	err = db.DB.BlacklistToken(ctx, user.Raw, expiration.Time)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	c.Set("user", nil)
	return c.NoContent(http.StatusOK)
}

func OAuthLogout(c echo.Context) error {
	// TODO: Implement OAuth logout.
	return c.NoContent(http.StatusNotImplemented)
}

func OAuthCallback(c echo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response().Writer, c.Request())
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	existingUser, err := db.DB.GetOAuthUser(ctx, user.Provider, user.UserID)
	if !errors.Is(err, db.ErrNotFound) {
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": err.Error(),
			})
		}
		token, err := utils.GenerateToken(
			existingUser.ID,
			config.Conf.Auth.JWT.AccessTokenExpiration,
			config.Conf.Auth.JWT.SigningMethod,
			config.Conf.Auth.JWT.SigningKey,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, echo.Map{
			"token": token,
		})
	}

	// TODO: Check for existing user (not OAuth).

	newUser := types.DatabaseUser{
		ID:          utils.GenerateID(config.Conf.General.IDLength),
		Username:    user.NickName,
		Email:       user.Email,
		DisplayName: user.Name,
	}

	// TODO: Profile picture.

	if err := db.DB.CreateUser(ctx, newUser); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	if err := db.DB.LinkOAuthAccount(ctx, user.Provider, newUser.ID, user.UserID); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	token, err := utils.GenerateToken(
		newUser.ID,
		config.Conf.Auth.JWT.AccessTokenExpiration,
		config.Conf.Auth.JWT.SigningMethod,
		config.Conf.Auth.JWT.SigningKey,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	// TODO: Make sure this is correct and working.
	// TODO: Maybe figure out frontend redirection since the login route is meant to be called as an API but the OAuth callback is redirected to by the OAuth provider.
	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
	})
}

func OAuth(c echo.Context) error {
	// provider := c.Param("provider")
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{"message": "OAuth callback failed", "error": err.Error()},
		)
	}

	// TODO

	return c.JSON(http.StatusOK, user)
}

func GetReservedUsernames() []string {
	return []string{
		"default",
	}
}

func IsUsernameReserved(username string) bool {
	return slices.Contains(GetReservedUsernames(), username) ||
		slices.ContainsFunc(config.Conf.General.ReservedUsernames, func(s string) bool {
			return strings.EqualFold(s, username)
		})
}
