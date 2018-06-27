package app

import (
	"reflect"
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

	cdc := NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app3Name, cdc, logger, db)

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey("acc")
	keyMain := sdk.NewKVStoreKey("main")
	keyFees := sdk.NewKVStoreKey("fee")

	// Set various mappers/keepers to interact easily with underlying stores
	accountMapper := auth.NewAccountMapper(cdc, keyAccount, &auth.BaseAccount{})
	accountKeeper := bank.NewKeeper(accountMapper)
	metadataMapper := NewApp3MetaDataMapper(keyMain)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFees)

	app.SetAnteHandler(auth.NewAnteHandler(accountMapper, feeKeeper))

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("bank", NewApp3Handler(accountKeeper, metadataMapper))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyMain, keyFees)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

func NewApp3Handler(accountKeeper bank.Keeper, metadataMapper MetaDataMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return betterHandleMsgSend(ctx, accountKeeper, msg)
		case MsgIssue:
			return betterHandleMsgIssue(ctx, metadataMapper, accountKeeper, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func betterHandleMsgSend(ctx sdk.Context, accountKeeper bank.Keeper, msg MsgSend) sdk.Result {
	// Subtract coins from sender account
	_, _, err := accountKeeper.SubtractCoins(ctx, msg.From, msg.Amount)
	if err != nil {
		// if error, return its result
		return err.Result()
	}

	// Add coins to receiver account
	_, _, err = accountKeeper.AddCoins(ctx, msg.To, msg.Amount)
	if err != nil {
		// if error, return its result
		return err.Result()
	}

	return sdk.Result{}
}

func betterHandleMsgIssue(ctx sdk.Context, metadataMapper MetaDataMapper, accountKeeper bank.Keeper, msg MsgIssue) sdk.Result {
	for _, o := range msg.Outputs {
		for _, coin := range o.Coins {
			metadata := metadataMapper.GetMetaData(ctx, coin.Denom)
			if len(metadata.Issuer) == 0 {
				// coin doesn't have issuer yet, set issuer to msg issuer
				metadata.Issuer = msg.Issuer
			}
			
			// Check that msg Issuer is authorized to issue these coins
			if !reflect.DeepEqual(metadata.Issuer, msg.Issuer) {
				return sdk.ErrUnauthorized(fmt.Sprintf("Msg Issuer cannot issue these coins: %s", coin.Denom)).Result()
			}

			// Issuer cannot issue more than remaining supply
			issuerSupply := metadata.TotalSupply.Sub(metadata.CurrentSupply)
			if coin.Amount.GT(issuerSupply) {
				return sdk.ErrInsufficientCoins(fmt.Sprintf("Issuer cannot issue that many coins. Current issuer supply: %d", issuerSupply.Int64())).Result()
			}

			// update metadata current circulating supply
			metadata.CurrentSupply = metadata.CurrentSupply.Add(coin.Amount)

			metadataMapper.SetMetaData(ctx, coin.Denom, metadata)
		}

		// Add newly issued coins to output address
		_, _, err := accountKeeper.AddCoins(ctx, o.Address, o.Coins)
		if err != nil {
			return err.Result()
		}
	}

	return sdk.Result{}
}



//------------------------------------------------------------------
// Mapper for Coin Metadata
// Example of a very simple user-defined mapper

type MetaDataMapper interface {
	GetMetaData(sdk.Context, string) CoinMetadata
	SetMetaData(sdk.Context, string, CoinMetadata)
}

type App3MetaDataMapper struct {
	mainKey *sdk.KVStoreKey
}

func NewApp3MetaDataMapper(key *sdk.KVStoreKey) App3MetaDataMapper {
	return App3MetaDataMapper{mainKey: key}
}

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
