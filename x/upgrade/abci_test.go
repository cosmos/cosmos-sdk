package upgrade_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type TestSuite struct {
	module  module.AppModule
	keeper  keeper.Keeper
	querier sdk.Querier
	handler govtypes.Handler
	ctx     sdk.Context
}

var s TestSuite

func setupTest(height int64, skip map[int64]bool) TestSuite {
	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, skip, simapp.DefaultNodeHome, 0, simapp.MakeTestEncodingConfig(), simapp.EmptyAppOptions{})
	genesisState := simapp.NewDefaultGenesisState(app.AppCodec())
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	if err != nil {
		panic(err)
	}
	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)

	s.keeper = app.UpgradeKeeper
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{Height: height, Time: time.Now()})

	s.module = upgrade.NewAppModule(s.keeper)
	s.querier = s.module.LegacyQuerierHandler(app.LegacyAmino())
	s.handler = upgrade.NewSoftwareUpgradeProposalHandler(s.keeper)
	return s
}

func TestRequireName(t *testing.T) {
	s := setupTest(10, map[int64]bool{})

	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{}})
	require.NotNil(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestRequireFutureTime(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: s.ctx.BlockHeader().Time}})
	require.NotNil(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestRequireFutureBlock(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight()}})
	require.NotNil(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestCantSetBothTimeAndHeight(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now(), Height: s.ctx.BlockHeight() + 1}})
	require.NotNil(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestDoTimeUpgrade(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify can schedule an upgrade")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now()}})
	require.Nil(t, err)

	VerifyDoUpgrade(t)
}

func TestDoHeightUpgrade(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify can schedule an upgrade")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	require.Nil(t, err)

	VerifyDoUpgrade(t)
}

func TestCanOverwriteScheduleUpgrade(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Can overwrite plan")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "bad_test", Height: s.ctx.BlockHeight() + 10}})
	require.Nil(t, err)
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	require.Nil(t, err)

	VerifyDoUpgrade(t)
}

func VerifyDoIBCLastBlock(t *testing.T) {
	t.Log("Verify that chain committed to consensus state on the last height it will commit")
	nextValsHash := []byte("nextValsHash")
	newCtx := s.ctx.WithBlockHeader(tmproto.Header{
		Height:             s.ctx.BlockHeight(),
		NextValidatorsHash: nextValsHash,
	})

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	s.module.BeginBlock(newCtx, req)

	// plan Height is at ctx.BlockHeight+1
	consState, err := s.keeper.GetUpgradedConsensusState(newCtx, s.ctx.BlockHeight()+1)
	require.NoError(t, err)
	require.Equal(t, &ibctmtypes.ConsensusState{Timestamp: newCtx.BlockTime(), NextValidatorsHash: nextValsHash}, consState)
}

func VerifyDoIBCUpgrade(t *testing.T) {
	t.Log("Verify that a panic happens at the upgrade time/height")
	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())

	// Check IBC state is set before upgrade using last height: s.ctx.BlockHeight()
	cs, err := s.keeper.GetUpgradedClient(newCtx, s.ctx.BlockHeight())
	require.NoError(t, err, "could not retrieve upgraded client before upgrade plan is applied")
	require.NotNil(t, cs, "IBC client is nil before upgrade")

	consState, err := s.keeper.GetUpgradedConsensusState(newCtx, s.ctx.BlockHeight())
	require.NoError(t, err, "could not retrieve upgraded consensus state before upgrade plan is applied")
	require.NotNil(t, consState, "IBC consensus state is nil before upgrade")

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler("test", func(ctx sdk.Context, plan types.Plan) {})
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyCleared(t, newCtx)

	// Check IBC state is cleared after upgrade using last height: s.ctx.BlockHeight()
	cs, err = s.keeper.GetUpgradedClient(newCtx, s.ctx.BlockHeight())
	require.Error(t, err, "retrieved upgraded client after upgrade plan is applied")
	require.Nil(t, cs, "IBC client is not-nil after upgrade")

	consState, err = s.keeper.GetUpgradedConsensusState(newCtx, s.ctx.BlockHeight())
	require.Error(t, err, "retrieved upgraded consensus state after upgrade plan is applied")
	require.Nil(t, consState, "IBC consensus state is not-nil after upgrade")
}

func VerifyDoUpgrade(t *testing.T) {
	t.Log("Verify that a panic happens at the upgrade time/height")
	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler("test", func(ctx sdk.Context, plan types.Plan) {})
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyCleared(t, newCtx)
}

