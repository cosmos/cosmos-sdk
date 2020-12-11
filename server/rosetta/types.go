package rosetta

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"
	"github.com/tendermint/cosmos-rosetta-gateway/service"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// list of supported operations
const (
	StatusReverted    = "Reverted"
	StatusSuccess     = "Success"
	OptionAddress     = "address"
	OptionGas         = "gas"
	OptionMemo        = "memo"
	Sequence          = "sequence"
	AccountNumber     = "account_number"
	ChainID           = "chain_id"
	OperationSend     = "send"
	OperationFee      = "fee"
	Log               = "log"
	OperationDelegate = "delegate"
)

// Synchronization stage constants used to determine if a node is synced or catching up
const (
	StageSynced  = "synced"
	StageSyncing = "syncing"
)

var SupportedOperations = []string{OperationSend, OperationFee, OperationDelegate}

// NewNetwork builds a rosetta gateway network
func NewNetwork(networkIdentifier *types.NetworkIdentifier, adapter crg.Adapter) service.Network {
	return service.Network{
		Properties: crg.NetworkProperties{
			Blockchain:          networkIdentifier.Blockchain,
			Network:             networkIdentifier.Network,
			AddrPrefix:          sdk.GetConfig().GetBech32AccountAddrPrefix(), // since we're inside cosmos sdk the config is supposed to be sealed
			SupportedOperations: SupportedOperations,
		},
		Adapter: adapter,
	}
}

// SdkTxWithHash wraps an sdk transaction with its hash and block identifier
type SdkTxWithHash struct {
	HexHash string
	Code    uint32
	Log     string
	Tx      sdk.Tx
}

// NodeClient defines the interface
// a client has to implement in order to
// interact with cosmos-sdk chains
type NodeClient interface {
	AccountInfo(ctx context.Context, addr string, height *int64) (auth.AccountI, error)
	// Balances fetches the balance of the given address
	// if height is not nil, then the balance will be displayed
	// at the provided height, otherwise last block balance will be returned
	Balances(ctx context.Context, addr string, height *int64) ([]*types.Amount, error)
	// BlockByHash gets a block and its transaction at the provided height
	BlockByHash(ctx context.Context, hash string) (*tmtypes.ResultBlock, []*SdkTxWithHash, error)
	// BlockByHeight gets a block given its height, if height is nil then last block is returned
	BlockByHeight(ctx context.Context, height *int64) (*tmtypes.ResultBlock, []*SdkTxWithHash, error)
	// BlockByHashAlt gets a block and its transaction at the provided height
	BlockByHashAlt(ctx context.Context, hash string) (BlockResponse, error)
	// BlockByHeightAlt gets a block given its height, if height is nil then last block is returned
	BlockByHeightAlt(ctx context.Context, height *int64) (BlockResponse, error)
	// BlockTransactionsByHash gets the block, parent block and transactions
	// given the block hash.
	BlockTransactionsByHash(ctx context.Context, hash string) (BlockTransactionsResponse, error)
	// BlockTransactionsByHash gets the block, parent block and transactions
	// given the block hash.
	BlockTransactionsByHeight(ctx context.Context, height *int64) (BlockTransactionsResponse, error)
	// GetTx gets a transaction given its hash
	GetTx(ctx context.Context, hash string) (*SdkTxWithHash, string, error)
	// GetUnconfirmedTx gets an unconfirmed Tx given its hash
	// NOTE(fdymylja): NOT IMPLEMENTED YET!
	GetUnconfirmedTx(ctx context.Context, hash string) (sdk.Tx, error)
	// Mempool returns the list of the current non confirmed transactions
	Mempool(ctx context.Context) (*tmtypes.ResultUnconfirmedTxs, error)
	// Peers gets the peers currently connected to the node
	Peers(ctx context.Context) ([]tmtypes.Peer, error)
	// Status returns the node status, such as sync data, version etc
	Status(ctx context.Context) (*tmtypes.ResultStatus, error)
	GetTxConfig() client.TxConfig
	PostTx(txBytes []byte) (res *sdk.TxResponse, err error)
}

type BlockTransactionsResponse struct {
	BlockResponse
	Transactions []*types.Transaction
}

type BlockResponse struct {
	Block                *types.BlockIdentifier
	ParentBlock          *types.BlockIdentifier
	MillisecondTimestamp int64
	TxCount              int64
}

// Msg interface is the interface that Cosmos SDK messages should implement if they want to
// be supported by the Rosetta service.
type Msg interface {
	ToOperations(withStatus bool, hasError bool) []*types.Operation
}

// Version returns the version for rosetta
// since this value is static, we can wrap it here
func Version() *types.Version {
	const rosettaSpecVersion = "1.4.6"
	const cosmosSdkVersion = "0.40.0-rc2"
	return &types.Version{
		RosettaVersion:    rosettaSpecVersion,
		NodeVersion:       cosmosSdkVersion,
		MiddlewareVersion: nil,
		Metadata:          nil,
	}
}

// Allow returns the allow operations
// and error information, since this is
// a static information we can club it here
func Allow() *types.Allow {
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
		OperationTypes:          SupportedOperations,
		Errors:                  AllowedErrors.RosettaErrors(),
		HistoricalBalanceLookup: true,
		TimestampStartIndex:     nil,
		CallMethods:             nil,
		BalanceExemptions:       nil,
	}
}
