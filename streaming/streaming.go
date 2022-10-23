package streaming

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci/grpc_abci_v1"
)

const pluginEnvKeyPrefix = "COSMOS_SDK"

// HandshakeMap contains a map of each supported streaming's handshake config
var HandshakeMap = map[string]plugin.HandshakeConfig{
	"grpc_abci_v1": grpc_abci_v1.Handshake,
}

// PluginMap contains a map of supported gRPC plugins
var PluginMap = map[string]plugin.Plugin{
	"grpc_abci_v1": &grpc_abci_v1.ABCIListenerGRPCPlugin{},
}

func GetPluginEnvKey(name string) string {
	return fmt.Sprintf("%s_%s", pluginEnvKeyPrefix, strings.ToUpper(name))
}

func NewStreamingPlugin(name string, logLevel string) (interface{}, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Output: hclog.DefaultOutput,
		Level:  toHclogLevel(logLevel),
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

func toHclogLevel(s string) hclog.Level {
	switch s {
	case "trace":
		return hclog.Trace
	case "debug":
		return hclog.Debug
	case "info":
		return hclog.Info
	case "warn":
		return hclog.Warn
	case "error":
		return hclog.Error
	default:
		return hclog.DefaultLevel
	}
}
