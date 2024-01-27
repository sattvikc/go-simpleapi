package simpleapi

import (
	"net/http"
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

func addSwaggerRoutes(app *App) {
	app.AddHandler("/docs", http.MethodGet, func(ctx *Context) error {
		return ctx.HTML(200, swaggerUI)
	})

	app.AddHandler("/openapi.json", http.MethodGet, func(ctx *Context) error {
		return ctx.JSON(200, app.swaggerJson)
	})
}
