package services

import (
	"context"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
)

// genesisBlockFetchTimeout defines a timeout to fetch the genesis block
const genesisBlockFetchTimeout = 15 * time.Second

// NewOnline builds a new online network
func NewOnline(client rosetta.CosmosClient, network *types.NetworkIdentifier) (rosetta.OnlineAPI, error) {
	ctx, cancel := context.WithTimeout(context.Background(), genesisBlockFetchTimeout)
	defer cancel()

	var genesisHeight int64 = 1
	block, err := client.BlockByHeight(ctx, &genesisHeight)
	if err != nil {
		return OnlineNetwork{}, err
	}

	return OnlineNetwork{
		client:                 client,
		network:                network,
		networkOptions:         &types.NetworkOptionsResponse{Version: rosetta.Version(client.NodeVersion()), Allow: rosetta.Allow(client.SupportedOperations())},
		genesisBlockIdentifier: block.Block,
	}, nil
}

// OnlineNetwork groups together all the components required for the full rosetta implementation
type OnlineNetwork struct {
	client rosetta.CosmosClient // used to query cosmos app + tendermint

	network        *types.NetworkIdentifier      // identifies the network, it's static
	networkOptions *types.NetworkOptionsResponse // identifies the network options, it's static

	genesisBlockIdentifier *types.BlockIdentifier // identifies genesis block, it's static
}

// AccountBalance retrieves the account balance of an address
// rosetta requires us to fetch the block information too
func (on OnlineNetwork) AccountBalance(ctx context.Context, request *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	var (
		height *int64
		block  rosetta.BlockResponse
		err    error
	)

	switch {
	case request.BlockIdentifier == nil:
		height = nil
		block, err = on.client.BlockByHeight(ctx, nil)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	case request.BlockIdentifier.Hash != nil:
		block, err = on.client.BlockByHash(ctx, *request.BlockIdentifier.Hash)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
		height = &block.Block.Index
	case request.BlockIdentifier.Index != nil:
		height = request.BlockIdentifier.Index
		block, err = on.client.BlockByHeight(ctx, height)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	}

	balances, err := on.client.Balances(ctx, request.AccountIdentifier.Address, height)

	return &types.AccountBalanceResponse{
		BlockIdentifier: block.Block,
		Balances:        balances,
	}, nil
}

// Block gets the transactions in the given block
func (on OnlineNetwork) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	var (
		blockResp rosetta.BlockTransactionsResponse
		err       error
	)
	// block identifier is assumed not to be nil as rosetta will do this check for us
	// check if we have to query via hash or block number
	switch {
	case request.BlockIdentifier.Hash != nil:
		blockResp, err = on.client.BlockTransactionsByHash(ctx, *request.BlockIdentifier.Hash)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	case request.BlockIdentifier.Index != nil:
		blockResp, err = on.client.BlockTransactionsByHeight(ctx, request.BlockIdentifier.Index)
		if err != nil {
			return nil, rosetta.ToRosettaError(err)
		}
	default:
		return nil, rosetta.WrapError(rosetta.ErrBadArgument, "at least one of hash or index needs to be specified").RosettaError()
	}

	resp := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       blockResp.Block,
			ParentBlockIdentifier: blockResp.ParentBlock,
			Timestamp:             blockResp.MillisecondTimestamp, // ts is required in milliseconds
			Transactions:          blockResp.Transactions,
		},
		OtherTransactions: nil,
	}

	// if height is 1, use the same block identifier as current block
	if resp.Block.BlockIdentifier.Index == 1 {
		resp.Block.ParentBlockIdentifier = resp.Block.BlockIdentifier
	}
	return resp, nil
}

// BlockTransaction gets the given transaction in the specified block, we do not need to check the block itself too
// due to the fact that tendermint achieves instant finality
func (on OnlineNetwork) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	tx, err := on.client.GetTransaction(ctx, request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.BlockTransactionResponse{
		Transaction: tx,
	}, nil
}

// Mempool fetches the transactions contained in the mempool
func (on OnlineNetwork) Mempool(ctx context.Context, _ *types.NetworkRequest) (*types.MempoolResponse, *types.Error) {
	txs, err := on.client.GetMempoolTransactions(ctx)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.MempoolResponse{
		TransactionIdentifiers: txs,
	}, nil
}

// MempoolTransaction fetches a single transaction in the mempool
// NOTE: it is not implemented yet
func (on OnlineNetwork) MempoolTransaction(ctx context.Context, request *types.MempoolTransactionRequest) (*types.MempoolTransactionResponse, *types.Error) {
	tx, err := on.client.GetMempoolTransaction(ctx, request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, rosetta.ToRosettaError(err)
	}

	return &types.MempoolTransactionResponse{
		Transaction: tx,
	}, nil
}

func (on OnlineNetwork) NetworkList(_ context.Context, _ *types.MetadataRequest) (*types.NetworkListResponse, *types.Error) {
	return &types.NetworkListResponse{NetworkIdentifiers: []*types.NetworkIdentifier{on.network}}, nil
}

func (on OnlineNetwork) NetworkOptions(_ context.Context, _ *types.NetworkRequest) (*types.NetworkOptionsResponse, *types.Error) {
	return on.networkOptions, nil
}

func (on OnlineNetwork) NetworkStatus(ctx context.Context, _ *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	block, err := on.client.BlockByHeight(ctx, nil)
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
		CurrentBlockIdentifier: block.Block,
		CurrentBlockTimestamp:  block.MillisecondTimestamp,
		GenesisBlockIdentifier: on.genesisBlockIdentifier,
		OldestBlockIdentifier:  nil,
		SyncStatus:             status,
		Peers:                  peers,
	}, nil
}
