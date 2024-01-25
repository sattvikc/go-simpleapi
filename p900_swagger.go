package fastapi

import (
	"mime/multipart"
	"reflect"
	"strings"
)

const swaggerUI = `<!DOCTYPE html>
<html>
<head>
<link type="text/css" rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-themes@3.0.1/themes/3.x/theme-muted.min.css" />
<link rel="shortcut icon" href="https://fastapi.tiangolo.com/img/favicon.png">
<title>Swagger UI</title>
</head>
<body>
<div id="swagger-ui">
</div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui-bundle.js"></script>
<!-- SwaggerUIBundle is now available on the page -->
<script>
const ui = SwaggerUIBundle({
	url: '/openapi.json',
"dom_id": "#swagger-ui",
"layout": "BaseLayout",
"deepLinking": true,
"showExtensions": true,
"showCommonExtensions": true,
oauth2RedirectUrl: window.location.origin + '/docs/oauth2-redirect',
presets: [
	SwaggerUIBundle.presets.apis,
	SwaggerUIBundle.SwaggerUIStandalonePreset
	],
})
</script>
</body>
</html>
`

func addSwaggerRoutes(server *Server) {
	server.GET("/docs", func(ctx *Context) error {
		return ctx.HTML(200, swaggerUI)
	})

	server.GET("/openapi.json", func(ctx *Context) error {
		return ctx.JSON(200, server.swaggerJson)
	})
}

func updateDefinitionFromhandler(definition map[string]interface{}, paramTypes []reflect.Type) {
	HEADER_EXCLUSIONS := map[string]bool{"content-type": true, "content-length": true, "user-agent": true}

	for _, paramType := range paramTypes {
		for i := 0; i < paramType.NumField(); i++ {
			field := paramType.Field(i)
			if field.Tag.Get("body") != "" {
				if field.Tag.Get("body") == "multipart" {
					properties := map[string]interface{}{}

					for j := 0; j < field.Type.NumField(); j++ {
						field := field.Type.Field(j)
						if field.Tag.Get("form") != "" {
							properties[field.Tag.Get("form")] = getSwaggerSchemaFromType(field.Type)
						}
					}

					bodyDefinition := map[string]interface{}{
						"content": map[string]interface{}{
							"multipart/form-data": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":       "object",
									"properties": properties,
								},
							},
						},
					}

					definition["requestBody"] = bodyDefinition

				} else if field.Tag.Get("body") == "urlencoded" {
					properties := map[string]interface{}{}

					for j := 0; j < field.Type.NumField(); j++ {
						field := field.Type.Field(j)
						if field.Tag.Get("form") != "" {
							properties[field.Tag.Get("form")] = getSwaggerSchemaFromType(field.Type)
						}
					}

					bodyDefinition := map[string]interface{}{
						"content": map[string]interface{}{
							"application/x-www-form-urlencoded": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":       "object",
									"properties": properties,
								},
							},
						},
					}

					definition["requestBody"] = bodyDefinition
				} else if field.Tag.Get("body") == "json" {
					bodyDefinition := map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": getSwaggerSchemaFromType(field.Type),
							},
						},
					}
					definition["requestBody"] = bodyDefinition
				}

			} else if field.Type.Kind() == reflect.Struct {
				updateDefinitionFromhandler(definition, []reflect.Type{field.Type})

			} else if field.Tag.Get("query") != "" {
				queryDefinition := map[string]interface{}{
					"in":       "query",
					"name":     field.Tag.Get("query"),
					"required": field.Type.Kind() != reflect.Ptr,
					"schema":   getSwaggerSchemaFromType(field.Type),
				}
				definition["parameters"] = append(definition["parameters"].([]interface{}), queryDefinition)

			} else if field.Tag.Get("header") != "" {
				if _, ok := HEADER_EXCLUSIONS[strings.ToLower(field.Tag.Get("header"))]; ok {
					continue
				}
				headerDefinition := map[string]interface{}{
					"in":       "header",
					"name":     field.Tag.Get("header"),
					"required": field.Type.Kind() != reflect.Ptr,
					"schema":   getSwaggerSchemaFromType(field.Type),
				}
				definition["parameters"] = append(definition["parameters"].([]interface{}), headerDefinition)

			} else if field.Tag.Get("path") != "" {
				pathDefinition := map[string]interface{}{
					"in":       "path",
					"name":     field.Tag.Get("path"),
					"required": true, // TODO
					"schema": map[string]interface{}{
						"type": "string",
					},
				}
				definition["parameters"] = append(definition["parameters"].([]interface{}), pathDefinition)
			}
		}
	}
}

func getSwaggerSchemaFromType(t reflect.Type) interface{} {
	fileType := reflect.TypeOf((*multipart.File)(nil)).Elem()
	if t.ConvertibleTo(fileType) {
		return map[string]interface{}{
			"type": "file",
		}
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct {
		properties := map[string]interface{}{}

		required := []string{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Tag.Get("json") != "" {
				properties[field.Tag.Get("json")] = getSwaggerSchemaFromType(field.Type)

				if field.Type.Kind() != reflect.Ptr {
					required = append(required, field.Tag.Get("json"))
				}
			}
		}

		return map[string]interface{}{
			"type":       "object",
			"properties": properties,
			"required":   required,
		}
	}

	if t.Kind() == reflect.Slice {
		return map[string]interface{}{
			"type":  "array",
			"items": getSwaggerSchemaFromType(t.Elem()),
		}
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]interface{}{
			"type": "string",
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]interface{}{
			"type": "integer",
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{
			"type": "integer",
		}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{
			"type": "number",
		}
	case reflect.Bool:
		return map[string]interface{}{
			"type": "boolean",
		}
	default:
		return map[string]interface{}{
			"type": "string",
		}
	}
}
