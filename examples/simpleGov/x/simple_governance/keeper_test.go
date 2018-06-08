package simpleGovernance

import (

	// 	"os"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/examples/democoin/app"
	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/assert"
	db "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

// func setupMultiStore(name string) (sdk.MultiStore, *sdk.KVStoreKey) {
// 	db := dbm.NewDB(name, backend, dir)
// 	storeKey := sdk.NewKVStoreKey(name)
// 	ms := store.NewCommitMultiStore(db)
// 	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
// 	ms.LoadLatestVersion()
// 	return ms, storeKey
// }

func loggerAndDB() (log.Logger, db.DB) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	dB := db.NewMemDB()
	return logger, dB
}

func newDemocoinApp() *app.DemocoinApp {
	logger, dB := loggerAndDB()
	return app.NewDemocoinApp(logger, dB)
}

func TestSimpleGovKeeper(t *testing.T) {

	// create Proposals
	title := "Photons at launch"
	description := "Should we include Photons at launch?"
	addr1 := sdk.Address([]byte{1, 2})
	multiCoins := sdk.Coins{{"atom", 123}, {"eth", 20}}

	proposal1 := NewProposal(title, description, addr1, 0, 20, multiCoins)

	authKey := sdk.NewKVStoreKey("authKey")

	cdc := wire.NewCodec()
	app := newDemocoinApp()
	ctx := app.NewContext(true, abci.Header{})
	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.BaseAccount{})
	coinKeeper := bank.NewKeeper(accountMapper)

	stakeKey := sdk.NewKVStoreKey("stakeKey")

	stakeKeeper := stake.NewKeeper(cdc, stakeKey, coinKeeper, DefaultCodespace)

	proposalKey := sdk.NewKVStoreKey("proposalKey")
	// ms.MountStoreWithDB() // XXX why this ?

	// new proposal Keeper
	proposalKeeper := NewKeeper(proposalKey, coinKeeper, stakeKeeper, DefaultCodespace)
	assert.NotNil(t, proposalKeeper)

	err := proposalKeeper.SetProposal(ctx, 1, proposal1)
	resProposal, err := proposalKeeper.GetProposal(ctx, 1)
	assert.NotNil(t, resProposal)
	assert.NoError(t, err)
	assert.Nil(t, err)
	assert.Equal(t, proposal1, resProposal)

	// new poposal KeeperRead

	proposalKeeperRead := NewKeeperRead(proposalKey, coinKeeper, stakeKeeper, DefaultCodespace)
	assert.NotNil(t, proposalKeeperRead)

}
