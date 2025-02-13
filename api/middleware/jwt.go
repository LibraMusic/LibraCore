package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/utils"
)

func JWTProtected(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var key interface{}
		switch config.Conf.Auth.JWT.SigningMethod {
		case "HS256", "HS384", "HS512":
			key = []byte(config.Conf.Auth.JWT.SigningKey)
		case "RS256", "RS384", "RS512", "PS256", "PS384", "PS512":
			key = utils.RSAPrivateKey.Public()
		case "ES256", "ES384", "ES512":
			key = utils.ECDSAPrivateKey.Public()
		case "EdDSA":
			key = utils.EdDSAPrivateKey.Public()
		}

		config := echojwt.Config{
			SigningKey:     key,
			SigningMethod:  config.Conf.Auth.JWT.SigningMethod,
			ContextKey:     "jwt",
			SuccessHandler: jwtSuccessHandler,
			ErrorHandler:   jwtErrorHandler,
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
	user := c.Get("jwt").(*jwt.Token)
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
