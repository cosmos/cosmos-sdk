package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	crypto "github.com/tendermint/go-crypto"
)

// initCapKeys, initBaseApp, initStores, initHandlers.
func (app *BasecoinApp) initBaseApp() {
	bapp := baseapp.NewBaseApp(appName)
	app.BaseApp = bapp
	app.router = bapp.Router()
	app.initBaseAppTxDecoder()
	app.initBaseAppInitStater()
}

func (app *BasecoinApp) initBaseAppTxDecoder() {
	cdc := makeTxCodec()
	app.BaseApp.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = sdk.StdTx{}
		// StdTx.Msg is an interface whose concrete
		// types are registered in app/msgs.go.
		err := cdc.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxParse("").TraceCause(err, "")
		}
		return tx, nil
	})
}

// define the custom logic for basecoin initialization
func (app *BasecoinApp) initBaseAppInitStater() {
	accountMapper := app.accountMapper

	app.BaseApp.SetInitStater(func(ctx sdk.Context, state json.RawMessage) sdk.Error {
		if state == nil {
			return nil
		}

		genesisState := new(GenesisState)
		err := json.Unmarshal(state, genesisState)
		if err != nil {
			return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, gacc := range genesisState.Accounts {
			acc, err := gacc.toAppAccount()
			if err != nil {
				return sdk.ErrGenesisParse("").TraceCause(err, "")
			}
			accountMapper.SetAccount(ctx, acc)
		}
		return nil
	})
}

//-----------------------------------------------------

type GenesisState struct {
	Accounts []*GenesisAccount `accounts`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Name    string         `json:"name"`
	Address crypto.Address `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
}

func NewGenesisAccount(aa *types.AppAccount) *GenesisAccount {
	return &GenesisAccount{
		Name:    aa.Name,
		Address: aa.Address,
		Coins:   aa.Coins,
	}
}

// convert GenesisAccount to AppAccount
func (ga *GenesisAccount) toAppAccount() (acc *types.AppAccount, err error) {
	baseAcc := auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins,
	}
	return &types.AppAccount{
		BaseAccount: baseAcc,
		Name:        ga.Name,
	}, nil
}
