package middleware

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"

	"github.com/libramusic/libracore"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/server/routes/auth"
)

var errTokenInvalidated = errors.New("token invalidated")

func JWTProtected(l *libracore.Libra) echo.MiddlewareFunc {
	key := auth.SigningKey(l.Config.Auth.JWT.SigningMethod, l.Config.Auth.JWT.SigningKey)

	return echojwt.WithConfig(echojwt.Config{
		SuccessHandler: jwtSuccessHandler(l.AuthDB),
		ErrorHandler:   jwtErrorHandler,
		SigningKey:     key,
		SigningMethod:  l.Config.Auth.JWT.SigningMethod,
		NewClaimsFunc: func(_ *echo.Context) jwt.Claims {
			return new(auth.TokenClaims)
		},
	})
}

func GlobalJWTProtected(l *libracore.Libra) echo.MiddlewareFunc {
	if l.Config.Auth.GlobalAPIRoutesRequireAuth {
		return JWTProtected(l)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}

func UserJWTProtected(l *libracore.Libra) echo.MiddlewareFunc {
	if l.Config.Auth.UserAPIRoutesRequireAuth {
		return JWTProtected(l)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}

func jwtSuccessHandler(authDB db.Database) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		isBlacklisted, err := authDB.IsTokenBlacklisted(c.Request().Context(), user.Raw)
		if err != nil {
			if jerr := c.JSON(http.StatusInternalServerError, map[string]any{
				"error": true,
				"msg":   err.Error(),
			}); jerr != nil {
				return jerr
			}
			return err
		}
		if isBlacklisted {
			if jerr := c.JSON(http.StatusUnauthorized, map[string]any{
				"error": true,
				"msg":   "Token invalidated",
			}); jerr != nil {
				return jerr
			}
			return errTokenInvalidated
		}
		return nil
	}
}

func jwtErrorHandler(c *echo.Context, err error) error {
	return c.JSON(http.StatusUnauthorized, map[string]any{
		"error": true,
		"msg":   err.Error(),
	})
}
