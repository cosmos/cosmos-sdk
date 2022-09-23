package streaming

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hashicorp/go-plugin"
)

const pluginEnvKeyPrefix = "COSMOS_SDK_STREAMING"

// HandshakeMap contains a map of each supported streaming's handshake config
var HandshakeMap = map[string]plugin.HandshakeConfig{
	"abci": abci.Handshake,
}

// PluginMap contains a map of supported gRPC plugins
var PluginMap = map[string]plugin.Plugin{
	"abci": &abci.ListenerGRPCPlugin{},
}

func GetPluginEnvKey(name string) string {
	return fmt.Sprintf("%s_%s", pluginEnvKeyPrefix, strings.ToUpper(name))
}

func NewStreamingPlugin(ctx sdk.Context, name string) (interface{}, error) {
	logger := ctx.Logger().With("streaming", name)

	// We're a host. Start by launching the streaming process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeMap[name],
		Plugins:         PluginMap,
		Cmd:             exec.Command("sh", "-c", os.Getenv(GetPluginEnvKey(name))),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	// Request the streaming
	raw, err := rpcClient.Dispense(name)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return raw, nil
}
