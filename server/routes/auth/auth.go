package auth

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"

	"charm.land/log/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"

	"github.com/libramusic/libracore"
	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/media"
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

func Register(c *echo.Context) error {
	l := c.Get("libra").(*libracore.Libra)

	if l.Config.Auth.DisableAccountCreation {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": "account creation is disabled",
		})
	}

	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": err.Error(),
		})
	}

	if IsUsernameReserved(&l.Config, req.Username) {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": "username is reserved",
		})
	}

	ctx := c.Request().Context()

	usernameExists, err := l.AuthDB.UsernameExists(ctx, req.Username)
	if err != nil {
		log.Error("error checking if username exists", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": "internal server error",
		})
	}
	if usernameExists {
		return c.JSON(http.StatusConflict, map[string]any{
			"message": "username already exists",
		})
	}

	if req.Email != "" {
		emailExists, err := l.AuthDB.EmailExists(ctx, req.Email)
		if err != nil {
			log.Error("error checking if email exists", "err", err)
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"message": "internal server error",
			})
		}
		if emailExists {
			return c.JSON(http.StatusConflict, map[string]any{
				"message": "email already exists",
			})
		}
	}

	user := media.DatabaseUser{
		ID:           media.GenerateID(l.Config.General.IDLength),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: generatePassword(req.Password),
	}
	err = l.AuthDB.AddUser(ctx, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	token, err := GenerateToken(
		user.ID,
		l.Config.Auth.JWT.AccessTokenExpiration,
		l.Config.Auth.JWT.SigningMethod,
		l.Config.Auth.JWT.SigningKey,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token": token,
	})
}

func Login(c *echo.Context) error {
	l := c.Get("libra").(*libracore.Libra)

	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": err.Error(),
		})
	}

	ctx := c.Request().Context()

	user, err := l.AuthDB.UserByUsername(ctx, req.Username)
	if errors.Is(err, db.ErrNotFound) {
		return c.JSON(http.StatusNotFound, map[string]any{
			"message": "user not found",
		})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}
	if !comparePassword(user.PasswordHash, req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]any{
			"message": "incorrect password",
		})
	}

	token, err := GenerateToken(
		user.ID,
		l.Config.Auth.JWT.AccessTokenExpiration,
		l.Config.Auth.JWT.SigningMethod,
		l.Config.Auth.JWT.SigningKey,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token": token,
	})
}

func LoginProvider(c *echo.Context) error {
	redirectURI := c.QueryParam(RedirectURIQueryParam)
	if redirectURI == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": RedirectURIQueryParam + " is required",
		})
	}

	if _, err := gothic.CompleteUserAuth(c.Response(), c.Request()); err == nil {
		return c.Redirect(http.StatusFound, redirectURI)
	}

	if err := gothic.StoreInSession(RedirectURIQueryParam, redirectURI, c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": "failed to store " + RedirectURIQueryParam + " in session: " + err.Error(),
		})
	}

	gothic.BeginAuthHandler(c.Response(), c.Request())
	return nil
}

func Logout(c *echo.Context) error {
	l := c.Get("libra").(*libracore.Libra)

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*TokenClaims)
	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	ctx := c.Request().Context()

	err = l.AuthDB.BlacklistToken(ctx, user.Raw, expiration.Time)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}
	c.Set("user", nil)
	return c.NoContent(http.StatusOK)
}

func ConnectProvider(c *echo.Context) error {
	redirectURI := c.QueryParam(RedirectURIQueryParam)
	if redirectURI == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": RedirectURIQueryParam + " is required",
		})
	}

	if _, err := gothic.CompleteUserAuth(c.Response(), c.Request()); err == nil {
		return c.Redirect(http.StatusFound, redirectURI)
	}

	if err := gothic.StoreInSession(RedirectURIQueryParam, redirectURI, c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": "failed to store " + RedirectURIQueryParam + " in session: " + err.Error(),
		})
	}

	gothic.BeginAuthHandler(c.Response(), c.Request())
	return nil
}

func ProviderCallback(c *echo.Context) error {
	l := c.Get("libra").(*libracore.Libra)

	providerUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": "failed to complete user auth: " + err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Check if the provider account is already linked to a user.
	user, err := l.AuthDB.ProviderUser(ctx, providerUser.Provider, providerUser.UserID)
	if errors.Is(err, db.ErrNotFound) {
		return handleNewProviderUser(ctx, c, providerUser)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	return handleExistingProviderUser(c, user)
}

func handleExistingProviderUser(c *echo.Context, user media.DatabaseUser) error {
	l := c.Get("libra").(*libracore.Libra)

	redirectURI, err := gothic.GetFromSession(RedirectURIQueryParam, c.Request())
	if err != nil || redirectURI == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": RedirectURIQueryParam + " not found in session",
		})
	}

	token, err := GenerateToken(
		user.ID,
		l.Config.Auth.JWT.AccessTokenExpiration,
		l.Config.Auth.JWT.SigningMethod,
		l.Config.Auth.JWT.SigningKey,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	redirectURL := buildRedirectURL(redirectURI, token)
	return c.Redirect(http.StatusFound, redirectURL)
}

func handleNewProviderUser(ctx context.Context, c *echo.Context, providerUser goth.User) error {
	l := c.Get("libra").(*libracore.Libra)

	newUser := media.DatabaseUser{
		ID:          media.GenerateID(l.Config.General.IDLength),
		Username:    providerUser.UserID,
		Email:       providerUser.Email,
		DisplayName: providerUser.Name,
	}

	// TODO: Profile picture.

	if err := l.AuthDB.AddUser(ctx, newUser); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	if err := l.AuthDB.LinkProviderAccount(
		ctx,
		providerUser.Provider,
		newUser.ID,
		providerUser.UserID,
	); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	redirectURI, err := gothic.GetFromSession(RedirectURIQueryParam, c.Request())
	if err != nil || redirectURI == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": RedirectURIQueryParam + " not found in session",
		})
	}

	token, err := GenerateToken(
		newUser.ID,
		l.Config.Auth.JWT.AccessTokenExpiration,
		l.Config.Auth.JWT.SigningMethod,
		l.Config.Auth.JWT.SigningKey,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	redirectURL := buildRedirectURL(redirectURI, token)
	return c.Redirect(http.StatusFound, redirectURL)
}

func buildRedirectURL(redirectURI, token string) string {
	if strings.Contains(redirectURI, "?") {
		redirectURI += "&token=" + token
	}
	return redirectURI + "?token=" + token
}

func DisconnectProvider(c *echo.Context) error {
	l := c.Get("libra").(*libracore.Libra)

	providerUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": "failed to complete user auth: " + err.Error(),
		})
	}
	_ = gothic.Logout(c.Response(), c.Request())

	ctx := c.Request().Context()

	_, err = l.AuthDB.ProviderUser(ctx, providerUser.Provider, providerUser.UserID)
	if errors.Is(err, db.ErrNotFound) {
		return c.JSON(http.StatusNotFound, map[string]any{
			"message": "provider account not found",
		})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}
	if err := l.AuthDB.DisconnectProviderAccount(ctx, providerUser.Provider, providerUser.UserID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"message": err.Error(),
		})
	}

	return c.NoContent(http.StatusOK)
}

func ReservedUsernames() []string {
	return []string{
		"default",
	}
}

func IsUsernameReserved(conf *config.Config, username string) bool {
	return slices.Contains(ReservedUsernames(), username) ||
		slices.ContainsFunc(conf.General.ReservedUsernames, func(s string) bool {
			return strings.EqualFold(s, username)
		})
}
