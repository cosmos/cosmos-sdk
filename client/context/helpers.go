package context

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/wire"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/lite"
	liteclient "github.com/tendermint/tendermint/lite/client"
	"github.com/tendermint/tendermint/lite/files"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/merkle"
	"github.com/cosmos/cosmos-sdk/store"
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
		return res, errors.Errorf("CheckTx failed: (%d) %s",
			res.CheckTx.Code,
			res.CheckTx.Log)
	}
	if res.DeliverTx.Code != uint32(0) {
		return res, errors.Errorf("DeliverTx failed: (%d) %s",
			res.DeliverTx.Code,
			res.DeliverTx.Log)
	}
	return res, err
}

// Query from Tendermint with the provided key and storename
func (ctx CoreContext) Query(key cmn.HexBytes, storeName string) (res []byte, err error) {
	resp, err := ctx.query(key, storeName, "key")
	if err != nil {
		return
	}
	if ctx.TrustNode {
		return
	}

	proof, err := merkle.DecodeProof(resp.Proof, store.WrapOpDecoder(merkle.DefaultOpDecoder))
	if err != nil {
		return
	}

	source := liteclient.NewHTTPProvider(ctx.NodeURI)
	trusted := lite.NewCacheProvider(lite.NewMemStoreProvider(), files.NewProvider(ctx.ProviderPath))

	genesis, err := tmtypes.GenesisDocFromFile(ctx.GenesisFile)
	if err != nil {
		return
	}

	vals := make([]*tmtypes.Validator, len(genesis.Validators))
	for i, val := range genesis.Validators {
		vals[i] = tmtypes.NewValidator(val.PubKey, val.Power)
	}
	valset := tmtypes.NewValidatorSet(vals)

	genfc := lite.FullCommit{
		Commit: lite.Commit{
			Header: &tmtypes.Header{
				Height:         0,
				ValidatorsHash: valset.Hash(),
			},
		},
		Validators: valset,
	}

	fmt.Printf("\n\n\n\n\n\n\n\n\n%+v\n", genfc)
	cert, err := lite.NewInquiringCertifier(ctx.ChainID, genfc, trusted, source)
	if err != nil {
		fmt.Printf("\n\n\n\n\n\n\n\n\n\n\n\ngenfc: %+v\nerr: %+v\n", genfc, err)
		return
	}

	fc, err := source.GetByHeight(resp.Height)
	if err != nil {
		return
	}

	err = cert.Certify(fc.Commit)
	if err != nil {
		return
	}

	fmt.Printf("\n\n\n\n\n\n\n\nproof: %+v\n", proof)
	err = proof.Verify(fc.Header.AppHash, [][]byte{res}, string(key), storeName)
	return
}

// Query from Tendermint with the provided storename and subspace
func (ctx CoreContext) QuerySubspace(cdc *wire.Codec, subspace []byte, storeName string) (res []sdk.KVPair, err error) {
	resp, err := ctx.query(subspace, storeName, "subspace")
	if err != nil {
		return res, err
	}
	cdc.MustUnmarshalBinary(resp.Value, &res)
	return
}

// Query from Tendermint with the provided storename and path
func (ctx CoreContext) query(key cmn.HexBytes, storeName, endPath string) (res abci.ResponseQuery, err error) {
	path := fmt.Sprintf("/store/%s/key", storeName)
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
	res = result.Response
	if res.Code != uint32(0) {
		return res, errors.Errorf("Query failed: (%d) %s", res.Code, res.Log)
	}
	return res, nil
}

// Get the from address from the name flag
func (ctx CoreContext) GetFromAddress() (from sdk.Address, err error) {

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
		return nil, errors.Errorf("No key for: %s", name)
	}

	return info.PubKey.Address(), nil
}

// sign and build the transaction from the msg
func (ctx CoreContext) SignAndBuild(name, passphrase string, msg sdk.Msg, cdc *wire.Codec) ([]byte, error) {

	// build the Sign Messsage from the Standard Message
	chainID := ctx.ChainID
	if chainID == "" {
		return nil, errors.Errorf("Chain ID required but not specified")
	}
	sequence := ctx.Sequence
	signMsg := sdk.StdSignMsg{
		ChainID:   chainID,
		Sequences: []int64{sequence},
		Msg:       msg,
		Fee:       sdk.NewStdFee(10000, sdk.Coin{}), // TODO run simulate to estimate gas?
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
	sigs := []sdk.StdSignature{{
		PubKey:    pubkey,
		Signature: sig,
		Sequence:  sequence,
	}}

	// marshal bytes
	tx := sdk.NewStdTx(signMsg.Msg, signMsg.Fee, sigs)

	return cdc.MarshalBinary(tx)
}

// sign and build the transaction from the msg
func (ctx CoreContext) EnsureSignBuildBroadcast(name string, msg sdk.Msg, cdc *wire.Codec) (res *ctypes.ResultBroadcastTxCommit, err error) {

	// default to next sequence number if none provided
	ctx, err = EnsureSequence(ctx)
	if err != nil {
		return nil, err
	}

	passphrase, err := ctx.GetPassphraseFromStdin(name)
	if err != nil {
		return nil, err
	}

	txBytes, err := ctx.SignAndBuild(name, passphrase, msg, cdc)
	if err != nil {
		return nil, err
	}

	return ctx.BroadcastTx(txBytes)
}

// get the next sequence for the account address
func (ctx CoreContext) NextSequence(address []byte) (int64, error) {
	if ctx.Decoder == nil {
		return 0, errors.New("AccountDecoder required but not provided")
	}

	res, err := ctx.Query(address, ctx.AccountStore)
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
		return nil, errors.New("Must define node URI")
	}
	return ctx.Client, nil
}
