package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
)

// These tests are initiated by TestKeeperTestSuite in keeper_test.go

func (s *TestSuite) TestSendRestrictionFn() {
	fundsHolder := s.keeper.GetFundsHolder()
	keeperWithoutFundsHolder := s.keeper.WithFundsHolder(nil)
	ctxWithBypass := quarantine.WithBypass(s.sdkCtx)

	cz := func(amt string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(amt)
		if err != nil {
			panic(err)
		}
		return rv
	}

	// addr1 opted in: auto-accept from addr2, auto-decline from addr3, unspecified from addr4.
	// addr5 NOT opted in, but has the same auto-responses defined as addr1.
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr1), "SetOptIn addr1")
	s.keeper.SetAutoResponse(s.sdkCtx, s.addr1, s.addr2, quarantine.AUTO_RESPONSE_ACCEPT)
	s.keeper.SetAutoResponse(s.sdkCtx, s.addr1, s.addr3, quarantine.AUTO_RESPONSE_DECLINE)
	s.keeper.SetAutoResponse(s.sdkCtx, s.addr1, s.addr4, quarantine.AUTO_RESPONSE_UNSPECIFIED)
	s.keeper.SetAutoResponse(s.sdkCtx, s.addr5, s.addr2, quarantine.AUTO_RESPONSE_ACCEPT)
	s.keeper.SetAutoResponse(s.sdkCtx, s.addr5, s.addr3, quarantine.AUTO_RESPONSE_DECLINE)
	s.keeper.SetAutoResponse(s.sdkCtx, s.addr5, s.addr4, quarantine.AUTO_RESPONSE_UNSPECIFIED)

	tests := []struct {
		name          string
		keeper        *keeper.Keeper
		ctx           *sdk.Context
		fromAddr      sdk.AccAddress
		toAddr        sdk.AccAddress
		amt           sdk.Coins
		expErr        []string
		expQuarantine bool
	}{
		{
			name:          "has bypass",
			ctx:           &ctxWithBypass,
			fromAddr:      s.addr4,
			toAddr:        s.addr1,
			amt:           cz("10acorns"),
			expQuarantine: false,
		},
		{
			name:          "from equals to",
			fromAddr:      s.addr2,
			toAddr:        s.addr2,
			amt:           cz("11bcorns"),
			expQuarantine: false,
		},
		{
			name:          "from equals funds holder",
			fromAddr:      fundsHolder,
			toAddr:        s.addr2,
			amt:           cz("12ccorns"),
			expQuarantine: false,
		},
		{
			name:          "to equals funds holder",
			fromAddr:      s.addr5,
			toAddr:        fundsHolder,
			amt:           cz("20kcorns"),
			expQuarantine: false,
		},
		{
			name:          "to address is not quarantined",
			fromAddr:      s.addr4,
			toAddr:        s.addr5,
			amt:           cz("13dcorns"),
			expQuarantine: false,
		},
		{
			name:          "to address is not quarantined but has auto-accept",
			fromAddr:      s.addr2,
			toAddr:        s.addr5,
			amt:           cz("14ecorns"),
			expQuarantine: false,
		},
		{
			name:          "to address is not quarantined but has auto-decline",
			fromAddr:      s.addr3,
			toAddr:        s.addr5,
			amt:           cz("15fcorns"),
			expQuarantine: false,
		},
		{
			name:          "to address is quarantined with auto-accept",
			fromAddr:      s.addr2,
			toAddr:        s.addr1,
			amt:           cz("16gcorns"),
			expQuarantine: false,
		},
		{
			name:          "to address is quarantined with auto-decline",
			fromAddr:      s.addr3,
			toAddr:        s.addr1,
			amt:           cz("17hcorns"),
			expQuarantine: true,
		},
		{
			name:          "to address is quarantined with no auto-response",
			fromAddr:      s.addr4,
			toAddr:        s.addr1,
			amt:           cz("18icorns"),
			expQuarantine: true,
		},
		{
			name:     "No quarantine funds holder",
			keeper:   &keeperWithoutFundsHolder,
			fromAddr: s.addr4,
			toAddr:   s.addr1,
			amt:      cz("19jcorns"),
			expErr:   []string{"no quarantine funds holder account defined", "unknown address"},
		},
		// AddQuarantinedCoins returns error
		//    As of writing this, the only reasons AddQuarantinedCoins returns an error is if
		//    the funds are already fully accepted, or if there's an error emitting an event.
		//    Neither are possible to trigger from here.
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			k := s.keeper
			if tc.keeper != nil {
				k = *tc.keeper
			}
			ctx := s.sdkCtx
			if tc.ctx != nil {
				ctx = *tc.ctx
			}

			expNewTo := tc.toAddr
			switch {
			case len(tc.expErr) != 0:
				expNewTo = nil
			case tc.expQuarantine:
				expNewTo = fundsHolder
			}

			newToAddr, err := k.SendRestrictionFn(ctx, tc.fromAddr, tc.toAddr, tc.amt)
			s.AssertErrorContents(err, tc.expErr, "SendRestrictionFn error")
			s.Assert().Equal(expNewTo, newToAddr, "SendRestrictionFn returned address")

			if tc.expQuarantine {
				qReq := s.keeper.GetQuarantineRecord(s.sdkCtx, tc.toAddr, tc.fromAddr)
				if s.Assert().NotNil(qReq, "GetQuarantineRecord") {
					qFunds := qReq.Coins
					s.Assert().Equal(tc.amt, qFunds, "amount quarantined")
					// Clear the record just in case a later tests uses the same addresses.
					qReq.AcceptedFromAddresses = append(qReq.AcceptedFromAddresses, qReq.UnacceptedFromAddresses...)
					qReq.UnacceptedFromAddresses = nil
					s.keeper.SetQuarantineRecord(s.sdkCtx, tc.toAddr, qReq)
				}
			}
		})
	}
}

