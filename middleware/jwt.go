package middleware

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/DevReaper0/libra/config"
	"github.com/DevReaper0/libra/db"
	"github.com/DevReaper0/libra/util"
)

func JWTProtected(c *fiber.Ctx) error {
	var key interface{}
	switch config.Conf.Auth.JWTSigningMethod {
	case "HS256", "HS384", "HS512":
		key = []byte(config.Conf.Auth.JWTSigningKey)
	case "RS256", "RS384", "RS512", "PS256", "PS384", "PS512":
		key = util.RSAPrivateKey.Public()
	case "ES256", "ES384", "ES512":
		key = util.ECDSAPrivateKey.Public()
	case "EdDSA":
		key = util.EdDSAPrivateKey.Public()
	}

	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{
			JWTAlg: config.Conf.Auth.JWTSigningMethod,
			Key:    key,
		},
		ContextKey: "jwt",
		SuccessHandler: func(c *fiber.Ctx) error {
			user := c.Locals("user").(*jwt.Token)
			isBlacklisted, err := db.DB.IsTokenBlacklisted(user.Raw)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": true,
					"msg":   err.Error(),
				})
			}
			if isBlacklisted {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": true,
					"msg":   "Token invalidated",
				})
			}
			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		},
	})(c)
}

func GlobalJWTProtected(c *fiber.Ctx) error {
	if config.Conf.Auth.GlobalAPIRoutesRequireAuth {
		return JWTProtected(c)
	}
	return c.Next()
}

func UserJWTProtected(c *fiber.Ctx) error {
	if config.Conf.Auth.UserAPIRoutesRequireAuth {
		return JWTProtected(c)
	}
	return c.Next()
}