func VerifyDoUpgradeWithCtx(t *testing.T, newCtx sdk.Context, proposalName string) {
	t.Log("Verify that a panic happens at the upgrade time/height")
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler(proposalName, func(ctx sdk.Context, plan types.Plan) {})
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyCleared(t, newCtx)
}

func TestHaltIfTooNew(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify that we don't panic with registered plan not in database at all")
	var called int
	s.keeper.SetUpgradeHandler("future", func(ctx sdk.Context, plan types.Plan) { called++ })

	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})
	require.Equal(t, 0, called)

	t.Log("Verify we panic if we have a registered handler ahead of time")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "future", Height: s.ctx.BlockHeight() + 3}})
	require.NoError(t, err)
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})
	require.Equal(t, 0, called)

	t.Log("Verify we no longer panic if the plan is on time")

	futCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 3).WithBlockTime(time.Now())
	req = abci.RequestBeginBlock{Header: futCtx.BlockHeader()}
	require.NotPanics(t, func() {
		s.module.BeginBlock(futCtx, req)
	})
	require.Equal(t, 1, called)

	VerifyCleared(t, futCtx)
}

func VerifyCleared(t *testing.T, newCtx sdk.Context) {
	t.Log("Verify that the upgrade plan has been cleared")
	bz, err := s.querier(newCtx, []string{types.QueryCurrent}, abci.RequestQuery{})
	require.NoError(t, err)
	require.Nil(t, bz)
}

func TestCanClear(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify upgrade is scheduled")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now()}})
	require.Nil(t, err)

	err = s.handler(s.ctx, &types.CancelSoftwareUpgradeProposal{Title: "cancel"})
	require.Nil(t, err)

	VerifyCleared(t, s.ctx)
}

func TestCantApplySameUpgradeTwice(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now()}})
	require.Nil(t, err)
	VerifyDoUpgrade(t)
	t.Log("Verify an executed upgrade \"test\" can't be rescheduled")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Time: time.Now()}})
	require.NotNil(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestNoSpuriousUpgrades(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	t.Log("Verify that no upgrade panic is triggered in the BeginBlocker when we haven't scheduled an upgrade")
	req := abci.RequestBeginBlock{Header: s.ctx.BlockHeader()}
	require.NotPanics(t, func() {
		s.module.BeginBlock(s.ctx, req)
	})
}

func TestPlanStringer(t *testing.T) {
	clientState := &ibctmtypes.ClientState{ChainId: "gaiachain"}
	cs, err := clienttypes.PackClientState(clientState)
	require.NoError(t, err)

	ti, err := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	require.Nil(t, err)
	require.Equal(t, `Upgrade Plan
  Name: test
  Time: 2020-01-01T00:00:00Z
  Info: .
  Upgraded IBC Client: no upgraded client provided`, types.Plan{Name: "test", Time: ti}.String())
	require.Equal(t, `Upgrade Plan
  Name: test
  Height: 100
  Info: .
  Upgraded IBC Client: no upgraded client provided`, types.Plan{Name: "test", Height: 100}.String())
	require.Equal(t, fmt.Sprintf(`Upgrade Plan
  Name: test
  Height: 100
  Info: .
  Upgraded IBC Client: %s`, clientState), types.Plan{Name: "test", Height: 100, UpgradedClientState: cs}.String())
}

func VerifyNotDone(t *testing.T, newCtx sdk.Context, name string) {
	t.Log("Verify that upgrade was not done")
	height := s.keeper.GetDoneHeight(newCtx, name)
	require.Zero(t, height)
}

func VerifyDone(t *testing.T, newCtx sdk.Context, name string) {
	t.Log("Verify that the upgrade plan has been executed")
	height := s.keeper.GetDoneHeight(newCtx, name)
	require.NotZero(t, height)
}

func VerifySet(t *testing.T, skipUpgradeHeights map[int64]bool) {
	t.Log("Verify if the skip upgrade has been set")

	for k := range skipUpgradeHeights {
		require.True(t, s.keeper.IsSkipHeight(k))
	}
}

func TestContains(t *testing.T) {
	var (
		skipOne int64 = 11
	)
	s := setupTest(10, map[int64]bool{skipOne: true})

	VerifySet(t, map[int64]bool{skipOne: true})
	t.Log("case where array contains the element")
	require.True(t, s.keeper.IsSkipHeight(11))

	t.Log("case where array doesn't contain the element")
	require.False(t, s.keeper.IsSkipHeight(4))
}

func TestSkipUpgradeSkippingAll(t *testing.T) {
	var (
		skipOne int64 = 11
		skipTwo int64 = 20
	)
	s := setupTest(10, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx := s.ctx

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}})
	require.NoError(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in both cases")
	VerifySet(t, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify a second proposal also is being cleared")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}})
	require.NoError(t, err)

	newCtx = newCtx.WithBlockHeight(skipTwo)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	// To ensure verification is being done only after both upgrades are cleared
	t.Log("Verify if both proposals are cleared")
	VerifyCleared(t, s.ctx)
	VerifyNotDone(t, s.ctx, "test")
	VerifyNotDone(t, s.ctx, "test2")
}

func TestUpgradeSkippingOne(t *testing.T) {
	var (
		skipOne int64 = 11
		skipTwo int64 = 20
	)
	s := setupTest(10, map[int64]bool{skipOne: true})

	newCtx := s.ctx

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}})
	require.Nil(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in one case and does upgrade on another")
	VerifySet(t, map[int64]bool{skipOne: true})

	// Setting block height of proposal test
	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify the second proposal is not skipped")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}})
	require.Nil(t, err)
	// Setting block height of proposal test2
	newCtx = newCtx.WithBlockHeight(skipTwo)
	VerifyDoUpgradeWithCtx(t, newCtx, "test2")

	t.Log("Verify first proposal is cleared and second is done")
	VerifyNotDone(t, s.ctx, "test")
	VerifyDone(t, s.ctx, "test2")
}

