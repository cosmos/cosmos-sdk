# Rosetta

Package rosetta implements the rosetta API for the current cosmos sdk release series.

## Extension

There are two ways in which you can customize and extend the implementation with your custom settings.

### Message extension

In order to make an `sdk.Msg` understandable by rosetta the only thing which is required is adding the methods to your message that satisfy the `rosetta.Msg` interface.
Examples on how to do so can be found in the staking types such as `MsgDelegate`, or in bank types such as `MsgSend`.

### Client interface override

In case more customization is required, it's possible to embed the Client type and override the methods which require customizations.

Example:

```go
package custom_client
import (

"context"
"github.com/coinbase/rosetta-sdk-go/types"
"github.com/cosmos/cosmos-sdk/server/rosetta/lib"
)

// CustomClient embeds the standard cosmos client
// which means that it implements the cosmos-rosetta-gateway Client
// interface while at the same time allowing to customize certain methods
type CustomClient struct {
    *rosetta.Client
}

func (c *CustomClient) ConstructionPayload(_ context.Context, request *types.ConstructionPayloadsRequest) (resp *types.ConstructionPayloadsResponse, err error) {
    // provide custom signature bytes
    panic("implement me")
}
```

### Error extension

Since rosetta requires to provide 'returned' errors to network options. In order to declare a new rosetta error, we use the `errors` package in cosmos-rosetta-gateway.

Example:

```go
package custom_errors
import crgerrs "github.com/cosmos/cosmos-sdk/server/rosetta/lib/errors"

var customErrRetriable = true
var CustomError = crgerrs.RegisterError(100, "custom message", customErrRetriable, "description")
```

Note: errors must be registered before cosmos-rosetta-gateway's `Server`.`Start` method is called. Otherwise the registration will be ignored. Errors with same code will be ignored too.

## Integration in app.go

To integrate rosetta as a command in your application, in app.go, in your root command simply use the `server.RosettaCommand` method.

Example:

```go
package app
import (

"github.com/cosmos/cosmos-sdk/server"
"github.com/spf13/cobra"
)

func buildAppCommand(rootCmd *cobra.Command) {
    // more app.go init stuff
	// ...
    // add rosetta command
	rootCmd.AddCommand(server.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Marshaler))
}
```

A full implementation example can be found in `simapp` package.

NOTE: when using a customized client, the command cannot be used as the constructors required **may** differ, so it's required to create a new one. We intend to provide a way to init a customized client without writing extra code in the future.
