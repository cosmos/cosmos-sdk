package tx

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"strconv"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"cosmossdk.io/core/address"
	authtypes "cosmossdk.io/x/auth/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

// TODO: move to internal

var _ AccountRetriever = accountRetriever{}

// Account defines a read-only version of the auth module's AccountI.
type Account interface {
	GetAddress() sdk.AccAddress
	GetPubKey() cryptotypes.PubKey // can return nil.
	GetAccountNumber() uint64
	GetSequence() uint64
}

// AccountRetriever defines the interfaces required by transactions to
// ensure an account exists and to be able to query for account fields necessary
// for signing.
type AccountRetriever interface {
	GetAccount(context.Context, sdk.AccAddress) (Account, error)
	GetAccountWithHeight(context.Context, sdk.AccAddress) (Account, int64, error)
	EnsureExists(context.Context, sdk.AccAddress) error
	GetAccountNumberSequence(context.Context, sdk.AccAddress) (accNum, accSeq uint64, err error)
}

type accountRetriever struct {
	ac       address.Codec
	conn     gogogrpc.ClientConn
	registry codectypes.InterfaceRegistry
}

func newAccountRetriever(ac address.Codec, conn gogogrpc.ClientConn, registry codectypes.InterfaceRegistry) *accountRetriever {
	return &accountRetriever{
		ac:       ac,
		conn:     conn,
		registry: registry,
	}
}

func (a accountRetriever) GetAccount(ctx context.Context, addr sdk.AccAddress) (Account, error) {
	acc, _, err := a.GetAccountWithHeight(ctx, addr)
	return acc, err
}

func (a accountRetriever) GetAccountWithHeight(ctx context.Context, addr sdk.AccAddress) (Account, int64, error) {
	var header metadata.MD

	qc := authtypes.NewQueryClient(a.conn)

	res, err := qc.Account(ctx, &authtypes.QueryAccountRequest{Address: addr.String()}, grpc.Header(&header))
	if err != nil {
		return nil, 0, err
	}

	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	if l := len(blockHeight); l != 1 {
		return nil, 0, fmt.Errorf("unexpected '%s' header length; got %d, expected: %d", grpctypes.GRPCBlockHeightHeader, l, 1)
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

func (a accountRetriever) EnsureExists(ctx context.Context, addr sdk.AccAddress) error {
	if _, err := a.GetAccount(ctx, addr); err != nil {
		return err
	}
	return nil
}

func (a accountRetriever) GetAccountNumberSequence(ctx context.Context, addr sdk.AccAddress) (accNum, accSeq uint64, err error) {
	acc, err := a.GetAccount(ctx, addr)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	return acc.GetAccountNumber(), acc.GetSequence(), nil
}
