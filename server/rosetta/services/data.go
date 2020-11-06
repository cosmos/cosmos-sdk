package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/cosmos/conversion"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
	"time"
)

// assert interface implementation
var _ crg.DataAPI = SingleNetwork{}

// genesisBlockFetchTimeout defines a timeout to fetch the genesis block
const genesisBlockFetchTimeout = 15 * time.Second

// NewSingleNetwork builds a single network client
// the client will attempt to fetch genesis block too
func NewSingleNetwork(client rosetta.DataAPIClient, network *types.NetworkIdentifier) (SingleNetwork, error) {
	ctx, cancel := context.WithTimeout(context.Background(), genesisBlockFetchTimeout)
	defer cancel()
	var genesisHeight int64 = 1
	block, _, err := client.BlockByHeight(ctx, &genesisHeight)
	if err != nil {
		return SingleNetwork{}, err
	}
	return SingleNetwork{
		client:                 client,
		network:                network,
		genesisBlockIdentifier: conversion.TendermintBlockToBlockIdentifier(block),
	}, nil
}

// SingleNetwork groups together all the components required for the full rosetta data API
// which is running on a single network
type SingleNetwork struct {
	client                 rosetta.DataAPIClient
	network                *types.NetworkIdentifier
	genesisBlockIdentifier *types.BlockIdentifier
}

// AccountBalance retrieves the account balance of an address
func (sn SingleNetwork) AccountBalance(ctx context.Context, request *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	var height *int64
	var blockHash *string
	// if block identifier is specified
	// then use it as reference for height
	if request.BlockIdentifier != nil {
		height = request.BlockIdentifier.Index
		blockHash = request.BlockIdentifier.Hash
	}
	var (
		balances sdk.Coins
		err      error
		block    *tmtypes.ResultBlock
	)
	// if the request is by hash then we need to get the height of the block first
	if blockHash != nil {
		block, _, err = sn.client.BlockByHash(ctx, *blockHash)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
		balances, err = sn.client.Balances(ctx, request.AccountIdentifier.Address, &block.Block.Height)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
		// otherwise get it by height, which if nil means we'll fetch the last block's balances
	} else {
		balances, err = sn.client.Balances(ctx, request.AccountIdentifier.Address, height)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	}
	resp := &types.AccountBalanceResponse{
		BlockIdentifier: partialBlockIdentifierToFull(request.BlockIdentifier),
		Balances:        conversion.CoinsToBalance(balances),
		Coins:           nil,
		Metadata:        nil,
	}
	return resp, nil
}

// Block gets the transactions in the given block
func (sn SingleNetwork) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	var (
		block *tmtypes.ResultBlock
		txs   []*rosetta.SdkTxWithHash
		err   error
	)
	if request.BlockIdentifier.Hash != nil {
		block, txs, err = sn.client.BlockByHash(ctx, *request.BlockIdentifier.Hash)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	} else if request.BlockIdentifier.Index != nil {
		block, txs, err = sn.client.BlockByHeight(ctx, request.BlockIdentifier.Index)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	} else {
		return nil, rosetta.WrapError(rosetta.ErrBadArgument, "at least one of hash or index needs to be specified").RosettaError()
	}
	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: block.Block.Height,
				Hash:  block.BlockID.Hash.String(),
			},
			ParentBlockIdentifier: conversion.TendermintBlockToBlockIdentifier(block),
			Timestamp:             conversion.TimeToMilliseconds(block.Block.Time), // ts is required in milliseconds
			Transactions:          conversion.ResultTxSearchToTransaction(txs),
			Metadata:              nil,
		},
		OtherTransactions: nil,
	}, nil
}

func (sn SingleNetwork) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	tx, err := sn.client.GetTx(ctx, request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return &types.BlockTransactionResponse{
		Transaction: &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{Hash: request.TransactionIdentifier.Hash},
			Operations:            conversion.SdkTxToOperations(tx, false, false),
			Metadata:              nil,
		},
	}, nil
}

func (sn SingleNetwork) Mempool(ctx context.Context, _ *types.NetworkRequest) (*types.MempoolResponse, *types.Error) {
	txs, err := sn.client.Mempool(ctx)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	return &types.MempoolResponse{
		TransactionIdentifiers: conversion.TendermintTxsToTxIdentifiers(txs.Txs),
	}, nil
}

func (sn SingleNetwork) MempoolTransaction(ctx context.Context, request *types.MempoolTransactionRequest) (*types.MempoolTransactionResponse, *types.Error) {
	tx, err := sn.client.GetUnconfirmedTx(ctx, request.TransactionIdentifier.Hash)
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

func (sn SingleNetwork) NetworkList(_ context.Context, _ *types.MetadataRequest) (*types.NetworkListResponse, *types.Error) {
	return &types.NetworkListResponse{NetworkIdentifiers: []*types.NetworkIdentifier{sn.network}}, nil
}

func (sn SingleNetwork) NetworkOptions(ctx context.Context, request *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion:    "",
			NodeVersion:       "",
			MiddlewareVersion: nil,
			Metadata:          nil,
		},
		Allow: &types.Allow{
			OperationStatuses:       nil,
			OperationTypes:          nil,
			Errors:                  nil,
			HistoricalBalanceLookup: false,
			TimestampStartIndex:     nil,
			CallMethods:             nil,
			BalanceExemptions:       nil,
		},
	}, nil
}

func (sn SingleNetwork) NetworkStatus(ctx context.Context, _ *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	block, _, err := sn.client.BlockByHeight(ctx, nil)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	resp := &types.NetworkStatusResponse{
		CurrentBlockIdentifier: conversion.TendermintBlockToBlockIdentifier(block),
		CurrentBlockTimestamp:  conversion.TimeToMilliseconds(block.Block.Time),
		GenesisBlockIdentifier: sn.genesisBlockIdentifier,
		OldestBlockIdentifier:  nil, // TODO what is this, most likely foresees that the node we're querying is not synced yet
		SyncStatus:             nil, // TODO what is this
		Peers: []*types.Peer{
			{
				PeerID:   "",
				Metadata: nil,
			},
		},
	}
	return resp, nil
}

// partialBlockIdentifierToFull converts a partial block identifier to a fully qualified block identifier
func partialBlockIdentifierToFull(partial *types.PartialBlockIdentifier) *types.BlockIdentifier {
	if partial == nil {
		return nil
	}
	full := &types.BlockIdentifier{}
	if partial.Hash != nil {
		full.Hash = *partial.Hash
	}
	if partial.Index != nil {
		full.Index = *partial.Index
	}
	return full
}
