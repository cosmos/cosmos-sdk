package app

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/examples/democoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

func setGenesis(sgApp *SimpleGovApp, nAccounts ...auth.BaseAccount) error {
	genesisAccounts := make([]*types.GenesisAccount, len(accs))
	for i, account := range nAccounts {
		genesisAccounts[i] = types.NewGenesisAccount(&types.AppAccount{account, "foobart"})
	}

	genesisState := types.GenesisState{
		Accounts: genesisAccounts,
	}

	stateBytes, err := wire.MarshalJSONIndent(sgApp.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.Validator{}
	sgApp.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: stateBytes})
	sgApp.Commit()

	return nil
}

func TestGenesis(t *testing.T) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	sgApp := NewSimpleGovApp(logger, db)

	// Construct some genesis bytes to reflect democoin/types/AppAccount
	pk := crypto.GenPrivKeyEd25519().PubKey()
	addr := pk.Address()
	coins, err := sdk.ParseCoins("77foocoin,99barcoin")
	require.Nil(t, err)
	baseAcc := auth.BaseAccount{
		Address: addr,
		Coins:   coins,
	}
	acc := &types.AppAccount{baseAcc, "foobart"}

	err = setGenesis(sgApp, baseAcc)
	require.Nil(t, err)
	// A checkTx context
	ctx := sgApp.BaseApp.NewContext(true, abci.Header{})
	res1 := sgApp.accountMapper.GetAccount(ctx, baseAcc.Address)
	require.Equal(t, acc, res1)

	// reload app and ensure the account is still there
	sgApp = NewSimpleGovApp(logger, db)
	sgApp.InitChain(abci.RequestInitChain{AppStateBytes: []byte("{}")})
	ctx = sgApp.BaseApp.NewContext(true, abci.Header{})
	res1 = sgApp.accountMapper.GetAccount(ctx, baseAcc.Address)
	require.Equal(t, acc, res1)
}
