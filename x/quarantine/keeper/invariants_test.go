package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	"github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

// These tests are initiated by TestKeeperTestSuite in keeper_test.go

func (s *TestSuite) TestFundsHolderBalanceInvariantHelper() {
	bk := NewMockBankKeeper()
	qk := s.keeper.WithBankKeeper(bk)

	s.Run("no quarantined funds no funds in holding account", func() {
		zqk := qk.WithFundsHolder(testutil.MakeTestAddr("fh", 0))
		msg, broken := keeper.FundsHolderBalanceInvariantHelper(s.sdkCtx, zqk)
		s.Assert().False(broken, "fundsHolderBalanceInvariantHelper broken")
		s.Assert().Equal("total funds quarantined is zero", msg, "fundsHolderBalanceInvariantHelper message")
	})

	dummyAddr0 := testutil.MakeTestAddr("dumfhbi", 0)
	dummyAddr1 := testutil.MakeTestAddr("dumfhbi", 1)
	dummyAddr2 := testutil.MakeTestAddr("dumfhbi", 2)
	dummyAddr3 := testutil.MakeTestAddr("dumfhbi", 3)
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, s.cz("4acoin"), dummyAddr0, dummyAddr1), "AddQuarantinedCoins 4acoin")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, s.cz("1acoin,9bcoin"), dummyAddr0, dummyAddr2), "AddQuarantinedCoins 1acoin,9bcoin")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, s.cz("2acoin,2bcoin,2ccoin"), dummyAddr0, dummyAddr3), "AddQuarantinedCoins 2acoin,2bcoin,2ccoin")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, s.cz("3acoin,11ccoin"), dummyAddr1, dummyAddr2), "AddQuarantinedCoins 3acoin,11ccoin")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, s.cz("5bcoin"), dummyAddr1, dummyAddr3), "AddQuarantinedCoins 5bcoin")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, s.cz("4bcoin,7ccoin"), dummyAddr2, dummyAddr3), "AddQuarantinedCoins 4bcoin,7ccoin")
	s.Require().NoError(s.keeper.AddQuarantinedCoins(s.sdkCtx, s.cz("10ccoin"), dummyAddr3, dummyAddr2, dummyAddr1, dummyAddr0), "AddQuarantinedCoins 10ccoin")
	// Quarantine records total: 10acoin,20bcoin,30ccoin

	makeFundedAddr := func(base string, coins sdk.Coins) sdk.AccAddress {
		addr := testutil.MakeTestAddr(base, 0)
		bk.AllBalances[string(addr)] = coins
		return addr
	}
	makeSufExpMsg := func(addr sdk.AccAddress) string {
		return "quarantine funds holder account " + addr.String() + " balance sufficient" +
			", have: " + bk.AllBalances[string(addr)].String() +
			", need: 10acoin,20bcoin,30ccoin"
	}
	makeInsufExpMsg := func(addr sdk.AccAddress, problems string) string {
		return "quarantine funds holder account " + addr.String() + " balance insufficient" +
			", have: " + bk.AllBalances[string(addr)].String() +
			", need: 10acoin,20bcoin,30ccoin" +
			", " + problems +
			": insufficient funds"
	}

	exact := makeFundedAddr("exact", s.cz("10acoin,20bcoin,30ccoin"))
	plenty := makeFundedAddr("plenty", s.cz("100acoin,200bcoin,300ccoin,400dcoin,500ecoin,600fcoin"))
	shorta := makeFundedAddr("ashort", s.cz("9acoin,20bcoin,30ccoin"))
	shortb := makeFundedAddr("bshort", s.cz("10acoin,19bcoin,30ccoin"))
	shortc := makeFundedAddr("cshort", s.cz("10acoin,20bcoin,29ccoin"))
	shortab := makeFundedAddr("abshort", s.cz("9acoin,19bcoin,30ccoin"))
	shortac := makeFundedAddr("acshort", s.cz("9acoin,20bcoin,29ccoin"))
	shortbc := makeFundedAddr("bcshort", s.cz("10acoin,19bcoin,29ccoin"))
	shortabc := makeFundedAddr("abcshort", s.cz("9acoin,19bcoin,29ccoin"))
	noa := makeFundedAddr("ano", s.cz("20bcoin,30ccoin"))
	nob := makeFundedAddr("bno", s.cz("10acoin,30ccoin"))
	noc := makeFundedAddr("cno", s.cz("10acoin,20bcoin"))
	noab := makeFundedAddr("abno", s.cz("30ccoin"))
	noac := makeFundedAddr("acno", s.cz("20bcoin"))
	nobc := makeFundedAddr("bcno", s.cz("10acoin"))
	noabc := makeFundedAddr("abcno", s.cz("40dcoin,50ecoin,60fcoin"))
	zero := makeFundedAddr("zero", sdk.Coins{})

	tests := []struct {
		name      string
		keeper    keeper.Keeper
		expMsg    string
		expBroken bool
	}{
		{
			name:      "exactly funded",
			keeper:    qk.WithFundsHolder(exact),
			expMsg:    makeSufExpMsg(exact),
			expBroken: false,
		},
		{
			name:      "overly funded",
			keeper:    qk.WithFundsHolder(plenty),
			expMsg:    makeSufExpMsg(plenty),
			expBroken: false,
		},
		{
			name:      "short a",
			keeper:    qk.WithFundsHolder(shorta),
			expMsg:    makeInsufExpMsg(shorta, "9acoin is less than 10acoin"),
			expBroken: true,
		},
		{
			name:      "short b",
			keeper:    qk.WithFundsHolder(shortb),
			expMsg:    makeInsufExpMsg(shortb, "19bcoin is less than 20bcoin"),
			expBroken: true,
		},
		{
			name:      "short c",
			keeper:    qk.WithFundsHolder(shortc),
			expMsg:    makeInsufExpMsg(shortc, "29ccoin is less than 30ccoin"),
			expBroken: true,
		},
		{
			name:      "short a b",
			keeper:    qk.WithFundsHolder(shortab),
			expMsg:    makeInsufExpMsg(shortab, "9acoin is less than 10acoin, 19bcoin is less than 20bcoin"),
			expBroken: true,
		},
		{
			name:      "short a c",
			keeper:    qk.WithFundsHolder(shortac),
			expMsg:    makeInsufExpMsg(shortac, "9acoin is less than 10acoin, 29ccoin is less than 30ccoin"),
			expBroken: true,
		},
		{
			name:      "short b c",
			keeper:    qk.WithFundsHolder(shortbc),
			expMsg:    makeInsufExpMsg(shortbc, "19bcoin is less than 20bcoin, 29ccoin is less than 30ccoin"),
			expBroken: true,
		},
		{
			name:      "short a b c",
			keeper:    qk.WithFundsHolder(shortabc),
			expMsg:    makeInsufExpMsg(shortabc, "9acoin is less than 10acoin, 19bcoin is less than 20bcoin, 29ccoin is less than 30ccoin"),
			expBroken: true,
		},
		{
			name:      "no a",
			keeper:    qk.WithFundsHolder(noa),
			expMsg:    makeInsufExpMsg(noa, "zero is less than 10acoin"),
			expBroken: true,
		},
		{
			name:      "no b",
			keeper:    qk.WithFundsHolder(nob),
			expMsg:    makeInsufExpMsg(nob, "zero is less than 20bcoin"),
			expBroken: true,
		},
		{
			name:      "no c",
			keeper:    qk.WithFundsHolder(noc),
			expMsg:    makeInsufExpMsg(noc, "zero is less than 30ccoin"),
			expBroken: true,
		},
		{
			name:      "no a b",
			keeper:    qk.WithFundsHolder(noab),
			expMsg:    makeInsufExpMsg(noab, "zero is less than 10acoin, zero is less than 20bcoin"),
			expBroken: true,
		},
		{
			name:      "no a c",
			keeper:    qk.WithFundsHolder(noac),
			expMsg:    makeInsufExpMsg(noac, "zero is less than 10acoin, zero is less than 30ccoin"),
			expBroken: true,
		},
		{
			name:      "no b c",
			keeper:    qk.WithFundsHolder(nobc),
			expMsg:    makeInsufExpMsg(nobc, "zero is less than 20bcoin, zero is less than 30ccoin"),
			expBroken: true,
		},
		{
			name:      "no a b c",
			keeper:    qk.WithFundsHolder(noabc),
			expMsg:    makeInsufExpMsg(noabc, "zero is less than 10acoin, zero is less than 20bcoin, zero is less than 30ccoin"),
			expBroken: true,
		},
		{
			name:   "zero",
			keeper: qk.WithFundsHolder(zero),
			expMsg: "quarantine funds holder account " + zero.String() + " balance insufficient" +
				", have: zero balance" +
				", need: 10acoin,20bcoin,30ccoin" +
				", zero is less than 10acoin, zero is less than 20bcoin, zero is less than 30ccoin" +
				": insufficient funds",
			expBroken: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msg, broken := keeper.FundsHolderBalanceInvariantHelper(s.sdkCtx, tc.keeper)
			s.Assert().Equal(tc.expBroken, broken, "fundsHolderBalanceInvariantHelper broken")
			s.Assert().Equal(tc.expMsg, msg, "fundsHolderBalanceInvariantHelper message")
		})
	}
}
