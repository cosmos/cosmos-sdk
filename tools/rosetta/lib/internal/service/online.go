package service

import (
	"context"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"

	"cosmossdk.io/log"
	crgerrs "cosmossdk.io/tools/rosetta/lib/errors"
	crgtypes "cosmossdk.io/tools/rosetta/lib/types"
)

// genesisBlockFetchTimeout defines a timeout to fetch the genesis block
const (
	genesisBlockFetchTimeout = 15 * time.Second
)

// NewOnlineNetwork builds a single network adapter.
// It will get the Genesis block on the beginning to avoid calling it everytime.
func NewOnlineNetwork(network *types.NetworkIdentifier, client crgtypes.Client, logger log.Logger) (crgtypes.API, error) {
	ctx, cancel := context.WithTimeout(context.Background(), genesisBlockFetchTimeout)
	defer cancel()

	var genesisHeight int64 = 1 // to get genesis block height
	genesisBlock, err := client.BlockByHeight(ctx, &genesisHeight)
	if err != nil {
		logger.Error("failed to get genesis block height", "err", err)
	}

	return OnlineNetwork{
		client:         client,
		network:        network,
		networkOptions: networkOptionsFromClient(client, genesisBlock.Block),
	}, nil
}

// OnlineNetwork groups together all the components required for the full rosetta implementation
type OnlineNetwork struct {
	client crgtypes.Client // used to query Cosmos app + CometBFT

	network        *types.NetworkIdentifier      // identifies the network, it's static
	networkOptions *types.NetworkOptionsResponse // identifies the network options, it's static
}

// networkOptionsFromClient builds network options given the client
func networkOptionsFromClient(client crgtypes.Client, genesisBlock *types.BlockIdentifier) *types.NetworkOptionsResponse {
	var tsi *int64
	if genesisBlock != nil {
		tsi = &(genesisBlock.Index)
	}
	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion: crgtypes.SpecVersion,
			NodeVersion:    client.Version(),
		},
		Allow: &types.Allow{
			OperationStatuses:       client.OperationStatuses(),
			OperationTypes:          client.SupportedOperations(),
			Errors:                  crgerrs.SealAndListErrors(),
			HistoricalBalanceLookup: true,
			TimestampStartIndex:     tsi,
		},
	}
}
