package cmd

import (
	"os"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/simapp/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// ProvideClientContext is a depinject Provider function which assembles and returns a client.Context.
func ProvideClientContext(
	appCodec codec.Codec,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfigOpts tx.ConfigOptions,
	legacyAmino registry.AminoRegistrar,
	addressCodec address.Codec,
	validatorAddressCodec address.ValidatorAddressCodec,
	consensusAddressCodec address.ConsensusAddressCodec,
) client.Context {
	var err error
	amino, ok := legacyAmino.(*codec.LegacyAmino)
	if !ok {
		panic("registry.AminoRegistrar must be an *codec.LegacyAmino instance for legacy ClientContext")
	}

	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithLegacyAmino(amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithAddressCodec(addressCodec).
		WithValidatorAddressCodec(validatorAddressCodec).
		WithConsensusAddressCodec(consensusAddressCodec).
		WithHomeDir(simapp.DefaultNodeHome).
		WithViper("") // uses by default the binary name as prefix

	// Read the config to overwrite the default values with the values from the config file
	customClientTemplate, customClientConfig := initClientConfig()
	clientCtx, err = config.CreateClientConfig(clientCtx, customClientTemplate, customClientConfig)
	if err != nil {
		panic(err)
	}

	// textual is enabled by default, we need to re-create the tx config grpc instead of bank keeper.
	txConfigOpts.TextualCoinMetadataQueryFn = authtxconfig.NewGRPCCoinMetadataQueryFn(clientCtx)
	txConfig, err := tx.NewTxConfigWithOptions(clientCtx.Codec, txConfigOpts)
	if err != nil {
		panic(err)
	}
	clientCtx = clientCtx.WithTxConfig(txConfig)

	return clientCtx
}
