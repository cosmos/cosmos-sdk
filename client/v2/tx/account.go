package tx

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
	authtypes "cosmossdk.io/x/auth/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: move to internal

const (
	// GRPCBlockHeightHeader is the gRPC header for block height.
	GRPCBlockHeightHeader = "x-cosmos-block-height"
)

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

func newAccountRetriever(ac address.Codec, conn gogogrpc.ClientConn, registry codectypes.InterfaceRegistry) *accountRetriever {
	return &accountRetriever{
		ac:       ac,
		conn:     conn,
		registry: registry,
	}
}

func (a accountRetriever) GetAccount(ctx context.Context, addr []byte) (Account, error) {
	acc, _, err := a.GetAccountWithHeight(ctx, addr)
	return acc, err
}

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
	if l := len(blockHeight); l != 1 {
		return nil, 0, fmt.Errorf("unexpected '%s' header length; got %d, expected: %d", GRPCBlockHeightHeader, l, 1)
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

func (a accountRetriever) EnsureExists(ctx context.Context, addr []byte) error {
	if _, err := a.GetAccount(ctx, addr); err != nil {
		return err
	}
	return nil
}

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
