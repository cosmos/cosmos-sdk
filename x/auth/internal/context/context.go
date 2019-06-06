package context

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AuthContext interface {
	WithAccountDecoder(*codec.Codec) AuthContext
	WithAccountStore(string) AuthContext
	GetAccount([]byte) (types.Account, error)
}

type CLIAuthContext struct {
	context.CLIContext
	AccDecoder   types.AccountDecoder
	AccountStore string
}

var _ AuthContext = CLIAuthContext{}

func NewCLIAuthContextFromCLIContext(cliCtx context.CLIContext) CLIAuthContext {
	return CLIAuthContext{CLIContext: cliCtx, AccountStore: types.StoreKey}
}

// WithAccountDecoder returns a copy of the context with an updated account
// decoder.
func (ctx CLIAuthContext) WithAccountDecoder(cdc *codec.Codec) AuthContext {
	ctx.AccDecoder = GetAccountDecoder(cdc)
	return ctx
}

// WithAccountStore returns a copy of the context with an updated AccountStore.
func (ctx CLIAuthContext) WithAccountStore(accountStore string) AuthContext {
	ctx.AccountStore = accountStore
	return ctx
}

// GetAccount queries for an account given an address and a block height. An
// error is returned if the query or decoding fails.
func (ctx CLIAuthContext) GetAccount(address []byte) (types.Account, error) {
	if ctx.AccDecoder == nil {
		return nil, errors.New("account decoder required but not provided")
	}

	res, err := ctx.queryAccount(address)
	if err != nil {
		return nil, err
	}

	var account types.Account
	if err := ctx.Codec.UnmarshalJSON(res, &account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccountNumber returns the next account number for the given account
// address.
func (ctx CLIAuthContext) GetAccountNumber(address []byte) (uint64, error) {
	account, err := ctx.GetAccount(address)
	if err != nil {
		return 0, err
	}

	return account.GetAccountNumber(), nil
}

// GetAccountSequence returns the sequence number for the given account
// address.
func (ctx CLIAuthContext) GetAccountSequence(address []byte) (uint64, error) {
	account, err := ctx.GetAccount(address)
	if err != nil {
		return 0, err
	}

	return account.GetSequence(), nil
}

// EnsureAccountExists ensures that an account exists for a given context. An
// error is returned if it does not.
func (ctx CLIAuthContext) EnsureAccountExists() error {
	addr := ctx.GetFromAddress()
	return ctx.EnsureAccountExistsFromAddr(addr)
}

// EnsureAccountExistsFromAddr ensures that an account exists for a given
// address. Instead of using the context's from name, a direct address is
// given. An error is returned if it does not.
func (ctx CLIAuthContext) EnsureAccountExistsFromAddr(addr sdk.AccAddress) error {
	_, err := ctx.queryAccount(addr)
	return err
}


// queryAccount queries an account using custom query endpoint of auth module
// returns an error if result is `null` otherwise account data
func (ctx CLIAuthContext) queryAccount(addr sdk.AccAddress) ([]byte, error) {
	bz, err := ctx.Codec.MarshalJSON(types.NewQueryAccountParams(addr))
	if err != nil {
		return nil, err
	}

	route := fmt.Sprintf("custom/%s/%s", ctx.AccountStore, types.QueryAccount)
	return ctx.QueryWithData(route, bz)
}

// GetAccountDecoder gets the account decoder for auth.DefaultAccount.
func GetAccountDecoder(cdc *codec.Codec) types.AccountDecoder {
	return func(accBytes []byte) (acct types.Account, err error) {
		err = cdc.UnmarshalBinaryBare(accBytes, &acct)
		if err != nil {
			panic(err)
		}

		return acct, err
	}
}
