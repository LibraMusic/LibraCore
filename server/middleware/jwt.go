package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/server/routes/auth"
)

func JWTProtected(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key := auth.SigningKey(config.Conf.Auth.JWT.SigningMethod, config.Conf.Auth.JWT.SigningKey)

		config := echojwt.Config{
			SuccessHandler: jwtSuccessHandler,
			ErrorHandler:   jwtErrorHandler,
			SigningKey:     key,
			SigningMethod:  config.Conf.Auth.JWT.SigningMethod,
			NewClaimsFunc: func(_ echo.Context) jwt.Claims {
				return new(auth.TokenClaims)
			},
		}

		return echojwt.WithConfig(config)(next)(c)
	}
}

func GlobalJWTProtected(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if config.Conf.Auth.GlobalAPIRoutesRequireAuth {
			return JWTProtected(next)(c)
		}
		return next(c)
	}
}

func UserJWTProtected(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if config.Conf.Auth.UserAPIRoutesRequireAuth {
			return JWTProtected(next)(c)
		}
		return next(c)
	}
}

func jwtSuccessHandler(c echo.Context) {
	user := c.Get("user").(*jwt.Token)
	isBlacklisted, err := db.DB.IsTokenBlacklisted(c.Request().Context(), user.Raw)
	if err != nil {
		_ = c.JSON(http.StatusInternalServerError, echo.Map{
			"error": true,
			"msg":   err.Error(),
		})
		return
	}
	if isBlacklisted {
		_ = c.JSON(http.StatusUnauthorized, echo.Map{
			"error": true,
			"msg":   "Token invalidated",
		})
	}
}

func jwtErrorHandler(c echo.Context, err error) error {
	return c.JSON(http.StatusUnauthorized, echo.Map{
		"error": true,
		"msg":   err.Error(),
	})
}
