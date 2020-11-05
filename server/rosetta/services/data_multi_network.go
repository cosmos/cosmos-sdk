package services

import (
	"context"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
)

// assert interface implementation
var _ rosetta.DataAPI = MultiNetwork{}

// idFromNetworkIdentifier builds an unique ID given a network identifier
func idFromNetworkIdentifier(id *types.NetworkIdentifier) string {
	return fmt.Sprintf("%s:%s", id.Network, id.Blockchain)
}

func NewMultiNetworkConfig(client rosetta.DataAPIClient, network *types.NetworkIdentifier) MultiNetworkConfig {
	return MultiNetworkConfig{
		Client:  client,
		Network: network,
	}
}

// MultiNetworkConfig
type MultiNetworkConfig struct {
	Client  rosetta.DataAPIClient
	Network *types.NetworkIdentifier
}

// NewMultiNetwork builds a new MultiNetwork rosetta API
func NewMultiNetwork(configs []MultiNetworkConfig) (MultiNetwork, error) {
	clients := make(map[string]SingleNetwork, len(configs))
	networks := make([]*types.NetworkIdentifier, len(configs))
	for i, conf := range configs {
		netID := idFromNetworkIdentifier(conf.Network)
		// check if it was already specified
		_, ok := clients[netID]
		if !ok {
			return MultiNetwork{}, fmt.Errorf("network already specified: %s", netID)
		}
		sn, err := NewSingleNetwork(conf.Client, conf.Network)
		if err != nil {
			return MultiNetwork{}, err
		}
		clients[netID] = sn
		networks[i] = conf.Network
	}
	return MultiNetwork{
		clients: clients,
		list:    networks,
	}, nil
}

// MultiNetwork implements a multi cosmos network rosetta API
// capable of querying multiple cosmos networks as long as they
// use the same cosmos-sdk and tendermint version used
// in the go.mod file (or a version which is forward/backwards compatible)
type MultiNetwork struct {
	clients map[string]SingleNetwork
	list    []*types.NetworkIdentifier
}

func (m MultiNetwork) getClient(req *types.NetworkIdentifier) (SingleNetwork, error) {
	c, ok := m.clients[idFromNetworkIdentifier(req)]
	if !ok {
		return SingleNetwork{}, rosetta.WrapError(rosetta.ErrNetworkNotSupported, fmt.Sprintf("'%s' '%s'", req.Blockchain, req.Network))
	}
	return c, nil
}

func (m MultiNetwork) AccountBalance(ctx context.Context, request *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	c, err := m.getClient(request.NetworkIdentifier)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return c.AccountBalance(ctx, request)
}

func (m MultiNetwork) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	c, err := m.getClient(request.NetworkIdentifier)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return c.Block(ctx, request)
}

func (m MultiNetwork) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	c, err := m.getClient(request.NetworkIdentifier)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return c.BlockTransaction(ctx, request)
}

func (m MultiNetwork) Mempool(ctx context.Context, request *types.NetworkRequest) (*types.MempoolResponse, *types.Error) {
	c, err := m.getClient(request.NetworkIdentifier)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return c.Mempool(ctx, request)
}

func (m MultiNetwork) MempoolTransaction(ctx context.Context, request *types.MempoolTransactionRequest) (*types.MempoolTransactionResponse, *types.Error) {
	c, err := m.getClient(request.NetworkIdentifier)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return c.MempoolTransaction(ctx, request)
}

func (m MultiNetwork) NetworkList(_ context.Context, _ *types.MetadataRequest) (*types.NetworkListResponse, *types.Error) {
	return &types.NetworkListResponse{NetworkIdentifiers: m.list}, nil
}

func (m MultiNetwork) NetworkOptions(ctx context.Context, request *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	c, err := m.getClient(request.NetworkIdentifier)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return c.NetworkOptions(ctx, request)
}

func (m MultiNetwork) NetworkStatus(ctx context.Context, request *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	c, err := m.getClient(request.NetworkIdentifier)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return c.NetworkStatus(ctx, request)
}
