package auth

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	authv1beta1 "cosmossdk.io/api/cosmos/auth/v1beta1"
)

// TestRekeyCommandsDoNotRequireUnreleasedAPIDescriptors keeps account
// rekeying usable by downstream applications before the next
// cosmossdk.io/api release. AutoCLI validates every configured RPC against
// that module, so new RPCs must remain custom commands until it is released.
func TestRekeyCommandsDoNotRequireUnreleasedAPIDescriptors(t *testing.T) {
	options := (AppModule{}).AutoCLIOptions()
	require.True(t, options.Query.EnhanceCustomCommand)
	require.True(t, options.Tx.EnhanceCustomCommand)

	queryMethods := make(map[string]bool)
	for _, method := range authv1beta1.Query_ServiceDesc.Methods {
		queryMethods[method.MethodName] = true
	}
	for _, rpc := range options.Query.RpcCommandOptions {
		require.True(t, queryMethods[rpc.RpcMethod], "query option names an unavailable API method: %s", rpc.RpcMethod)
	}

	msgMethods := make(map[string]bool)
	for _, method := range authv1beta1.Msg_ServiceDesc.Methods {
		msgMethods[method.MethodName] = true
	}
	for _, rpc := range options.Tx.RpcCommandOptions {
		require.True(t, msgMethods[rpc.RpcMethod], "tx option names an unavailable API method: %s", rpc.RpcMethod)
	}

	queryCmd := (AppModuleBasic{}).GetQueryCmd()
	require.NotNil(t, queryCmd)
	require.NotNil(t, findCommand(queryCmd, "rekeyed-accounts"))
	require.NotNil(t, findCommand(queryCmd, "pubkey-history"))

	txCmd := (AppModuleBasic{}).GetTxCmd()
	require.NotNil(t, findCommand(txCmd, "change-pubkey"))
}

func findCommand(root interface{ Commands() []*cobra.Command }, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
