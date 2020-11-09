package rosetta

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"
	"github.com/tendermint/cosmos-rosetta-gateway/service"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// list of supported operations
const (
	StatusReverted   = "Reverted"
	StatusSuccess    = "Success"
	OperationMsgSend = "send"
)

// Synchronization stage constants used to determine if a node is synced or catching up
const (
	StageSynced  = "synced"
	StageSyncing = "syncing"
)

// NewNetwork builds a rosetta gateway network
func NewNetwork(networkIdentifier *types.NetworkIdentifier, adapter crg.Adapter) service.Network {
	return service.Network{
		Properties: crg.NetworkProperties{
			Blockchain:          networkIdentifier.Blockchain,
			Network:             networkIdentifier.Network,
			SupportedOperations: []string{OperationMsgSend},
		},
		Adapter: adapter,
	}
}

// SdkTxWithHash wraps an sdk transaction with its hash and block identifier
type SdkTxWithHash struct {
	HexHash string
	Tx      sdk.Tx
}

// DataAPIClient defines the interface
// a client has to implement in order to
// interact with cosmos-sdk chains
type DataAPIClient interface {
	// Balances fetches the balance of the given address
	// if height is not nil, then the balance will be displayed
	// at the provided height, otherwise last block balance will be returned
	Balances(ctx context.Context, addr string, height *int64) ([]sdk.Coin, error)
	// BlockByHash gets a block and its transaction at the provided height
	BlockByHash(ctx context.Context, hash string) (*tmtypes.ResultBlock, []*SdkTxWithHash, error)
	// BlockByHeight gets a block given its height, if height is nil then last block is returned
	BlockByHeight(ctx context.Context, height *int64) (*tmtypes.ResultBlock, []*SdkTxWithHash, error)
	// GetTx gets a transaction given its hash
	GetTx(ctx context.Context, hash string) (sdk.Tx, error)
	// GetUnconfirmedTx gets an unconfirmed Tx given its hash
	// NOTE(fdymylja): NOT IMPLEMENTED YET!
	GetUnconfirmedTx(ctx context.Context, hash string) (sdk.Tx, error)
	// Mempool returns the list of the current non confirmed transactions
	Mempool(ctx context.Context) (*tmtypes.ResultUnconfirmedTxs, error)
	// Peers gets the peers currently connected to the node
	Peers(ctx context.Context) ([]tmtypes.Peer, error)
	// Status returns the node status, such as sync data, version etc
	Status(ctx context.Context) (*tmtypes.ResultStatus, error)
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
		OperationTypes:          []string{OperationMsgSend},
		Errors:                  AllowedErrors.RosettaErrors(),
		HistoricalBalanceLookup: false,
		TimestampStartIndex:     nil,
		CallMethods:             nil,
		BalanceExemptions:       nil,
	}
}
