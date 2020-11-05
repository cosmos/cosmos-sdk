package rosetta

import (
	"context"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crg "github.com/tendermint/cosmos-rosetta-gateway/rosetta"
	"github.com/tendermint/cosmos-rosetta-gateway/service"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
)

// list of supported operations
const (
	StatusReverted = "Reverted"
	StatusSuccess  = "Success"

	OperationTransfer = "Transfer"

	OptionAddress = "address"
	OptionGas     = "gas"
)

// NewNetwork builds a rosetta gateway network
func NewNetwork(networkIdentifier *types.NetworkIdentifier, adapter crg.Adapter) service.Network {
	return service.Network{
		Properties: crg.NetworkProperties{
			Blockchain:          networkIdentifier.Blockchain,
			Network:             networkIdentifier.Network,
			AddrPrefix:          sdk.GetConfig().GetBech32AccountAddrPrefix(),                                         // since we're inside cosmos sdk the config is supposed to be sealed
			SupportedOperations: []string{StatusReverted, StatusSuccess, OperationTransfer, OptionAddress, OptionGas}, // TODO are this defaults always true?
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
}

// Error defines an error that can be converted to a rosetta error
type Error interface {
	error
	// RosettaError returns the error as a rosetta error
	RosettaError() *types.Error
}

type errorWrapper struct {
	err *types.Error
}

func (e errorWrapper) Error() string {
	return fmt.Sprintf("rosetta: %d %s", e.err.Code, e.err.Message)
}

func (e errorWrapper) RosettaError() *types.Error {
	return e.err
}

// NewError instantiates a new rosetta error
func NewError(code int32, message string, retry bool) Error {
	return errorWrapper{
		err: &types.Error{
			Code:        code,
			Message:     message,
			Description: nil,
			Retriable:   retry,
			Details:     nil,
		},
	}
}

// WrapError wraps the rosetta error with additional context
func WrapError(err Error, msg string) Error {
	e := err.RosettaError()
	e.Message = fmt.Sprintf("%s: %s", e.Message, msg)
	return err
}

// ToRosettaError attempts to converting an error into a rosetta
// error, if the error cannot be converted it will be parsed as unknown
func ToRosettaError(err error) *types.Error {
	rosErr, ok := err.(Error)
	if !ok {
		return WrapError(ErrUnknown, err.Error()).RosettaError()
	}
	return rosErr.RosettaError()
}

// FromGRPCToRosettaError converts a gRPC error to rosetta error
func FromGRPCToRosettaError(err error) Error {
	return errorWrapper{err: &types.Error{
		Code:        0,
		Message:     err.Error(),
		Description: nil,
		Retriable:   false,
		Details:     nil,
	}}
}

// Default error list
var (
	// ErrUnknown defines an unknown error, if this is returned it means
	// the library is ignoring an error
	ErrUnknown = NewError(0, "unknown", true)
	// ErrBadArgument is returned when the request is malformed
	ErrBadArgument = NewError(400, "bad argument", false)
	// ErrBadGateway is returned when there are problems interacting with the nodes
	ErrBadGateway = NewError(502, "bad gateway", true)
	// ErrOffline is returned when there is an attempt to query an endpoint in offline mode
	ErrOffline = NewError(1, "cannot query endpoint in offline mode", false)
	// ErrNetworkNotSupported is returned when there is an attempt to query a network which is not supported
	ErrNetworkNotSupported = NewError(2, "network is not supported", false)
	// ErrCodec is returned when there's an error while marshalling or unmarshalling data
	ErrCodec = NewError(3, "encode/decode error", true)
	// ErrNotImplemented is returned when a method is not implemented yet
	ErrNotImplemented = NewError(14, "not implemented", false)
	// ErrUnsupportedCurve is returned when the curve specified is not supported
	ErrUnsupportedCurve = NewError(15, "unsupported curve, expected secp256k1", false)
	// ErrInvalidPubkey is returned when the public key is invalid
	ErrInvalidPubkey = NewError(8, "invalid pubkey", false)
)
