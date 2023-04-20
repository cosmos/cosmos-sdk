package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
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
			expErr:   []string{"account is sanctioned", addrSanctioned.String()},
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

func (s *SendRestrictionTestSuite) TestBankSendUsesSendRestrictionFn() {
	// This specifically does NOT mock the bank keeper because it's testing
	// that the bank keeper is paying attention to quarantine.

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
		expErr := sanctionedAddr.String() + ": account is sanctioned"
		err := s.App.BankKeeper.SendCoins(s.SdkCtx, sanctionedAddr, otherAddr, cz(5_000_000_000))
		s.Assert().EqualError(err, expErr, "SendCoins from sanctioned address error")
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
