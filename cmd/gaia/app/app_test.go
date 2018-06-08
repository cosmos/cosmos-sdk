package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/auth/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

var (
	chainID = "" // TODO

	priv1 = crypto.GenPrivKeyEd25519()
	addr1 = priv1.PubKey().Address()
	priv2 = crypto.GenPrivKeyEd25519()
	addr2 = priv2.PubKey().Address()
)

func loggerAndDB() (log.Logger, dbm.DB) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	return logger, db
}

func setGenesis(gapp *GaiaApp, accs ...*auth.BaseAccount) error {
	genaccs := make([]GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = NewGenesisAccount(acc)
	}

	genesisState := GenesisState{
		Accounts:  genaccs,
		StakeData: stake.DefaultGenesisState(),
	}

	stateBytes, err := wire.MarshalJSONIndent(gapp.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.Validator{}
	gapp.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: stateBytes})
	gapp.Commit()

	return nil
}

//_______________________________________________________________________

func TestGenesis(t *testing.T) {
	logger, dbs := loggerAndDB()
	gapp := NewGaiaApp(logger, dbs)

	// Construct some genesis bytes to reflect GaiaAccount
	pk := crypto.GenPrivKeyEd25519().PubKey()
	addr := pk.Address()
	coins, err := sdk.ParseCoins("77foocoin,99barcoin")
	require.Nil(t, err)
	baseAcc := &auth.BaseAccount{
		Address: addr,
		Coins:   coins,
	}

	err = setGenesis(gapp, baseAcc)
	require.Nil(t, err)

	// A checkTx context
	ctx := gapp.BaseApp.NewContext(true, abci.Header{})
	res1 := gapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, baseAcc, res1)

	// reload app and ensure the account is still there
	gapp = NewGaiaApp(logger, dbs)
	ctx = gapp.BaseApp.NewContext(true, abci.Header{})
	res1 = gapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, baseAcc, res1)
}

func TestExportValidators(t *testing.T) {
	logger, dbs := loggerAndDB()
	gapp := NewGaiaApp(logger, dbs)

	genCoins, err := sdk.ParseCoins("42steak")
	require.Nil(t, err)
	bondCoin, err := sdk.ParseCoin("10steak")
	require.Nil(t, err)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   genCoins,
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   genCoins,
	}

	err = setGenesis(gapp, acc1, acc2)
	require.Nil(t, err)

	// Create Validator
	description := stake.NewDescription("foo_moniker", "", "", "")
	createValidatorMsg := stake.NewMsgCreateValidator(
		addr1, priv1.PubKey(), bondCoin, description,
	)
	mock.SignCheckDeliver(t, gapp.BaseApp, createValidatorMsg, []int64{0}, true, priv1)
	gapp.Commit()

	// Export validator set
	_, validators, err := gapp.ExportAppStateAndValidators()
	require.Nil(t, err)
	require.Equal(t, 1, len(validators)) // 1 validator
	require.Equal(t, priv1.PubKey(), validators[0].PubKey)
	require.Equal(t, int64(10), validators[0].Power)
}
