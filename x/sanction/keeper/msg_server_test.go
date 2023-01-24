package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

type MsgServerTestSuite struct {
	BaseTestSuite

	quotedAuthority string
}

func (s *MsgServerTestSuite) SetupTest() {
	s.BaseSetup()

	s.quotedAuthority = `"` + s.Keeper.GetAuthority() + `"`
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) TestKeeper_Sanction() {
	addr1 := sdk.AccAddress("1_addr_sanction_test")
	addr2 := sdk.AccAddress("2_addr_sanction_test")
	addr3 := sdk.AccAddress("3_addr_sanction_test")
	addr4 := sdk.AccAddress("4_addr_sanction_test")
	addr5 := sdk.AccAddress("5_addr_sanction_test")
	addr6 := sdk.AccAddress("6_addr_sanction_test")

	tests := []struct {
		name     string
		iniState *sanction.GenesisState
		req      *sanction.MsgSanction
		expErr   []string
		expState *sanction.GenesisState
	}{
		{
			name: "empty authority",
			req: &sanction.MsgSanction{
				Addresses: nil,
				Authority: "",
			},
			expErr: []string{"expected gov account as only signer for proposal message", `""`, s.quotedAuthority},
		},
		{
			name: "wrong authority quoted",
			req: &sanction.MsgSanction{
				Addresses: nil,
				Authority: s.quotedAuthority,
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"\"` + s.Keeper.GetAuthority() + `\""`, s.quotedAuthority},
		},
		{
			name: "wrong authority space at end",
			req: &sanction.MsgSanction{
				Addresses: nil,
				Authority: s.Keeper.GetAuthority() + " ",
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"` + s.Keeper.GetAuthority() + ` "`, s.quotedAuthority},
		},
		{
			name: "wrong authority space at front",
			req: &sanction.MsgSanction{
				Addresses: nil,
				Authority: " " + s.Keeper.GetAuthority(),
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`" ` + s.Keeper.GetAuthority() + `"`, s.quotedAuthority},
		},
		{
			name: "wrong authority missing first char",
			req: &sanction.MsgSanction{
				Addresses: nil,
				Authority: s.Keeper.GetAuthority()[1:],
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"` + s.Keeper.GetAuthority()[1:] + `"`, s.quotedAuthority},
		},
		{
			name: "wrong authority missing last char",
			req: &sanction.MsgSanction{
				Addresses: nil,
				Authority: s.Keeper.GetAuthority()[:len(s.Keeper.GetAuthority())-1],
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"` + s.Keeper.GetAuthority()[:len(s.Keeper.GetAuthority())-1] + `"`, s.quotedAuthority},
		},
		{
			name: "six addrs invalid first",
			req: &sanction.MsgSanction{
				Addresses: []string{
					"notanaddr",
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expErr:   []string{"invalid address", "invalid address[0]", "decoding bech32 failed"},
			expState: sanction.DefaultGenesisState(),
		},
		{
			name: "six addrs invalid third",
			req: &sanction.MsgSanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					"addrnogood",
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expErr:   []string{"invalid address", "invalid address[2]", "decoding bech32 failed"},
			expState: sanction.DefaultGenesisState(),
		},
		{
			name: "six addrs invalid sixth",
			req: &sanction.MsgSanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					"badaddrbad",
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expErr:   []string{"invalid address", "invalid address[5]", "decoding bech32 failed"},
			expState: sanction.DefaultGenesisState(),
		},
		{
			name: "six new addrs",
			req: &sanction.MsgSanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				TemporaryEntries: nil,
			},
		},
		{
			name: "temp entries cleared for addr",
			iniState: &sanction.GenesisState{
				Params:              sanction.DefaultParams(),
				SanctionedAddresses: nil,
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, true),
					newTempEntry(addr3, 1, true),
					newTempEntry(addr3, 2, false),
					newTempEntry(addr3, 3, false),
					newTempEntry(addr6, 1, true),
				},
			},
			req: &sanction.MsgSanction{
				Addresses: []string{addr3.String()},
				Authority: s.Keeper.GetAuthority(),
			},
			expState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr3.String(),
				},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, true),
					newTempEntry(addr6, 1, true),
				},
			},
		},
		{
			name: "already sanctioned addr",
			iniState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr4.String(),
				},
				TemporaryEntries: nil,
			},
			req: &sanction.MsgSanction{
				Addresses: []string{addr4.String()},
				Authority: s.Keeper.GetAuthority(),
			},
			expState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr4.String(),
				},
				TemporaryEntries: nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ClearState()
			if tc.iniState != nil {
				s.Require().NotPanics(func() {
					s.Keeper.InitGenesis(s.SdkCtx, tc.iniState)
				}, "InitGenesis")
			}

			var err error
			testFunc := func() {
				_, err = s.Keeper.Sanction(s.StdlibCtx, tc.req)
			}
			s.Require().NotPanics(testFunc, "Sanction")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "Sanction error")
			if tc.expState != nil {
				s.ExportAndCheck(tc.expState)
			}
		})
	}

}

