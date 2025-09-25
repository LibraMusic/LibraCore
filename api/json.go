package api

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// JSONV2Serializer implements JSON encoding using encoding/json/v2.
type JSONV2Serializer struct{}

// Serialize converts an interface into a json and writes it to the response.
// You can optionally use the indent parameter to produce pretty JSONs.
func (JSONV2Serializer) Serialize(c echo.Context, i any, indent string) error {
	if indent != "" {
		return json.MarshalWrite(c.Response(), i, jsontext.WithIndent(indent))
	}
	return json.MarshalWrite(c.Response(), i)
}

// Deserialize reads a JSON from a request body and converts it into an interface.
func (JSONV2Serializer) Deserialize(c echo.Context, i any) error {
	err := json.UnmarshalRead(c.Request().Body, i)

	var synctaticError *jsontext.SyntacticError
	var semanticError *json.SemanticError

	if errors.As(err, &synctaticError) {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", synctaticError.ByteOffset, synctaticError.Error())).
			SetInternal(err)
	} else if errors.As(err, &semanticError) {
		if semanticError.GoType != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", semanticError.GoType, semanticError.JSONKind, semanticError.JSONPointer, semanticError.ByteOffset)).SetInternal(err)
		}
	}
	return err
}
