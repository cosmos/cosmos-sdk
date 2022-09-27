package streaming

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci"
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

func NewStreamingPlugin(name string) (interface{}, error) {
	// todo need to figure out a way to hook into the sdk logger
	logger := hclog.New(&hclog.LoggerOptions{
		Output: hclog.DefaultOutput,
		Level:  hclog.Trace,
		Name:   fmt.Sprintf("plugin.%s", name),
	})

	// We're a host. Start by launching the streaming process.
	env := os.Getenv(GetPluginEnvKey(name))
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeMap[name],
		Plugins:         PluginMap,
		Cmd:             exec.Command("sh", "-c", env),
		Logger:          logger,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request streaming plugin
	return rpcClient.Dispense(name)
}
