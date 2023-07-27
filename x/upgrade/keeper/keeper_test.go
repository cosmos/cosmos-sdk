package keeper_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type KeeperTestSuite struct {
	suite.Suite

	homeDir string
	app     *simapp.SimApp
	ctx     sdk.Context
}

func (s *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	homeDir := filepath.Join(s.T().TempDir(), "x_upgrade_keeper_test")
	app.UpgradeKeeper = keeper.NewKeeper( // recreate keeper in order to use a custom home path
		make(map[int64]bool), app.GetKey(types.StoreKey), app.AppCodec(), homeDir, app.BaseApp,
	)
	s.T().Log("home dir:", homeDir)
	s.homeDir = homeDir
	s.app = app
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{
		Time:   time.Now(),
		Height: 10,
	})
}

func (s *KeeperTestSuite) TestReadUpgradeInfoFromDisk() {
	// require no error when the upgrade info file does not exist
	_, err := s.app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	s.Require().NoError(err)

	expected := store.UpgradeInfo{
		Name:   "test_upgrade",
		Height: 100,
	}

	// create an upgrade info file
	s.Require().NoError(s.app.UpgradeKeeper.DumpUpgradeInfoToDisk(expected.Height, expected.Name))

	ui, err := s.app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	s.Require().NoError(err)
	s.Require().Equal(expected, ui)
}

func (s *KeeperTestSuite) TestScheduleUpgrade() {
	cases := []struct {
		name    string
		plan    types.Plan
		setup   func()
		expPass bool
	}{
		{
			name: "successful height schedule",
			plan: types.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setup:   func() {},
			expPass: true,
		},
		{
			name: "successful overwrite",
			plan: types.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setup: func() {
				err := s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, types.Plan{
					Name:   "alt-good",
					Info:   "new text here",
					Height: 543210000,
				})
				s.Require().NoError(err)
			},
			expPass: true,
		},
		{
			name: "unsuccessful schedule: invalid plan",
			plan: types.Plan{
				Height: 123450000,
			},
			setup:   func() {},
			expPass: false,
		},
		{
			name: "unsuccessful height schedule: due date in past",
			plan: types.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 1,
			},
			setup:   func() {},
			expPass: false,
		},
		{
			name: "unsuccessful schedule: schedule already executed",
			plan: types.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setup: func() {
				s.app.UpgradeKeeper.SetUpgradeHandler("all-good", func(ctx sdk.Context, plan types.Plan, vm module.VersionMap) (module.VersionMap, error) {
					return vm, nil
				})
				s.app.UpgradeKeeper.ApplyUpgrade(s.ctx, types.Plan{
					Name:   "all-good",
					Info:   "some text here",
					Height: 123450000,
				})
			},
			expPass: false,
		},
	}

	for _, tc := range cases {
		tc := tc

		s.Run(tc.name, func() {
			// reset suite
			s.SetupTest()

			// setup test case
			tc.setup()

			err := s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, tc.plan)

			if tc.expPass {
				s.Require().NoError(err, "valid test case failed")
			} else {
				s.Require().Error(err, "invalid test case passed")
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetUpgradedClient() {
	cs := []byte("IBC client state")

	cases := []struct {
		name   string
		height int64
		setup  func()
		exists bool
	}{
		{
			name:   "no upgraded client exists",
			height: 10,
			setup:  func() {},
			exists: false,
		},
		{
			name:   "success",
			height: 10,
			setup: func() {
				s.app.UpgradeKeeper.SetUpgradedClient(s.ctx, 10, cs)
			},
			exists: true,
		},
	}

	for _, tc := range cases {
		// reset suite
		s.SetupTest()

		// setup test case
		tc.setup()

		gotCs, exists := s.app.UpgradeKeeper.GetUpgradedClient(s.ctx, tc.height)
		if tc.exists {
			s.Require().Equal(cs, gotCs, "valid case: %s did not retrieve correct client state", tc.name)
			s.Require().True(exists, "valid case: %s did not retrieve client state", tc.name)
		} else {
			s.Require().Nil(gotCs, "invalid case: %s retrieved valid client state", tc.name)
			s.Require().False(exists, "invalid case: %s retrieved valid client state", tc.name)
		}
	}
}

// Test that the protocol version successfully increments after an
// upgrade and is successfully set on BaseApp's appVersion.
func (s *KeeperTestSuite) TestIncrementProtocolVersion() {
	oldProtocolVersion := s.app.BaseApp.AppVersion()
	s.app.UpgradeKeeper.SetUpgradeHandler("dummy", func(_ sdk.Context, _ types.Plan, vm module.VersionMap) (module.VersionMap, error) { return vm, nil })
	dummyPlan := types.Plan{
		Name:   "dummy",
		Info:   "some text here",
		Height: 100,
	}
	s.app.UpgradeKeeper.ApplyUpgrade(s.ctx, dummyPlan)
	upgradedProtocolVersion := s.app.BaseApp.AppVersion()

	s.Require().Equal(oldProtocolVersion+1, upgradedProtocolVersion)
}

// Tests that the underlying state of x/upgrade is set correctly after
// an upgrade.
func (s *KeeperTestSuite) TestMigrations() {
	initialVM := module.VersionMap{"bank": uint64(1)}
	s.app.UpgradeKeeper.SetModuleVersionMap(s.ctx, initialVM)
	vmBefore := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)
	s.app.UpgradeKeeper.SetUpgradeHandler("dummy", func(_ sdk.Context, _ types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// simulate upgrading the bank module
		vm["bank"] = vm["bank"] + 1
		return vm, nil
	})
	dummyPlan := types.Plan{
		Name:   "dummy",
		Info:   "some text here",
		Height: 123450000,
	}

	s.app.UpgradeKeeper.ApplyUpgrade(s.ctx, dummyPlan)
	vm := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)
	s.Require().Equal(vmBefore["bank"]+1, vm["bank"])
}

func (s *KeeperTestSuite) TestLastCompletedUpgrade() {
	keeper := s.app.UpgradeKeeper
	require := s.Require()

	s.T().Log("verify empty name if applied upgrades are empty")
	name, height := keeper.GetLastCompletedUpgrade(s.ctx)
	require.Equal("", name)
	require.Equal(int64(0), height)

	keeper.SetUpgradeHandler("test-v0.9", func(_ sdk.Context, _ types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})

	keeper.ApplyUpgrade(s.ctx, types.Plan{
		Name:   "test-v0.9",
		Height: 10,
	})

	s.T().Log("verify valid upgrade name and height")
	name, height = keeper.GetLastCompletedUpgrade(s.ctx)
	require.Equal("test-v0.9", name)
	require.Equal(int64(10), height)

	keeper.SetUpgradeHandler("test-v0.10", func(_ sdk.Context, _ types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})

	newCtx := s.ctx.WithBlockHeight(15)
	keeper.ApplyUpgrade(newCtx, types.Plan{
		Name:   "test-v0.10",
		Height: 15,
	})

	s.T().Log("verify valid upgrade name and height with multiple upgrades")
	name, height = keeper.GetLastCompletedUpgrade(newCtx)
	require.Equal("test-v0.10", name)
	require.Equal(int64(15), height)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
