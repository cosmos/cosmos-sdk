package app

import (
	"encoding/json"
	"reflect"
	"fmt"

	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
	abci "github.com/tendermint/abci/types"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

const (
	app4Name = "App4"
)

func NewApp4(logger log.Logger, db dbm.DB) *bapp.BaseApp {

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
	metadataMapper := NewApp4MetaDataMapper(keyMain)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFees)

	app.SetAnteHandler(auth.NewAnteHandler(accountMapper, feeKeeper))

	// Set InitChainer
	app.SetInitChainer(NewInitChainer(cdc, accountMapper, metadataMapper))

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("bank", NewApp4Handler(accountKeeper, metadataMapper))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyMain, keyFees)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

type GenesisState struct {
	Accounts  []*GenesisAccount  `json:"accounts"`
	Coins     []*GenesisCoin     `json:"coins"`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Address sdk.Address `json:"address"`
	Coins   sdk.Coins   `json:"coins"`
}

func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount, err error) {
	baseAcc := auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}
	return &baseAcc, nil
}

// GenesisCoin enforces CurrentSupply is 0 at genesis.
type GenesisCoin struct {
	Issuer sdk.Address `json:"issuer"`
	TotalSupply sdk.Int `json:"total_supply`
	Decimal     uint64   `json:"decimals"`
}

func (gc *GenesisCoin) ToMetaData() CoinMetadata {
	return CoinMetadata{
		Issuer: gc.Issuer,
		TotalSupply: gc.TotalSupply,
		Decimal: gc.Decimal,
	}
}

// InitChainer will set initial balances for accounts as well as initial coin metadata
// MsgIssue can no longer be used to create new coin
func NewInitChainer(cdc *wire.Codec, accountMapper auth.AccountMapper, metadataMapper MetaDataMapper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		stateJSON := req.AppStateBytes

		genesisState := new(GenesisState)
		err := cdc.UnmarshalJSON(stateJSON, genesisState)
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			// return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, gacc := range genesisState.Accounts {
			acc, err := gacc.ToAccount()
			if err != nil {
				panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
				//	return sdk.ErrGenesisParse("").TraceCause(err, "")
			}
			acc.AccountNumber = accountMapper.GetNextAccountNumber(ctx)
			accountMapper.SetAccount(ctx, acc)
		}

		return abci.ResponseInitChain{}

	}
}

//---------------------------------------------------------------------------------------------
// Now that initializing coin metadata is done in InitChainer we can simplifiy handleMsgIssue

func NewApp4Handler(accountKeeper bank.Keeper, metadataMapper MetaDataMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return betterHandleMsgSend(ctx, accountKeeper, msg)
		case MsgIssue:
			// use new MsgIssue handler
			return evenBetterHandleMsgIssue(ctx, metadataMapper, accountKeeper, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func evenBetterHandleMsgIssue(ctx sdk.Context, metadataMapper MetaDataMapper, accountKeeper bank.Keeper, msg MsgIssue) sdk.Result {
	for _, o := range msg.Outputs {
		for _, coin := range o.Coins {
			// Metadata is no longer created on the fly since it is initalized at genesis with InitChain
			metadata := metadataMapper.GetMetaData(ctx, coin.Denom)
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


//---------------------------------------------------------------------------------------------
// Simpler MetaDataMapper no longer able to initalize default CoinMetaData

type App4MetaDataMapper struct {
	mainKey *sdk.KVStoreKey
}

func NewApp4MetaDataMapper(key *sdk.KVStoreKey) App4MetaDataMapper {
	return App4MetaDataMapper{mainKey: key}
}

func (mdm App4MetaDataMapper) GetMetaData(ctx sdk.Context, denom string) CoinMetadata {
	store := ctx.KVStore(mdm.mainKey)

	bz := store.Get([]byte(denom))
	if bz == nil {
		// Coin metadata doesn't exist, create new metadata with default params
		return CoinMetadata{}
	}

	var metadata CoinMetadata
	err := json.Unmarshal(bz, &metadata)
	if err != nil {
		panic(err)
	}

	return metadata
}

func (mdm App4MetaDataMapper) SetMetaData(ctx sdk.Context, denom string, metadata CoinMetadata) {
	store := ctx.KVStore(mdm.mainKey)

	val, err := json.Marshal(metadata)
	if err != nil {
		panic(err)
	}

	store.Set([]byte(denom), val)
}

//------------------------------------------------------------------
// AccountMapper

//------------------------------------------------------------------
// CoinsKeeper
