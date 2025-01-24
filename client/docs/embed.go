package docs

import (
	"embed"
	"io/fs"
)

//go:embed swagger-ui
var SwaggerUI embed.FS

// GetSwaggerFS returns the embedded Swagger UI filesystem
func GetSwaggerFS() fs.FS {
	root, err := fs.Sub(SwaggerUI, "swagger-ui")
	if err != nil {
		panic(err)
	}

	return root
}
