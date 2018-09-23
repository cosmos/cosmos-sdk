package context

import (
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/pkg/errors"

	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/lite"
	tmliteErr "github.com/tendermint/tendermint/lite/errors"
	tmliteProxy "github.com/tendermint/tendermint/lite/proxy"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GetNode returns an RPC client. If the context's client is not defined, an
// error is returned.
func (ctx CLIContext) GetNode() (rpcclient.Client, error) {
	if ctx.Client == nil {
		return nil, errors.New("no RPC client defined")
	}

	return ctx.Client, nil
}

// Query performs a query for information about the connected node.
func (ctx CLIContext) Query(path string, data cmn.HexBytes) (res []byte, err error) {
	return ctx.query(path, data)
}

// Query information about the connected node with a data payload
func (ctx CLIContext) QueryWithData(path string, data []byte) (res []byte, err error) {
	return ctx.query(path, data)
}

// QueryStore performs a query from a Tendermint node with the provided key and
// store name.
func (ctx CLIContext) QueryStore(key cmn.HexBytes, storeName string) (res []byte, err error) {
	return ctx.queryStore(key, storeName, "key")
}

// QuerySubspace performs a query from a Tendermint node with the provided
// store name and subspace.
func (ctx CLIContext) QuerySubspace(subspace []byte, storeName string) (res []sdk.KVPair, err error) {
	resRaw, err := ctx.queryStore(subspace, storeName, "subspace")
	if err != nil {
		return res, err
	}

	ctx.Codec.MustUnmarshalBinary(resRaw, &res)
	return
}

// GetAccount queries for an account given an address and a block height. An
// error is returned if the query or decoding fails.
func (ctx CLIContext) GetAccount(address []byte) (auth.Account, error) {
	if ctx.AccDecoder == nil {
		return nil, errors.New("account decoder required but not provided")
	}

	res, err := ctx.QueryStore(auth.AddressStoreKey(address), ctx.AccountStore)
	if err != nil {
		return nil, err
	} else if len(res) == 0 {
		return nil, err
	}

	account, err := ctx.AccDecoder(res)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetFromAddress returns the from address from the context's name.
func (ctx CLIContext) GetFromAddress() (from sdk.AccAddress, err error) {
	if ctx.FromAddressName == "" {
		return nil, errors.Errorf("must provide a from address name")
	}

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	info, err := keybase.Get(ctx.FromAddressName)
	if err != nil {
		return nil, errors.Errorf("no key for: %s", ctx.FromAddressName)
	}

	return sdk.AccAddress(info.GetPubKey().Address()), nil
}

// GetAccountNumber returns the next account number for the given account
// address.
func (ctx CLIContext) GetAccountNumber(address []byte) (int64, error) {
	account, err := ctx.GetAccount(address)
	if err != nil {
		return 0, err
	}

	return account.GetAccountNumber(), nil
}

// GetAccountSequence returns the sequence number for the given account
// address.
func (ctx CLIContext) GetAccountSequence(address []byte) (int64, error) {
	account, err := ctx.GetAccount(address)
	if err != nil {
		return 0, err
	}

	return account.GetSequence(), nil
}

// BroadcastTx broadcasts transaction bytes to a Tendermint node.
func (ctx CLIContext) BroadcastTx(tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return res, err
	}

	if !res.CheckTx.IsOK() {
		return res, errors.Errorf("checkTx failed: (%d) %s",
			res.CheckTx.Code,
			res.CheckTx.Log)
	}

	if !res.DeliverTx.IsOK() {
		return res, errors.Errorf("deliverTx failed: (%d) %s",
			res.DeliverTx.Code,
			res.DeliverTx.Log)
	}

	return res, err
}

// BroadcastTxAsync broadcasts transaction bytes to a Tendermint node
// asynchronously.
func (ctx CLIContext) BroadcastTxAsync(tx []byte) (*ctypes.ResultBroadcastTx, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxAsync(tx)
	if err != nil {
		return res, err
	}

	return res, err
}

// BroadcastTxSync broadcasts transaction bytes to a Tendermint node
// synchronously.
func (ctx CLIContext) BroadcastTxSync(tx []byte) (*ctypes.ResultBroadcastTx, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxSync(tx)
	if err != nil {
		return res, err
	}

	return res, err
}

// EnsureAccountExists ensures that an account exists for a given context. An
// error is returned if it does not.
func (ctx CLIContext) EnsureAccountExists() error {
	addr, err := ctx.GetFromAddress()
	if err != nil {
		return err
	}

	accountBytes, err := ctx.QueryStore(auth.AddressStoreKey(addr), ctx.AccountStore)
	if err != nil {
		return err
	}

	if len(accountBytes) == 0 {
		return ErrInvalidAccount(addr)
	}

	return nil
}

// EnsureAccountExistsFromAddr ensures that an account exists for a given
// address. Instead of using the context's from name, a direct address is
// given. An error is returned if it does not.
func (ctx CLIContext) EnsureAccountExistsFromAddr(addr sdk.AccAddress) error {
	accountBytes, err := ctx.QueryStore(auth.AddressStoreKey(addr), ctx.AccountStore)
	if err != nil {
		return err
	}

	if len(accountBytes) == 0 {
		return ErrInvalidAccount(addr)
	}

	return nil
}

// EnsureBroadcastTx broadcasts a transactions either synchronously or
// asynchronously based on the context parameters. The result of the broadcast
// is parsed into an intermediate structure which is logged if the context has
// a logger defined.
func (ctx CLIContext) EnsureBroadcastTx(txBytes []byte) error {
	if ctx.Async {
		return ctx.ensureBroadcastTxAsync(txBytes)
	}

	return ctx.ensureBroadcastTx(txBytes)
}

func (ctx CLIContext) ensureBroadcastTxAsync(txBytes []byte) error {
	res, err := ctx.BroadcastTxAsync(txBytes)
	if err != nil {
		return err
	}

	if ctx.JSON {
		type toJSON struct {
			TxHash string
		}

		if ctx.Logger != nil {
			resJSON := toJSON{res.Hash.String()}
			bz, err := ctx.Codec.MarshalJSON(resJSON)
			if err != nil {
				return err
			}

			ctx.Logger.Write(bz)
			io.WriteString(ctx.Logger, "\n")
		}
	} else {
		if ctx.Logger != nil {
			io.WriteString(ctx.Logger, fmt.Sprintf("Async tx sent (tx hash: %s)\n", res.Hash))
		}
	}

	return nil
}

func (ctx CLIContext) ensureBroadcastTx(txBytes []byte) error {
	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	if ctx.JSON {
		// since JSON is intended for automated scripts, always include
		// response in JSON mode.
		type toJSON struct {
			Height   int64
			TxHash   string
			Response abci.ResponseDeliverTx
		}

		if ctx.Logger != nil {
			resJSON := toJSON{res.Height, res.Hash.String(), res.DeliverTx}
			bz, err := ctx.Codec.MarshalJSON(resJSON)
			if err != nil {
				return err
			}

			ctx.Logger.Write(bz)
			io.WriteString(ctx.Logger, "\n")
		}

		return nil
	}

	if ctx.Logger != nil {
		resStr := fmt.Sprintf("Committed at block %d (tx hash: %s)\n", res.Height, res.Hash.String())

		if ctx.PrintResponse {
			resStr = fmt.Sprintf("Committed at block %d (tx hash: %s, response: %+v)\n",
				res.Height, res.Hash.String(), res.DeliverTx,
			)
		}

		io.WriteString(ctx.Logger, resStr)
	}

	return nil
}

// query performs a query from a Tendermint node with the provided store name
// and path.
func (ctx CLIContext) query(path string, key cmn.HexBytes) (res []byte, err error) {
	node, err := ctx.GetNode()
	if err != nil {
		return res, err
	}

	opts := rpcclient.ABCIQueryOptions{
		Height:  ctx.Height,
		Trusted: ctx.TrustNode,
	}

	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return res, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return res, errors.Errorf("query failed: (%d) %s", resp.Code, resp.Log)
	}

	// Data from trusted node or subspace query doesn't need verification.
	if ctx.TrustNode || !isQueryStoreWithProof(path) {
		return resp.Value, nil
	}

	err = ctx.verifyProof(path, resp)
	if err != nil {
		return nil, err
	}

	return resp.Value, nil
}

