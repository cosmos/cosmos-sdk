package services

import (
	"context"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"github.com/cosmos/cosmos-sdk/server/rosetta/cosmos/conversion"
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
		networkOptions:         &types.NetworkOptionsResponse{Version: rosetta.Version(), Allow: rosetta.Allow()},
		genesisBlockIdentifier: conversion.TendermintBlockToBlockIdentifier(block),
	}, nil
}

// SingleNetwork groups together all the components required for the full rosetta data API
// which is running on a single network
type SingleNetwork struct {
	client rosetta.DataAPIClient // used to query cosmos app + tendermint

	network        *types.NetworkIdentifier      // identifies the network, it's static
	networkOptions *types.NetworkOptionsResponse // identifies the network options, it's static

	genesisBlockIdentifier *types.BlockIdentifier // identifies genesis block, it's static

}

// AccountBalance retrieves the account balance of an address
// NOTE(fdymylja): for now historical balance is not supported
func (sn SingleNetwork) AccountBalance(ctx context.Context, request *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	// we need to provide block information, so get last block
	block, _, err := sn.client.BlockByHeight(ctx, nil)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	// now get balance
	balances, err := sn.client.Balances(ctx, request.AccountIdentifier.Address, nil)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	resp := &types.AccountBalanceResponse{
		BlockIdentifier: conversion.TendermintBlockToBlockIdentifier(block),
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
			BlockIdentifier:       conversion.TendermintBlockToBlockIdentifier(block),
			ParentBlockIdentifier: conversion.ParentBlockIdentifierFromLastBlock(block),
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

func (sn SingleNetwork) NetworkOptions(_ context.Context, _ *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	return sn.networkOptions, nil
}

func (sn SingleNetwork) NetworkStatus(ctx context.Context, _ *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	block, _, err := sn.client.BlockByHeight(ctx, nil)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	peers, err := sn.client.Peers(ctx)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	status, err := sn.client.Status(ctx)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}
	resp := &types.NetworkStatusResponse{
		CurrentBlockIdentifier: conversion.TendermintBlockToBlockIdentifier(block),
		CurrentBlockTimestamp:  conversion.TimeToMilliseconds(block.Block.Time),
		GenesisBlockIdentifier: sn.genesisBlockIdentifier,
		OldestBlockIdentifier:  nil,
		SyncStatus:             conversion.TendermintStatusToSync(status),
		Peers:                  conversion.TmPeersToRosettaPeers(peers),
	}
	return resp, nil
}
