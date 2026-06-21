package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

// swaggerDarkUI is a self-rendered Swagger UI page wired up with a dark theme.
// The stock swaggo/gin-swagger handler embeds the light-only UI and exposes no
// hook for custom CSS, so we serve our own page instead and point it at the
// generated spec at /swagger/doc.json.
const swaggerDarkUI = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Fracture API — Swagger UI</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui.css">
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/Amoenus/SwaggerDark@master/SwaggerDark.css">
  <style>
    body { margin: 0; background: #1b1b1b; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "doc.json",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        layout: "StandaloneLayout",
      });
    };
  </script>
</body>
</html>`

// registerSwagger mounts the dark-themed Swagger UI and the OpenAPI spec.
func registerSwagger(r *gin.Engine) {
	serveUI := func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerDarkUI))
	}

	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/")
	})
	r.GET("/swagger/", serveUI)
	r.GET("/swagger/index.html", serveUI)

	r.GET("/swagger/doc.json", func(c *gin.Context) {
		doc, err := swag.ReadDoc()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(doc))
	})
}
