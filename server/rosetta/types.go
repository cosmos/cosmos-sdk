package rosetta

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
)

const (
	SpecVersion = "1.4.6"
)

const (
	StatusSuccess  = "Success"
	StatusReverted = "Reverted"
)

// BlockTransactionsResponse is a convenience wrapper
type BlockTransactionsResponse struct {
	BlockResponse
	Transactions []*types.Transaction
}

// BlockResponse is a convenience wrapper
// since some rosetta API calls require
// the block timestamp too
type BlockResponse struct {
	Block                *types.BlockIdentifier
	ParentBlock          *types.BlockIdentifier
	MillisecondTimestamp int64
	TxCount              int64
}

type CosmosClient interface {
	CosmosDataAPIClient
	CosmosConstructionAPIClient
}

// CosmosConstructionAPIClient defines the interface cosmos sdk implementation
// must satisfy in order to provide access to the construction API
type CosmosConstructionAPIClient interface {
	AccountIdentifierFromPubKeyBytes(curveType string, pkBytes []byte) (account *types.AccountIdentifier, err error)
	TransactionIdentifierFromHexBytes(hexBytes []byte) (txIdentifier *types.TransactionIdentifier, err error)
	TxOperationsAndSignersAccountIdentifiers(signed bool, hexBytes []byte) (ops []*types.Operation, signers []*types.AccountIdentifier, err error)
	PostTxBytes(ctx context.Context, txBytes []byte) (txResp *types.TransactionIdentifier, meta map[string]interface{}, err error)
	ConstructionMetadataFromOptions(ctx context.Context, options map[string]interface{}) (meta map[string]interface{}, err error)
	SignedTx(ctx context.Context, txBytes []byte, sigs []*types.Signature) (signedTxBytes []byte, err error)
	OperationsToMetadata(ctx context.Context, ops []*types.Operation) (meta map[string]interface{}, err error)
	ConstructionPayload(ctx context.Context, req *types.ConstructionPayloadsRequest) (resp *types.ConstructionPayloadsResponse, err error)
}

// CosmosDataAPIClient defines the cosmos client that
// is used to interact with the data API service
// it returns rosetta types relative to the version
// of the rosetta module used.
type CosmosDataAPIClient interface {
	// Balances fetches the balance given an account identifier
	// if height is nil, last height balance must be returned
	Balances(ctx context.Context, address string, height *int64) ([]*types.Amount, error)
	// BlockByHeight fetches a block given its height, if height is nil
	// last block must be fetched
	BlockByHeight(ctx context.Context, height *int64) (BlockResponse, error)
	// BlockByHash fetches a block given its hash
	BlockByHash(ctx context.Context, hash string) (BlockResponse, error)
	// BlockTransactionsByHeight gets the block, parent block and transactions
	// given the block height, if height is nil then last height is used
	BlockTransactionsByHeight(ctx context.Context, height *int64) (BlockTransactionsResponse, error)
	// BlockTransactionsByHash gets the block, parent block and transactions
	// given the block hash
	BlockTransactionsByHash(ctx context.Context, hash string) (BlockTransactionsResponse, error)
	// GetTransaction gets a transaction given its hash
	GetTransaction(ctx context.Context, hash string) (tx *types.Transaction, err error)
	// GetMempoolTransactions returns the transactions from the mempool
	GetMempoolTransactions(ctx context.Context) (txs []*types.TransactionIdentifier, err error)
	// GetMempoolTransaction returns the full transaction by hash in the mempool
	GetMempoolTransaction(ctx context.Context, hash string) (tx *types.Transaction, err error)
	// Peers returns the peers
	Peers(ctx context.Context) (peers []*types.Peer, err error)
	// Status returns the current synchronization status
	Status(ctx context.Context) (status *types.SyncStatus, err error)
	// SupportedOperations returns the list of supported ops
	SupportedOperations() []string
	// NodeVersion returns the cosmos sdk version
	// and possibly the tendermint version
	NodeVersion() string
}

// OnlineAPI defines the exposed APIs
// if the service is online
type OnlineAPI interface {
	DataAPI
	ConstructionAPI
}

type OfflineAPI interface {
	ConstructionAPI
}

// DataAPI defines the full data OnlineAPI implementation
type DataAPI interface {
	server.NetworkAPIServicer
	server.AccountAPIServicer
	server.BlockAPIServicer
	server.MempoolAPIServicer
}

// ConstructionAPI defines the construction OnlineAPI implementation
type ConstructionAPI interface {
	server.ConstructionAPIServicer
}

// Version returns the version for rosetta
// which contains static data for rosetta spec
// version and the variable data regarding the node
func Version(nodeVersion string) *types.Version {
	return &types.Version{
		RosettaVersion:    SpecVersion,
		NodeVersion:       nodeVersion,
		MiddlewareVersion: nil,
		Metadata:          nil,
	}
}

// Allow returns the allow operations
// some values are default, as statuses
// others depend on version of the node
func Allow(supportedOperations []string) *types.Allow {
	return &types.Allow{
		OperationStatuses: []*types.OperationStatus{
			{
				Status:     StatusSuccess,
				Successful: true,
			},
			{
				Status:     StatusReverted,
				Successful: false,
			},
		},
		OperationTypes:          supportedOperations,
		Errors:                  AllowedErrors.RosettaErrors(),
		HistoricalBalanceLookup: false,
		TimestampStartIndex:     nil,
		CallMethods:             nil,
		BalanceExemptions:       nil,
	}
}
