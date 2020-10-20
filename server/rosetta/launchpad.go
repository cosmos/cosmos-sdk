package rosetta

import (
	"fmt"

	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"
	"github.com/tendermint/cosmos-rosetta-gateway/service"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Options struct {
	// CosmosEndpoint is the endpoint that exposes the cosmos rpc in a cosmos app.
	CosmosEndpoint string

	// CosmosEndpoint is the endpoint that exposes the tendermint rpc in a cosmos app.
	TendermintEndpoint string

	// Blockchain represents the name of the blockchain, it is used for NetworkList endpoint.
	Blockchain string

	// Network represents the name of the network, it is used for NetworkList endpoint.
	Network string

	// AddrPrefix is the prefix used for bech32 addresses.
	AddrPrefix string

	// Offline mode forces to run without querying the node. Some endpoints won't work.
	OfflineMode bool
}

type properties struct {
	// Blockchain represents the name of the blockchain, it is used for NetworkList endpoint.
	Blockchain string

	// Network represents the name of the network, it is used for NetworkList endpoint.
	Network string

	// AddrPrefix is the prefix used for bech32 addresses.
	AddrPrefix string

	// Offline mode forces to run without querying the node. Some endpoints won't work.
	OfflineMode bool
}

type launchpad struct {
	tendermint TendermintClient
	cosmos     SdkClient

	properties properties
}

// NewLaunchpadNetwork returns a configured network to work in a Launchpad version.
func NewLaunchpadNetwork(options Options) service.Network {
	cosmosClient := cosmos.NewClient(fmt.Sprintf("http://%s", options.CosmosEndpoint))
	tendermintClient := tendermint.NewClient(fmt.Sprintf("http://%s", options.TendermintEndpoint))

	return service.Network{
		Properties: rosetta.NetworkProperties{
			Blockchain:          options.Blockchain,
			Network:             options.Network,
			SupportedOperations: []string{OperationTransfer},
		},
		Adapter: newAdapter(
			cosmosClient,
			tendermintClient,
			properties{
				Blockchain:  options.Blockchain,
				Network:     options.Network,
				AddrPrefix:  options.AddrPrefix,
				OfflineMode: options.OfflineMode,
			},
		),
	}
}

func newAdapter(cosmos SdkClient, tendermint TendermintClient, options properties) rosetta.Adapter {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(
		options.AddrPrefix,
		options.AddrPrefix+sdk.PrefixPublic)

	return &launchpad{
		cosmos:     cosmos,
		tendermint: tendermint,
		properties: options,
	}
}
