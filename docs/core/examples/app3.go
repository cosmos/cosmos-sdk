package app

import (
	"bytes"
	"encoding/json"
	"fmt"

	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

const (
	app3Name = "App3"
)

func NewApp3(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	// Create the codec with registered Msg types
	cdc := NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app3Name, cdc, logger, db)

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey("acc")
	keyMain := sdk.NewKVStoreKey("main")
	keyFees := sdk.NewKVStoreKey("fee")

	// Set various mappers/keepers to interact easily with underlying stores
	// TODO: Need to register Account interface or use different Codec
	accountMapper := auth.NewAccountMapper(cdc, keyAccount, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)
	metadataMapper := NewApp3MetaDataMapper(keyMain)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFees)

	app.SetAnteHandler(auth.NewAnteHandler(accountMapper, feeKeeper))

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("send", betterHandleMsgSend(accountKeeper)).
		AddRoute("issue", betterHandleMsgIssue(metadataMapper, accountKeeper))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyMain, keyFees)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

func betterHandleMsgSend(accountKeeper bank.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		sendMsg, ok := msg.(MsgSend)
		if !ok {
			return sdk.NewError(2, 1, "Send Message Malformed").Result()
		}

		// Subtract coins from sender account
		_, _, err := accountKeeper.SubtractCoins(ctx, sendMsg.From, sendMsg.Amount)
		if err != nil {
			// if error, return its result
			return err.Result()
		}

		// Add coins to receiver account
		_, _, err = accountKeeper.AddCoins(ctx, sendMsg.To, sendMsg.Amount)
		if err != nil {
			// if error, return its result
			return err.Result()
		}

		return sdk.Result{
			Tags: sendMsg.Tags(),
		}
	}
}

func betterHandleMsgIssue(metadataMapper MetaDataMapper, accountKeeper bank.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		issueMsg, ok := msg.(MsgIssue)
		if !ok {
			return sdk.NewError(2, 1, "Issue Message Malformed").Result()
		}

		// Handle updating metadata
		if res := betterHandleMetaData(ctx, metadataMapper, issueMsg.Issuer, issueMsg.Coin); !res.IsOK() {
			return res
		}

		// Add newly issued coins to output address
		_, _, err := accountKeeper.AddCoins(ctx, issueMsg.Receiver, []sdk.Coin{issueMsg.Coin})
		if err != nil {
			return err.Result()
		}

		return sdk.Result{
			// Return result with Issue msg tags
			Tags: issueMsg.Tags(),
		}
	}
}

func betterHandleMetaData(ctx sdk.Context, metadataMapper MetaDataMapper, issuer sdk.Address, coin sdk.Coin) sdk.Result {
	metadata := metadataMapper.GetMetaData(ctx, coin.Denom)

	// Metadata was created fresh, should set issuer to msg issuer
	if len(metadata.Issuer) == 0 {
		metadata.Issuer = issuer
	}

	// Msg Issuer is not authorized to issue these coins
	if !bytes.Equal(metadata.Issuer, issuer) {
		return sdk.ErrUnauthorized(fmt.Sprintf("Msg Issuer cannot issue tokens: %s", coin.Denom)).Result()
	}

	// Update current circulating supply
	metadata.CurrentSupply = metadata.CurrentSupply.Add(coin.Amount)

	// Current supply cannot exceed total supply
	if metadata.TotalSupply.LT(metadata.CurrentSupply) {
		return sdk.ErrInsufficientCoins("Issuer cannot issue more than total supply of coin").Result()
	}

	metadataMapper.SetMetaData(ctx, coin.Denom, metadata)
	return sdk.Result{}
}

//------------------------------------------------------------------
// Mapper for Coin Metadata

// Example of a very simple user-defined mapper interface
type MetaDataMapper interface {
	GetMetaData(sdk.Context, string) CoinMetadata
	SetMetaData(sdk.Context, string, CoinMetadata)
}

// Implements MetaDataMapper
type App3MetaDataMapper struct {
	mainKey *sdk.KVStoreKey
}

// Construct new App3MetaDataMapper
func NewApp3MetaDataMapper(key *sdk.KVStoreKey) App3MetaDataMapper {
	return App3MetaDataMapper{mainKey: key}
}

// Implements MetaDataMpper. Returns metadata for coin
// If metadata does not exist in store, function creates default metadata and returns it
// without adding it to the store.
func (mdm App3MetaDataMapper) GetMetaData(ctx sdk.Context, denom string) CoinMetadata {
	store := ctx.KVStore(mdm.mainKey)

	bz := store.Get([]byte(denom))
	if bz == nil {
		// Coin metadata doesn't exist, create new metadata with default params
		return CoinMetadata{
			TotalSupply: sdk.NewInt(1000000),
		}
	}

	var metadata CoinMetadata
	err := json.Unmarshal(bz, &metadata)
	if err != nil {
		panic(err)
	}

	return metadata
}

// Implements MetaDataMapper. Sets metadata in store with key equal to denom.
func (mdm App3MetaDataMapper) SetMetaData(ctx sdk.Context, denom string, metadata CoinMetadata) {
	store := ctx.KVStore(mdm.mainKey)

	val, err := json.Marshal(metadata)
	if err != nil {
		panic(err)
	}

	store.Set([]byte(denom), val)
}

//------------------------------------------------------------------
// StdTx

//------------------------------------------------------------------
// Account

//------------------------------------------------------------------
// Ante Handler
