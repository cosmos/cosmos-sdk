package app

import (
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

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

	// Set various mappers/keepers to interact easily with underlying stores
	accountMapper := auth.NewAccountMapper(cdc, keyAccount, &auth.BaseAccount{})
	coinKeeper := bank.NewKeeper(accountMapper)

	// TODO
	keyFees := sdk.NewKVStoreKey("fee")
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFees)

	app.SetAnteHandler(auth.NewAnteHandler(accountMapper, feeKeeper))

	// Set InitChainer
	app.SetInitChainer(NewInitChainer(cdc, accountMapper))

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("send", bank.NewHandler(coinKeeper))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyFees)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// Application state at Genesis has accounts with starting balances
type GenesisState struct {
	Accounts []*GenesisAccount `json:"accounts"`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Address sdk.AccAddress `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
}

// Converts GenesisAccount to auth.BaseAccount for storage in account store
func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount, err error) {
	baseAcc := auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}
	return &baseAcc, nil
}

// InitChainer will set initial balances for accounts as well as initial coin metadata
// MsgIssue can no longer be used to create new coin
func NewInitChainer(cdc *wire.Codec, accountMapper auth.AccountMapper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		stateJSON := req.AppStateBytes

		genesisState := new(GenesisState)
		err := cdc.UnmarshalJSON(stateJSON, genesisState)
		if err != nil {
			panic(err)
		}

		for _, gacc := range genesisState.Accounts {
			acc, err := gacc.ToAccount()
			if err != nil {
				panic(err)
			}
			acc.AccountNumber = accountMapper.GetNextAccountNumber(ctx)
			accountMapper.SetAccount(ctx, acc)
		}

		return abci.ResponseInitChain{}
	}
}
