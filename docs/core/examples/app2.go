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
		AddRoute("bank", NewApp2Handler(keyAccount, keyMain))

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

// Create Output struct to allow single message to issue arbitrary coins to multiple users
type Output struct {
	Address sdk.Address
	Coins   sdk.Coins
}

// Single permissioned issuer can issue multiple outputs
// Implements sdk.Msg Interface
type MsgIssue struct {
	Issuer  sdk.Address
	Outputs []Output
}

// nolint
func (msg MsgIssue) Type() string { return "bank" }

func (msg MsgIssue) ValidateBasic() sdk.Error {
	if len(msg.Issuer) == 0 {
		return sdk.ErrInvalidAddress("Issuer address cannot be empty")
	}

	for _, o := range msg.Outputs {
		if len(o.Address) == 0 {
			return sdk.ErrInvalidAddress("Output address cannot be empty")
		}
		// Cannot issue zero or negative coins
		if !o.Coins.IsPositive() {
			return sdk.ErrInvalidCoins("Cannot issue 0 or negative coin amounts")
		}
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

//------------------------------------------------------------------
// Handler for the message

func NewApp2Handler(keyAcc *sdk.KVStoreKey, keyMain *sdk.KVStoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return handleMsgSend(ctx, keyAcc, msg)
		case MsgIssue:
			return handleMsgIssue(ctx, keyMain, keyAcc, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle Msg Issue
func handleMsgIssue(ctx sdk.Context, keyMain *sdk.KVStoreKey, keyAcc *sdk.KVStoreKey, msg MsgIssue) sdk.Result {
	store := ctx.KVStore(keyMain)
	accStore := ctx.KVStore(keyAcc)

	for _, o := range msg.Outputs {
		for _, coin := range o.Coins {
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
					Issuer:        msg.Issuer,
					Decimal:       10,
				}
			} else {
				// Decode coin metadata
				err := json.Unmarshal(bz, &metadata)
				if err != nil {
					return sdk.ErrInternal("Decoding coin metadata failed").Result()
				}
			}

			// Return error result if msg Issuer is not equal to coin issuer
			if !reflect.DeepEqual(metadata.Issuer, msg.Issuer) {
				return sdk.ErrUnauthorized(fmt.Sprintf("Msg issuer cannot issue these coins: %s", coin.Denom)).Result()
			}

			// Issuer cannot issue more than remaining supply
			issuerSupply := metadata.TotalSupply.Sub(metadata.CurrentSupply)
			if coin.Amount.GT(issuerSupply) {
				return sdk.ErrInsufficientCoins(fmt.Sprintf("Issuer cannot issue that many coins. Current issuer supply: %d", issuerSupply.Int64())).Result()
			}

			// Update coin metadata
			metadata.CurrentSupply = metadata.CurrentSupply.Add(coin.Amount)

			val, err := json.Marshal(metadata)
			if err != nil {
				return sdk.ErrInternal("Encoding coin metadata failed").Result()
			}

			// Update coin metadata in store
			store.Set([]byte(coin.Denom), val)
		}

		// Add coins to receiver account
		bz := accStore.Get(o.Address)
		var acc appAccount
		if bz == nil {
			// Receiver account does not already exist, create a new one.
			acc = appAccount{}
		} else {
			// Receiver account already exists. Retrieve and decode it.
			err := json.Unmarshal(bz, &acc)
			if err != nil {
				return sdk.ErrInternal("Account decoding error").Result()
			}
		}

		// Add amount to receiver's old coins
		receiverCoins := acc.Coins.Plus(o.Coins)

		// Update receiver account
		acc.Coins = receiverCoins

		// Encode receiver account
		val, err := json.Marshal(acc)
		if err != nil {
			return sdk.ErrInternal("Account encoding error").Result()
		}

		// set account with new issued coins in store
		store.Set(o.Address, val)
	}

	return sdk.Result{
	// TODO: Tags
	}

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
