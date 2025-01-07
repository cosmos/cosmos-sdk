/*
Package swagger provides Swagger UI support for server/v2.

Example usage in commands.go:

	// Create Swagger server
	swaggerServer, err := swaggerv2.New[T](
		logger.With(log.ModuleKey, "swagger"),
		deps.GlobalConfig,
		swaggerv2.CfgOption(func(cfg *swaggerv2.Config) {
			cfg.SwaggerUI = docs.SwaggerUI
		}),
	)

	// Add server to your application
	return serverv2.AddCommands[T](
		// ...other servers...,
		swaggerServer,
	)

Configuration options:
  - enable: Enable/disable the Swagger server (default: true)
  - address: Server address (default: localhost:8080)
  - path: UI endpoint path (default: /swagger/)
*/
package swagger 