func TestUpgradeSkippingOnlyTwo(t *testing.T) {
	var (
		skipOne   int64 = 11
		skipTwo   int64 = 20
		skipThree int64 = 25
	)
	s := setupTest(10, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx := s.ctx

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}})
	require.Nil(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in both cases and does third upgrade")
	VerifySet(t, map[int64]bool{skipOne: true, skipTwo: true})

	// Setting block height of proposal test
	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	// A new proposal with height in skipUpgradeHeights
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}})
	require.Nil(t, err)
	// Setting block height of proposal test2
	newCtx = newCtx.WithBlockHeight(skipTwo)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify a new proposal is not skipped")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop3", Plan: types.Plan{Name: "test3", Height: skipThree}})
	require.Nil(t, err)
	newCtx = newCtx.WithBlockHeight(skipThree)
	VerifyDoUpgradeWithCtx(t, newCtx, "test3")

	t.Log("Verify two proposals are cleared and third is done")
	VerifyNotDone(t, s.ctx, "test")
	VerifyNotDone(t, s.ctx, "test2")
	VerifyDone(t, s.ctx, "test3")
}

func TestUpgradeWithoutSkip(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	require.Nil(t, err)
	t.Log("Verify if upgrade happens without skip upgrade")
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyDoUpgrade(t)
	VerifyDone(t, s.ctx, "test")
}

func TestIBCUpgradeWithoutSkip(t *testing.T) {
	s := setupTest(10, map[int64]bool{})
	cs, err := clienttypes.PackClientState(&ibctmtypes.ClientState{})
	require.NoError(t, err)
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{
		Title: "prop",
		Plan: types.Plan{
			Name:                "test",
			Height:              s.ctx.BlockHeight() + 1,
			UpgradedClientState: cs,
		},
	})
	require.Nil(t, err)

	t.Log("Verify if last height stores consensus state")
	VerifyDoIBCLastBlock(t)

	VerifyDoUpgrade(t)
	VerifyDone(t, s.ctx, "test")
}

func TestDumpUpgradeInfoToFile(t *testing.T) {
	s := setupTest(10, map[int64]bool{})

	planHeight := s.ctx.BlockHeight() + 1
	name := "test"
	t.Log("verify if upgrade height is dumped to file")
	err := s.keeper.DumpUpgradeInfoToDisk(planHeight, name)
	require.Nil(t, err)

	upgradeInfoFilePath, err := s.keeper.GetUpgradeInfoPath()
	require.Nil(t, err)

	data, err := ioutil.ReadFile(upgradeInfoFilePath)
	require.NoError(t, err)

	var upgradeInfo storetypes.UpgradeInfo
	err = json.Unmarshal(data, &upgradeInfo)
	require.Nil(t, err)

	t.Log("Verify upgrade height from file matches ")
	require.Equal(t, upgradeInfo.Height, planHeight)

	// clear the test file
	err = os.Remove(upgradeInfoFilePath)
	require.Nil(t, err)
}
