package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
)

// assert interface implementation
var _ rosetta.Rosetta = OfflineNetwork{}

// NewOffline instantiates the instance of an offline network
// whilst the offline network does not support the DataAPI,
// it supports a subset of the construction API.
func NewOffline(network *types.NetworkIdentifier) OfflineNetwork {
	return OfflineNetwork{
		SingleNetwork{
			client:                 nil,
			network:                network,
			genesisBlockIdentifier: nil,
		},
	}
}

// OfflineNetwork implements an offline data API
// which is basically a data API that constantly
// returns errors, because it cannot be used if offline
type OfflineNetwork struct {
	SingleNetwork
}

// Implement DataAPI in offline mode, which means no method is available
func (o OfflineNetwork) AccountBalance(_ context.Context, _ *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	return nil, rosetta.ErrOffline.RosettaError()
}

func (o OfflineNetwork) Block(_ context.Context, _ *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	return nil, rosetta.ErrOffline.RosettaError()
}

func (o OfflineNetwork) BlockTransaction(_ context.Context, _ *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	return nil, rosetta.ErrOffline.RosettaError()
}

func (o OfflineNetwork) Mempool(_ context.Context, _ *types.NetworkRequest) (*types.MempoolResponse, *types.Error) {
	return nil, rosetta.ErrOffline.RosettaError()
}

func (o OfflineNetwork) MempoolTransaction(_ context.Context, _ *types.MempoolTransactionRequest) (*types.MempoolTransactionResponse, *types.Error) {
	return nil, rosetta.ErrOffline.RosettaError()
}

func (o OfflineNetwork) NetworkList(_ context.Context, _ *types.MetadataRequest) (*types.NetworkListResponse, *types.Error) {
	return nil, rosetta.ErrOffline.RosettaError()
}

func (o OfflineNetwork) NetworkOptions(_ context.Context, _ *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	return nil, rosetta.ErrOffline.RosettaError()
}

func (o OfflineNetwork) NetworkStatus(_ context.Context, _ *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	return nil, rosetta.ErrOffline.RosettaError()
}

// ConstructionSubmit is the only endpoint which is not supported in offline mode, hence we are overriding it here and disabling it.
func (o OfflineNetwork) ConstructionSubmit(_ context.Context, _ *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return nil, rosetta.ErrOffline.RosettaError()
}
