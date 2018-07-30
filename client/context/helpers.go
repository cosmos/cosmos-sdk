package context

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/common"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	cmn "github.com/tendermint/tendermint/libs/common"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Broadcast the transaction bytes to Tendermint
func (ctx CoreContext) BroadcastTx(tx []byte) (*ctypes.ResultBroadcastTxCommit, error) {

	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return res, err
	}

	if res.CheckTx.Code != uint32(0) {
		return res, errors.Errorf("checkTx failed: (%d) %s",
			res.CheckTx.Code,
			res.CheckTx.Log)
	}
	if res.DeliverTx.Code != uint32(0) {
		return res, errors.Errorf("deliverTx failed: (%d) %s",
			res.DeliverTx.Code,
			res.DeliverTx.Log)
	}
	return res, err
}

// Broadcast the transaction bytes to Tendermint
func (ctx CoreContext) BroadcastTxAsync(tx []byte) (*ctypes.ResultBroadcastTx, error) {

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

// Query information about the connected node
func (ctx CoreContext) Query(path string) (res []byte, err error) {
	return ctx.query(path, nil)
}

// QueryStore from Tendermint with the provided key and storename
func (ctx CoreContext) QueryStore(key cmn.HexBytes, storeName string) (res []byte, err error) {
	return ctx.queryStore(key, storeName, "key")
}

// Query from Tendermint with the provided storename and subspace
func (ctx CoreContext) QuerySubspace(cdc *wire.Codec, subspace []byte, storeName string) (res []sdk.KVPair, err error) {
	resRaw, err := ctx.queryStore(subspace, storeName, "subspace")
	if err != nil {
		return res, err
	}
	cdc.MustUnmarshalBinary(resRaw, &res)
	return
}

// Query from Tendermint with the provided storename and path
func (ctx CoreContext) query(path string, key common.HexBytes) (res []byte, err error) {
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
	if resp.Code != uint32(0) {
		return res, errors.Errorf("query failed: (%d) %s", resp.Code, resp.Log)
	}
	return resp.Value, nil
}

// Query from Tendermint with the provided storename and path
func (ctx CoreContext) queryStore(key cmn.HexBytes, storeName, endPath string) (res []byte, err error) {
	path := fmt.Sprintf("/store/%s/%s", storeName, endPath)
	return ctx.query(path, key)
}

// Get the from address from the name flag
func (ctx CoreContext) GetFromAddress() (from sdk.AccAddress, err error) {

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	name := ctx.FromAddressName
	if name == "" {
		return nil, errors.Errorf("must provide a from address name")
	}

	info, err := keybase.Get(name)
	if err != nil {
		return nil, errors.Errorf("no key for: %s", name)
	}

	return sdk.AccAddress(info.GetPubKey().Address()), nil
}

// sign and build the transaction from the msg
func (ctx CoreContext) SignAndBuild(name, passphrase string, msgs []sdk.Msg, cdc *wire.Codec) ([]byte, error) {

	// build the Sign Messsage from the Standard Message
	chainID := ctx.ChainID
	if chainID == "" {
		return nil, errors.Errorf("chain ID required but not specified")
	}
	accnum := ctx.AccountNumber
	sequence := ctx.Sequence
	memo := ctx.Memo

	fee := sdk.Coin{}
	if ctx.Fee != "" {
		parsedFee, err := sdk.ParseCoin(ctx.Fee)
		if err != nil {
			return nil, err
		}
		fee = parsedFee
	}

	signMsg := auth.StdSignMsg{
		ChainID:       chainID,
		AccountNumber: accnum,
		Sequence:      sequence,
		Msgs:          msgs,
		Memo:          memo,
		Fee:           auth.NewStdFee(ctx.Gas, fee), // TODO run simulate to estimate gas?
	}

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	// sign and build
	bz := signMsg.Bytes()

	sig, pubkey, err := keybase.Sign(name, passphrase, bz)
	if err != nil {
		return nil, err
	}
	sigs := []auth.StdSignature{{
		PubKey:        pubkey,
		Signature:     sig,
		AccountNumber: accnum,
		Sequence:      sequence,
	}}

	// marshal bytes
	tx := auth.NewStdTx(signMsg.Msgs, signMsg.Fee, sigs, memo)

	return cdc.MarshalBinary(tx)
}

// sign and build the transaction from the msg
func (ctx CoreContext) ensureSignBuild(name string, msgs []sdk.Msg, cdc *wire.Codec) (tyBytes []byte, err error) {
	err = EnsureAccountExists(ctx, name)
	if err != nil {
		return nil, err
	}

	ctx, err = EnsureAccountNumber(ctx)
	if err != nil {
		return nil, err
	}
	// default to next sequence number if none provided
	ctx, err = EnsureSequence(ctx)
	if err != nil {
		return nil, err
	}

	var txBytes []byte

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	info, err := keybase.Get(name)
	if err != nil {
		return nil, err
	}
	var passphrase string
	// Only need a passphrase for locally-stored keys
	if info.GetType() == "local" {
		passphrase, err = ctx.GetPassphraseFromStdin(name)
		if err != nil {
			return nil, fmt.Errorf("Error fetching passphrase: %v", err)
		}
	}
	txBytes, err = ctx.SignAndBuild(name, passphrase, msgs, cdc)
	if err != nil {
		return nil, fmt.Errorf("Error signing transaction: %v", err)
	}

	return txBytes, err
}

// sign and build the transaction from the msg
func (ctx CoreContext) EnsureSignBuildBroadcast(name string, msgs []sdk.Msg, cdc *wire.Codec) (err error) {

	txBytes, err := ctx.ensureSignBuild(name, msgs, cdc)
	if err != nil {
		return err
	}

	if ctx.Async {
		res, err := ctx.BroadcastTxAsync(txBytes)
		if err != nil {
			return err
		}
		if ctx.JSON {
			type toJSON struct {
				TxHash string
			}
			valueToJSON := toJSON{res.Hash.String()}
			JSON, err := cdc.MarshalJSON(valueToJSON)
			if err != nil {
				return err
			}
			fmt.Println(string(JSON))
		} else {
			fmt.Println("Async tx sent. tx hash: ", res.Hash.String())
		}
		return nil
	}
	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}
	if ctx.JSON {
		// Since JSON is intended for automated scripts, always include response in JSON mode
		type toJSON struct {
			Height   int64
			TxHash   string
			Response string
		}
		valueToJSON := toJSON{res.Height, res.Hash.String(), fmt.Sprintf("%+v", res.DeliverTx)}
		JSON, err := cdc.MarshalJSON(valueToJSON)
		if err != nil {
			return err
		}
		fmt.Println(string(JSON))
		return nil
	}
	if ctx.PrintResponse {
		fmt.Printf("Committed at block %d. Hash: %s Response:%+v \n", res.Height, res.Hash.String(), res.DeliverTx)
	} else {
		fmt.Printf("Committed at block %d. Hash: %s \n", res.Height, res.Hash.String())
	}
	return nil
}

