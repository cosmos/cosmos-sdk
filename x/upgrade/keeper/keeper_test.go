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
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	ibcexported "github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
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
		make(map[int64]bool), app.GetKey(types.StoreKey), app.AppCodec(), homeDir,
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
	clientState := &ibctmtypes.ClientState{ChainId: "gaiachain"}
	cs, err := clienttypes.PackClientState(clientState)
	s.Require().NoError(err)

	altClientState := &ibctmtypes.ClientState{ChainId: "ethermint"}
	altCs, err := clienttypes.PackClientState(altClientState)
	s.Require().NoError(err)

	consState := ibctmtypes.NewConsensusState(time.Now(), commitmenttypes.NewMerkleRoot([]byte("app_hash")), []byte("next_vals_hash"))
	consAny, err := clienttypes.PackConsensusState(consState)
	s.Require().NoError(err)

	cases := []struct {
		name    string
		plan    types.Plan
		setup   func()
		expPass bool
	}{
		{
			name: "successful time schedule",
			plan: types.Plan{
				Name: "all-good",
				Info: "some text here",
				Time: s.ctx.BlockTime().Add(time.Hour),
			},
			setup:   func() {},
			expPass: true,
		},
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
			name: "successful ibc schedule",
			plan: types.Plan{
				Name:                "all-good",
				Info:                "some text here",
				Height:              123450000,
				UpgradedClientState: cs,
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
				s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, types.Plan{
					Name:   "alt-good",
					Info:   "new text here",
					Height: 543210000,
				})
			},
			expPass: true,
		},
		{
			name: "successful IBC overwrite",
			plan: types.Plan{
				Name:                "all-good",
				Info:                "some text here",
				Height:              123450000,
				UpgradedClientState: cs,
			},
			setup: func() {
				s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, types.Plan{
					Name:                "alt-good",
					Info:                "new text here",
					Height:              543210000,
					UpgradedClientState: altCs,
				})
			},
			expPass: true,
		},
		{
			name: "successful IBC overwrite with non IBC plan",
			plan: types.Plan{
				Name:   "all-good",
				Info:   "some text here",
				Height: 123450000,
			},
			setup: func() {
				s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, types.Plan{
					Name:                "alt-good",
					Info:                "new text here",
					Height:              543210000,
					UpgradedClientState: altCs,
				})
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
			name: "unsuccessful time schedule: due date in past",
			plan: types.Plan{
				Name: "all-good",
				Info: "some text here",
				Time: s.ctx.BlockTime(),
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
				s.app.UpgradeKeeper.SetUpgradeHandler("all-good", func(_ sdk.Context, _ types.Plan) {})
				s.app.UpgradeKeeper.ApplyUpgrade(s.ctx, types.Plan{
					Name:   "all-good",
					Info:   "some text here",
					Height: 123450000,
				})
			},
			expPass: false,
		},
		{
			name: "unsuccessful IBC schedule: UpgradedClientState is not valid client state",
			plan: types.Plan{
				Name:                "all-good",
				Info:                "some text here",
				Height:              123450000,
				UpgradedClientState: consAny,
			},
			setup:   func() {},
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
				if tc.plan.UpgradedClientState != nil {
					got, height, err := s.app.UpgradeKeeper.GetUpgradedClient(s.ctx)
					s.Require().NoError(err)
					s.Require().Equal(tc.plan.Height, height, "upgradedClient not stored at correct upgrade height")
					s.Require().Equal(clientState, got, "upgradedClient not equal to expected value")
				} else {
					// check that upgraded client is empty if latest plan does not specify an upgraded client
					got, _, err := s.app.UpgradeKeeper.GetUpgradedClient(s.ctx)
					s.Require().Error(err)
					s.Require().Nil(got)
				}
			} else {
				s.Require().Error(err, "invalid test case passed")
			}
		})
	}
}

func (s *TestKeeperTestSuite) TestGetUpgradedClient() {
	var (
		clientState ibcexported.ClientState
		height      int64
	)
	cases := []struct {
		name   string
		setup  func()
		exists bool
		panics bool
	}{
		{
			name:   "no upgraded client exists",
			setup:  func() {},
			exists: false,
			panics: false,
		},
		{
			name: "success",
			setup: func() {
				clientState = &ibctmtypes.ClientState{ChainId: "gaiachain"}
				height = 10

				s.app.UpgradeKeeper.SetUpgradedClient(s.app.GetContext(), 10, clientState)
			},
			exists: true,
			panics: false,
		},
		{
			name: "GetUpgradedClient panics if multiple clients exist in store",
			setup: func() {

				clientState = &ibctmtypes.ClientState{ChainId: "gaiachain"}
				altCs := &ibctmtypes.ClientState{ChainId: "ethermint"}

				height = 10

				s.app.UpgradeKeeper.SetUpgradedClient(s.app.GetContext(), 10, clientState)
				s.app.UpgradeKeeper.SetUpgradedClient(s.app.GetContext(), 50, altCs)
			},
			exists: true,
			panics: true,
		},
	}

	for _, tc := range cases {
		// reset suite
		s.SetupTest()

		// setup test case
		tc.setup()

		if tc.panics {
			s.Require().Panics(func() {
				s.app.UpgradeKeeper.GetUpgradedClient(s.app.GetContext())
			},
				"case: %s did not panic as expected", tc.name)
			continue
		}

		gotCs, gotHeight, err := s.app.UpgradeKeeper.GetUpgradedClient(s.app.GetContext())
		if tc.exists {
			s.Require().Equal(clientState, gotCs, "valid case: %s did not retrieve correct client state", tc.name)
			s.Require().Equal(height, gotHeight, "valid case: %s did not retrieve correct upgrade height", tc.name)
			s.Require().NoError(err, "valid case: %s returned error")
		} else {
			s.Require().Nil(gotCs, "invalid case: %s retrieved valid client state", tc.name)
			s.Require().Equal(0, gotHeight, "invalid case: %s retrieved valid upgrade height", tc.name)
			s.Require().Error(err, "invalid case: %s did not return error", tc.name)
		}
	}

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
