// Code generated by swaggo/swag. DO NOT EDIT.

package server

import "github.com/swaggo/swag/v2"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "components": {"schemas":{"routes.fakePlayable":{"type":"object"},"types.Album":{"properties":{"addition_date":{"example":1634296980,"type":"integer"},"additional_meta":{"additionalProperties":{},"type":"object"},"artist_ids":{"example":["h3r3VpPvSq8","R2QTLKbHamW"],"items":{"type":"string"},"type":"array","uniqueItems":false},"description":{"example":"Lorem ipsum dolor sit amet.","type":"string"},"favorite_count":{"example":5,"type":"integer"},"id":{"example":"BhRpYVlrMo8","type":"string"},"linked_item_ids":{"items":{"type":"string"},"type":"array","uniqueItems":false},"listen_count":{"example":150,"type":"integer"},"metadata_source":{"type":"string"},"permissions":{"additionalProperties":{"type":"string"},"type":"object"},"release_date":{"example":"2023-10-01","type":"string"},"tags":{"items":{"type":"string"},"type":"array","uniqueItems":false},"title":{"example":"Lorem Ipsum","type":"string"},"track_ids":{"example":["7nTwkcl51u4","OBTwkAXODLd"],"items":{"type":"string"},"type":"array","uniqueItems":false},"upc":{"example":"012345678905","type":"string"},"user_id":{"example":"TPkrKcIZRRq","type":"string"}},"type":"object"},"types.Artist":{"properties":{"addition_date":{"example":1634296980,"type":"integer"},"additional_meta":{"additionalProperties":{},"type":"object"},"album_ids":{"example":["BhRpYVlrMo8","poFEUbgBuwJ"],"items":{"type":"string"},"type":"array","uniqueItems":false},"creation_date":{"example":"2023-10-01","type":"string"},"description":{"example":"Artist description here.","type":"string"},"favorite_count":{"example":5,"type":"integer"},"id":{"example":"h3r3VpPvSq8","type":"string"},"linked_item_ids":{"items":{"type":"string"},"type":"array","uniqueItems":false},"listen_count":{"example":150,"type":"integer"},"metadata_source":{"type":"string"},"name":{"example":"John Doe","type":"string"},"permissions":{"additionalProperties":{"type":"string"},"type":"object"},"tags":{"items":{"type":"string"},"type":"array","uniqueItems":false},"track_ids":{"example":["7nTwkcl51u4","OBTwkAXODLd"],"items":{"type":"string"},"type":"array","uniqueItems":false},"user_id":{"example":"TPkrKcIZRRq","type":"string"}},"type":"object"},"types.Playlist":{"properties":{"addition_date":{"example":1634296980,"type":"integer"},"additional_meta":{"additionalProperties":{},"type":"object"},"creation_date":{"example":"2023-10-01","type":"string"},"description":{"example":"Lorem ipsum dolor sit amet.","type":"string"},"favorite_count":{"example":25,"type":"integer"},"id":{"example":"JpXHsNCAATt","type":"string"},"listen_count":{"example":150,"type":"integer"},"metadata_source":{"type":"string"},"permissions":{"additionalProperties":{"type":"string"},"type":"object"},"tags":{"items":{"type":"string"},"type":"array","uniqueItems":false},"title":{"example":"Lorem Ipsum Playlist","type":"string"},"track_ids":{"example":["7nTwkcl51u4","OBTwkAXODLd"],"items":{"type":"string"},"type":"array","uniqueItems":false},"user_id":{"example":"TPkrKcIZRRq","type":"string"}},"type":"object"},"types.Track":{"properties":{"addition_date":{"example":1634296980,"type":"integer"},"additional_meta":{"additionalProperties":{},"type":"object"},"album_ids":{"example":["BhRpYVlrMo8","poFEUbgBuwJ"],"items":{"type":"string"},"type":"array","uniqueItems":false},"artist_ids":{"example":["h3r3VpPvSq8","R2QTLKbHamW"],"items":{"type":"string"},"type":"array","uniqueItems":false},"content_source":{"type":"string"},"description":{"example":"Lorem ipsum dolor sit amet.","type":"string"},"duration":{"example":300,"type":"integer"},"favorite_count":{"example":5,"type":"integer"},"id":{"example":"7nTwkcl51u4","type":"string"},"isrc":{"example":"USSKG1912345","type":"string"},"linked_item_ids":{"items":{"type":"string"},"type":"array","uniqueItems":false},"listen_count":{"example":150,"type":"integer"},"lyric_sources":{"additionalProperties":{"type":"string"},"type":"object"},"lyrics":{"additionalProperties":{"type":"string"},"type":"object"},"metadata_source":{"type":"string"},"permissions":{"additionalProperties":{"type":"string"},"type":"object"},"primary_album_id":{"example":"BhRpYVlrMo8","type":"string"},"release_date":{"example":"2023-10-01","type":"string"},"tags":{"items":{"type":"string"},"type":"array","uniqueItems":false},"title":{"example":"Lorem","type":"string"},"track_number":{"example":1,"type":"integer"},"user_id":{"example":"TPkrKcIZRRq","type":"string"}},"type":"object"},"types.User":{"properties":{"creation_date":{"type":"integer"},"description":{"example":"I am a person.","type":"string"},"display_name":{"example":"John Doe","type":"string"},"email":{"example":"john.doe@example.com","type":"string"},"favorites":{"items":{"type":"string"},"type":"array","uniqueItems":false},"id":{"example":"TPkrKcIZRRq","type":"string"},"linked_artist_id":{"example":"h3r3VpPvSq8","type":"string"},"linked_sources":{"additionalProperties":{"type":"string"},"type":"object"},"listened_to":{"additionalProperties":{"type":"integer"},"type":"object"},"password_hash":{"type":"string"},"permissions":{"additionalProperties":{"type":"string"},"type":"object"},"public_view_count":{"example":519,"type":"integer"},"username":{"example":"JohnDoe","type":"string"}},"type":"object"},"types.Video":{"properties":{"addition_date":{"type":"integer"},"additional_meta":{"additionalProperties":{},"type":"object"},"artist_ids":{"example":["h3r3VpPvSq8","R2QTLKbHamW"],"items":{"type":"string"},"type":"array","uniqueItems":false},"content_source":{"type":"string"},"description":{"example":"Lorem ipsum dolor sit amet.","type":"string"},"duration":{"example":300,"type":"integer"},"favorite_count":{"example":10,"type":"integer"},"id":{"example":"hCNchWdmbro","type":"string"},"linked_item_ids":{"items":{"type":"string"},"type":"array","uniqueItems":false},"lyric_sources":{"additionalProperties":{"type":"string"},"type":"object"},"metadata_source":{"type":"string"},"permissions":{"additionalProperties":{"type":"string"},"type":"object"},"release_date":{"example":"2023-10-01","type":"string"},"subtitles":{"additionalProperties":{"type":"string"},"type":"object"},"tags":{"items":{"type":"string"},"type":"array","uniqueItems":false},"title":{"example":"Dolor Sit Amet","type":"string"},"user_id":{"example":"TPkrKcIZRRq","type":"string"},"watch_count":{"example":185,"type":"integer"}},"type":"object"}}},
    "info": {"contact":{"name":"Libra Team","url":"https://github.com/LibraMusic/LibraCore"},"description":"{{escape .Description}}","license":{"name":"MIT","url":"https://opensource.org/licenses/MIT"},"title":"{{.Title}}","version":"{{.Version}}"},
    "externalDocs": {"description":"","url":""},
    "paths": {"/fake":{"get":{"responses":{"200":{"content":{"application/json":{"schema":{"oneOf":[{"$ref":"#/components/schemas/types.Track"},{"$ref":"#/components/schemas/types.Album"},{"$ref":"#/components/schemas/types.Video"},{"$ref":"#/components/schemas/types.Playlist"},{"$ref":"#/components/schemas/types.Artist"},{"$ref":"#/components/schemas/types.User"}]}}},"description":"OK"}}}},"/playables":{"get":{"operationId":"getAllPlayables","responses":{"200":{"content":{"application/json":{"schema":{"items":{"$ref":"#/components/schemas/routes.fakePlayable"},"type":"array"}}},"description":"Returns a list of all playables"},"500":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"Internal Server Error"}},"summary":"Get all playables"}},"/playables/{id}":{"get":{"operationId":"getUserPlayables","parameters":[{"description":"User ID","in":"path","name":"id","required":true,"schema":{"type":"string"}}],"responses":{"200":{"content":{"application/json":{"schema":{"items":{"$ref":"#/components/schemas/routes.fakePlayable"},"type":"array"}}},"description":"Returns a list of user's playables"},"500":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"Internal Server Error"}},"summary":"Get user's playables"}},"/search":{"get":{"operationId":"searchPlayables","parameters":[{"description":"Search query","in":"query","name":"q","required":true,"schema":{"type":"string"}}],"responses":{"200":{"content":{"application/json":{"schema":{"items":{"$ref":"#/components/schemas/routes.fakePlayable"},"type":"array"}}},"description":"Returns a list of playables matching the search query"},"500":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"Internal Server Error"}},"summary":"Search for playables by query"}}},
    "openapi": "3.1.0",
    "servers": [
        {"url":"http://localhost:8080/api/v1"}
    ]
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1.0-DEV",
	Title:            "Libra API",
	Description:      "Libra Core API providing music streaming and management capabilities.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