// get the next sequence for the account address
func (ctx CoreContext) GetAccountNumber(address []byte) (int64, error) {
	if ctx.Decoder == nil {
		return 0, errors.New("accountDecoder required but not provided")
	}

	res, err := ctx.QueryStore(auth.AddressStoreKey(address), ctx.AccountStore)
	if err != nil {
		return 0, err
	}

	if len(res) == 0 {
		fmt.Printf("No account found.  Returning 0.\n")
		return 0, err
	}

	account, err := ctx.Decoder(res)
	if err != nil {
		panic(err)
	}

	return account.GetAccountNumber(), nil
}

// get the next sequence for the account address
func (ctx CoreContext) NextSequence(address []byte) (int64, error) {
	if ctx.Decoder == nil {
		return 0, errors.New("accountDecoder required but not provided")
	}

	res, err := ctx.QueryStore(auth.AddressStoreKey(address), ctx.AccountStore)
	if err != nil {
		return 0, err
	}

	if len(res) == 0 {
		fmt.Printf("No account found, defaulting to sequence 0\n")
		return 0, err
	}

	account, err := ctx.Decoder(res)
	if err != nil {
		panic(err)
	}

	return account.GetSequence(), nil
}

// get passphrase from std input
func (ctx CoreContext) GetPassphraseFromStdin(name string) (pass string, err error) {
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)
	return client.GetPassword(prompt, buf)
}

// GetNode prepares a simple rpc.Client
func (ctx CoreContext) GetNode() (rpcclient.Client, error) {
	if ctx.Client == nil {
		return nil, errors.New("must define node URI")
	}
	return ctx.Client, nil
}
