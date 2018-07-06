package app

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/examples/democoin/types"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/cool"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

func setGenesis(bapp *DemocoinApp, trend string, accs ...auth.BaseAccount) error {
	genaccs := make([]*types.GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = types.NewGenesisAccount(&types.AppAccount{acc, "foobart"})
	}

	genesisState := types.GenesisState{
		Accounts:    genaccs,
		CoolGenesis: cool.Genesis{trend},
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

func TestGenesis(t *testing.T) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	config := cfg.DefaultConfig()
	sctx := sdk.NewServerContext(config, logger)
	db := dbm.NewMemDB()
	bapp := NewDemocoinApp(sctx, db)

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

	err = setGenesis(bapp, "ice-cold", baseAcc)
	require.Nil(t, err)
	// A checkTx context
	ctx := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	require.Equal(t, acc, res1)

	// reload app and ensure the account is still there
	bapp = NewDemocoinApp(sctx, db)
	bapp.InitChain(abci.RequestInitChain{AppStateBytes: []byte("{}")})
	ctx = bapp.BaseApp.NewContext(true, abci.Header{})
	res1 = bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	require.Equal(t, acc, res1)
}
