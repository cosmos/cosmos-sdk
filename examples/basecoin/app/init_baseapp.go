package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	crypto "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
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

// We use GenesisAccount instead of types.AppAccount for cleaner json input of PubKey
type GenesisAccount struct {
	Name     string         `json:"name"`
	Address  crypto.Address `json:"address"`
	Coins    sdk.Coins      `json:"coins"`
	PubKey   cmn.HexBytes   `json:"public_key"`
	Sequence int64          `json:"sequence"`
}

func NewGenesisAccount(aa types.AppAccount) *GenesisAccount {
	return &GenesisAccount{
		Name:     aa.Name,
		Address:  aa.Address,
		Coins:    aa.Coins,
		PubKey:   aa.PubKey.Bytes(),
		Sequence: aa.Sequence,
	}
}

// convert GenesisAccount to AppAccount
func (ga *GenesisAccount) toAppAccount() (acc types.AppAccount, err error) {

	pk, err := crypto.PubKeyFromBytes(ga.PubKey)
	if err != nil {
		return
	}
	baseAcc := auth.BaseAccount{
		Address:  ga.Address,
		Coins:    ga.Coins,
		PubKey:   pk,
		Sequence: ga.Sequence,
	}
	return types.AppAccount{
		BaseAccount: baseAcc,
		Name:        "foobart",
	}, nil
}

// define the custom logic for basecoin initialization
func (app *BasecoinApp) initBaseAppInitStater() {
	accountMapper := app.accountMapper

	app.BaseApp.SetInitStater(func(ctxCheckTx, ctxDeliverTx sdk.Context, state json.RawMessage) sdk.Error {
		if state == nil {
			return nil
		}

		var gaccs []*GenesisAccount

		err := json.Unmarshal(state, &gaccs)
		if err != nil {
			return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, gacc := range gaccs {
			acc, err := gacc.toAppAccount()
			if err != nil {
				return sdk.ErrGenesisParse("").TraceCause(err, "")
			}

			//panic(fmt.Sprintf("debug acc: %s\n", acc))
			accountMapper.SetAccount(ctxCheckTx, &acc.BaseAccount)
			accountMapper.SetAccount(ctxDeliverTx, &acc.BaseAccount)
		}
		return nil
	})
}
