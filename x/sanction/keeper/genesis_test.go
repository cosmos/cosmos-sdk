package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

type GenesisTestSuite struct {
	BaseTestSuite
}

func (s *GenesisTestSuite) SetupTest() {
	s.BaseSetup()
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) TestKeeper_InitGenesis() {
	// Set some different (hopefully unique) default param values.
	origSanctMinDep := sanction.DefaultImmediateSanctionMinDeposit
	origUnsanctMinDep := sanction.DefaultImmediateUnsanctionMinDeposit
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origSanctMinDep
		sanction.DefaultImmediateUnsanctionMinDeposit = origUnsanctMinDep
	}()
	sanction.DefaultImmediateSanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("jounce", 31367))
	sanction.DefaultImmediateUnsanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("crackle", 56476))

	addr1 := sdk.AccAddress("1st_init_tester_addr")
	addr2 := sdk.AccAddress("2nd_init_tester_addr")
	addr3 := sdk.AccAddress("3rd_init_tester_addr")
	addr4 := sdk.AccAddress("4th_init_tester_addr")
	addr5 := sdk.AccAddress("5th_init_tester_addr")
	addr6 := sdk.AccAddress("6th_init_tester_addr")
	addr7 := sdk.AccAddress("7th_init_tester_addr")
	addr8 := sdk.AccAddress("8th_init_tester_addr")

	tests := []struct {
		name      string
		setup     func(s *GenesisTestSuite)
		genState  *sanction.GenesisState
		expExport *sanction.GenesisState
		expPanic  []string
	}{
		{
			name:     "nil gen state nothing yet in state",
			genState: nil,
			expExport: &sanction.GenesisState{
				Params:              sanction.DefaultParams(),
				SanctionedAddresses: nil,
				TemporaryEntries:    nil,
			},
		},
		{
			name: "nil gen state stuff already in state",
			setup: func(s *GenesisTestSuite) {
				s.ReqOKSetParams(&sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("snap", 4)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("pop", 6)),
				})
				s.ReqOKAddPermSanct("addr1", addr1)
				s.ReqOKAddTempSanct(55, "addr2", addr2)
				s.ReqOKAddTempUnsanct(111, "addr3", addr3)
			},
			genState: nil,
			expExport: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("snap", 4)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("pop", 6)),
				},
				SanctionedAddresses: []string{addr1.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr2.String(), ProposalId: 55, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr3.String(), ProposalId: 111, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
		},
		{
			name: "filled gen state nothing yet in state",
			genState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("bop", 71)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("slide", 95)),
				},
				SanctionedAddresses: []string{
					addr7.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr1.String(),
					addr6.String(),
					addr8.String(),
					addr5.String(),
				},
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr4.String(), ProposalId: 5555, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr7.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr1.String(), ProposalId: 24, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr3.String(), ProposalId: 99, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr1.String(), ProposalId: 23, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr2.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_SANCTIONED},
				},
			},
			expExport: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("bop", 71)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("slide", 95)),
				},
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
					addr7.String(),
					addr8.String(),
				},
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 23, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr1.String(), ProposalId: 24, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr2.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr3.String(), ProposalId: 99, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr4.String(), ProposalId: 5555, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr7.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
		},
		{
			name: "filled gen state other stuff already in state",
			setup: func(s *GenesisTestSuite) {
				s.ReqOKSetParams(&sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("oops", 954845)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("wrong", 6)),
				})
				s.ReqOKAddPermSanct("addr2, addr4, addr6, addr8", addr2, addr4, addr6, addr8)
				s.ReqOKAddTempSanct(99, "addr3, addr5", addr3, addr5)
				s.ReqOKAddTempUnsanct(12, "addr7, addr2, addr6", addr7, addr2, addr6)
			},
			genState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("plop", 1984)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("skewer", 5532)),
				},
				SanctionedAddresses: []string{
					addr7.String(),
					addr3.String(),
					addr1.String(),
					addr6.String(),
					addr8.String(),
				},
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr4.String(), ProposalId: 5555, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr1.String(), ProposalId: 24, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr2.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr7.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr1.String(), ProposalId: 23, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr3.String(), ProposalId: 99, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
			expExport: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("plop", 1984)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("skewer", 5532)),
				},
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr6.String(),
					addr7.String(),
					addr8.String(),
				},
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 23, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr1.String(), ProposalId: 24, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr2.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr3.String(), ProposalId: 99, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr4.String(), ProposalId: 5555, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr5.String(), ProposalId: 99, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr7.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					// Because addr6 is in SanctionedAddresses, the temporary entry for it (that is added during setup) is cleared.
				},
			},
		},
		{
			name: "nil params",
			setup: func(s *GenesisTestSuite) {
				s.ReqOKSetParams(&sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("these", 3302), sdk.NewInt64Coin("should", 9)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("goaway", 551515)),
				})
			},
			genState: &sanction.GenesisState{
				Params:              nil,
				SanctionedAddresses: []string{addr1.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr2.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr7.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
			expExport: &sanction.GenesisState{
				Params:              sanction.DefaultParams(),
				SanctionedAddresses: []string{addr1.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr2.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr7.String(), ProposalId: 12, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
		},
		{
			name: "six addrs first bad",
			genState: &sanction.GenesisState{
				SanctionedAddresses: []string{
					"addrOneString",
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
			},
			expPanic: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "six addrs third bad",
			genState: &sanction.GenesisState{
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					"addrThreeString",
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
			},
			expPanic: []string{"invalid address[2]", "decoding bech32 failed"},
		},
		{
			name: "six addrs, sixth bad",
			genState: &sanction.GenesisState{
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					"addrSixString",
				},
			},
			expPanic: []string{"invalid address[5]", "decoding bech32 failed"},
		},
		{
			name: "six temps first with bad addr",
			genState: &sanction.GenesisState{
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: "addrOneString", ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr3.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr5.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr6.String(), ProposalId: 6, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
			expPanic: []string{"invalid temp entry[0]", "invalid address", "decoding bech32 failed"},
		},
		{
			name: "six temps first with bad status",
			genState: &sanction.GenesisState{
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 1, Status: -12},
					{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr3.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr5.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr6.String(), ProposalId: 6, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
			expPanic: []string{"invalid temp entry[0]", "invalid status", "-12"},
		},
		{
			name: "six temps first with unspecified status",
			genState: &sanction.GenesisState{
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_UNSPECIFIED},
					{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr3.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr5.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr6.String(), ProposalId: 6, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
			expPanic: []string{"invalid temp entry[0]", "invalid status", "TEMP_STATUS_UNSPECIFIED"},
		},
		{
			name: "six temps third with bad addr",
			genState: &sanction.GenesisState{
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: "addrThreeString", ProposalId: 3, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr5.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr6.String(), ProposalId: 6, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
			expPanic: []string{"invalid temp entry[2]", "invalid address", "decoding bech32 failed"},
		},
		{
			name: "six temps third with bad status",
			genState: &sanction.GenesisState{
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr3.String(), ProposalId: 3, Status: 55},
					{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr5.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr6.String(), ProposalId: 6, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
			expPanic: []string{"invalid temp entry[2]", "invalid status", "55"},
		},
		{
			name: "six temps third with unspecified status",
			genState: &sanction.GenesisState{
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr3.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_UNSPECIFIED},
					{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr5.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr6.String(), ProposalId: 6, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
			expPanic: []string{"invalid temp entry[2]", "invalid status", "TEMP_STATUS_UNSPECIFIED"},
		},
		{
			name: "six temps sixth with bad addr",
			genState: &sanction.GenesisState{
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr3.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr5.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: "addrSixString", ProposalId: 5, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				},
			},
			expPanic: []string{"invalid temp entry[5]", "invalid address", "decoding bech32 failed"},
		},
		{
			name: "six temps sixth with bad status",
			genState: &sanction.GenesisState{
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr3.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr5.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr6.String(), ProposalId: 6, Status: 800},
				},
			},
			expPanic: []string{"invalid temp entry[5]", "invalid status", "800"},
		},
		{
			name: "six temps sixth with unspecified status",
			genState: &sanction.GenesisState{
				TemporaryEntries: []*sanction.TemporaryEntry{
					{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr3.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
					{Address: addr5.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_SANCTIONED},
					{Address: addr6.String(), ProposalId: 6, Status: sanction.TEMP_STATUS_UNSPECIFIED},
				},
			},
			expPanic: []string{"invalid temp entry[5]", "invalid status", "TEMP_STATUS_UNSPECIFIED"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ClearState()
			if tc.setup != nil {
				tc.setup(s)
			}

			em := sdk.NewEventManager()
			ctx := s.SdkCtx.WithEventManager(em)
			testFuncInit := func() {
				s.Keeper.InitGenesis(ctx, tc.genState)
			}
			testutil.RequirePanicContents(s.T(), tc.expPanic, testFuncInit, "InitGenesis")
			events := em.Events()
			s.Assert().Empty(events, "events emitted during InitGenesis")
			if tc.expExport != nil {
				s.ExportAndCheck(tc.expExport)
			}
		})
	}
}

func (s *GenesisTestSuite) TestKeeper_ExportGenesis() {
	s.ClearState()

	addr1 := sdk.AccAddress("1st_export_test_addr")
	addr2 := sdk.AccAddress("2nd_export_test_addr")
	addr3 := sdk.AccAddress("3rd_export_test_addr")
	addr4 := sdk.AccAddress("4th_export_test_addr")
	addr5 := sdk.AccAddress("5th_export_test_addr")
	addr6 := sdk.AccAddress("6th_export_test_addr")
	addr7 := sdk.AccAddress("7th_export_test_addr")
	addr8 := sdk.AccAddress("8th_export_test_addr")

	// Set some different (hopefully unique) default param values.
	origSanctMinDep := sanction.DefaultImmediateSanctionMinDeposit
	origUnsanctMinDep := sanction.DefaultImmediateUnsanctionMinDeposit
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origSanctMinDep
		sanction.DefaultImmediateUnsanctionMinDeposit = origUnsanctMinDep
	}()
	sanction.DefaultImmediateSanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("gurgle", 129))
	sanction.DefaultImmediateUnsanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("yodle", 94221))

	s.Run("nothing in state", func() {
		expected := &sanction.GenesisState{
			Params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sanction.DefaultImmediateSanctionMinDeposit,
				ImmediateUnsanctionMinDeposit: sanction.DefaultImmediateUnsanctionMinDeposit,
			},
			SanctionedAddresses: nil,
			TemporaryEntries:    nil,
		}

		s.ExportAndCheck(expected)
	})

	s.Run("a little of everything in state", func() {
		expected := &sanction.GenesisState{
			Params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("scoina", 54), sdk.NewInt64Coin("scoinb", 76)),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("ucoiny", 17), sdk.NewInt64Coin("ucoinz", 39)),
			},
			SanctionedAddresses: []string{
				addr1.String(),
				addr2.String(),
				addr3.String(),
				addr4.String(),
				addr5.String(),
				addr6.String(),
				addr7.String(),
				addr8.String(),
			},
			TemporaryEntries: []*sanction.TemporaryEntry{
				{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},

				{Address: addr2.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_UNSANCTIONED},

				{Address: addr3.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_SANCTIONED},
				{Address: addr3.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_UNSANCTIONED},

				{Address: addr4.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_SANCTIONED},

				{Address: addr5.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},
				{Address: addr5.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_SANCTIONED},

				{Address: addr6.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				{Address: addr6.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},

				{Address: addr7.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_SANCTIONED},
				{Address: addr7.String(), ProposalId: 47, Status: sanction.TEMP_STATUS_SANCTIONED},
				{Address: addr7.String(), ProposalId: 88, Status: sanction.TEMP_STATUS_UNSANCTIONED},

				{Address: addr8.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
				{Address: addr8.String(), ProposalId: 47, Status: sanction.TEMP_STATUS_SANCTIONED},
				{Address: addr8.String(), ProposalId: 88, Status: sanction.TEMP_STATUS_UNSANCTIONED},
			},
		}

		s.ReqOKSetParams(expected.Params)
		s.ReqOKAddPermSanct("addr1, addr2, addr3, addr4, addr5, addr6, addr7, addr8", addr1, addr2, addr3, addr4, addr5, addr6, addr7, addr8)
		s.ReqOKAddTempSanct(1, "addr1, addr5", addr1, addr5)
		s.ReqOKAddTempUnsanct(1, "addr2, addr6", addr2, addr6)
		s.ReqOKAddTempSanct(2, "addr3, addr5, addr7", addr3, addr5, addr7)
		s.ReqOKAddTempUnsanct(2, "addr4, addr6, addr8", addr4, addr6, addr8)
		s.ReqOKAddTempUnsanct(3, "addr3", addr3)
		s.ReqOKAddTempSanct(4, "addr4", addr4)
		s.ReqOKAddTempSanct(47, "addr7, addr8", addr7, addr8)
		s.ReqOKAddTempUnsanct(88, "addr7, addr8", addr7, addr8)

		s.ExportAndCheck(expected)
	})

	s.Run("with some entries removed", func() {
		expected := &sanction.GenesisState{
			Params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("scoina", 54), sdk.NewInt64Coin("scoinb", 76)),
				ImmediateUnsanctionMinDeposit: sanction.DefaultImmediateUnsanctionMinDeposit,
			},
			SanctionedAddresses: []string{
				addr1.String(),
				addr3.String(),
				addr5.String(),
				addr7.String(),
			},
			TemporaryEntries: []*sanction.TemporaryEntry{
				{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},

				{Address: addr2.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_UNSANCTIONED},

				{Address: addr3.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_UNSANCTIONED},

				{Address: addr4.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_SANCTIONED},

				{Address: addr6.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_UNSANCTIONED},

				{Address: addr8.String(), ProposalId: 47, Status: sanction.TEMP_STATUS_SANCTIONED},
				{Address: addr8.String(), ProposalId: 88, Status: sanction.TEMP_STATUS_UNSANCTIONED},
			},
		}

		store := s.GetStore()
		s.Require().NotPanics(func() {
			s.Keeper.OnlyTestsDeleteParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit)
		}, "deleting the ParamNameImmediateUnsanctionMinDeposit param entry")
		// Note: Not using UnsanctionAddresses here since I want to keep the temp entries for these addrs.
		s.Require().NotPanics(func() {
			for _, addr := range []sdk.AccAddress{addr2, addr4, addr6, addr8} {
				key := keeper.CreateSanctionedAddrKey(addr)
				store.Delete(key)
			}
		}, "deleting sanctioned addr entries for addr2, addr4, addr6, addr8")
		s.ReqOKDelAddrTemp("addr5, addr7", addr5, addr7)
		s.ReqOKDelPropTemp(2)

		s.ExportAndCheck(expected)
	})
}

