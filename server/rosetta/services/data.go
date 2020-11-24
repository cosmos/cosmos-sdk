package services

import (
	"context"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"

	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"

	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/cosmos/conversion"
)

// genesisBlockFetchTimeout defines a timeout to fetch the genesis block
const genesisBlockFetchTimeout = 15 * time.Second

// NewSingleNetwork builds a single network client
// the client will attempt to fetch genesis block too
func NewSingleNetwork(client rosetta.NodeClient, network *types.NetworkIdentifier) (crg.Adapter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), genesisBlockFetchTimeout)
	defer cancel()

	var genesisHeight int64 = 1
	block, _, err := client.BlockByHeight(ctx, &genesisHeight)
	if err != nil {
		return OnlineNetwork{}, err
	}

	return OnlineNetwork{
		client:                 client,
		network:                network,
		networkOptions:         &types.NetworkOptionsResponse{Version: rosetta.Version(), Allow: rosetta.Allow()},
		genesisBlockIdentifier: conversion.TMBlockToRosettaBlockIdentifier(block),
	}, nil
}

// OnlineNetwork groups together all the components required for the full rosetta implementation
type OnlineNetwork struct {
	client rosetta.NodeClient // used to query cosmos app + tendermint

	network        *types.NetworkIdentifier      // identifies the network, it's static
	networkOptions *types.NetworkOptionsResponse // identifies the network options, it's static

	genesisBlockIdentifier *types.BlockIdentifier // identifies genesis block, it's static
}

// AccountBalance retrieves the account balance of an address
// rosetta requires us to fetch the block information too
func (on OnlineNetwork) AccountBalance(ctx context.Context, request *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	var (
		height *int64
		block  *tmtypes.ResultBlock
		err    error
	)

	switch {
	case request.BlockIdentifier == nil:
		height = nil
		block, _, err = on.client.BlockByHeight(ctx, nil)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	case request.BlockIdentifier.Hash != nil:
		block, _, err = on.client.BlockByHash(ctx, *request.BlockIdentifier.Hash)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
		height = &block.Block.Height
	case request.BlockIdentifier.Index != nil:
		height = request.BlockIdentifier.Index
		block, _, err = on.client.BlockByHeight(ctx, height)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	}

	accountCoins, err := on.client.Balances(ctx, request.AccountIdentifier.Address, height)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	availableCoins, err := on.client.Coins(ctx)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: conversion.TMBlockToRosettaBlockIdentifier(block),
		Balances:        conversion.SdkCoinsToRosettaAmounts(accountCoins, availableCoins),
		Coins:           nil,
		Metadata:        nil,
	}, nil
}

// Block gets the transactions in the given block
func (on OnlineNetwork) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	var (
		block *tmtypes.ResultBlock
		txs   []*rosetta.SdkTxWithHash
		err   error
	)
	// block identifier is assumed not to be nil as rosetta will do this check for us
	// check if we have to query via hash or block number
	switch {
	case request.BlockIdentifier.Hash != nil:
		block, txs, err = on.client.BlockByHash(ctx, *request.BlockIdentifier.Hash)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	case request.BlockIdentifier.Index != nil:
		block, txs, err = on.client.BlockByHeight(ctx, request.BlockIdentifier.Index)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	default:
		return nil, rosetta.WrapError(rosetta.ErrBadArgument, "at least one of hash or index needs to be specified").RosettaError()
	}

	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       conversion.TMBlockToRosettaBlockIdentifier(block),
			ParentBlockIdentifier: conversion.TMBlockToRosettaParentBlockIdentifier(block),
			Timestamp:             conversion.TimeToMilliseconds(block.Block.Time), // ts is required in milliseconds
			Transactions:          conversion.SdkTxsWithHashToRosettaTxs(txs),
			Metadata:              nil,
		},
		OtherTransactions: nil,
	}, nil
}

// BlockTransaction gets the given transaction in the specified block, we do not need to check the block itself too
// due to the fact that tendermint achieves instant finality
func (on OnlineNetwork) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	tx, log, err := on.client.GetTx(ctx, request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.BlockTransactionResponse{
		Transaction: &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{Hash: request.TransactionIdentifier.Hash},
			Operations:            conversion.SdkTxToOperations(tx, false, false),
			Metadata: map[string]interface{}{
				rosetta.Log: log,
			},
		},
	}, nil
}

// Mempool fetches the transactions contained in the mempool
func (on OnlineNetwork) Mempool(ctx context.Context, _ *types.NetworkRequest) (*types.MempoolResponse, *types.Error) {
	txs, err := on.client.Mempool(ctx)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.MempoolResponse{
		TransactionIdentifiers: conversion.TMTxsToRosettaTxsIdentifiers(txs.Txs),
	}, nil
}

// MempoolTransaction fetches a single transaction in the mempool
// NOTE: it is not implemented yet
func (on OnlineNetwork) MempoolTransaction(ctx context.Context, request *types.MempoolTransactionRequest) (*types.MempoolTransactionResponse, *types.Error) {
	tx, err := on.client.GetUnconfirmedTx(ctx, request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.MempoolTransactionResponse{
		Transaction: &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{Hash: request.TransactionIdentifier.Hash},
			Operations:            conversion.SdkTxToOperations(tx, false, false),
			Metadata:              nil,
		},
	}, nil
}

func (on OnlineNetwork) NetworkList(_ context.Context, _ *types.MetadataRequest) (*types.NetworkListResponse, *types.Error) {
	return &types.NetworkListResponse{NetworkIdentifiers: []*types.NetworkIdentifier{on.network}}, nil
}

func (on OnlineNetwork) NetworkOptions(_ context.Context, _ *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	return on.networkOptions, nil
}

func (on OnlineNetwork) NetworkStatus(ctx context.Context, _ *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	block, _, err := on.client.BlockByHeight(ctx, nil)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	peers, err := on.client.Peers(ctx)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	status, err := on.client.Status(ctx)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.NetworkStatusResponse{
		CurrentBlockIdentifier: conversion.TMBlockToRosettaBlockIdentifier(block),
		CurrentBlockTimestamp:  conversion.TimeToMilliseconds(block.Block.Time),
		GenesisBlockIdentifier: on.genesisBlockIdentifier,
		OldestBlockIdentifier:  nil,
		SyncStatus:             conversion.TMStatusToRosettaSyncStatus(status),
		Peers:                  conversion.TmPeersToRosettaPeers(peers),
	}, nil
}
