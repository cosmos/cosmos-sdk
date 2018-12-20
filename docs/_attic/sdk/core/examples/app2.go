package app

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	app2Name = "App2"
)

var (
	issuer = ed25519.GenPrivKey().PubKey().Address()
)

func NewCodec() *codec.Codec {
	cdc := codec.New()
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(MsgSend{}, "example/MsgSend", nil)
	cdc.RegisterConcrete(MsgIssue{}, "example/MsgIssue", nil)
	cryptoAmino.RegisterAmino(cdc)
	return cdc
}

func NewApp2(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	cdc := NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app2Name, logger, db, tx2Decoder(cdc))

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey(auth.StoreKey)
	// Create a key for accessing the issue store.
	keyIssue := sdk.NewKVStoreKey("issue")

	// set antehandler function
	app.SetAnteHandler(antehandler)

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("send", handleMsgSend(keyAccount)).
		AddRoute("issue", handleMsgIssue(keyAccount, keyIssue))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyIssue)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

//------------------------------------------------------------------
// Msgs

// MsgIssue to allow a registered issuer
// to issue new coins.
type MsgIssue struct {
	Issuer   sdk.AccAddress
	Receiver sdk.AccAddress
	Coin     sdk.Coin
}

// Implements Msg.
// nolint
func (msg MsgIssue) Route() string { return "issue" }
func (msg MsgIssue) Type() string  { return "issue" }

// Implements Msg. Ensures addresses are valid and Coin is positive
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

// Implements Msg. Get canonical sign bytes for MsgIssue
func (msg MsgIssue) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bz)
}

// Implements Msg. Return the signer.
func (msg MsgIssue) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Issuer}
}

// Returns the sdk.Tags for the message
func (msg MsgIssue) Tags() sdk.Tags {
	return sdk.NewTags("issuer", []byte(msg.Issuer.String())).
		AppendTag("receiver", []byte(msg.Receiver.String()))
}

//------------------------------------------------------------------
// Handler for the message

// Handle MsgIssue.
func handleMsgIssue(keyIssue *sdk.KVStoreKey, keyAcc *sdk.KVStoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		issueMsg, ok := msg.(MsgIssue)
		if !ok {
			return sdk.NewError(bankCodespace, 1, "MsgIssue is malformed").Result()
		}

		// Retrieve stores
		issueStore := ctx.KVStore(keyIssue)
		accStore := ctx.KVStore(keyAcc)

		// Handle updating coin info
		if res := handleIssuer(issueStore, issueMsg.Issuer, issueMsg.Coin); !res.IsOK() {
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

func handleIssuer(store sdk.KVStore, issuer sdk.AccAddress, coin sdk.Coin) sdk.Result {
	// the issuer address is stored directly under the coin denomination
	denom := []byte(coin.Denom)
	infoBytes := store.Get(denom)
	if infoBytes == nil {
		return sdk.ErrInvalidCoins(fmt.Sprintf("Unknown coin type %s", coin.Denom)).Result()
	}

	var coinInfo coinInfo
	err := json.Unmarshal(infoBytes, &coinInfo)
	if err != nil {
		return sdk.ErrInternal("Error when deserializing coinInfo").Result()
	}

	// Msg Issuer is not authorized to issue these coins
	if !bytes.Equal(coinInfo.Issuer, issuer) {
		return sdk.ErrUnauthorized(fmt.Sprintf("Msg Issuer cannot issue tokens: %s", coin.Denom)).Result()
	}

	return sdk.Result{}
}

// coinInfo stores meta data about a coin
type coinInfo struct {
	Issuer sdk.AccAddress `json:"issuer"`
}

//------------------------------------------------------------------
// Tx

// Simple tx to wrap the Msg.
type app2Tx struct {
	sdk.Msg

	PubKey    crypto.PubKey
	Signature []byte
}

// This tx only has one Msg.
func (tx app2Tx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx.Msg}
}

func (tx app2Tx) GetSignature() []byte {
	return tx.Signature
}

// Amino decode app2Tx. Capable of decoding both MsgSend and MsgIssue
func tx2Decoder(cdc *codec.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx app2Tx
		err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxDecode(err.Error())
		}
		return tx, nil
	}
}

//------------------------------------------------------------------

// Simple anteHandler that ensures msg signers have signed.
// Provides no replay protection.
func antehandler(ctx sdk.Context, tx sdk.Tx, simulate bool) (_ sdk.Context, _ sdk.Result, abort bool) {
	appTx, ok := tx.(app2Tx)
	if !ok {
		// set abort boolean to true so that we don't continue to process failed tx
		return ctx, sdk.ErrTxDecode("Tx must be of format app2Tx").Result(), true
	}

	// expect only one msg and one signer in app2Tx
	msg := tx.GetMsgs()[0]
	signerAddr := msg.GetSigners()[0]

	signBytes := msg.GetSignBytes()

	sig := appTx.GetSignature()

	// check that submitted pubkey belongs to required address
	if !bytes.Equal(appTx.PubKey.Address(), signerAddr) {
		return ctx, sdk.ErrUnauthorized("Provided Pubkey does not match required address").Result(), true
	}

	// check that signature is over expected signBytes
	if !appTx.PubKey.VerifyBytes(signBytes, sig) {
		return ctx, sdk.ErrUnauthorized("Signature verification failed").Result(), true
	}

	// authentication passed, app to continue processing by sending msg to handler
	return ctx, sdk.Result{}, false
}
