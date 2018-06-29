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
	coinKeeper := bank.NewKeeper(accountMapper)
	infoMapper := NewCoinInfoMapper(keyMain)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFees)

	app.SetAnteHandler(auth.NewAnteHandler(accountMapper, feeKeeper))

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("send", handleMsgSendWithKeeper(coinKeeper)).
		AddRoute("issue", handleMsgIssueWithInfoMapper(infoMapper, coinKeeper))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyMain, keyFees)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

func handleMsgSendWithKeeper(coinKeeper bank.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		sendMsg, ok := msg.(MsgSend)
		if !ok {
			return sdk.NewError(2, 1, "Send Message Malformed").Result()
		}

		// Subtract coins from sender account
		_, _, err := coinKeeper.SubtractCoins(ctx, sendMsg.From, sendMsg.Amount)
		if err != nil {
			// if error, return its result
			return err.Result()
		}

		// Add coins to receiver account
		_, _, err = coinKeeper.AddCoins(ctx, sendMsg.To, sendMsg.Amount)
		if err != nil {
			// if error, return its result
			return err.Result()
		}

		return sdk.Result{
			Tags: sendMsg.Tags(),
		}
	}
}

func handleMsgIssueWithInfoMapper(infoMapper coinInfoMapper, coinKeeper bank.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		issueMsg, ok := msg.(MsgIssue)
		if !ok {
			return sdk.NewError(2, 1, "Issue Message Malformed").Result()
		}

		// Handle updating metadata
		if res := handleCoinInfoWithMapper(ctx, infoMapper, issueMsg.Issuer, issueMsg.Coin); !res.IsOK() {
			return res
		}

		// Add newly issued coins to output address
		_, _, err := coinKeeper.AddCoins(ctx, issueMsg.Receiver, []sdk.Coin{issueMsg.Coin})
		if err != nil {
			return err.Result()
		}

		return sdk.Result{
			Tags: issueMsg.Tags(),
		}
	}
}

func handleCoinInfoWithMapper(ctx sdk.Context, infoMapper CoinInfoMapper, issuer sdk.Address, coin sdk.Coin) sdk.Result {
	coinInfo := infoMapper.GetInfo(ctx, coin.Denom)

	// Metadata was created fresh, should set issuer to msg issuer
	if len(coinInfo.Issuer) == 0 {
		coinInfo.Issuer = issuer
	}

	// Msg Issuer is not authorized to issue these coins
	if !bytes.Equal(coinInfo.Issuer, issuer) {
		return sdk.ErrUnauthorized(fmt.Sprintf("Msg Issuer cannot issue tokens: %s", coin.Denom)).Result()
	}

	return sdk.Result{}
}

//------------------------------------------------------------------
// Mapper for CoinInfo

// Example of a very simple user-defined read-only mapper interface.
type CoinInfoMapper interface {
	GetInfo(sdk.Context, string) coinInfo
}

// Implements CoinInfoMapper.
type coinInfoMapper struct {
	key *sdk.KVStoreKey
}

// Construct new CoinInfoMapper.
func NewCoinInfoMapper(key *sdk.KVStoreKey) coinInfoMapper {
	return coinInfoMapper{key: key}
}

// Implements CoinInfoMapper. Returns info for coin.
func (cim coinInfoMapper) GetInfo(ctx sdk.Context, denom string) coinInfo {
	store := ctx.KVStore(cim.key)

	infoBytes := store.Get([]byte(denom))
	if infoBytes == nil {
		// TODO
	}

	var coinInfo coinInfo
	err := json.Unmarshal(infoBytes, &coinInfo)
	if err != nil {
		panic(err)
	}

	return coinInfo
}
