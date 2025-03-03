package server

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"

	"github.com/libramusic/libracore/api/routes"
	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/utils"
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
	SwaggerInfo.Version = utils.LibraVersion.String()
	v1SpecJSON := SwaggerInfo.ReadDoc()
	var v1Spec echo.Map
	if err := json.Unmarshal([]byte(v1SpecJSON), &v1Spec); err != nil {
		log.Fatal("Error unmarshalling OpenAPI spec", "err", err)
	}

	// Update server URL
	servers := v1Spec["servers"].([]interface{})
	server := servers[0].(map[string]interface{})
	server["url"] = strings.ReplaceAll(
		server["url"].(string),
		"http://localhost:8080",
		config.Conf.Application.PublicURL,
	)

	return v1Spec
}

// processSchemas cleans up schema references and creates a proper Playable interface schema.
func processSchemas(v1Spec echo.Map) {
	schemas := v1Spec["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	// Remove "types." prefix from schema names
	for key := range schemas {
		if strings.HasPrefix(key, "types.") {
			newKey := strings.TrimPrefix(key, "types.")
			schemas[newKey] = schemas[key]
			delete(schemas, key)
		}
	}

	// Replace the fakePlayable schema with the actual Playable interface schema
	playableSchema := map[string]interface{}{
		"oneOf": []interface{}{
			map[string]interface{}{"$ref": "#/components/schemas/Track"},
			map[string]interface{}{"$ref": "#/components/schemas/Album"},
			map[string]interface{}{"$ref": "#/components/schemas/Video"},
			map[string]interface{}{"$ref": "#/components/schemas/Playlist"},
			map[string]interface{}{"$ref": "#/components/schemas/Artist"},
			map[string]interface{}{"$ref": "#/components/schemas/User"},
		},
	}
	schemas["Playable"] = playableSchema
	delete(schemas, "routes.fakePlayable")
}

// processPaths removes placeholder paths and standardizes path definitions.
func processPaths(v1Spec echo.Map) {
	paths := v1Spec["paths"].(map[string]interface{})
	delete(paths, "/fake")

	for _, pathObj := range paths {
		objMap, ok := pathObj.(map[string]interface{})
		if !ok {
			continue
		}
		processPathOperations(objMap)
	}
}

// processPathOperations updates response schema references in API operations.
func processPathOperations(pathObj map[string]interface{}) {
	for _, op := range pathObj {
		opMap, ok := op.(map[string]interface{})
		if !ok {
			continue
		}
		responses, ok := opMap["responses"].(map[string]interface{})
		if !ok {
			continue
		}

		for _, response := range responses {
			processResponseObject(response)
		}
	}
}

// processResponseObject updates schema references in JSON response content.
func processResponseObject(response interface{}) {
	resp, ok := response.(map[string]interface{})
	if !ok {
		return
	}

	content, ok := resp["content"].(map[string]interface{})
	if !ok {
		return
	}

	jsonContent, ok := content["application/json"].(map[string]interface{})
	if !ok {
		return
	}

	schema, ok := jsonContent["schema"].(map[string]interface{})
	if !ok {
		return
	}

	processSchema(schema)
}

// processSchema normalizes references in schema objects, oneOf lists, and array items.
func processSchema(schema map[string]interface{}) {
	if ref, ok := schema["$ref"].(string); ok {
		schema["$ref"] = updateReference(ref)
	} else if oneOf, ok := schema["oneOf"].([]interface{}); ok {
		for i, item := range oneOf {
			itemRef, ok := item.(string)
			if !ok {
				continue
			}
			oneOf[i] = updateReference(itemRef)
		}
	} else if items, ok := schema["items"].(map[string]interface{}); ok {
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
	case strings.HasPrefix(ref, "#/components/schemas/types."):
		newRef := strings.TrimPrefix(ref, "#/components/schemas/types.")
		return "#/components/schemas/" + newRef
	default:
		return ref
	}
}

// addFeedRoutesDocumentation adds feed endpoints (RSS, Atom, JSON) to the API spec.
func addFeedRoutesDocumentation(v1Spec echo.Map) {
	paths := v1Spec["paths"].(map[string]interface{})

	for _, feedRoute := range routes.FeedRoutesDoc {
		content := createFeedRouteContent(feedRoute.Type)
		description := createFeedRouteDescription(feedRoute.Type)

		path := map[string]interface{}{
			"get": map[string]interface{}{
				"summary": feedRoute.Summary,
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
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
func createFeedRouteContent(feedType string) map[string]interface{} {
	var contentType, description string
	var schema map[string]interface{}

	if feedType == "JSON" {
		contentType = "application/json"
		description = "JSON feed response"
		schema = map[string]interface{}{
			"type": "object",
		}
	} else {
		contentType = "application/xml"
		description = feedType + " XML feed response"
		schema = map[string]interface{}{
			"type": "string",
		}
	}

	return map[string]interface{}{
		contentType: map[string]interface{}{
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
func copyBasePathInfo(paths map[string]interface{}, basePath string, targetPath map[string]interface{}) {
	basePathObj, ok := paths[basePath].(map[string]interface{})
	if !ok {
		return
	}

	baseRouteGet, ok := basePathObj["get"].(map[string]interface{})
	if !ok {
		return
	}

	// Copy parameters
	if parameters, ok := baseRouteGet["parameters"].([]interface{}); ok {
		targetPath["get"].(map[string]interface{})["parameters"] = parameters
	}

	// Copy non-200 responses
	if responses, ok := baseRouteGet["responses"].(map[string]interface{}); ok {
		targetResponses := targetPath["get"].(map[string]interface{})["responses"].(map[string]interface{})
		for key, value := range responses {
			if key == "200" {
				continue
			}
			targetResponses[key] = value
		}
	}
}
