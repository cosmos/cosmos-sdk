package app

import (
	"os"
	"testing"

	"github.com/YunSuk-Yeo/cosmos-sdk/x/bank"
	"github.com/YunSuk-Yeo/cosmos-sdk/x/crisis"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/YunSuk-Yeo/cosmos-sdk/codec"
	"github.com/YunSuk-Yeo/cosmos-sdk/x/auth"
	distr "github.com/YunSuk-Yeo/cosmos-sdk/x/distribution"
	"github.com/YunSuk-Yeo/cosmos-sdk/x/gov"
	"github.com/YunSuk-Yeo/cosmos-sdk/x/mint"
	"github.com/YunSuk-Yeo/cosmos-sdk/x/slashing"
	"github.com/YunSuk-Yeo/cosmos-sdk/x/staking"

	abci "github.com/tendermint/tendermint/abci/types"
)

func setGenesis(gapp *GaiaApp, accs ...*auth.BaseAccount) error {
	genaccs := make([]GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = NewGenesisAccount(acc)
	}

	genesisState := NewGenesisState(
		genaccs,
		auth.DefaultGenesisState(),
		bank.DefaultGenesisState(),
		staking.DefaultGenesisState(),
		mint.DefaultGenesisState(),
		distr.DefaultGenesisState(),
		gov.DefaultGenesisState(),
		crisis.DefaultGenesisState(),
		slashing.DefaultGenesisState(),
	)

	stateBytes, err := codec.MarshalJSONIndent(gapp.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.ValidatorUpdate{}
	gapp.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: stateBytes})
	gapp.Commit()

	return nil
}

func TestGaiadExport(t *testing.T) {
	db := db.NewMemDB()
	gapp := NewGaiaApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	setGenesis(gapp)

	// Making a new app object with the db, so that initchain hasn't been called
	newGapp := NewGaiaApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	_, _, err := newGapp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
