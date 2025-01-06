/*
Package swagger provides Swagger UI support for server/v2.

Example usage:

	import (
		"cosmossdk.io/client/docs"
		"cosmossdk.io/core/server"
		"cosmossdk.io/log"
		swaggerv2 "cosmossdk.io/server/v2/api/swagger"
	)

	// Create a logger
	logger := log.NewLogger()

	// Configure Swagger server
	swaggerCfg := server.ConfigMap{
		"swagger": map[string]any{
			"enable":  true,
			"address": "localhost:8080",
			"path":    "/swagger/",
		},
	}

	// Create new Swagger server with the default SDK Swagger UI
	swaggerServer, err := swaggerv2.New[YourTxType](
		logger.With(log.ModuleKey, "swagger"),
		swaggerCfg,
		swaggerv2.CfgOption(func(cfg *swaggerv2.Config) {
			cfg.SwaggerUI = docs.SwaggerUI // Use the default SDK Swagger UI
		}),
	)
	if err != nil {
		// Handle error
	}

	// Add Swagger server to your application
	app.AddServer(swaggerServer)

The server will serve Swagger UI documentation at the configured path (default: /swagger/).
Users can customize the configuration through the following options:
  - enable: Enable/disable the Swagger server
  - address: The address to listen on (default: localhost:8080)
  - path: The path to serve Swagger UI at (default: /swagger/)
  - SwaggerUI: The http.FileSystem containing Swagger UI files
*/
package swagger 