func (s *GenesisTestSuite) TestKeeper_GetAllSanctionedAddresses() {
	addr1 := sdk.AccAddress("1st_get_all_perm_address_in_test")
	addr2 := sdk.AccAddress("2nd_get_all_perm_address_in_test")
	addr3 := sdk.AccAddress("3rd_get_all_perm_address_in_test")
	addr4 := sdk.AccAddress("4th_get_all_perm_address_in_test")
	addr5 := sdk.AccAddress("5th_get_all_perm_address_in_test")
	addr6 := sdk.AccAddress("6th_get_all_perm_address_in_test")
	addr7 := sdk.AccAddress("7th_get_all_perm_address_in_test")
	addr8 := sdk.AccAddress("8th_get_all_perm_address_in_test")

	// Add some temporary sanctions to help ensure they don't matter here.
	s.ReqOKAddTempSanct(1, "addr1, addr5", addr1, addr5)
	s.ReqOKAddTempUnsanct(1, "addr2, addr6", addr2, addr6)
	s.ReqOKAddTempSanct(2, "addr1, addr3, addr5, addr7", addr1, addr3, addr5, addr7)
	s.ReqOKAddTempUnsanct(2, "addr2, addr4, addr6, addr8", addr2, addr4, addr6, addr8)
	s.ReqOKAddTempUnsanct(3, "addr3", addr3)
	s.ReqOKAddTempSanct(4, "addr4", addr4)
	s.ReqOKAddTempSanct(47, "addr7, addr8", addr7, addr8)
	s.ReqOKAddTempUnsanct(88, "addr7, addr8", addr7, addr8)
	// Then delete a few in the really off chance that matters in here.
	s.ReqOKDelAddrTemp("addr1", addr1)
	s.ReqOKDelAddrTemp("addr2", addr2)
	s.ReqOKDelPropTemp(2)

	s.Run("no entries", func() {
		var actual []string
		testFunc := func() {
			actual = s.Keeper.GetAllSanctionedAddresses(s.SdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllSanctionedAddresses")
		s.Assert().Empty(actual, "GetAllSanctionedAddresses result")
	})

	s.Run("one entry", func() {
		expected := []string{addr2.String()}
		s.ReqOKAddPermSanct("addr2", addr2)

		var actual []string
		testFunc := func() {
			actual = s.Keeper.GetAllSanctionedAddresses(s.SdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllSanctionedAddresses")
		s.Assert().Equal(expected, actual, "GetAllSanctionedAddresses result")
	})

	s.Run("several entries", func() {
		expected := []string{
			addr1.String(),
			addr2.String(),
			addr3.String(),
			addr4.String(),
			addr5.String(),
			addr6.String(),
			addr7.String(),
			addr8.String(),
		}
		// Note: addr2 was sanctioned in the "one entry" test.
		s.ReqOKAddPermSanct("addr1, addr3, addr4, addr5, addr6, addr7, addr8", addr1, addr3, addr4, addr5, addr6, addr7, addr8)

		var actual []string
		testFunc := func() {
			actual = s.Keeper.GetAllSanctionedAddresses(s.SdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllSanctionedAddresses")
		s.Assert().Equal(expected, actual, "GetAllSanctionedAddresses result")
	})

	s.Run("after some unsanctions", func() {
		// Note: all addrs should be sanctioned by now. Unsanction the odd ones.
		expected := []string{
			addr2.String(),
			addr4.String(),
			addr6.String(),
			addr8.String(),
		}

		s.ReqOKAddPermUnsanct("addr1, addr3, addr5, addr7", addr1, addr3, addr5, addr7)

		var actual []string
		testFunc := func() {
			actual = s.Keeper.GetAllSanctionedAddresses(s.SdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllSanctionedAddresses")
		s.Assert().Equal(expected, actual, "GetAllSanctionedAddresses result")
	})
}

func (s *GenesisTestSuite) TestKeeper_GetAllTemporaryEntries() {
	addr1 := sdk.AccAddress("1st_get_all_temp_address_in_test")
	addr2 := sdk.AccAddress("2nd_get_all_temp_address_in_test")
	addr3 := sdk.AccAddress("3rd_get_all_temp_address_in_test")
	addr4 := sdk.AccAddress("4th_get_all_temp_address_in_test")
	addr5 := sdk.AccAddress("5th_get_all_temp_address_in_test")
	addr6 := sdk.AccAddress("6th_get_all_temp_address_in_test")
	addr7 := sdk.AccAddress("7th_get_all_temp_address_in_test")
	addr8 := sdk.AccAddress("8th_get_all_temp_address_in_test")

	// Set permanent sanctions for the even addresses to help ensure they don't matter here.
	s.ReqOKAddPermSanct("addr2, addr4, addr6, addr8", addr2, addr4, addr6, addr8)
	// Unsanction the odd addrs just in case in the really off chance it does weird things in here.
	s.ReqOKAddPermUnsanct("addr1, addr3, addr5, addr7", addr1, addr3, addr5, addr7)

	s.Run("no entries", func() {
		var actual []*sanction.TemporaryEntry
		testFunc := func() {
			actual = s.Keeper.GetAllTemporaryEntries(s.SdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllTemporaryEntries")
		s.Assert().Empty(actual, "GetAllTemporaryEntries result")
	})

	s.Run("one entry", func() {
		expected := []*sanction.TemporaryEntry{
			{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_SANCTIONED},
		}
		s.ReqOKAddTempSanct(2, "addr2", addr2)

		var actual []*sanction.TemporaryEntry
		testFunc := func() {
			actual = s.Keeper.GetAllTemporaryEntries(s.SdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllTemporaryEntries")
		s.Assert().Equal(expected, actual, "GetAllTemporaryEntries result")
	})

	s.Run("several entries", func() {
		// Setup
		//	addr1: sanctioned prop 1, unsanctioned prop 2, sanctioned prop 3
		//	addr2: unsanctioned prop 1, sanctioned prop 2, unsanctioned prop 3
		//	addr3: sanctioned prop 1
		//	addr4: unsanctioned prop 1
		//	addr5: sanctioned props 2, 3, 4, 6
		//	addr6: unsanctioned props 3, 4, 5
		//	addr7: sanctioned prop 10
		//	addr8: unsanctioned prop 15
		// I.e.:
		//	Prop 1:  sanct: addr1, addr3; unsanct: addr2, addr4
		//	Prop 2:  sanct: addr2, addr5; unsanct: addr1
		//	Prop 3:  sanct: addr1, addr5; unsanct: addr2, addr6
		//	Prop 4:  sanct: addr5;        unsanct: addr6
		//	Prop 5:                       unsanct: addr6
		//  Prop 6:  sanct: addr5
		//	Prop 10: sanct: addr7;
		//	Prop 15:                      unsanct: addr8

		expected := []*sanction.TemporaryEntry{
			{Address: addr1.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},
			{Address: addr1.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_UNSANCTIONED},
			{Address: addr1.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_SANCTIONED},

			{Address: addr2.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_UNSANCTIONED},
			{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_SANCTIONED},
			{Address: addr2.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_UNSANCTIONED},

			{Address: addr3.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},

			{Address: addr4.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_UNSANCTIONED},

			{Address: addr5.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_SANCTIONED},
			{Address: addr5.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_SANCTIONED},
			{Address: addr5.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_SANCTIONED},
			{Address: addr5.String(), ProposalId: 6, Status: sanction.TEMP_STATUS_SANCTIONED},

			{Address: addr6.String(), ProposalId: 3, Status: sanction.TEMP_STATUS_UNSANCTIONED},
			{Address: addr6.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
			{Address: addr6.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_UNSANCTIONED},

			{Address: addr7.String(), ProposalId: 10, Status: sanction.TEMP_STATUS_SANCTIONED},

			{Address: addr8.String(), ProposalId: 15, Status: sanction.TEMP_STATUS_UNSANCTIONED},
		}

		s.ReqOKAddTempSanct(1, "addr1, addr3", addr1, addr3)
		s.ReqOKAddTempUnsanct(1, "addr2, addr4", addr2, addr4)
		// Note that the prop 2, addr2 sanction was already set in the "one entry" test.
		s.ReqOKAddTempSanct(2, "addr5", addr5)
		s.ReqOKAddTempUnsanct(2, "addr1", addr1)
		s.ReqOKAddTempSanct(3, "addr1, addr5", addr1, addr5)
		s.ReqOKAddTempUnsanct(3, "addr2, addr6", addr2, addr6)
		s.ReqOKAddTempSanct(4, "addr5", addr5)
		s.ReqOKAddTempUnsanct(4, "addr6", addr6)
		s.ReqOKAddTempUnsanct(5, "addr6", addr6)
		s.ReqOKAddTempSanct(6, "addr5", addr5)
		s.ReqOKAddTempSanct(10, "addr7", addr7)
		s.ReqOKAddTempUnsanct(15, "addr8", addr8)

		var actual []*sanction.TemporaryEntry
		testFunc := func() {
			actual = s.Keeper.GetAllTemporaryEntries(s.SdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllTemporaryEntries")
		s.Assert().Equal(expected, actual, "GetAllTemporaryEntries result")
	})

	s.Run("after some entries are deleted", func() {
		// Note: There should several entries by now. Delete all the addr1 entries and gov prop 3 entries.
		expected := []*sanction.TemporaryEntry{
			{Address: addr2.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_UNSANCTIONED},
			{Address: addr2.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_SANCTIONED},

			{Address: addr3.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_SANCTIONED},

			{Address: addr4.String(), ProposalId: 1, Status: sanction.TEMP_STATUS_UNSANCTIONED},

			{Address: addr5.String(), ProposalId: 2, Status: sanction.TEMP_STATUS_SANCTIONED},
			{Address: addr5.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_SANCTIONED},
			{Address: addr5.String(), ProposalId: 6, Status: sanction.TEMP_STATUS_SANCTIONED},

			{Address: addr6.String(), ProposalId: 4, Status: sanction.TEMP_STATUS_UNSANCTIONED},
			{Address: addr6.String(), ProposalId: 5, Status: sanction.TEMP_STATUS_UNSANCTIONED},

			{Address: addr7.String(), ProposalId: 10, Status: sanction.TEMP_STATUS_SANCTIONED},

			{Address: addr8.String(), ProposalId: 15, Status: sanction.TEMP_STATUS_UNSANCTIONED},
		}

		s.ReqOKDelAddrTemp("addr1", addr1)
		s.ReqOKDelPropTemp(3)

		var actual []*sanction.TemporaryEntry
		testFunc := func() {
			actual = s.Keeper.GetAllTemporaryEntries(s.SdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllTemporaryEntries")
		s.Assert().Equal(expected, actual, "GetAllTemporaryEntries result")
	})
}
