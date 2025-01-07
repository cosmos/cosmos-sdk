/*
Package swagger provides Swagger UI support for server/v2.

Example usage in commands.go:

	// Create Swagger server
	swaggerServer, err := swaggerv2.New[T](
		logger.With(log.ModuleKey, "swagger"),
		deps.GlobalConfig,
	)

Configuration options:
  - enable: Enable/disable the Swagger UI server (default: true)
  - address: Server address (default: localhost:8090)
*/
package swagger 
