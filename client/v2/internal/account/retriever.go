package account

import (
	"context"
	"fmt"
	"strconv"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"cosmossdk.io/core/address"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GRPCBlockHeightHeader represents the gRPC header for block height.
const GRPCBlockHeightHeader = "x-cosmos-block-height"

var _ AccountRetriever = accountRetriever{}

// Account provides a read-only abstraction over the auth module's AccountI.
type Account interface {
	GetAddress() sdk.AccAddress
	GetPubKey() cryptotypes.PubKey // can return nil.
	GetAccountNumber() uint64
	GetSequence() uint64
}

// AccountRetriever defines methods required to retrieve account details necessary for transaction signing.
type AccountRetriever interface {
	GetAccount(context.Context, []byte) (Account, error)
	GetAccountWithHeight(context.Context, []byte) (Account, int64, error)
	EnsureExists(context.Context, []byte) error
	GetAccountNumberSequence(context.Context, []byte) (accNum, accSeq uint64, err error)
}

type accountRetriever struct {
	ac       address.Codec
	conn     gogogrpc.ClientConn
	registry codectypes.InterfaceRegistry
}

// NewAccountRetriever creates a new instance of accountRetriever.
func NewAccountRetriever(ac address.Codec, conn gogogrpc.ClientConn, registry codectypes.InterfaceRegistry) *accountRetriever {
	return &accountRetriever{
		ac:       ac,
		conn:     conn,
		registry: registry,
	}
}

// GetAccount retrieves an account using its address.
func (a accountRetriever) GetAccount(ctx context.Context, addr []byte) (Account, error) {
	acc, _, err := a.GetAccountWithHeight(ctx, addr)
	return acc, err
}

// GetAccountWithHeight retrieves an account and its associated block height using the account's address.
func (a accountRetriever) GetAccountWithHeight(ctx context.Context, addr []byte) (Account, int64, error) {
	var header metadata.MD
	qc := authtypes.NewQueryClient(a.conn)

	addrStr, err := a.ac.BytesToString(addr)
	if err != nil {
		return nil, 0, err
	}

	res, err := qc.Account(ctx, &authtypes.QueryAccountRequest{Address: addrStr}, grpc.Header(&header))
	if err != nil {
		return nil, 0, err
	}

	blockHeight := header.Get(GRPCBlockHeightHeader)
	if len(blockHeight) != 1 {
		return nil, 0, fmt.Errorf("unexpected '%s' header length; got %d, expected 1", GRPCBlockHeightHeader, len(blockHeight))
	}

	nBlockHeight, err := strconv.Atoi(blockHeight[0])
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse block height: %w", err)
	}

	var acc Account
	if err := a.registry.UnpackAny(res.Account, &acc); err != nil {
		return nil, 0, err
	}

	return acc, int64(nBlockHeight), nil
}

// EnsureExists checks if an account exists using its address.
func (a accountRetriever) EnsureExists(ctx context.Context, addr []byte) error {
	if _, err := a.GetAccount(ctx, addr); err != nil {
		return err
	}
	return nil
}

// GetAccountNumberSequence retrieves the account number and sequence for an account using its address.
func (a accountRetriever) GetAccountNumberSequence(ctx context.Context, addr []byte) (accNum, accSeq uint64, err error) {
	acc, err := a.GetAccount(ctx, addr)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	return acc.GetAccountNumber(), acc.GetSequence(), nil
}
