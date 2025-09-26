package server

import (
	"encoding/json/v2"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/libramusic/libracore"
	"github.com/libramusic/libracore/api/routes"
	"github.com/libramusic/libracore/config"
)

func GetOpenAPISpec() echo.Map {
	v1Spec := loadBaseOpenAPISpec()

	processSchemas(v1Spec)
	processPaths(v1Spec)

	if config.Conf.General.DocumentFeedRoutes {
		addFeedRoutesDocumentation(v1Spec)
	}

	return v1Spec
}

// loadBaseOpenAPISpec loads the OpenAPI spec and updates version and server URL info.
func loadBaseOpenAPISpec() echo.Map {
	SwaggerInfo.Version = libracore.LibraVersion.String()
	v1SpecJSON := SwaggerInfo.ReadDoc()
	var v1Spec echo.Map
	if err := json.Unmarshal([]byte(v1SpecJSON), &v1Spec); err != nil {
		log.Fatal("Error unmarshalling OpenAPI spec", "err", err)
	}

	// Update server URL
	servers := v1Spec["servers"].([]any)
	server := servers[0].(map[string]any)
	server["url"] = strings.ReplaceAll(
		server["url"].(string),
		"http://localhost:8080",
		config.Conf.Application.PublicURL,
	)

	return v1Spec
}

// processSchemas cleans up schema references and creates a proper Playable interface schema.
func processSchemas(v1Spec echo.Map) {
	schemas := v1Spec["components"].(map[string]any)["schemas"].(map[string]any)

	// Remove "media." prefix from schema names
	for key := range schemas {
		if strings.HasPrefix(key, "media.") {
			newKey := strings.TrimPrefix(key, "media.")
			schemas[newKey] = schemas[key]
			delete(schemas, key)
		}
	}

	// Replace the fakePlayable schema with the actual Playable interface schema
	playableSchema := map[string]any{
		"oneOf": []any{
			map[string]any{"$ref": "#/components/schemas/Track"},
			map[string]any{"$ref": "#/components/schemas/Album"},
			map[string]any{"$ref": "#/components/schemas/Video"},
			map[string]any{"$ref": "#/components/schemas/Playlist"},
			map[string]any{"$ref": "#/components/schemas/Artist"},
			map[string]any{"$ref": "#/components/schemas/User"},
		},
	}
	schemas["Playable"] = playableSchema
	delete(schemas, "routes.fakePlayable")
}

// processPaths removes placeholder paths and standardizes path definitions.
func processPaths(v1Spec echo.Map) {
	paths := v1Spec["paths"].(map[string]any)
	delete(paths, "/fake")

	for _, pathObj := range paths {
		objMap, ok := pathObj.(map[string]any)
		if !ok {
			continue
		}
		processPathOperations(objMap)
	}
}

// processPathOperations updates response schema references in API operations.
func processPathOperations(pathObj map[string]any) {
	for _, op := range pathObj {
		opMap, ok := op.(map[string]any)
		if !ok {
			continue
		}
		responses, ok := opMap["responses"].(map[string]any)
		if !ok {
			continue
		}

		for _, response := range responses {
			processResponseObject(response)
		}
	}
}

// processResponseObject updates schema references in JSON response content.
func processResponseObject(response any) {
	resp, ok := response.(map[string]any)
	if !ok {
		return
	}

	content, ok := resp["content"].(map[string]any)
	if !ok {
		return
	}

	jsonContent, ok := content["application/json"].(map[string]any)
	if !ok {
		return
	}

	schema, ok := jsonContent["schema"].(map[string]any)
	if !ok {
		return
	}

	processSchema(schema)
}

// processSchema normalizes references in schema objects, oneOf lists, and array items.
func processSchema(schema map[string]any) {
	if ref, ok := schema["$ref"].(string); ok {
		schema["$ref"] = updateReference(ref)
	} else if oneOf, ok := schema["oneOf"].([]any); ok {
		for i, item := range oneOf {
			itemRef, ok := item.(string)
			if !ok {
				continue
			}
			oneOf[i] = updateReference(itemRef)
		}
	} else if items, ok := schema["items"].(map[string]any); ok {
		if ref, ok := items["$ref"].(string); ok {
			items["$ref"] = updateReference(ref)
		}
	}
}

// updateReference converts schema references to their canonical form.
func updateReference(ref string) string {
	switch {
	case ref == "#/components/schemas/routes.fakePlayable":
		return "#/components/schemas/Playable"
	case strings.HasPrefix(ref, "#/components/schemas/media."):
		newRef := strings.TrimPrefix(ref, "#/components/schemas/media.")
		return "#/components/schemas/" + newRef
	default:
		return ref
	}
}

// addFeedRoutesDocumentation adds feed endpoints (RSS, Atom, JSON) to the API spec.
func addFeedRoutesDocumentation(v1Spec echo.Map) {
	paths := v1Spec["paths"].(map[string]any)

	for _, feedRoute := range routes.FeedRoutesDoc {
		content := createFeedRouteContent(feedRoute.Type)
		description := createFeedRouteDescription(feedRoute.Type)

		path := map[string]any{
			"get": map[string]any{
				"summary": feedRoute.Summary,
				"responses": map[string]any{
					"200": map[string]any{
						"description": description,
						"content":     content,
					},
				},
			},
		}

		copyBasePathInfo(paths, feedRoute.BasePath, path)
		paths[feedRoute.Path] = path
	}
}

// createFeedRouteContent generates content type and schema for feed routes.
func createFeedRouteContent(feedType string) map[string]any {
	var contentType, description string
	var schema map[string]any

	if feedType == "JSON" {
		contentType = "application/json"
		description = "JSON feed response"
		schema = map[string]any{
			"type": "object",
		}
	} else {
		contentType = "application/xml"
		description = feedType + " XML feed response"
		schema = map[string]any{
			"type": "string",
		}
	}

	return map[string]any{
		contentType: map[string]any{
			"schema":      schema,
			"description": description,
		},
	}
}

// createFeedRouteDescription generates descriptions for feed endpoints.
func createFeedRouteDescription(feedType string) string {
	if feedType == "JSON" {
		return "JSON feed response"
	}
	return feedType + " XML feed response"
}

// copyBasePathInfo inherits parameters and error responses from a base path.
func copyBasePathInfo(paths map[string]any, basePath string, targetPath map[string]any) {
	basePathObj, ok := paths[basePath].(map[string]any)
	if !ok {
		return
	}

	baseRouteGet, ok := basePathObj["get"].(map[string]any)
	if !ok {
		return
	}

	// Copy parameters
	if parameters, ok := baseRouteGet["parameters"].([]any); ok {
		targetPath["get"].(map[string]any)["parameters"] = parameters
	}

	// Copy non-200 responses
	if responses, ok := baseRouteGet["responses"].(map[string]any); ok {
		targetResponses := targetPath["get"].(map[string]any)["responses"].(map[string]any)
		for key, value := range responses {
			if key == "200" {
				continue
			}
			targetResponses[key] = value
		}
	}
}
