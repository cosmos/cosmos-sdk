package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

func setGenesis(bapp *BasecoinApp, accs ...auth.BaseAccount) error {
	genaccs := make([]*types.GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = types.NewGenesisAccount(&types.AppAccount{acc, "foobart"})
	}

	genesisState := types.GenesisState{
		Accounts:  genaccs,
		StakeData: stake.DefaultGenesisState(),
	}

	stateBytes, err := wire.MarshalJSONIndent(bapp.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: stateBytes})
	bapp.Commit()

	return nil
}

//_______________________________________________________________________

func TestGenesis(t *testing.T) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	bapp := NewBasecoinApp(logger, db)

	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	pk := crypto.GenPrivKeyEd25519().PubKey()
	addr := pk.Address()
	coins, err := sdk.ParseCoins("77foocoin,99barcoin")
	require.Nil(t, err)
	baseAcc := auth.BaseAccount{
		Address: addr,
		Coins:   coins,
	}
	acc := &types.AppAccount{baseAcc, "foobart"}

	err = setGenesis(bapp, baseAcc)
	require.Nil(t, err)

	// A checkTx context
	ctx := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, acc, res1)

	// reload app and ensure the account is still there
	bapp = NewBasecoinApp(logger, db)
	ctx = bapp.BaseApp.NewContext(true, abci.Header{})
	res1 = bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, acc, res1)
}
