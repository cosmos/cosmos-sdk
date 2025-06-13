package client

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/v2/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/v2/rpc/client"
	"github.com/cometbft/cometbft/v2/rpc/client/mock"
	coretypes "github.com/cometbft/cometbft/v2/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/v2/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ AccountRetriever = TestAccountRetriever{}
	_ Account          = TestAccount{}
)

// TestAccount represents a client Account that can be used in unit tests
type TestAccount struct {
	Address sdk.AccAddress
	Num     uint64
	Seq     uint64
}

// GetAddress implements client Account.GetAddress
func (t TestAccount) GetAddress() sdk.AccAddress {
	return t.Address
}

// GetPubKey implements client Account.GetPubKey
func (t TestAccount) GetPubKey() cryptotypes.PubKey {
	return nil
}

// GetAccountNumber implements client Account.GetAccountNumber
func (t TestAccount) GetAccountNumber() uint64 {
	return t.Num
}

// GetSequence implements client Account.GetSequence
func (t TestAccount) GetSequence() uint64 {
	return t.Seq
}

// TestAccountRetriever is an AccountRetriever that can be used in unit tests
type TestAccountRetriever struct {
	Accounts map[string]TestAccount
}

// GetAccount implements AccountRetriever.GetAccount
func (t TestAccountRetriever) GetAccount(_ Context, addr sdk.AccAddress) (Account, error) {
	acc, ok := t.Accounts[addr.String()]
	if !ok {
		return nil, fmt.Errorf("account %s not found", addr)
	}

	return acc, nil
}

// GetAccountWithHeight implements AccountRetriever.GetAccountWithHeight
func (t TestAccountRetriever) GetAccountWithHeight(clientCtx Context, addr sdk.AccAddress) (Account, int64, error) {
	acc, err := t.GetAccount(clientCtx, addr)
	if err != nil {
		return nil, 0, err
	}

	return acc, 0, nil
}

// EnsureExists implements AccountRetriever.EnsureExists
func (t TestAccountRetriever) EnsureExists(_ Context, addr sdk.AccAddress) error {
	_, ok := t.Accounts[addr.String()]
	if !ok {
		return fmt.Errorf("account %s not found", addr)
	}
	return nil
}

// GetAccountNumberSequence implements AccountRetriever.GetAccountNumberSequence
func (t TestAccountRetriever) GetAccountNumberSequence(_ Context, addr sdk.AccAddress) (accNum, accSeq uint64, err error) {
	acc, ok := t.Accounts[addr.String()]
	if !ok {
		return 0, 0, fmt.Errorf("account %s not found", addr)
	}
	return acc.Num, acc.Seq, nil
}

type MockClient struct {
	mock.Client
	err error
}

func (c MockClient) ABCIQueryWithOptions(
	ctx context.Context,
	_ string,
	_ bytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return handleError[*coretypes.ResultABCIQuery](ctx, c.err)
}

func (c MockClient) BlockSearch(
	ctx context.Context,
	_ string,
	_, _ *int,
	_ string,
) (*coretypes.ResultBlockSearch, error) {
	return handleError[*coretypes.ResultBlockSearch](ctx, c.err)
}

func (c MockClient) BroadcastTxAsync(ctx context.Context, _ cmttypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	return handleError[*coretypes.ResultBroadcastTx](ctx, c.err)
}

func (c MockClient) BroadcastTxSync(ctx context.Context, _ cmttypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	return handleError[*coretypes.ResultBroadcastTx](ctx, c.err)
}

func (c MockClient) Block(ctx context.Context, _ *int64) (*coretypes.ResultBlock, error) {
	return handleError[*coretypes.ResultBlock](ctx, c.err)
}

func (c MockClient) BlockByHash(ctx context.Context, _ []byte) (*coretypes.ResultBlock, error) {
	return handleError[*coretypes.ResultBlock](ctx, c.err)
}

func (c MockClient) Status(ctx context.Context) (*coretypes.ResultStatus, error) {
	return handleError[*coretypes.ResultStatus](ctx, c.err)
}

func (c MockClient) Tx(ctx context.Context, _ []byte, _ bool) (*coretypes.ResultTx, error) {
	return handleError[*coretypes.ResultTx](ctx, c.err)
}

func (c MockClient) TxSearch(
	ctx context.Context,
	_ string,
	_ bool,
	_, _ *int,
	_ string,
) (*coretypes.ResultTxSearch, error) {
	return handleError[*coretypes.ResultTxSearch](ctx, c.err)
}

func handleError[T any](ctx context.Context, err error) (T, error) {
	var ret T
	if ctx != nil && ctx.Err() != nil {
		return ret, ctx.Err()
	} else {
		return ret, err
	}
}