func (s *MsgServerTestSuite) TestKeeper_Unsanction() {
	addr1 := sdk.AccAddress("1_addr_unsanction_test")
	addr2 := sdk.AccAddress("2_addr_unsanction_test")
	addr3 := sdk.AccAddress("3_addr_unsanction_test")
	addr4 := sdk.AccAddress("4_addr_unsanction_test")
	addr5 := sdk.AccAddress("5_addr_unsanction_test")
	addr6 := sdk.AccAddress("6_addr_unsanction_test")

	tests := []struct {
		name     string
		iniState *sanction.GenesisState
		req      *sanction.MsgUnsanction
		expErr   []string
		expState *sanction.GenesisState
	}{
		{
			name: "empty authority",
			req: &sanction.MsgUnsanction{
				Addresses: nil,
				Authority: "",
			},
			expErr: []string{"expected gov account as only signer for proposal message", `""`, s.quotedAuthority},
		},
		{
			name: "wrong authority quoted",
			req: &sanction.MsgUnsanction{
				Addresses: nil,
				Authority: s.quotedAuthority,
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"\"` + s.Keeper.GetAuthority() + `\""`, s.quotedAuthority},
		},
		{
			name: "wrong authority space at end",
			req: &sanction.MsgUnsanction{
				Addresses: nil,
				Authority: s.Keeper.GetAuthority() + " ",
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"` + s.Keeper.GetAuthority() + ` "`, s.quotedAuthority},
		},
		{
			name: "wrong authority space at front",
			req: &sanction.MsgUnsanction{
				Addresses: nil,
				Authority: " " + s.Keeper.GetAuthority(),
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`" ` + s.Keeper.GetAuthority() + `"`, s.quotedAuthority},
		},
		{
			name: "wrong authority missing first char",
			req: &sanction.MsgUnsanction{
				Addresses: nil,
				Authority: s.Keeper.GetAuthority()[1:],
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"` + s.Keeper.GetAuthority()[1:] + `"`, s.quotedAuthority},
		},
		{
			name: "wrong authority missing last char",
			req: &sanction.MsgUnsanction{
				Addresses: nil,
				Authority: s.Keeper.GetAuthority()[:len(s.Keeper.GetAuthority())-1],
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"` + s.Keeper.GetAuthority()[:len(s.Keeper.GetAuthority())-1] + `"`, s.quotedAuthority},
		},
		{
			name: "six addrs invalid first",
			iniState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
			},
			req: &sanction.MsgUnsanction{
				Addresses: []string{
					"notanaddr",
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expErr: []string{"invalid address", "invalid address[0]", "decoding bech32 failed"},
			expState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
			},
		},
		{
			name: "six addrs invalid third",
			iniState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
			},
			req: &sanction.MsgUnsanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					"addrnogood",
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expErr: []string{"invalid address", "invalid address[2]", "decoding bech32 failed"},
			expState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
			},
		},
		{
			name: "six addrs invalid sixth",
			iniState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
			},
			req: &sanction.MsgUnsanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					"badaddrbad",
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expErr: []string{"invalid address", "invalid address[5]", "decoding bech32 failed"},
			expState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
			},
		},
		{
			name: "six previously sanctioned addrs",
			req: &sanction.MsgUnsanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expState: &sanction.GenesisState{
				Params:              sanction.DefaultParams(),
				SanctionedAddresses: nil,
				TemporaryEntries:    nil,
			},
		},
		{
			name: "temp entries cleared for addr",
			iniState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
				SanctionedAddresses: []string{
					addr3.String(),
				},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, true),
					newTempEntry(addr3, 1, true),
					newTempEntry(addr3, 2, false),
					newTempEntry(addr3, 3, false),
					newTempEntry(addr6, 1, true),
				},
			},
			req: &sanction.MsgUnsanction{
				Addresses: []string{addr3.String()},
				Authority: s.Keeper.GetAuthority(),
			},
			expState: &sanction.GenesisState{
				Params:              sanction.DefaultParams(),
				SanctionedAddresses: nil,
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, 1, true),
					newTempEntry(addr6, 1, true),
				},
			},
		},
		{
			name: "not sanctioned addr",
			req: &sanction.MsgUnsanction{
				Addresses: []string{addr4.String()},
				Authority: s.Keeper.GetAuthority(),
			},
			expState: sanction.DefaultGenesisState(),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ClearState()
			if tc.iniState != nil {
				s.Require().NotPanics(func() {
					s.Keeper.InitGenesis(s.SdkCtx, tc.iniState)
				}, "InitGenesis")
			}

			var err error
			testFunc := func() {
				_, err = s.Keeper.Unsanction(s.StdlibCtx, tc.req)
			}
			s.Require().NotPanics(testFunc, "Unsanction")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "Unsanction error")
			if tc.expState != nil {
				s.ExportAndCheck(tc.expState)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestKeeper_UpdateParams() {
	origMinSanct := sanction.DefaultImmediateSanctionMinDeposit
	origMinUnsanct := sanction.DefaultImmediateUnsanctionMinDeposit
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origMinSanct
		sanction.DefaultImmediateUnsanctionMinDeposit = origMinUnsanct
	}()
	sanction.DefaultImmediateSanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("spcoin", 15))
	sanction.DefaultImmediateUnsanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("upcoin", 17))

	tests := []struct {
		name     string
		iniState *sanction.GenesisState
		req      *sanction.MsgUpdateParams
		expErr   []string
		expState *sanction.GenesisState
	}{
		{
			name: "empty authority",
			req: &sanction.MsgUpdateParams{
				Authority: "",
			},
			expErr: []string{"expected gov account as only signer for proposal message", `""`, s.quotedAuthority},
		},
		{
			name: "wrong authority quoted",
			req: &sanction.MsgUpdateParams{
				Authority: s.quotedAuthority,
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"\"` + s.Keeper.GetAuthority() + `\""`, s.quotedAuthority},
		},
		{
			name: "wrong authority space at end",
			req: &sanction.MsgUpdateParams{
				Authority: s.Keeper.GetAuthority() + " ",
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"` + s.Keeper.GetAuthority() + ` "`, s.quotedAuthority},
		},
		{
			name: "wrong authority space at front",
			req: &sanction.MsgUpdateParams{
				Authority: " " + s.Keeper.GetAuthority(),
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`" ` + s.Keeper.GetAuthority() + `"`, s.quotedAuthority},
		},
		{
			name: "wrong authority missing first char",
			req: &sanction.MsgUpdateParams{
				Authority: s.Keeper.GetAuthority()[1:],
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"` + s.Keeper.GetAuthority()[1:] + `"`, s.quotedAuthority},
		},
		{
			name: "wrong authority missing last char",
			req: &sanction.MsgUpdateParams{
				Authority: s.Keeper.GetAuthority()[:len(s.Keeper.GetAuthority())-1],
			},
			expErr: []string{"expected gov account as only signer for proposal message",
				`"` + s.Keeper.GetAuthority()[:len(s.Keeper.GetAuthority())-1] + `"`, s.quotedAuthority},
		},
		{
			name: "invalid params",
			req: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.Coins{sdk.Coin{Denom: "a", Amount: sdk.NewInt(5)}},
					ImmediateUnsanctionMinDeposit: nil,
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expErr:   []string{"invalid params", "invalid immediate sanction min deposit", "invalid denom"},
			expState: sanction.DefaultGenesisState(),
		},
		{
			name: "params not previously set",
			req: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("doitcoin", 3)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("nahcoin", 5)),
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expErr: nil,
			expState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("doitcoin", 3)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("nahcoin", 5)),
				},
			},
		},
		{
			name: "params being overwritten",
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("oldscoin", 22)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("olducoin", 44)),
				},
			},
			req: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("doitcoin", 3)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("nahcoin", 5)),
				},
				Authority: s.Keeper.GetAuthority(),
			},
			expErr: nil,
			expState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("doitcoin", 3)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("nahcoin", 5)),
				},
			},
		},
		{
			name: "nil params over previously defined",
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("oldscoin", 22)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("olducoin", 44)),
				},
			},
			req: &sanction.MsgUpdateParams{
				Params:    nil,
				Authority: s.Keeper.GetAuthority(),
			},
			expErr: nil,
			expState: &sanction.GenesisState{
				Params: sanction.DefaultParams(),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ClearState()
			if tc.iniState != nil {
				s.Require().NotPanics(func() {
					s.Keeper.InitGenesis(s.SdkCtx, tc.iniState)
				}, "InitGenesis")
			}

			var err error
			testFunc := func() {
				_, err = s.Keeper.UpdateParams(s.StdlibCtx, tc.req)
			}
			s.Require().NotPanics(testFunc, "UpdateParams")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "UpdateParams error")
			if tc.expState != nil {
				s.ExportAndCheck(tc.expState)
			}
		})
	}
}