// Certify verifies the consensus proof at given height
func (ctx CLIContext) Certify(height int64) (lite.Commit, error) {
	check, err := tmliteProxy.GetCertifiedCommit(height, ctx.Client, ctx.Certifier)
	if tmliteErr.IsCommitNotFoundErr(err) {
		return lite.Commit{}, ErrVerifyCommit(height)
	} else if err != nil {
		return lite.Commit{}, err
	}
	return check, nil
}

// verifyProof perform response proof verification
// nolint: unparam
func (ctx CLIContext) verifyProof(path string, resp abci.ResponseQuery) error {

	if ctx.Certifier == nil {
		return fmt.Errorf("missing valid certifier to verify data from untrusted node")
	}

	// AppHash for height H is in header H+1
	commit, err := ctx.Certify(resp.Height + 1)
	if err != nil {
		return err
	}

	var multiStoreProof store.MultiStoreProof
	cdc := codec.New()
	err = cdc.UnmarshalBinary(resp.Proof, &multiStoreProof)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshalBinary rangeProof")
	}

	// Verify the substore commit hash against trusted appHash
	substoreCommitHash, err := store.VerifyMultiStoreCommitInfo(
		multiStoreProof.StoreName, multiStoreProof.StoreInfos, commit.Header.AppHash)
	if err != nil {
		return errors.Wrap(err, "failed in verifying the proof against appHash")
	}

	err = store.VerifyRangeProof(resp.Key, resp.Value, substoreCommitHash, &multiStoreProof.RangeProof)
	if err != nil {
		return errors.Wrap(err, "failed in the range proof verification")
	}

	return nil
}

// queryStore performs a query from a Tendermint node with the provided a store
// name and path.
func (ctx CLIContext) queryStore(key cmn.HexBytes, storeName, endPath string) ([]byte, error) {
	path := fmt.Sprintf("/store/%s/%s", storeName, endPath)
	return ctx.query(path, key)
}

// isQueryStoreWithProof expects a format like /<queryType>/<storeName>/<subpath>
// queryType can be app or store
func isQueryStoreWithProof(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}
	paths := strings.SplitN(path[1:], "/", 3)
	if len(paths) != 3 {
		return false
	}

	if store.RequireProof("/" + paths[2]) {
		return true
	}
	return false
}
