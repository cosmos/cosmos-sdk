/*
The rest package contains the "RestServer" type which serves as a means to create
and start a RESTful service for applications. In addition, applications can create
and mount a command which starts a RESTful service through the ServeCommand function.
The ServeCommand function accepts a "func(*RestServer)" parameter which allows
applications to register their own routes and handlers. Typically, this occur
through the use of the `ModuleManager` and will look like as follows:

  import "github.com/cosmos/cosmos-sdk/client/rest"

  func registerRoutes(rs *rest.RestServer) {
    app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
    ...
  }

  rootCmd.AddCommand(
    rest.ServeCommand(cdc, registerRoutes),
    ...
  )

The "ServeCommand" will additionally automatically mount RESTful documentation
via Swagger under the "/swagger/" endpoint. Documentation is achieved through
annotating each handler via the Swag library. See https://github.com/swaggo/swag
for further documentation and annotation APIs. Application developers may wish
to add main API details to the Swagger documentation such as the title and base
URL to perform HTTP request directly from the documentation.

  var (
    SwaggerHost        string
    SwaggerVersion     string
    SwaggerBasePath    string
    SwaggerTitle       string
    SwaggerDescription string
  )

These variables can be set in one of two ways:

1. Directly by setting them prior to starting the REST server:

	import "github.com/cosmos/cosmos-sdk/client/rest"

	rest.SwaggerHost = "your API host"
	rest.SwaggerTitle = "your API title"
	// ...

2. Setting them at build-time via ldflags:

  ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=gaia \
    -X github.com/cosmos/cosmos-sdk/version.ServerName=appd \
    -X github.com/cosmos/cosmos-sdk/version.ClientName=appcli \
    -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
    -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
    ...
    -X github.com/cosmos/cosmos-sdk/client/rest.SwaggerTitle=$(API_TITLE)

  BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
  go install -mod=readonly $(BUILD_FLAGS) ./cmd/appd

*/
package rest
