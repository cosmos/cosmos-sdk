package context

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/pkg/errors"

	"strings"

	cmn "github.com/tendermint/tendermint/libs/common"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/store"
)

// Query performs a query for information about the connected node.
func (ctx CLIContext) Query(path string, data cmn.HexBytes) (res []byte, err error) {
	return ctx.query(path, data)
}

// Query information about the connected node with a data payload
func (ctx CLIContext) QueryWithData(path string, data []byte) (res []byte, err error) {
	return ctx.query(path, data)
}

// QueryStore performs a query from a Tendermint node with the provided key and store name.
func (ctx CLIContext) QueryStore(key cmn.HexBytes, storeName string) (res []byte, err error) {
	return ctx.queryStore(key, storeName, "key")
}

// QuerySubspace performs a query from a Tendermint node with the provided store and subspace.
func (ctx CLIContext) QuerySubspace(subspace []byte, storeName string) (res []sdk.KVPair, err error) {
	resRaw, err := ctx.queryStore(subspace, storeName, "subspace")
	if err != nil {
		return res, err
	}

	ctx.Codec.MustUnmarshalBinaryLengthPrefixed(resRaw, &res)
	return
}

// FetchAccount queries for an account given an address and a block height. An
// error is returned if the query or decoding fails.
func (ctx CLIContext) FetchAccount(address []byte) (auth.Account, error) {
	if ctx.AccDecoder == nil {
		ctx.WithAccountDecoder()
	}

	res, err := ctx.QueryStore(auth.AddressStoreKey(address), ctx.AccountStore)
	if err != nil {
		return nil, err
	} else if len(res) == 0 {
		return nil, ErrInvalidAccount(address)
	}

	account, err := ctx.AccDecoder(res)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccAndSeqNums fetches the account and sequence number for a given address
func (ctx CLIContext) FetchAccAndSeqNums(address sdk.Address) (uint64, uint64, error) {
	account, err := ctx.FetchAccount(address.Bytes())
	if err != nil {
		return 0, 0, err
	}
	return account.GetAccountNumber(), account.GetSequence(), err
}

// EnsureAccountExists ensures that an account exists for a given
// address. Instead of using the context's from name, a direct address is
// given. An error is returned if it does not.
func (ctx CLIContext) EnsureAccountExists(addr sdk.AccAddress) error {
	accountBytes, err := ctx.QueryStore(auth.AddressStoreKey(addr), ctx.AccountStore)
	if err != nil {
		return err
	}

	if len(accountBytes) == 0 {
		return ErrInvalidAccount(addr)
	}

	return nil
}

// query performs a query from a Tendermint node with the provided store name
// and path.
func (ctx CLIContext) query(path string, key cmn.HexBytes) (res []byte, err error) {
	node, err := ctx.GetClient()
	if err != nil {
		return res, err
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: ctx.Height,
		Prove:  !ctx.TrustNode,
	}

	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return res, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return res, errors.Errorf(resp.Log)
	}

	// data from trusted node or subspace query doesn't need verification
	if ctx.TrustNode || !isQueryStoreWithProof(path) {
		return resp.Value, nil
	}

	err = ctx.verifyProof(path, resp)
	if err != nil {
		return nil, err
	}

	return resp.Value, nil
}

// queryStore performs a query from a Tendermint node with the provided a store
// name and path.
func (ctx CLIContext) queryStore(key cmn.HexBytes, storeName, endPath string) ([]byte, error) {
	path := fmt.Sprintf("/store/%s/%s", storeName, endPath)
	return ctx.query(path, key)
}

// isQueryStoreWithProof expects a format like /<queryType>/<storeName>/<subpath>
// queryType must be "store" and subpath must be "key" to require a proof.
func isQueryStoreWithProof(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}

	paths := strings.SplitN(path[1:], "/", 3)
	switch {
	case len(paths) != 3:
		return false
	case paths[0] != "store":
		return false
	case store.RequireProof("/" + paths[2]):
		return true
	}

	return false
}
