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
	"github.com/libramusic/libracore/media"
	"github.com/libramusic/libracore/utils"
)

const RedirectURIQueryParam = "redirect_uri"

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

	user := media.DatabaseUser{
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

func LoginProvider(c echo.Context) error {
	redirectURI := c.QueryParam(RedirectURIQueryParam)
	if redirectURI == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": RedirectURIQueryParam + " is required",
		})
	}

	if _, err := gothic.CompleteUserAuth(c.Response(), c.Request()); err == nil {
		return c.Redirect(http.StatusFound, redirectURI)
	}

	if err := gothic.StoreInSession(RedirectURIQueryParam, redirectURI, c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to store " + RedirectURIQueryParam + " in session: " + err.Error(),
		})
	}

	gothic.BeginAuthHandler(c.Response(), c.Request())
	return nil
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

func ConnectProvider(c echo.Context) error {
	redirectURI := c.QueryParam(RedirectURIQueryParam)
	if redirectURI == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": RedirectURIQueryParam + " is required",
		})
	}

	if _, err := gothic.CompleteUserAuth(c.Response(), c.Request()); err == nil {
		return c.Redirect(http.StatusFound, redirectURI)
	}

	if err := gothic.StoreInSession(RedirectURIQueryParam, redirectURI, c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to store " + RedirectURIQueryParam + " in session: " + err.Error(),
		})
	}

	gothic.BeginAuthHandler(c.Response(), c.Request())
	return nil
}

func ProviderCallback(c echo.Context) error {
	providerUser, err := gothic.CompleteUserAuth(c.Response().Writer, c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to complete user auth: " + err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Check if the provider account is already linked to a user.
	user, err := db.DB.GetProviderUser(ctx, providerUser.Provider, providerUser.UserID)
	if !errors.Is(err, db.ErrNotFound) {
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": err.Error(),
			})
		}

		redirectURI, err := gothic.GetFromSession(RedirectURIQueryParam, c.Request())
		if err != nil || redirectURI == "" {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": RedirectURIQueryParam + " not found in session",
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

		if strings.Contains(redirectURI, "?") {
			redirectURI += "&token=" + token
		} else {
			redirectURI += "?token=" + token
		}

		return c.Redirect(http.StatusFound, redirectURI)
	}

	newUser := media.DatabaseUser{
		ID:          utils.GenerateID(config.Conf.General.IDLength),
		Username:    providerUser.UserID,
		Email:       providerUser.Email,
		DisplayName: providerUser.Name,
	}

	// TODO: Profile picture.

	if err := db.DB.CreateUser(ctx, newUser); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	if err := db.DB.LinkProviderAccount(ctx, providerUser.Provider, newUser.ID, providerUser.UserID); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	redirectURI, err := gothic.GetFromSession(RedirectURIQueryParam, c.Request())
	if err != nil || redirectURI == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": RedirectURIQueryParam + " not found in session",
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

	if strings.Contains(redirectURI, "?") {
		redirectURI += "&token=" + token
	} else {
		redirectURI += "?token=" + token
	}

	return c.Redirect(http.StatusFound, redirectURI)
}

func DisconnectProvider(c echo.Context) error {
	providerUser, err := gothic.CompleteUserAuth(c.Response().Writer, c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to complete user auth: " + err.Error(),
		})
	}
	gothic.Logout(c.Response().Writer, c.Request())

	ctx := c.Request().Context()

	_, err = db.DB.GetProviderUser(ctx, providerUser.Provider, providerUser.UserID)
	if errors.Is(err, db.ErrNotFound) {
		return c.JSON(http.StatusNotFound, echo.Map{
			"message": "provider account not found",
		})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	if err := db.DB.DisconnectProviderAccount(ctx, providerUser.Provider, providerUser.UserID); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	return c.NoContent(http.StatusOK)
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