func (s *TestSuite) TestBankSendUsesSendRestrictionFn() {
	// This specifically does NOT mock the bank keeper because it's testing
	// that the bank keeper is paying attention to quarantine.

	denom := "greatcoin"
	cz := func(amt int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(denom, amt))
	}

	// Set up addr1 to be quarantined.
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr1), "SetOptIn addr1")
	// Give addr2 some funds and send them to addr1.
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.sdkCtx, s.addr2, cz(888)), "FundAccount addr2 888%s", denom)

	// Do a Send from addr2 to addr1
	s.Require().NoError(s.app.BankKeeper.SendCoins(s.sdkCtx, s.addr2, s.addr1, cz(88)), "SendCoins 88%s from addr2 to addr1", denom)

	s.Run("funds do not go into addr1's account", func() {
		addr1Bal := s.app.BankKeeper.GetBalance(s.sdkCtx, s.addr1, denom)
		s.Assert().Equal("0"+denom, addr1Bal.String(), "addr1's balances")
	})

	s.Run("funds came out of addr2's account", func() {
		addr2Bal := s.app.BankKeeper.GetBalance(s.sdkCtx, s.addr2, denom)
		s.Assert().Equal("800"+denom, addr2Bal.String(), "addr2's balances")
	})

	s.Run("the funds holder account has them", func() {
		fundsHolderBal := s.app.BankKeeper.GetBalance(s.sdkCtx, s.keeper.GetFundsHolder(), denom)
		s.Assert().Equal("88"+denom, fundsHolderBal.String(), "quarantine funds holder balance")
	})

	s.Run("there's a record of the quarantined funds", func() {
		qReq := s.keeper.GetQuarantineRecord(s.sdkCtx, s.addr1, s.addr2)
		if s.Assert().NotNil(qReq, "GetQuarantineRecord to addr1 from addr2") {
			qCoins := qReq.Coins
			s.Assert().Equal("88"+denom, qCoins.String(), "amount quarantined")
		}
	})
}
