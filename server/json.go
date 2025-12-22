package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
)

// GoJSONSerializer implements JSON encoding using github.com/goccy/go-json.
type GoJSONSerializer struct{}

// Serialize converts an interface into a json and writes it to the response.
// You can optionally use the indent parameter to produce pretty JSONs.
func (GoJSONSerializer) Serialize(c echo.Context, i any, indent string) error {
	enc := json.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

// Deserialize reads a JSON from a request body and converts it into an interface.
func (GoJSONSerializer) Deserialize(c echo.Context, i any) error {
	err := json.NewDecoder(c.Request().Body).Decode(i)
	var ute *json.UnmarshalTypeError
	var se *json.SyntaxError
	if errors.As(err, &ute) {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)).
			SetInternal(err)
	} else if errors.As(err, &se) {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())).SetInternal(err)
	}
	return err
}
