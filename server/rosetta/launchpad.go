package rosetta

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"
	"github.com/tendermint/cosmos-rosetta-gateway/service"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"
)

type Options struct {
	// AppEndpoint is the endpoint that exposes the cosmos rpc in a cosmos app.
	AppEndpoint string

	// TendermintEndpoint is the endpoint that exposes the tendermint rpc in a cosmos app.
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

	cdc        *codec.Codec
	properties properties
}

// NewNetwork returns a configured network to work in a Launchpad version.
func NewNetwork(cdc *codec.Codec, options Options) service.Network {
	cosmosClient := cosmos.NewClient(fmt.Sprintf("http://%s", options.AppEndpoint), cdc)
	tendermintClient := tendermint.NewClient(fmt.Sprintf("http://%s", options.TendermintEndpoint))

	return service.Network{
		Properties: rosetta.NetworkProperties{
			Blockchain:          options.Blockchain,
			Network:             options.Network,
			SupportedOperations: []string{OperationTransfer},
		},
		Adapter: newAdapter(
			cdc,
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

func newAdapter(cdc *codec.Codec, cosmos SdkClient, tendermint TendermintClient, options properties) rosetta.Adapter {
	return &launchpad{
		cosmos:     cosmos,
		tendermint: tendermint,
		properties: options,
		cdc:        cdc,
	}
}
