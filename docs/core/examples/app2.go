package app

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	app2Name = "App2"
)

var (
	issuer = crypto.GenPrivKeyEd25519().PubKey().Address()
)

func NewCodec() *wire.Codec {
	cdc := wire.NewCodec()
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(MsgSend{}, "example/MsgSend", nil)
	cdc.RegisterConcrete(MsgIssue{}, "example/MsgIssue", nil)
	return cdc
}

func NewApp2(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	cdc := NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app2Name, cdc, logger, db)

	// Create a key for accessing the account store.
	keyMain := sdk.NewKVStoreKey("main")
	keyAccount := sdk.NewKVStoreKey("acc")

	// set antehandler function
	app.SetAnteHandler(antehandler)

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("send", handleMsgSend(keyAccount)).
		AddRoute("issue", handleMsgIssue(keyAccount, keyMain))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyMain)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// Coin Metadata
type CoinMetadata struct {
	TotalSupply   sdk.Int
	CurrentSupply sdk.Int
	Issuer        sdk.Address
	Decimal       uint64
}

//------------------------------------------------------------------
// Msgs

// Single permissioned issuer can issue Coin to Receiver
// if he is the issuer in Coin Metadata
// Implements sdk.Msg Interface
type MsgIssue struct {
	Issuer  sdk.Address
	Receiver sdk.Address
	Coin sdk.Coin
}

// nolint
func (msg MsgIssue) Type() string { return "issue" }

func (msg MsgIssue) ValidateBasic() sdk.Error {
	if len(msg.Issuer) == 0 {
		return sdk.ErrInvalidAddress("Issuer address cannot be empty")
	}

	if len(msg.Receiver) == 0 {
		return sdk.ErrInvalidAddress("Receiver address cannot be empty")
	}

	// Cannot issue zero or negative coins
	if !msg.Coin.IsPositive() {
		return sdk.ErrInvalidCoins("Cannot issue 0 or negative coin amounts")
	}

	return nil
}

func (msg MsgIssue) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// Implements Msg. Return the signer.
func (msg MsgIssue) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Issuer}
}

// Returns the sdk.Tags for the message
func (msg MsgIssue) Tags() sdk.Tags {
	return sdk.NewTags("issuer", []byte(msg.Issuer.String())).
		AppendTag("receiver", []byte(msg.Receiver.String()))
}

//------------------------------------------------------------------
// Handler for the message

// Handle Msg Issue
func handleMsgIssue(keyMain *sdk.KVStoreKey, keyAcc *sdk.KVStoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		issueMsg, ok := msg.(MsgIssue)
		if !ok {
			return sdk.NewError(2, 1, "IssueMsg is malformed").Result()
		}

		// Retrieve stores
		store := ctx.KVStore(keyMain)
		accStore := ctx.KVStore(keyAcc)

		// Handle updating metadata
		if res := handleMetaData(store, issueMsg.Issuer, issueMsg.Coin); !res.IsOK() {
			return res
		}

		// Issue coins to receiver using previously defined handleTo function
		if res := handleTo(accStore, issueMsg.Receiver, []sdk.Coin{issueMsg.Coin}); !res.IsOK() {
			return res
		}

		return sdk.Result{
			// Return result with Issue msg tags
			Tags: issueMsg.Tags(),
		}
	}
}

func handleMetaData(store sdk.KVStore, issuer sdk.Address, coin sdk.Coin) sdk.Result {
	bz := store.Get([]byte(coin.Denom))
	var metadata CoinMetadata

	if bz == nil {
		// Coin not set yet, initialize with issuer and default values
		// Coin amount can't be above default value
		if coin.Amount.GT(sdk.NewInt(1000000)) {
			return sdk.ErrInvalidCoins("Cannot issue that many new coins").Result()
		}
		metadata = CoinMetadata{
			TotalSupply:   sdk.NewInt(1000000),
			CurrentSupply: sdk.NewInt(0),
			Issuer:        issuer,
			Decimal:       10,
		}
	} else {
		// Decode coin metadata
		err := json.Unmarshal(bz, &metadata)
		if err != nil {
			return sdk.ErrInternal("Decoding coin metadata failed").Result()
		}
	}

	// Msg Issuer is not authorized to issue these coins
	if !reflect.DeepEqual(metadata.Issuer, issuer) {
		return sdk.ErrUnauthorized(fmt.Sprintf("Msg Issuer cannot issue tokens: %s", coin.Denom)).Result()
	}

	// Update coin current circulating supply
	metadata.CurrentSupply = metadata.CurrentSupply.Add(coin.Amount)

	// Current supply cannot exceed total supply
	if metadata.TotalSupply.LT(metadata.CurrentSupply) {
		return sdk.ErrInsufficientCoins("Issuer cannot issue more than total supply of coin").Result()
	}

	val, err := json.Marshal(metadata)
	if err != nil {
		return sdk.ErrInternal(fmt.Sprintf("Error encoding metadata: %s", err.Error())).Result()
	}

	// Update store with new metadata
	store.Set([]byte(coin.Denom), val)
	
	return sdk.Result{}
}


//------------------------------------------------------------------
// Tx

// Simple tx to wrap the Msg.
type app2Tx struct {
	sdk.Msg
	Signatures []auth.StdSignature
}

// This tx only has one Msg.
func (tx app2Tx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx.Msg}
}

func (tx app2Tx) GetSignatures() []auth.StdSignature {
	return tx.Signatures
}

//------------------------------------------------------------------

// Simple antehandler that ensures msg signers has signed over msg signBytes w/ no replay protection
// Implement sdk.AnteHandler interface
func antehandler(ctx sdk.Context, tx sdk.Tx) (_ sdk.Context, _ sdk.Result, abort bool) {
	appTx, ok := tx.(app2Tx)
	if !ok {
		// set abort boolean to true so that we don't continue to process failed tx
		return ctx, sdk.ErrTxDecode("Tx must be of format app2Tx").Result(), true
	}

	// expect only one msg in app2Tx
	msg := tx.GetMsgs()[0]

	signerAddrs := msg.GetSigners()
	signBytes := msg.GetSignBytes()

	if len(signerAddrs) != len(appTx.GetSignatures()) {
		return ctx, sdk.ErrUnauthorized("Number of signatures do not match required amount").Result(), true
	}

	for i, addr := range signerAddrs {
		sig := appTx.GetSignatures()[i]

		// check that submitted pubkey belongs to required address
		if !reflect.DeepEqual(sig.PubKey.Address(), addr) {
			return ctx, sdk.ErrUnauthorized("Provided Pubkey does not match required address").Result(), true
		}

		// check that signature is over expected signBytes
		if !sig.PubKey.VerifyBytes(signBytes, sig.Signature) {
			return ctx, sdk.ErrUnauthorized("Signature verification failed").Result(), true
		}
	}

	// authentication passed, app to continue processing by sending msg to handler
	return ctx, sdk.Result{}, false
}
