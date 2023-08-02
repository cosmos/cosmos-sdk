package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

type SendRestrictionTestSuite struct {
	BaseTestSuite
}

func (s *SendRestrictionTestSuite) SetupTest() {
	s.BaseSetup()
}

func TestSendRestrictionTestSuite(t *testing.T) {
	suite.Run(t, new(SendRestrictionTestSuite))
}

func (s *SendRestrictionTestSuite) TestSendRestrictionFn() {
	addrSanctioned := sdk.AccAddress("addrSanctioned______")
	addrUnsanctioned := sdk.AccAddress("addrUnsanctioned____")
	addrOther := sdk.AccAddress("addrOther___________")
	ctxWithBypass := sanction.WithBypass(s.SdkCtx)

	s.ReqOKAddPermSanct("addrSanctioned", addrSanctioned)
	s.ReqOKAddPermUnsanct("addrUnsanctioned", addrUnsanctioned)

	tests := []struct {
		name     string
		ctx      *sdk.Context
		fromAddr sdk.AccAddress
		toAddr   sdk.AccAddress
		amt      sdk.Coins
		expErr   []string
	}{
		{
			name:     "has bypass",
			ctx:      &ctxWithBypass,
			fromAddr: addrSanctioned,
			toAddr:   addrOther,
		},
		{
			name:     "from sanctioned address",
			fromAddr: addrSanctioned,
			toAddr:   addrOther,
			expErr:   []string{"account is sanctioned", "cannot send from " + addrSanctioned.String()},
		},
		{
			name:     "from unsanctioned address",
			fromAddr: addrUnsanctioned,
			toAddr:   addrOther,
		},
		{
			name:     "to sanctioned address",
			fromAddr: addrOther,
			toAddr:   addrSanctioned,
		},
		{
			name:     "to unsanctioned address",
			fromAddr: addrOther,
			toAddr:   addrUnsanctioned,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx := s.SdkCtx
			if tc.ctx != nil {
				ctx = *tc.ctx
			}
			var expNewTo sdk.AccAddress
			if len(tc.expErr) == 0 {
				expNewTo = tc.toAddr
			}

			newTo, err := s.Keeper.SendRestrictionFn(ctx, tc.fromAddr, tc.toAddr, tc.amt)
			s.AssertErrorContents(err, tc.expErr, "SendRestrictionFn error")
			s.Assert().Equal(expNewTo, newTo, "SendRestrictionFn returned address")
		})
	}
}

func (s *SendRestrictionTestSuite) TestBankSendCoinsUsesSendRestrictionFn() {
	// This specifically does NOT mock the bank keeper because it's testing
	// that the bank keeper is applying this module's send restriction.

	denom := "greatcoin"
	cz := func(amt int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(denom, amt))
	}

	sanctionedAddr := sdk.AccAddress("sanctionedAddr______")
	otherAddr := sdk.AccAddress("otherAddr___________")

	// Fund the addresses
	s.Require().NoError(testutil.FundAccount(s.App.BankKeeper, s.SdkCtx, sanctionedAddr, cz(1_000_005_000_000_000)), "FundAccount sanctionedAddr")
	s.Require().NoError(testutil.FundAccount(s.App.BankKeeper, s.SdkCtx, otherAddr, cz(3_000)), "FundAccount otherAddr")

	// Sanction the account.
	s.ReqOKAddPermSanct("sanctionedAddr", sanctionedAddr)

	s.Run("SendCoins from sanctioned addr returns error", func() {
		ctx, writeCache := s.SdkCtx.CacheContext()
		expErr := "cannot send from " + sanctionedAddr.String() + ": account is sanctioned"
		err := s.App.BankKeeper.SendCoins(ctx, sanctionedAddr, otherAddr, cz(5_000_000_000))
		s.Assert().EqualError(err, expErr, "SendCoins from sanctioned address error")
		if err == nil {
			writeCache()
		}
	})

	s.Run("SendCoins to sanctioned addr does not return an error", func() {
		err := s.App.BankKeeper.SendCoins(s.SdkCtx, otherAddr, sanctionedAddr, cz(3_000))
		s.Assert().NoError(err, "SendCoins to sanctioned address error")
	})

	s.Run("sanctioned address has expected balance", func() {
		bal := s.App.BankKeeper.GetBalance(s.SdkCtx, sanctionedAddr, denom)
		s.Assert().Equal(cz(1_000_005_000_003_000).String(), bal.String(), "GetBalance sanctionedAddr")
	})

	s.Run("other address has expected balance", func() {
		bal := s.App.BankKeeper.GetBalance(s.SdkCtx, otherAddr, denom)
		s.Assert().Equal("0"+denom, bal.String(), "GetBalance otherAddr")
	})
}

