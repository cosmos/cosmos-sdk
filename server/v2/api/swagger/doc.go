/*
Package swagger provides Swagger UI server/v2 component.

Example usage in commands.go:

	swaggerServer, err := swaggerv2.New[T](
		logger.With(log.ModuleKey, "swagger"),
		deps.GlobalConfig,
		swaggerv2.CfgOption(func(cfg *swaggerv2.Config) {
			cfg.SwaggerUI = docs.SwaggerUI
		}),
	)
*/
package swagger
