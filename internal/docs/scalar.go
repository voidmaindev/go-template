// Package docs provides API documentation handlers.
package docs

import "github.com/gofiber/fiber/v2"

// ScalarHandler returns a Fiber handler that serves the Scalar API documentation UI.
// specURL should be the path to the OpenAPI specification endpoint.
func ScalarHandler(specURL string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Go Template API - Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
</head>
<body>
    <script id="api-reference" data-url="` + specURL + `"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}