func (s *SendRestrictionTestSuite) TestBankInputOutputCoinsUsesSendRestrictionFn() {
	// This specifically does NOT mock the bank keeper because it's testing
	// that the bank keeper is applying this module's send restriction.

	denom := "goodcoin"
	cz := func(amt int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(denom, amt))
	}

	sanctionedAddr := sdk.AccAddress("sanctionedAddr______")
	otherAddr1 := sdk.AccAddress("otherAddr1__________")
	otherAddr2 := sdk.AccAddress("otherAddr2__________")
	otherAddr3 := sdk.AccAddress("otherAddr3__________")

	// Fund the addresses
	s.Require().NoError(testutil.FundAccount(s.App.BankKeeper, s.SdkCtx, sanctionedAddr, cz(6_006)), "FundAccount sanctionedAddr")
	s.Require().NoError(testutil.FundAccount(s.App.BankKeeper, s.SdkCtx, otherAddr1, cz(1)), "FundAccount otherAddr1")
	s.Require().NoError(testutil.FundAccount(s.App.BankKeeper, s.SdkCtx, otherAddr2, cz(2)), "FundAccount otherAddr2")
	s.Require().NoError(testutil.FundAccount(s.App.BankKeeper, s.SdkCtx, otherAddr3, cz(3)), "FundAccount otherAddr3")

	// Sanction the account.
	s.ReqOKAddPermSanct("sanctionedAddr", sanctionedAddr)

	// Do an InputOutputCoins from the sanctioned address to the others.
	inputs := []banktypes.Input{{Address: sanctionedAddr.String(), Coins: cz(6_000)}}
	outputs := []banktypes.Output{
		{Address: otherAddr1.String(), Coins: cz(1_000)},
		{Address: otherAddr2.String(), Coins: cz(2_000)},
		{Address: otherAddr3.String(), Coins: cz(3_000)},
	}
	err := s.App.BankKeeper.InputOutputCoins(s.SdkCtx, inputs, outputs)

	s.Run("error is as expected", func() {
		exp := "cannot send from " + sanctionedAddr.String() + ": account is sanctioned"
		s.Assert().EqualError(err, exp, "InputOutputCoins")
	})

	// Note: In InputOutputCoins, the funds are removed from the input before calling the restriction function.
	//       This is okay because it's usually being called in a transaction where an error will cause a rollback.
	//       Rather than having a test that passes, but technically contrary to desired behavior,
	//       the input balance is just not checked.

	expBals := []struct {
		name string
		addr sdk.AccAddress
		exp  sdk.Coins
	}{
		// not checking input balance (see comment above).
		{name: "funds not removed from output[0]", addr: otherAddr1, exp: cz(1)},
		{name: "funds not removed from output[1]", addr: otherAddr2, exp: cz(2)},
		{name: "funds not removed from output[2]", addr: otherAddr3, exp: cz(3)},
	}

	for _, tc := range expBals {
		s.Run(tc.name, func() {
			bal := s.App.BankKeeper.GetBalance(s.SdkCtx, tc.addr, denom)
			s.Assert().Equal(tc.exp.String(), bal.String(), "GetBalance")
		})
	}
}
