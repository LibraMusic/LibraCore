package httpcontext

import (
	"github.com/labstack/echo/v4"

	"github.com/libramusic/libracore"
)

// Context is a custom context that extends echo.Context with Libra-specific fields.
type Context struct {
	echo.Context

	Libra *libracore.Libra
}
