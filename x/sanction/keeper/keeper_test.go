package keeper_test

import (
	"bytes"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

type KeeperTestSuite struct {
	BaseTestSuite

	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
	addr3 sdk.AccAddress
	addr4 sdk.AccAddress
	addr5 sdk.AccAddress
}

func (s *KeeperTestSuite) SetupTest() {
	s.BaseSetup()

	addrs := simapp.AddTestAddrsIncremental(s.App, s.SdkCtx, 5, sdk.NewInt(1_000_000_000))
	s.addr1 = addrs[0]
	s.addr2 = addrs[1]
	s.addr3 = addrs[2]
	s.addr4 = addrs[3]
	s.addr5 = addrs[4]
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestKeeperMsgUrls() {
	vals := []struct {
		name string
		val  string
		exp  string
	}{
		{
			name: "msgSanctionTypeURL",
			val:  s.Keeper.OnlyTestsGetMsgSanctionTypeURL(),
			exp:  "/cosmos.sanction.v1beta1.MsgSanction",
		},
		{
			name: "msgUnsanctionTypeURL",
			val:  s.Keeper.OnlyTestsGetMsgUnsanctionTypeURL(),
			exp:  "/cosmos.sanction.v1beta1.MsgUnsanction",
		},
		{
			name: "msgExecLegacyContentTypeURL",
			val:  s.Keeper.OnlyTestsGetMsgExecLegacyContentTypeURL(),
			exp:  "/cosmos.gov.v1.MsgExecLegacyContent",
		},
	}

	for i, val := range vals {
		s.Run(val.name, func() {
			s.Assert().Equal(val.exp, val.val, "field value")

			for j, val2 := range vals {
				if i == j {
					continue
				}
				s.Assert().NotEqual(val.val, val2.val, "%q = %s = %s", val.val, val.name, val2.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_GetAuthority() {
	s.Run("default", func() {
		expected := authtypes.NewModuleAddress(govtypes.ModuleName).String()
		actual := s.Keeper.GetAuthority()
		s.Assert().Equal(expected, actual, "GetAuthority result")
	})

	tests := []string{"something", "something else"}
	for _, tc := range tests {
		s.Run(tc, func() {
			k := s.Keeper.OnlyTestsWithAuthority(tc)
			actual := k.GetAuthority()
			s.Assert().Equal(tc, actual, "GetAuthority result")
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_IsSanctionedAddr() {
	// Setup:
	// addr1 will be sanctioned.
	// addr2 will be sanctioned, but have a temp unsanction.
	// addr3 will have a temp sanction.
	// addr4 will have a temp sanction then temp unsanction.
	// addr5 will be sanctioned and have a temp unsanction then a temp sanction.
	// addrUnsanctionable will have a sanction in place, but be one of the unsanctionable addresses.
	addrUnsanctionable := sdk.AccAddress("unsanctionable_addr_")
	s.ReqOKAddPermSanct("s.addr1, s.addr2, s.addr5, addrUnsanctionable", s.addr1, s.addr2, s.addr5, addrUnsanctionable)
	s.ReqOKAddTempSanct(1, "s.addr3, s.addr4", s.addr3, s.addr4)
	s.ReqOKAddTempUnsanct(2, "s.addr2, s.addr4, s.addr5", s.addr2, s.addr4, s.addr5)
	s.ReqOKAddTempSanct(3, "s.addr5", s.addr5)

	k := s.Keeper.OnlyTestsWithUnsanctionableAddrs(map[string]bool{string(addrUnsanctionable): true})

	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  bool
	}{
		{
			name: "nil",
			addr: nil,
			exp:  false,
		},
		{
			name: "empty",
			addr: sdk.AccAddress{},
			exp:  false,
		},
		{
			name: "unknown address",
			addr: sdk.AccAddress("an__unknown__address"),
			exp:  false,
		},
		{
			name: "sanctioned addr",
			addr: s.addr1,
			exp:  true,
		},
		{
			name: "sanctioned with temp unsanction",
			addr: s.addr2,
			exp:  false,
		},
		{
			name: "temp sanction",
			addr: s.addr3,
			exp:  true,
		},
		{
			name: "temp sanction then temp unsanction",
			addr: s.addr4,
			exp:  false,
		},
		{
			name: "sanctioned with temp unsanction then temp sanction",
			addr: s.addr5,
			exp:  true,
		},
		{
			name: "first byte of sanctioned addr",
			addr: sdk.AccAddress{s.addr1[0]},
			exp:  false,
		},
		{
			name: "sanctioned addr plus 1 byte at end",
			addr: append(append([]byte{}, s.addr1...), 'f'),
			exp:  false,
		},
		{
			name: "sanctioned addr plus 1 byte at front",
			addr: append([]byte{'g'}, s.addr1...),
			exp:  false,
		},
		{
			name: "sanctioned addr minus last byte",
			addr: s.addr1[:len(s.addr1)-1],
			exp:  false,
		},
		{
			name: "sanctioned addr minus first byte",
			addr: s.addr1[1:],
			exp:  false,
		},
		{
			name: "sanctioned addr that is now unsanctionable",
			addr: addrUnsanctionable,
			exp:  false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var origAddr, origAtFullCap sdk.AccAddress
			if tc.addr != nil {
				origAddr = make(sdk.AccAddress, len(tc.addr), cap(tc.addr))
				copy(origAddr, tc.addr[:cap(tc.addr)])
				origAtFullCap = tc.addr[:cap(tc.addr)]
			}
			var actual bool
			testFunc := func() {
				actual = k.IsSanctionedAddr(s.SdkCtx, tc.addr)
			}
			s.Require().NotPanics(testFunc, "IsSanctionedAddr")
			s.Assert().Equal(tc.exp, actual, "IsSanctionedAddr result")
			s.Assert().Equal(origAddr, tc.addr, "provided addr before and after")
			s.Assert().Equal(cap(origAddr), cap(tc.addr), "provided addr capacity before and after")
			var addrAtFullCap sdk.AccAddress
			if tc.addr != nil {
				addrAtFullCap = tc.addr[:cap(tc.addr)]
			}
			s.Assert().Equal(origAtFullCap, addrAtFullCap, "provided addr at full capacity before and after")
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_SanctionAddresses() {
	makeEvents := func(addrs ...sdk.AccAddress) sdk.Events {
		rv := sdk.Events{}
		for _, addr := range addrs {
			event, err := sdk.TypedEventToEvent(sanction.NewEventAddressSanctioned(addr))
			s.Require().NoError(err, "TypedEventToEvent NewEventAddressSanctioned")
			rv = append(rv, event)
		}
		return rv
	}

	addrUnsanctionable := sdk.AccAddress("unsanctionable_addr_")
	k := s.Keeper.OnlyTestsWithUnsanctionableAddrs(map[string]bool{string(addrUnsanctionable): true})

	tests := []struct {
		name               string
		addrs              []sdk.AccAddress
		expEvents          sdk.Events
		expErr             []string
		checkSanctioned    []sdk.AccAddress
		checkNotSanctioned []sdk.AccAddress
	}{
		{
			name:               "no addresses",
			addrs:              []sdk.AccAddress{},
			expEvents:          sdk.Events{},
			checkSanctioned:    []sdk.AccAddress{},
			checkNotSanctioned: []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5, addrUnsanctionable},
		},
		{
			name:               "one addr",
			addrs:              []sdk.AccAddress{s.addr1},
			expEvents:          makeEvents(s.addr1),
			checkSanctioned:    []sdk.AccAddress{s.addr1},
			checkNotSanctioned: []sdk.AccAddress{s.addr2, s.addr3, s.addr4, s.addr5, addrUnsanctionable},
		},
		{
			name:               "already sanctioned addr",
			addrs:              []sdk.AccAddress{s.addr1},
			expEvents:          makeEvents(s.addr1),
			checkSanctioned:    []sdk.AccAddress{s.addr1},
			checkNotSanctioned: []sdk.AccAddress{s.addr2, s.addr3, s.addr4, s.addr5, addrUnsanctionable},
		},
		{
			name:               "two new addrs",
			addrs:              []sdk.AccAddress{s.addr2, s.addr3},
			expEvents:          makeEvents(s.addr2, s.addr3),
			checkSanctioned:    []sdk.AccAddress{s.addr1, s.addr2, s.addr3},
			checkNotSanctioned: []sdk.AccAddress{s.addr4, s.addr5, addrUnsanctionable},
		},
		{
			name:               "three addrs all already sanctioned",
			addrs:              []sdk.AccAddress{s.addr1, s.addr2, s.addr3},
			expEvents:          makeEvents(s.addr1, s.addr2, s.addr3),
			checkSanctioned:    []sdk.AccAddress{s.addr1, s.addr2, s.addr3},
			checkNotSanctioned: []sdk.AccAddress{s.addr4, s.addr5, addrUnsanctionable},
		},
		{
			name:               "two addrs one already sanctioned",
			addrs:              []sdk.AccAddress{s.addr1, s.addr4},
			expEvents:          makeEvents(s.addr1, s.addr4),
			checkSanctioned:    []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4},
			checkNotSanctioned: []sdk.AccAddress{s.addr5, addrUnsanctionable},
		},
		{
			name:               "unsanctionable addr",
			addrs:              []sdk.AccAddress{addrUnsanctionable},
			expEvents:          sdk.Events{},
			expErr:             []string{addrUnsanctionable.String(), "address cannot be sanctioned"},
			checkSanctioned:    []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4},
			checkNotSanctioned: []sdk.AccAddress{s.addr5, addrUnsanctionable},
		},
		{
			name:               "three addrs first unsanctionable",
			addrs:              []sdk.AccAddress{addrUnsanctionable, s.addr4, s.addr5},
			expEvents:          sdk.Events{},
			expErr:             []string{addrUnsanctionable.String(), "address cannot be sanctioned"},
			checkSanctioned:    []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4},
			checkNotSanctioned: []sdk.AccAddress{s.addr5, addrUnsanctionable},
		},
		{
			name:               "three addrs second unsanctionable",
			addrs:              []sdk.AccAddress{s.addr1, addrUnsanctionable, s.addr5},
			expEvents:          makeEvents(s.addr1),
			expErr:             []string{addrUnsanctionable.String(), "address cannot be sanctioned"},
			checkSanctioned:    []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4},
			checkNotSanctioned: []sdk.AccAddress{s.addr5, addrUnsanctionable},
		},
		{
			name:               "three addrs third unsanctionable",
			addrs:              []sdk.AccAddress{s.addr1, s.addr2, addrUnsanctionable},
			expEvents:          makeEvents(s.addr1, s.addr2),
			expErr:             []string{addrUnsanctionable.String(), "address cannot be sanctioned"},
			checkSanctioned:    []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4},
			checkNotSanctioned: []sdk.AccAddress{s.addr5, addrUnsanctionable},
		},
	}

	var isSanctioned bool
	testIsSanction := func(addr sdk.AccAddress) func() {
		return func() {
			isSanctioned = k.IsSanctionedAddr(s.SdkCtx, addr)
		}
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := s.SdkCtx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = k.SanctionAddresses(ctx, tc.addrs...)
			}
			s.Require().NotPanics(testFunc, "SanctionAddresses")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "SanctionAddresses error")
			events := em.Events()
			s.Assert().Equal(tc.expEvents, events, "events emitted during SanctionAddresses")
			for _, addr := range tc.checkSanctioned {
				if s.Assert().NotPanics(testIsSanction(addr), "IsSanctionedAddr") {
					s.Assert().True(isSanctioned, "IsSanctionedAddr result")
				}
			}
			for _, addr := range tc.checkNotSanctioned {
				if s.Assert().NotPanics(testIsSanction(addr), "IsSanctionedAddr") {
					s.Assert().False(isSanctioned, "IsSanctionedAddr result")
				}
			}
		})
	}

	s.Run("temp entries are deleted", func() {
		s.ReqOKAddTempSanct(1, "s.addr1, s.addr2", s.addr1, s.addr2)
		s.ReqOKAddTempSanct(2, "s.addr1, s.addr3", s.addr1, s.addr3)
		s.ReqOKAddTempUnsanct(3, "s.addr2, s.addr4, s.addr5", s.addr2, s.addr4, s.addr5)

		testFunc := func() error {
			return k.SanctionAddresses(s.SdkCtx, s.addr5, s.addr3, s.addr1, s.addr2, s.addr4)
		}
		s.RequireNotPanicsNoError(testFunc, "SanctionAddresses")

		tempEntries := s.GetAllTempEntries()
		s.Assert().Empty(tempEntries, "temporary entries still in the store")

		tempIndEntries := s.GetAllIndexTempEntries()
		s.Assert().Empty(tempIndEntries, "proposal index temporary entries still in the store")
	})
}

func (s *KeeperTestSuite) TestKeeper_UnsanctionAddresses() {
	makeEvents := func(addrs ...sdk.AccAddress) sdk.Events {
		rv := sdk.Events{}
		for _, addr := range addrs {
			event, err := sdk.TypedEventToEvent(sanction.NewEventAddressUnsanctioned(addr))
			s.Require().NoError(err, "TypedEventToEvent NewEventAddressUnsanctioned")
			rv = append(rv, event)
		}
		return rv
	}

	// Setup: Sanction all 5 addrs plus a new one that will end up being unsanctionable.
	addrUnsanctionable := sdk.AccAddress("unsanctionable_addr_")
	addrRandom := sdk.AccAddress("just_a_random_addr")
	s.ReqOKAddPermSanct("s.addr1, s.addr2, s.addr3, s.addr4, s.addr5, addrUnsanctionable",
		s.addr1, s.addr2, s.addr3, s.addr4, s.addr5, addrUnsanctionable)
	k := s.Keeper.OnlyTestsWithUnsanctionableAddrs(map[string]bool{string(addrUnsanctionable): true})

	tests := []struct {
		name               string
		addrs              []sdk.AccAddress
		expEvents          sdk.Events
		checkSanctioned    []sdk.AccAddress
		checkNotSanctioned []sdk.AccAddress
	}{
		{
			name:               "no addresses",
			addrs:              []sdk.AccAddress{},
			expEvents:          sdk.Events{},
			checkSanctioned:    []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5},
			checkNotSanctioned: []sdk.AccAddress{addrUnsanctionable, addrRandom},
		},
		{
			name:               "one addr never sanctioned",
			addrs:              []sdk.AccAddress{addrRandom},
			expEvents:          makeEvents(addrRandom),
			checkSanctioned:    []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5},
			checkNotSanctioned: []sdk.AccAddress{addrUnsanctionable, addrRandom},
		},
		{
			name:               "one addr",
			addrs:              []sdk.AccAddress{s.addr1},
			expEvents:          makeEvents(s.addr1),
			checkSanctioned:    []sdk.AccAddress{s.addr2, s.addr3, s.addr4, s.addr5},
			checkNotSanctioned: []sdk.AccAddress{s.addr1, addrUnsanctionable, addrRandom},
		},
		{
			name:               "already unsanctioned",
			addrs:              []sdk.AccAddress{s.addr1},
			expEvents:          makeEvents(s.addr1),
			checkSanctioned:    []sdk.AccAddress{s.addr2, s.addr3, s.addr4, s.addr5},
			checkNotSanctioned: []sdk.AccAddress{s.addr1, addrUnsanctionable, addrRandom},
		},
		{
			name:               "two new addrs",
			addrs:              []sdk.AccAddress{s.addr2, s.addr3},
			expEvents:          makeEvents(s.addr2, s.addr3),
			checkSanctioned:    []sdk.AccAddress{s.addr4, s.addr5},
			checkNotSanctioned: []sdk.AccAddress{s.addr1, s.addr2, s.addr3, addrUnsanctionable, addrRandom},
		},
		{
			name:               "three addrs all already unsanctioned",
			addrs:              []sdk.AccAddress{s.addr1, s.addr2, s.addr3},
			expEvents:          makeEvents(s.addr1, s.addr2, s.addr3),
			checkSanctioned:    []sdk.AccAddress{s.addr4, s.addr5},
			checkNotSanctioned: []sdk.AccAddress{s.addr1, s.addr2, s.addr3, addrUnsanctionable, addrRandom},
		},
		{
			name:               "two addrs one already unsanctioned",
			addrs:              []sdk.AccAddress{s.addr4, s.addr1},
			expEvents:          makeEvents(s.addr4, s.addr1),
			checkSanctioned:    []sdk.AccAddress{s.addr5},
			checkNotSanctioned: []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, addrUnsanctionable, addrRandom},
		},
		{
			name:               "three addrs one unsanctionable",
			addrs:              []sdk.AccAddress{addrUnsanctionable, s.addr4, s.addr1},
			expEvents:          makeEvents(addrUnsanctionable, s.addr4, s.addr1),
			checkSanctioned:    []sdk.AccAddress{s.addr5},
			checkNotSanctioned: []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, addrUnsanctionable, addrRandom},
		},
	}

	var isSanctioned bool
	testIsSanction := func(addr sdk.AccAddress) func() {
		return func() {
			isSanctioned = k.IsSanctionedAddr(s.SdkCtx, addr)
		}
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := s.SdkCtx.WithEventManager(em)
			testFunc := func() error {
				return k.UnsanctionAddresses(ctx, tc.addrs...)
			}
			s.RequireNotPanicsNoError(testFunc, "UnsanctionAddresses")
			events := em.Events()
			s.Assert().Equal(tc.expEvents, events, "events emitted during UnsanctionAddresses")
			for i, addr := range tc.checkSanctioned {
				if s.Assert().NotPanics(testIsSanction(addr), "checkSanctioned[%d] IsSanctionedAddr", i) {
					s.Assert().True(isSanctioned, "checkSanctioned[%d] IsSanctionedAddr result", i)
				}
			}
			for i, addr := range tc.checkNotSanctioned {
				if s.Assert().NotPanics(testIsSanction(addr), "checkNotSanctioned[%d] IsSanctionedAddr", i) {
					s.Assert().False(isSanctioned, "checkNotSanctioned[%d] IsSanctionedAddr result", i)
				}
			}
		})
	}

	s.Run("temp entries are deleted", func() {
		s.ReqOKAddTempSanct(1, "s.addr1, s.addr2", s.addr1, s.addr2)
		s.ReqOKAddTempSanct(2, "s.addr1, s.addr3", s.addr1, s.addr3)
		s.ReqOKAddTempUnsanct(3, "s.addr2, s.addr4, s.addr5", s.addr2, s.addr4, s.addr5)

		testFunc := func() error {
			return k.UnsanctionAddresses(s.SdkCtx, s.addr5, s.addr3, s.addr1, s.addr2, s.addr4)
		}
		s.RequireNotPanicsNoError(testFunc, "UnsanctionAddresses")

		tempEntries := s.GetAllTempEntries()
		s.Assert().Empty(tempEntries, "temporary entries still in the store")

		tempIndEntries := s.GetAllIndexTempEntries()
		s.Assert().Empty(tempIndEntries, "proposal index temporary entries still in the store")
	})
}

func (s *KeeperTestSuite) TestKeeper_AddTemporarySanction() {
	makeEvents := func(addrs ...sdk.AccAddress) sdk.Events {
		rv := sdk.Events{}
		for _, addr := range addrs {
			event, err := sdk.TypedEventToEvent(keeper.NewTempEvent(keeper.SanctionB, addr))
			s.Require().NoError(err, "TypedEventToEvent temp event")
			rv = append(rv, event)
		}
		return rv
	}

	var previousTempEntries []*sanction.TemporaryEntry
	var previousIndEntries []*sanction.TemporaryEntry

	getNewEntries := func(previous, now []*sanction.TemporaryEntry) []*sanction.TemporaryEntry {
		var rv []*sanction.TemporaryEntry
		for _, entry := range now {
			found := false
			for _, known := range previous {
				if entry.Address == known.Address && entry.ProposalId == known.ProposalId && entry.Status == known.Status {
					found = true
					break
				}
			}
			if !found {
				rv = append(rv, entry)
			}
		}
		return rv
	}

	// Start with addr5 having a temp unsanction entry.
	s.ReqOKAddTempUnsanct(100, "s.addr5", s.addr5)

	addrUnsanctionable := sdk.AccAddress("unsanctionable_addr_")
	k := s.Keeper.OnlyTestsWithUnsanctionableAddrs(map[string]bool{string(addrUnsanctionable): true})

	tests := []struct {
		name             string
		govPropID        uint64
		addrs            []sdk.AccAddress
		expEvents        sdk.Events
		expErr           []string
		addedTempEntries []*sanction.TemporaryEntry
		addedIndEntries  []*sanction.TemporaryEntry
	}{
		{
			name:      "no addrs",
			addrs:     []sdk.AccAddress{},
			expEvents: sdk.Events{},
		},
		{
			name:             "one addr",
			govPropID:        1,
			addrs:            []sdk.AccAddress{s.addr1},
			expEvents:        makeEvents(s.addr1),
			addedTempEntries: []*sanction.TemporaryEntry{newTempEntry(s.addr1, 1, true)},
			addedIndEntries:  []*sanction.TemporaryEntry{newIndTempEntry(1, s.addr1)},
		},
		{
			name:      "same addr and gov prop as before",
			govPropID: 1,
			addrs:     []sdk.AccAddress{s.addr1},
			expEvents: makeEvents(s.addr1),
		},
		{
			name:             "same addr new gov prop",
			govPropID:        2,
			addrs:            []sdk.AccAddress{s.addr1},
			expEvents:        makeEvents(s.addr1),
			addedTempEntries: []*sanction.TemporaryEntry{newTempEntry(s.addr1, 2, true)},
			addedIndEntries:  []*sanction.TemporaryEntry{newIndTempEntry(2, s.addr1)},
		},
		{
			name:             "previous gov prop new addr",
			govPropID:        1,
			addrs:            []sdk.AccAddress{s.addr2},
			expEvents:        makeEvents(s.addr2),
			addedTempEntries: []*sanction.TemporaryEntry{newTempEntry(s.addr2, 1, true)},
			addedIndEntries:  []*sanction.TemporaryEntry{newIndTempEntry(1, s.addr2)},
		},
		{
			name:      "three addrs",
			govPropID: 3,
			addrs:     []sdk.AccAddress{s.addr1, s.addr2, s.addr3},
			expEvents: makeEvents(s.addr1, s.addr2, s.addr3),
			addedTempEntries: []*sanction.TemporaryEntry{
				newTempEntry(s.addr1, 3, true),
				newTempEntry(s.addr2, 3, true),
				newTempEntry(s.addr3, 3, true),
			},
			addedIndEntries: []*sanction.TemporaryEntry{
				newIndTempEntry(3, s.addr1),
				newIndTempEntry(3, s.addr2),
				newIndTempEntry(3, s.addr3),
			},
		},
		{
			name:      "five addrs first unsanctionable",
			govPropID: 4,
			addrs:     []sdk.AccAddress{addrUnsanctionable, s.addr2, s.addr3, s.addr4, s.addr5},
			expEvents: sdk.Events{},
			expErr:    []string{addrUnsanctionable.String(), "address cannot be sanctioned"},
		},
		{
			name:      "five addrs third unsanctionable",
			govPropID: 5,
			addrs:     []sdk.AccAddress{s.addr1, s.addr2, addrUnsanctionable, s.addr4, s.addr5},
			expEvents: makeEvents(s.addr1, s.addr2),
			expErr:    []string{addrUnsanctionable.String(), "address cannot be sanctioned"},
			addedTempEntries: []*sanction.TemporaryEntry{
				newTempEntry(s.addr1, 5, true),
				newTempEntry(s.addr2, 5, true),
			},
			addedIndEntries: []*sanction.TemporaryEntry{
				newIndTempEntry(5, s.addr1),
				newIndTempEntry(5, s.addr2),
			},
		},
		{
			name:      "five addrs fifth unsanctionable",
			govPropID: 6,
			addrs:     []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, addrUnsanctionable},
			expEvents: makeEvents(s.addr1, s.addr2, s.addr3, s.addr4),
			expErr:    []string{addrUnsanctionable.String(), "address cannot be sanctioned"},
			addedTempEntries: []*sanction.TemporaryEntry{
				newTempEntry(s.addr1, 6, true),
				newTempEntry(s.addr2, 6, true),
				newTempEntry(s.addr3, 6, true),
				newTempEntry(s.addr4, 6, true),
			},
			addedIndEntries: []*sanction.TemporaryEntry{
				newIndTempEntry(6, s.addr1),
				newIndTempEntry(6, s.addr2),
				newIndTempEntry(6, s.addr3),
				newIndTempEntry(6, s.addr4),
			},
		},
		{
			name:             "previous entry overwritten",
			govPropID:        100,
			addrs:            []sdk.AccAddress{s.addr5},
			expEvents:        makeEvents(s.addr5),
			addedTempEntries: []*sanction.TemporaryEntry{newTempEntry(s.addr5, 100, true)},
			addedIndEntries:  nil,
		},
	}

	previousTempEntries = s.GetAllTempEntries()
	previousIndEntries = s.GetAllIndexTempEntries()

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := s.SdkCtx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = k.AddTemporarySanction(ctx, tc.govPropID, tc.addrs...)
			}
			s.Require().NotPanics(testFunc, "AddTemporarySanction")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "AddTemporarySanction error")

			events := em.Events()
			s.Assert().Equal(tc.expEvents, events, "events emitted during AddTemporarySanction")

			currentTempEntries := s.GetAllTempEntries()
			newTempEntries := getNewEntries(previousTempEntries, currentTempEntries)
			s.Assert().ElementsMatch(tc.addedTempEntries, newTempEntries, "new temp entries, A = expected, B = actual")
			previousTempEntries = currentTempEntries

			currentIndEntries := s.GetAllIndexTempEntries()
			newIndEntries := getNewEntries(previousIndEntries, currentIndEntries)
			s.Assert().ElementsMatch(tc.addedIndEntries, newIndEntries, "new index entries, A = expected, B = actual")
			previousIndEntries = currentIndEntries
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_AddTemporaryUnsanction() {
	makeEvents := func(addrs ...sdk.AccAddress) sdk.Events {
		rv := sdk.Events{}
		for _, addr := range addrs {
			event, err := sdk.TypedEventToEvent(keeper.NewTempEvent(keeper.UnsanctionB, addr))
			s.Require().NoError(err, "TypedEventToEvent temp event")
			rv = append(rv, event)
		}
		return rv
	}

	var previousTempEntries []*sanction.TemporaryEntry
	var previousIndEntries []*sanction.TemporaryEntry

	getNewEntries := func(previous, now []*sanction.TemporaryEntry) []*sanction.TemporaryEntry {
		var rv []*sanction.TemporaryEntry
		for _, entry := range now {
			found := false
			for _, known := range previous {
				if entry.Address == known.Address && entry.ProposalId == known.ProposalId && entry.Status == known.Status {
					found = true
					break
				}
			}
			if !found {
				rv = append(rv, entry)
			}
		}
		return rv
	}

	// Start with addr5 having a temp sanction entry.
	s.ReqOKAddTempSanct(100, "s.addr5", s.addr5)

	addrUnsanctionable := sdk.AccAddress("unsanctionable_addr_")
	k := s.Keeper.OnlyTestsWithUnsanctionableAddrs(map[string]bool{string(addrUnsanctionable): true})

	tests := []struct {
		name             string
		govPropID        uint64
		addrs            []sdk.AccAddress
		expEvents        sdk.Events
		addedTempEntries []*sanction.TemporaryEntry
		addedIndEntries  []*sanction.TemporaryEntry
	}{
		{
			name:      "no addrs",
			addrs:     []sdk.AccAddress{},
			expEvents: sdk.Events{},
		},
		{
			name:             "one addr",
			govPropID:        1,
			addrs:            []sdk.AccAddress{s.addr1},
			expEvents:        makeEvents(s.addr1),
			addedTempEntries: []*sanction.TemporaryEntry{newTempEntry(s.addr1, 1, false)},
			addedIndEntries:  []*sanction.TemporaryEntry{newIndTempEntry(1, s.addr1)},
		},
		{
			name:      "same addr and gov prop as before",
			govPropID: 1,
			addrs:     []sdk.AccAddress{s.addr1},
			expEvents: makeEvents(s.addr1),
		},
		{
			name:             "same addr new gov prop",
			govPropID:        2,
			addrs:            []sdk.AccAddress{s.addr1},
			expEvents:        makeEvents(s.addr1),
			addedTempEntries: []*sanction.TemporaryEntry{newTempEntry(s.addr1, 2, false)},
			addedIndEntries:  []*sanction.TemporaryEntry{newIndTempEntry(2, s.addr1)},
		},
		{
			name:             "previous gov prop new addr",
			govPropID:        1,
			addrs:            []sdk.AccAddress{s.addr2},
			expEvents:        makeEvents(s.addr2),
			addedTempEntries: []*sanction.TemporaryEntry{newTempEntry(s.addr2, 1, false)},
			addedIndEntries:  []*sanction.TemporaryEntry{newIndTempEntry(1, s.addr2)},
		},
		{
			name:      "three addrs",
			govPropID: 3,
			addrs:     []sdk.AccAddress{s.addr1, s.addr2, s.addr3},
			expEvents: makeEvents(s.addr1, s.addr2, s.addr3),
			addedTempEntries: []*sanction.TemporaryEntry{
				newTempEntry(s.addr1, 3, false),
				newTempEntry(s.addr2, 3, false),
				newTempEntry(s.addr3, 3, false),
			},
			addedIndEntries: []*sanction.TemporaryEntry{
				newIndTempEntry(3, s.addr1),
				newIndTempEntry(3, s.addr2),
				newIndTempEntry(3, s.addr3),
			},
		},
		{
			name:      "five addrs first unsanctionable",
			govPropID: 4,
			addrs:     []sdk.AccAddress{addrUnsanctionable, s.addr2, s.addr3, s.addr4, s.addr5},
			expEvents: makeEvents(addrUnsanctionable, s.addr2, s.addr3, s.addr4, s.addr5),
			addedTempEntries: []*sanction.TemporaryEntry{
				newTempEntry(addrUnsanctionable, 4, false),
				newTempEntry(s.addr2, 4, false),
				newTempEntry(s.addr3, 4, false),
				newTempEntry(s.addr4, 4, false),
				newTempEntry(s.addr5, 4, false),
			},
			addedIndEntries: []*sanction.TemporaryEntry{
				newIndTempEntry(4, addrUnsanctionable),
				newIndTempEntry(4, s.addr2),
				newIndTempEntry(4, s.addr3),
				newIndTempEntry(4, s.addr4),
				newIndTempEntry(4, s.addr5),
			},
		},
		{
			name:      "five addrs third unsanctionable",
			govPropID: 5,
			addrs:     []sdk.AccAddress{s.addr1, s.addr2, addrUnsanctionable, s.addr4, s.addr5},
			expEvents: makeEvents(s.addr1, s.addr2, addrUnsanctionable, s.addr4, s.addr5),
			addedTempEntries: []*sanction.TemporaryEntry{
				newTempEntry(s.addr1, 5, false),
				newTempEntry(s.addr2, 5, false),
				newTempEntry(addrUnsanctionable, 5, false),
				newTempEntry(s.addr4, 5, false),
				newTempEntry(s.addr5, 5, false),
			},
			addedIndEntries: []*sanction.TemporaryEntry{
				newIndTempEntry(5, s.addr1),
				newIndTempEntry(5, s.addr2),
				newIndTempEntry(5, addrUnsanctionable),
				newIndTempEntry(5, s.addr4),
				newIndTempEntry(5, s.addr5),
			},
		},
		{
			name:      "five addrs fifth unsanctionable",
			govPropID: 6,
			addrs:     []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, addrUnsanctionable},
			expEvents: makeEvents(s.addr1, s.addr2, s.addr3, s.addr4, addrUnsanctionable),
			addedTempEntries: []*sanction.TemporaryEntry{
				newTempEntry(s.addr1, 6, false),
				newTempEntry(s.addr2, 6, false),
				newTempEntry(s.addr3, 6, false),
				newTempEntry(s.addr4, 6, false),
				newTempEntry(addrUnsanctionable, 6, false),
			},
			addedIndEntries: []*sanction.TemporaryEntry{
				newIndTempEntry(6, s.addr1),
				newIndTempEntry(6, s.addr2),
				newIndTempEntry(6, s.addr3),
				newIndTempEntry(6, s.addr4),
				newIndTempEntry(6, addrUnsanctionable),
			},
		},
		{
			name:             "previous entry overwritten",
			govPropID:        100,
			addrs:            []sdk.AccAddress{s.addr5},
			expEvents:        makeEvents(s.addr5),
			addedTempEntries: []*sanction.TemporaryEntry{newTempEntry(s.addr5, 100, false)},
			addedIndEntries:  nil,
		},
	}

	previousTempEntries = s.GetAllTempEntries()
	previousIndEntries = s.GetAllIndexTempEntries()

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := s.SdkCtx.WithEventManager(em)
			testFunc := func() error {
				return k.AddTemporaryUnsanction(ctx, tc.govPropID, tc.addrs...)
			}
			s.RequireNotPanicsNoError(testFunc, "AddTemporaryUnsanction")

			events := em.Events()
			s.Assert().Equal(tc.expEvents, events, "events emitted during AddTemporaryUnsanction")

			currentTempEntries := s.GetAllTempEntries()
			newTempEntries := getNewEntries(previousTempEntries, currentTempEntries)
			s.Assert().ElementsMatch(tc.addedTempEntries, newTempEntries, "new temp entries, A = expected, B = actual")
			previousTempEntries = currentTempEntries

			currentIndEntries := s.GetAllIndexTempEntries()
			newIndEntries := getNewEntries(previousIndEntries, currentIndEntries)
			s.Assert().ElementsMatch(tc.addedIndEntries, newIndEntries, "new index entries, A = expected, B = actual")
			previousIndEntries = currentIndEntries
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_getLatestTempEntry() {
	store := s.GetStore()
	// Add a few random entries with weird values, so they're easy to identify.
	randAddr1 := sdk.AccAddress{0, 0, 0, 0, 0}
	randAddr2 := s.addr1[:len(s.addr1)-1]
	randAddr3 := s.addr1[1:]
	randAddr4 := sdk.AccAddress{255, 255, 255, 255, 255, 255}
	val := uint8(39)
	for _, id := range []uint64{18, 19, 55, 100000} {
		for _, addr := range []sdk.AccAddress{randAddr1, randAddr2, randAddr3, randAddr4} {
			val += 1
			store.Set(keeper.CreateTemporaryKey(addr, id), []byte{val})
			store.Set(keeper.CreateProposalTempIndexKey(id, addr), []byte{val})
		}
	}

	s.Run("nil addr", func() {
		var expected []byte
		var actual []byte
		testFunc := func() {
			actual = s.Keeper.OnlyTestsGetLatestTempEntry(store, nil)
		}
		s.Require().NotPanics(testFunc, "getLatestTempEntry")
		s.Assert().Equal(expected, actual, "getLatestTempEntry result")
	})

	s.Run("empty addr", func() {
		var expected []byte
		var actual []byte
		testFunc := func() {
			actual = s.Keeper.OnlyTestsGetLatestTempEntry(store, sdk.AccAddress{})
		}
		s.Require().NotPanics(testFunc, "getLatestTempEntry")
		s.Assert().Equal(expected, actual, "getLatestTempEntry result")
	})

	s.Run("no entries", func() {
		var expected []byte
		var actual []byte
		testFunc := func() {
			actual = s.Keeper.OnlyTestsGetLatestTempEntry(store, s.addr1)
		}
		s.Require().NotPanics(testFunc, "getLatestTempEntry")
		s.Assert().Equal(expected, actual, "getLatestTempEntry result")
	})

	s.Run("one sanction entry", func() {
		addr := sdk.AccAddress("one_entry_test_addr")
		s.ReqOKAddTempSanct(1, "addr", addr)

		expected := []byte{keeper.SanctionB}
		var actual []byte
		testFunc := func() {
			actual = s.Keeper.OnlyTestsGetLatestTempEntry(store, addr)
		}
		s.Require().NotPanics(testFunc, "getLatestTempEntry")
		s.Assert().Equal(expected, actual, "getLatestTempEntry result")
	})

	s.Run("one unsanction entry", func() {
		addr := sdk.AccAddress("one_entry_test_addr2")
		s.ReqOKAddTempUnsanct(2, "addr", addr)

		expected := []byte{keeper.UnsanctionB}
		var actual []byte
		testFunc := func() {
			actual = s.Keeper.OnlyTestsGetLatestTempEntry(store, addr)
		}
		s.Require().NotPanics(testFunc, "getLatestTempEntry")
		s.Assert().Equal(expected, actual, "getLatestTempEntry result")
	})

	s.Run("three entries last sanction", func() {
		addr := sdk.AccAddress("three_entry_sanctioned")
		// Writing the one with the largest prop id first to show that later writes with smaller prop ids don't mess it up.
		s.ReqOKAddTempSanct(5, "addr", addr)
		s.ReqOKAddTempUnsanct(3, "addr", addr)
		s.ReqOKAddTempUnsanct(4, "addr", addr)

		expected := []byte{keeper.SanctionB}
		var actual []byte
		testFunc := func() {
			actual = s.Keeper.OnlyTestsGetLatestTempEntry(store, addr)
		}
		s.Require().NotPanics(testFunc, "getLatestTempEntry")
		s.Assert().Equal(expected, actual, "getLatestTempEntry result")
	})

	s.Run("three entries last unsanction", func() {
		addr := sdk.AccAddress("three_entry_unsanctioned")
		// Writing the one with the largest prop id first to show that later writes with smaller prop ids don't mess it up.
		s.ReqOKAddTempUnsanct(8, "addr", addr)
		s.ReqOKAddTempSanct(7, "addr", addr)
		s.ReqOKAddTempSanct(6, "addr", addr)

		expected := []byte{keeper.UnsanctionB}
		var actual []byte
		testFunc := func() {
			actual = s.Keeper.OnlyTestsGetLatestTempEntry(store, addr)
		}
		s.Require().NotPanics(testFunc, "getLatestTempEntry")
		s.Assert().Equal(expected, actual, "getLatestTempEntry result")
	})
}

func (s *KeeperTestSuite) TestKeeper_DeleteGovPropTempEntries() {
	// Add several temp entries for multiple gov props.
	addrs := []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5}
	for id := uint64(1); id <= 10; id++ {
		if id%2 == 1 {
			s.ReqOKAddTempSanct(id, "addrs...", addrs...)
		} else {
			s.ReqOKAddTempUnsanct(id, "addrs...", addrs...)
		}
	}

	s.Run("unknown gov prop id", func() {
		origTempEntries := s.GetAllTempEntries()
		origIndEntries := s.GetAllIndexTempEntries()

		testFunc := func() {
			s.Keeper.DeleteGovPropTempEntries(s.SdkCtx, 382892)
		}
		s.Require().NotPanics(testFunc, "DeleteGovPropTempEntries")

		finalTempEntries := s.GetAllTempEntries()
		finalIndEntries := s.GetAllIndexTempEntries()

		s.Assert().ElementsMatch(origTempEntries, finalTempEntries, "temp entries, A = orig, B = after delete")
		s.Assert().ElementsMatch(origIndEntries, finalIndEntries, "index entries, A = orig, B = after delete")
	})

	s.Run("id with sanction entries", func() {
		origTempEntries := s.GetAllTempEntries()
		origIndEntries := s.GetAllIndexTempEntries()

		idToDelete := uint64(5)
		var expTempEntries []*sanction.TemporaryEntry
		for _, entry := range origTempEntries {
			if entry.ProposalId != idToDelete {
				expTempEntries = append(expTempEntries, entry)
			}
		}
		var expIndEntries []*sanction.TemporaryEntry
		for _, entry := range origIndEntries {
			if entry.ProposalId != idToDelete {
				expIndEntries = append(expIndEntries, entry)
			}
		}

		testFunc := func() {
			s.Keeper.DeleteGovPropTempEntries(s.SdkCtx, idToDelete)
		}
		s.Require().NotPanics(testFunc, "DeleteGovPropTempEntries")

		finalTempEntries := s.GetAllTempEntries()
		finalIndEntries := s.GetAllIndexTempEntries()

		s.Assert().ElementsMatch(expTempEntries, finalTempEntries, "temp entries, A = expected, B = after delete")
		s.Assert().ElementsMatch(expIndEntries, finalIndEntries, "index entries, A = expected, B = after delete")
	})

	s.Run("id with unsanction entries", func() {
		origTempEntries := s.GetAllTempEntries()
		origIndEntries := s.GetAllIndexTempEntries()

		idToDelete := uint64(2)
		var expTempEntries []*sanction.TemporaryEntry
		for _, entry := range origTempEntries {
			if entry.ProposalId != idToDelete {
				expTempEntries = append(expTempEntries, entry)
			}
		}
		var expIndEntries []*sanction.TemporaryEntry
		for _, entry := range origIndEntries {
			if entry.ProposalId != idToDelete {
				expIndEntries = append(expIndEntries, entry)
			}
		}

		testFunc := func() {
			s.Keeper.DeleteGovPropTempEntries(s.SdkCtx, idToDelete)
		}
		s.Require().NotPanics(testFunc, "DeleteGovPropTempEntries")

		finalTempEntries := s.GetAllTempEntries()
		finalIndEntries := s.GetAllIndexTempEntries()

		s.Assert().ElementsMatch(expTempEntries, finalTempEntries, "temp entries, A = expected, B = after delete")
		s.Assert().ElementsMatch(expIndEntries, finalIndEntries, "index entries, A = expected, B = after delete")
	})
}

func (s *KeeperTestSuite) TestKeeper_DeleteAddrTempEntries() {
	// Add several temp entries for multiple gov props.
	addrs := []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5}
	for id := uint64(1); id <= 10; id++ {
		if id%2 == 1 {
			s.ReqOKAddTempSanct(id, "addrs...", addrs...)
		} else {
			s.ReqOKAddTempUnsanct(id, "addrs...", addrs...)
		}
	}

	s.Run("unknown address", func() {
		origTempEntries := s.GetAllTempEntries()
		origIndEntries := s.GetAllIndexTempEntries()

		testFunc := func() {
			s.Keeper.DeleteAddrTempEntries(s.SdkCtx, sdk.AccAddress("unknown_test_address"))
		}
		s.Require().NotPanics(testFunc, "DeleteAddrTempEntries")

		finalTempEntries := s.GetAllTempEntries()
		finalIndEntries := s.GetAllIndexTempEntries()

		s.Assert().ElementsMatch(origTempEntries, finalTempEntries, "temp entries, A = orig, B = after delete")
		s.Assert().ElementsMatch(origIndEntries, finalIndEntries, "index entries, A = orig, B = after delete")
	})

	s.Run("known addr", func() {
		origTempEntries := s.GetAllTempEntries()
		origIndEntries := s.GetAllIndexTempEntries()

		addrToDelete := s.addr3
		addrToDeleteStr := addrToDelete.String()

		var expTempEntries []*sanction.TemporaryEntry
		for _, entry := range origTempEntries {
			if entry.Address != addrToDeleteStr {
				expTempEntries = append(expTempEntries, entry)
			}
		}
		var expIndEntries []*sanction.TemporaryEntry
		for _, entry := range origIndEntries {
			if entry.Address != addrToDeleteStr {
				expIndEntries = append(expIndEntries, entry)
			}
		}

		testFunc := func() {
			s.Keeper.DeleteAddrTempEntries(s.SdkCtx, addrToDelete)
		}
		s.Require().NotPanics(testFunc, "DeleteAddrTempEntries")

		finalTempEntries := s.GetAllTempEntries()
		finalIndEntries := s.GetAllIndexTempEntries()

		s.Assert().ElementsMatch(expTempEntries, finalTempEntries, "temp entries, A = expected, B = after delete")
		s.Assert().ElementsMatch(expIndEntries, finalIndEntries, "index entries, A = expected, B = after delete")
	})
}

func (s *KeeperTestSuite) TestKeeper_IterateSanctionedAddresses() {
	s.Run("nothing to iterate", func() {
		var addrs []sdk.AccAddress
		cb := func(addr sdk.AccAddress) bool {
			addrs = append(addrs, addr)
			return false
		}
		testFunc := func() {
			s.Keeper.IterateSanctionedAddresses(s.SdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateSanctionedAddresses")
		s.Require().Empty(addrs, "addresses iterated")
	})

	// rAddr makes a "random" address that starts with 0xFF, 0xFF and an index.
	// hopefully that makes them last when iterating, putting other entries first.
	rAddr := func(i uint8) sdk.AccAddress {
		return append(sdk.AccAddress{255, 255, i}, "_random_test_addr"...)
	}
	randomAddrs := []sdk.AccAddress{rAddr(0), rAddr(1), rAddr(2), rAddr(3), rAddr(4)}
	// Setup:
	// all the randomAddrs = sanctioned
	// addr1 = sanctioned
	// addr2 = sanctioned then unsanctioned
	// addr3 = temp sanctioned
	// addr4 = temp unsanctioned
	s.ReqOKAddPermSanct("s.addr1, s.addr2", s.addr1, s.addr2)
	s.ReqOKAddPermSanct("randomAddrs...", randomAddrs...)
	s.ReqOKAddPermUnsanct("s.addr2", s.addr2)
	s.ReqOKAddTempSanct(1, "s.addr3", s.addr3)
	s.ReqOKAddTempUnsanct(1, "s.addr4", s.addr4)

	s.Run("get all entries", func() {
		expected := []sdk.AccAddress{s.addr1}
		expected = append(expected, randomAddrs...)
		var addrs []sdk.AccAddress
		cb := func(addr sdk.AccAddress) bool {
			addrs = append(addrs, addr)
			return false
		}
		testFunc := func() {
			s.Keeper.IterateSanctionedAddresses(s.SdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateSanctionedAddresses")
		s.Assert().Equal(expected, addrs, "sanctioned addresses iterated")
	})

	s.Run("stop after third", func() {
		expected := []sdk.AccAddress{s.addr1}
		expected = append(expected, randomAddrs...)
		expected = expected[:3]
		var addrs []sdk.AccAddress
		cb := func(addr sdk.AccAddress) bool {
			addrs = append(addrs, addr)
			return len(addrs) >= 3
		}
		testFunc := func() {
			s.Keeper.IterateSanctionedAddresses(s.SdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateSanctionedAddresses")
		s.Assert().Equal(expected, addrs, "sanctioned addresses iterated")
	})

	s.Run("stop after first", func() {
		expected := []sdk.AccAddress{s.addr1}
		var addrs []sdk.AccAddress
		cb := func(addr sdk.AccAddress) bool {
			addrs = append(addrs, addr)
			return true
		}
		testFunc := func() {
			s.Keeper.IterateSanctionedAddresses(s.SdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateSanctionedAddresses")
		s.Assert().Equal(expected, addrs, "sanctioned addresses iterated")
	})
}

func (s *KeeperTestSuite) TestKeeper_IterateTemporaryEntries() {
	s.Run("nothing to iterate", func() {
		var addrs []sdk.AccAddress
		cb := func(addr sdk.AccAddress, _ uint64, _ bool) bool {
			addrs = append(addrs, addr)
			return false
		}
		testFunc := func() {
			s.Keeper.IterateTemporaryEntries(s.SdkCtx, nil, cb)
		}
		s.Require().NotPanics(testFunc, "IterateTemporaryEntries")
		s.Require().Empty(addrs, "addresses iterated")
	})

	// rAddr makes a "random" address that starts with 0xFF, 0xFF and an index.
	// hopefully that makes them last when iterating, putting other entries first.
	rAddr := func(i uint8) sdk.AccAddress {
		return append(sdk.AccAddress{255, 255, i}, "_random_test_addr"...)
	}
	randomSanctAddrs := []sdk.AccAddress{rAddr(0), rAddr(1), rAddr(2), rAddr(3), rAddr(4)}
	randomUnsanctAddrs := []sdk.AccAddress{rAddr(5), rAddr(6), rAddr(7), rAddr(8), rAddr(9)}
	mixedAddr := rAddr(10)
	// Setup:
	// addr1 = sanctioned
	// addr2 = sanctioned then unsanctioned
	// addr3 = temp sanctioned id 1
	// addr4 = temp unsanctioned id 1
	// all the randomSanctAddrs = temp sanctioned for gov prop 1 and 2
	// first two randomSanctAddrs = temp unsanctioned for gov prop 3 too
	// all the randomUnsanctAddrs = temp unsanctioned for gov prop 1 and 2
	// first two randomUnsanctAddrs = temp unsanctioned for gov prop 3 too
	// mixedAddr = temp sanction for 1 and 3, temp unsanction for 2
	s.ReqOKAddPermSanct("s.addr1, s.addr2", s.addr1, s.addr2)
	s.ReqOKAddPermUnsanct("s.addr2", s.addr2)
	s.ReqOKAddTempSanct(1, "s.addr3", s.addr3)
	s.ReqOKAddTempUnsanct(1, "s.addr4", s.addr4)
	s.ReqOKAddTempSanct(1, "randomSanctAddrs...", randomSanctAddrs...)
	s.ReqOKAddTempSanct(2, "randomSanctAddrs...", randomSanctAddrs...)
	s.ReqOKAddTempSanct(3, "randomSanctAddrs[:2]...", randomSanctAddrs[:2]...)
	s.ReqOKAddTempUnsanct(1, "randomUnsanctAddrs...", randomUnsanctAddrs...)
	s.ReqOKAddTempUnsanct(2, "randomUnsanctAddrs...", randomUnsanctAddrs...)
	s.ReqOKAddTempUnsanct(3, "randomUnsanctAddrs[:2]...", randomUnsanctAddrs[:2]...)
	s.ReqOKAddTempSanct(1, "mixedAddr", mixedAddr)
	s.ReqOKAddTempUnsanct(2, "mixedAddr", mixedAddr)
	s.ReqOKAddTempSanct(3, "mixedAddr", mixedAddr)

	// sortEntries sorts the provided entries in place and also returns that slice.
	// They are ordered the same way they're expected to be in state.
	// This is horribly inefficient. Do not use it outside unit tests.
	sortEntries := func(entries []*sanction.TemporaryEntry) []*sanction.TemporaryEntry {
		sort.Slice(entries, func(i, j int) bool {
			addrI, err := sdk.AccAddressFromBech32(entries[i].Address)
			s.Require().NoError(err, "AccAddressFromBech32(%q)", entries[i].Address)
			addrJ, err := sdk.AccAddressFromBech32(entries[j].Address)
			s.Require().NoError(err, "AccAddressFromBech32(%q)", entries[j].Address)
			addrCmp := bytes.Compare(addrI, addrJ)
			if addrCmp < 0 {
				return true
			}
			return addrCmp == 0 && entries[i].ProposalId < entries[j].ProposalId
		})
		return entries
	}

	addr3Entries := sortEntries([]*sanction.TemporaryEntry{
		newTempEntry(s.addr3, 1, true),
	})
	addr4Entries := sortEntries([]*sanction.TemporaryEntry{
		newTempEntry(s.addr4, 1, false),
	})
	randomSanctEntries := [][]*sanction.TemporaryEntry{
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomSanctAddrs[0], 1, true),
			newTempEntry(randomSanctAddrs[0], 2, true),
			newTempEntry(randomSanctAddrs[0], 3, true),
		}),
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomSanctAddrs[1], 1, true),
			newTempEntry(randomSanctAddrs[1], 2, true),
			newTempEntry(randomSanctAddrs[1], 3, true),
		}),
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomSanctAddrs[2], 1, true),
			newTempEntry(randomSanctAddrs[2], 2, true),
		}),
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomSanctAddrs[3], 1, true),
			newTempEntry(randomSanctAddrs[3], 2, true),
		}),
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomSanctAddrs[4], 1, true),
			newTempEntry(randomSanctAddrs[4], 2, true),
		}),
	}
	randomUnsanctEntries := [][]*sanction.TemporaryEntry{
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomUnsanctAddrs[0], 1, false),
			newTempEntry(randomUnsanctAddrs[0], 2, false),
			newTempEntry(randomUnsanctAddrs[0], 3, false),
		}),
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomUnsanctAddrs[1], 1, false),
			newTempEntry(randomUnsanctAddrs[1], 2, false),
			newTempEntry(randomUnsanctAddrs[1], 3, false),
		}),
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomUnsanctAddrs[2], 1, false),
			newTempEntry(randomUnsanctAddrs[2], 2, false),
		}),
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomUnsanctAddrs[3], 1, false),
			newTempEntry(randomUnsanctAddrs[3], 2, false),
		}),
		sortEntries([]*sanction.TemporaryEntry{
			newTempEntry(randomUnsanctAddrs[4], 1, false),
			newTempEntry(randomUnsanctAddrs[4], 2, false),
		}),
	}
	mixedEntries := sortEntries([]*sanction.TemporaryEntry{
		newTempEntry(mixedAddr, 1, true),
		newTempEntry(mixedAddr, 2, false),
		newTempEntry(mixedAddr, 3, true),
	})

	var allEntries []*sanction.TemporaryEntry
	allEntries = append(allEntries, addr3Entries...)
	allEntries = append(allEntries, addr4Entries...)
	for _, entries := range randomSanctEntries {
		allEntries = append(allEntries, entries...)
	}
	for _, entries := range randomUnsanctEntries {
		allEntries = append(allEntries, entries...)
	}
	allEntries = append(allEntries, mixedEntries...)
	allEntries = sortEntries(allEntries)

	s.Run("stop after third", func() {
		expected := allEntries[:3]
		var entries []*sanction.TemporaryEntry
		cb := func(addr sdk.AccAddress, govPropId uint64, isSanctioned bool) bool {
			entries = append(entries, newTempEntry(addr, govPropId, isSanctioned))
			return len(entries) >= 3
		}
		testFunc := func() {
			s.Keeper.IterateTemporaryEntries(s.SdkCtx, nil, cb)
		}
		s.Require().NotPanics(testFunc, "IterateTemporaryEntries")
		s.Assert().Equal(expected, entries, "entries iterated")
	})

	s.Run("stop after first", func() {
		expected := allEntries[:1]
		var entries []*sanction.TemporaryEntry
		cb := func(addr sdk.AccAddress, govPropId uint64, isSanctioned bool) bool {
			entries = append(entries, newTempEntry(addr, govPropId, isSanctioned))
			return true
		}
		testFunc := func() {
			s.Keeper.IterateTemporaryEntries(s.SdkCtx, nil, cb)
		}
		s.Require().NotPanics(testFunc, "IterateTemporaryEntries")
		s.Assert().Equal(expected, entries, "entries iterated")
	})

	tests := []struct {
		name     string
		addr     sdk.AccAddress
		expected []*sanction.TemporaryEntry
	}{
		{
			name:     "nil addr",
			addr:     nil,
			expected: allEntries,
		},
		{
			name:     "addr with only one is sanction",
			addr:     s.addr3,
			expected: addr3Entries,
		},
		{
			name:     "addr with only one is unsanction",
			addr:     s.addr4,
			expected: addr4Entries,
		},
		{
			name:     "addr with 3 sanction entries",
			addr:     randomSanctAddrs[0],
			expected: randomSanctEntries[0],
		},
		{
			name:     "addr with 3 unsanction entries",
			addr:     randomUnsanctAddrs[0],
			expected: randomUnsanctEntries[0],
		},
		{
			name:     "addr with mixed entries",
			addr:     mixedAddr,
			expected: mixedEntries,
		},
		{
			name:     "first byte of a random addr",
			addr:     sdk.AccAddress{randomSanctAddrs[0][0]},
			expected: nil,
		},
		{
			name:     "first byte of addr 3",
			addr:     sdk.AccAddress{s.addr3[0]},
			expected: nil,
		},
		{
			name:     "first byte of addr 4",
			addr:     sdk.AccAddress{s.addr4[0]},
			expected: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var entries []*sanction.TemporaryEntry
			cb := func(addr sdk.AccAddress, govPropId uint64, isSanctioned bool) bool {
				entries = append(entries, newTempEntry(addr, govPropId, isSanctioned))
				return false
			}
			testFunc := func() {
				s.Keeper.IterateTemporaryEntries(s.SdkCtx, tc.addr, cb)
			}
			s.Require().NotPanics(testFunc, "IterateTemporaryEntries")
			s.Assert().Equal(tc.expected, entries, "entries iterated")
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_IterateProposalIndexEntries() {
	s.Run("nothing to iterate", func() {
		var addrs []sdk.AccAddress
		cb := func(_ uint64, addr sdk.AccAddress) bool {
			addrs = append(addrs, addr)
			return false
		}
		testFunc := func() {
			s.Keeper.IterateProposalIndexEntries(s.SdkCtx, nil, cb)
		}
		s.Require().NotPanics(testFunc, "IterateProposalIndexEntries")
		s.Require().Empty(addrs, "addresses iterated")
	})

	// rAddr makes a "random" address that starts with 0xFF, 0xFF and an index.
	// hopefully that makes them last when iterating, putting other entries first.
	rAddr := func(i uint8) sdk.AccAddress {
		return append(sdk.AccAddress{255, 255, i}, "_random_test_addr"...)
	}
	randomSanctAddrs := []sdk.AccAddress{rAddr(0), rAddr(1), rAddr(2), rAddr(3), rAddr(4)}
	randomUnsanctAddrs := []sdk.AccAddress{rAddr(5), rAddr(6), rAddr(7), rAddr(8), rAddr(9)}
	mixedAddr := rAddr(10)

	// Setup:
	// id 1 = sanctioned: addr3, all randomSanctAddrs mixed addr, unsanctioned: addr4, all randomUnsanctAddrs
	// id 2 = sanctioned: addr3, all randomSanctAddrs mixed addr
	// id 3 = unsanctioned: addr4, all randomUnsanctAddrs, mixed addr
	s.ReqOKAddPermSanct("s.addr1, s.addr2", s.addr1, s.addr2)
	s.ReqOKAddPermUnsanct("s.addr2", s.addr2)
	s.ReqOKAddTempSanct(1, "s.addr3, mixedAddr", s.addr3, mixedAddr)
	s.ReqOKAddTempSanct(1, "randomSanctAddrs...", randomSanctAddrs...)
	s.ReqOKAddTempUnsanct(1, "s.addr4", s.addr4)
	s.ReqOKAddTempUnsanct(1, "randomUnsanctAddrs...", randomUnsanctAddrs...)
	s.ReqOKAddTempSanct(2, "s.addr3, mixedAddr", s.addr3, mixedAddr)
	s.ReqOKAddTempSanct(2, "randomSanctAddrs...", randomSanctAddrs...)
	s.ReqOKAddTempUnsanct(3, "s.addr4, mixedAddr", s.addr4, mixedAddr)
	s.ReqOKAddTempUnsanct(3, "randomUnsanctAddrs...", randomUnsanctAddrs...)

	// sortEntries sorts the provided entries in place and also returns that slice.
	// They are ordered the same way they're expected to be in state.
	// This is horribly inefficient. Do not use it outside unit tests.
	sortEntries := func(entries []*sanction.TemporaryEntry) []*sanction.TemporaryEntry {
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].ProposalId < entries[j].ProposalId {
				return true
			}
			if entries[i].ProposalId > entries[j].ProposalId {
				return false
			}
			addrI, err := sdk.AccAddressFromBech32(entries[i].Address)
			s.Require().NoError(err, "AccAddressFromBech32(%q)", entries[i].Address)
			addrJ, err := sdk.AccAddressFromBech32(entries[j].Address)
			s.Require().NoError(err, "AccAddressFromBech32(%q)", entries[j].Address)
			return bytes.Compare(addrI, addrJ) < 0
		})
		return entries
	}

	prop1Entries := sortEntries([]*sanction.TemporaryEntry{
		newIndTempEntry(1, s.addr3),
		newIndTempEntry(1, s.addr4),
		newIndTempEntry(1, mixedAddr),
		newIndTempEntry(1, randomSanctAddrs[0]),
		newIndTempEntry(1, randomSanctAddrs[1]),
		newIndTempEntry(1, randomSanctAddrs[2]),
		newIndTempEntry(1, randomSanctAddrs[3]),
		newIndTempEntry(1, randomSanctAddrs[4]),
		newIndTempEntry(1, randomUnsanctAddrs[0]),
		newIndTempEntry(1, randomUnsanctAddrs[1]),
		newIndTempEntry(1, randomUnsanctAddrs[2]),
		newIndTempEntry(1, randomUnsanctAddrs[3]),
		newIndTempEntry(1, randomUnsanctAddrs[4]),
	})
	prop2Entries := sortEntries([]*sanction.TemporaryEntry{
		newIndTempEntry(2, s.addr3),
		newIndTempEntry(2, mixedAddr),
		newIndTempEntry(2, randomSanctAddrs[0]),
		newIndTempEntry(2, randomSanctAddrs[1]),
		newIndTempEntry(2, randomSanctAddrs[2]),
		newIndTempEntry(2, randomSanctAddrs[3]),
		newIndTempEntry(2, randomSanctAddrs[4]),
	})
	prop3Entries := sortEntries([]*sanction.TemporaryEntry{
		newIndTempEntry(3, s.addr4),
		newIndTempEntry(3, mixedAddr),
		newIndTempEntry(3, randomUnsanctAddrs[0]),
		newIndTempEntry(3, randomUnsanctAddrs[1]),
		newIndTempEntry(3, randomUnsanctAddrs[2]),
		newIndTempEntry(3, randomUnsanctAddrs[3]),
		newIndTempEntry(3, randomUnsanctAddrs[4]),
	})

	var allEntries []*sanction.TemporaryEntry
	allEntries = append(allEntries, prop1Entries...)
	allEntries = append(allEntries, prop2Entries...)
	allEntries = append(allEntries, prop3Entries...)
	allEntries = sortEntries(allEntries)

	s.Run("stop after third", func() {
		expected := allEntries[:3]
		var entries []*sanction.TemporaryEntry
		cb := func(govPropId uint64, addr sdk.AccAddress) bool {
			entries = append(entries, newIndTempEntry(govPropId, addr))
			return len(entries) >= 3
		}
		testFunc := func() {
			s.Keeper.IterateProposalIndexEntries(s.SdkCtx, nil, cb)
		}
		s.Require().NotPanics(testFunc, "IterateProposalIndexEntries")
		s.Assert().Equal(expected, entries, "entries iterated")
	})

	s.Run("stop after first", func() {
		expected := allEntries[:1]
		var entries []*sanction.TemporaryEntry
		cb := func(govPropId uint64, addr sdk.AccAddress) bool {
			entries = append(entries, newIndTempEntry(govPropId, addr))
			return true
		}
		testFunc := func() {
			s.Keeper.IterateProposalIndexEntries(s.SdkCtx, nil, cb)
		}
		s.Require().NotPanics(testFunc, "IterateProposalIndexEntries")
		s.Assert().Equal(expected, entries, "entries iterated")
	})

	id := func(i uint64) *uint64 {
		return &i
	}

	tests := []struct {
		name      string
		govPropId *uint64
		expected  []*sanction.TemporaryEntry
	}{
		{
			name:      "nil id",
			govPropId: nil,
			expected:  allEntries,
		},
		{
			name:      "id without entries.",
			govPropId: id(392023),
			expected:  nil,
		},
		{
			name:      "id with mixed entries.",
			govPropId: id(1),
			expected:  prop1Entries,
		},
		{
			name:      "id with only sanctions",
			govPropId: id(2),
			expected:  prop2Entries,
		},
		{
			name:      "id with only unsanctions",
			govPropId: id(3),
			expected:  prop3Entries,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var entries []*sanction.TemporaryEntry
			cb := func(govPropId uint64, addr sdk.AccAddress) bool {
				entries = append(entries, newIndTempEntry(govPropId, addr))
				return false
			}
			testFunc := func() {
				s.Keeper.IterateProposalIndexEntries(s.SdkCtx, tc.govPropId, cb)
			}
			s.Require().NotPanics(testFunc, "IterateProposalIndexEntries")
			s.Assert().Equal(tc.expected, entries, "entries iterated")
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_IsAddrThatCannotBeSanctioned() {
	k := s.Keeper.OnlyTestsWithUnsanctionableAddrs(map[string]bool{
		string(s.addr1): true,
		string(s.addr2): true,
		string(s.addr3): false, // I'm not sure how this would happen, but whatever.
	})

	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  bool
	}{
		{
			name: "unsanctionable addr 1",
			addr: s.addr1,
			exp:  true,
		},
		{
			name: "unsanctionable addr 2",
			addr: s.addr2,
			exp:  true,
		},
		{
			name: "sanctionable addr",
			addr: s.addr3,
			exp:  false,
		},
		{
			name: "nil",
			addr: nil,
			exp:  false,
		},
		{
			name: "empty",
			addr: nil,
			exp:  false,
		},
		{
			name: "random",
			addr: sdk.AccAddress("random"),
			exp:  false,
		},
		{
			name: "other addr",
			addr: s.addr5,
			exp:  false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual bool
			testFunc := func() {
				actual = k.IsAddrThatCannotBeSanctioned(tc.addr)
			}
			s.Require().NotPanics(testFunc, "IsAddrThatCannotBeSanctioned")
			s.Assert().Equal(tc.exp, actual, "IsAddrThatCannotBeSanctioned result")
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_GetSetParams() {
	// Change the defaults from their norm so, we know they've got values we can check against.
	origSanct := sanction.DefaultImmediateSanctionMinDeposit
	origUnsanct := sanction.DefaultImmediateUnsanctionMinDeposit
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origSanct
		sanction.DefaultImmediateUnsanctionMinDeposit = origUnsanct
	}()
	sanction.DefaultImmediateSanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("sanct", 93))
	sanction.DefaultImmediateUnsanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("usanct", 49))

	store := s.GetStore()
	s.Require().NotPanics(func() {
		s.Keeper.OnlyTestsDeleteParam(store, keeper.ParamNameImmediateSanctionMinDeposit)
	}, "deleteParam(%q)", keeper.ParamNameImmediateSanctionMinDeposit)
	s.Require().NotPanics(func() {
		s.Keeper.OnlyTestsDeleteParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit)
	}, "deleteParam(%q)", keeper.ParamNameImmediateUnsanctionMinDeposit)

	s.Run("get with no entries in store", func() {
		expected := sanction.DefaultParams()
		var actual *sanction.Params
		testGet := func() {
			actual = s.Keeper.GetParams(s.SdkCtx)
		}
		s.Require().NotPanics(testGet, "GetParams")
		s.Assert().Equal(expected, actual, "GetParams result")
	})

	tests := []struct {
		name      string
		setInput  *sanction.Params
		getOutput *sanction.Params
	}{
		{
			name: "params with nils",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: nil,
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: nil,
			},
		},
		{
			name: "empty coins",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.Coins{},
				ImmediateUnsanctionMinDeposit: sdk.Coins{},
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: nil,
			},
		},
		{
			name: "only sanction",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("sanct", 66)),
				ImmediateUnsanctionMinDeposit: nil,
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("sanct", 66)),
				ImmediateUnsanctionMinDeposit: nil,
			},
		},
		{
			name: "only unsanction",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("unsuns", 5555)),
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("unsuns", 5555)),
			},
		},
		{
			name: "with both",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("sss", 123)),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("uuu", 456)),
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("sss", 123)),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("uuu", 456)),
			},
		},
		{
			name:      "nil",
			setInput:  nil,
			getOutput: sanction.DefaultParams(),
		},
	}

	paramsUpdatedEvent, eventErr := sdk.TypedEventToEvent(&sanction.EventParamsUpdated{})
	s.Require().NoError(eventErr, "sdk.TypedEventToEvent(&sanction.EventParamsUpdated{})")
	expectedEvents := sdk.Events{paramsUpdatedEvent}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := s.SdkCtx.WithEventManager(em)
			s.RequireNotPanicsNoError(func() error {
				return s.Keeper.SetParams(ctx, tc.setInput)
			}, "SetParams")
			actualEvents := em.Events()
			s.Assert().Equal(expectedEvents, actualEvents, "events emitted during SetParams")
			var actual *sanction.Params
			testGet := func() {
				actual = s.Keeper.GetParams(ctx)
			}
			s.Require().NotPanics(testGet, "GetParams")
			if !s.Assert().Equal(tc.getOutput, actual, "GetParams result") {
				if actual != nil {
					// it failed, but the coins aren't easy to read in that output, so be helpful here.
					s.Assert().Equal(tc.getOutput.ImmediateSanctionMinDeposit.String(),
						actual.ImmediateSanctionMinDeposit.String(), "ImmediateSanctionMinDeposit")
					s.Assert().Equal(tc.getOutput.ImmediateUnsanctionMinDeposit.String(),
						actual.ImmediateUnsanctionMinDeposit.String(), "ImmediateUnsanctionMinDeposit")
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_IterateParams() {
	type kvPair struct {
		key   string
		value string
	}

	store := s.GetStore()
	s.Require().NotPanics(func() {
		s.Keeper.OnlyTestsDeleteParam(store, keeper.ParamNameImmediateSanctionMinDeposit)
	}, "deleteParam(%q)", keeper.ParamNameImmediateSanctionMinDeposit)
	s.Require().NotPanics(func() {
		s.Keeper.OnlyTestsDeleteParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit)
	}, "deleteParam(%q)", keeper.ParamNameImmediateUnsanctionMinDeposit)

	s.Run("no entries", func() {
		var actual []kvPair
		cb := func(name, value string) bool {
			actual = append(actual, kvPair{key: name, value: value})
			return false
		}
		testFunc := func() {
			s.Keeper.IterateParams(s.SdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateParams")
		s.Assert().Empty(actual, "params iterated")
	})

	// They should be iterated in alphabetical order by key, so they're ordered as such here.
	expected := []kvPair{
		{key: "param1", value: "value for param1"},
		{key: "param2", value: "param2 value"},
		{key: "param3", value: "the param3 value"},
		{key: "param4", value: "This is param4's value."},
		{key: "param5", value: "5valuecoin"},
	}
	// Write them in reverse order from expected.
	for i := len(expected) - 1; i >= 0; i-- {
		s.Require().NotPanics(func() {
			s.Keeper.OnlyTestsSetParam(store, expected[i].key, expected[i].value)
		}, "setParam(%q, %q)", expected[i].key, expected[i].value)
	}

	s.Run("full iteration", func() {
		var actual []kvPair
		cb := func(name, value string) bool {
			actual = append(actual, kvPair{key: name, value: value})
			return false
		}
		testFunc := func() {
			s.Keeper.IterateParams(s.SdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateParams")
		s.Assert().Equal(expected, actual, "params iterated")
	})

	s.Run("stop after 3", func() {
		exp := []kvPair{expected[0], expected[1], expected[2]}
		var actual []kvPair
		cb := func(name, value string) bool {
			actual = append(actual, kvPair{key: name, value: value})
			return len(actual) >= len(exp)
		}
		testFunc := func() {
			s.Keeper.IterateParams(s.SdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateParams")
		s.Assert().Equal(exp, actual, "params iterated")
	})

	s.Run("stop after 1", func() {
		exp := []kvPair{expected[0]}
		var actual []kvPair
		cb := func(name, value string) bool {
			actual = append(actual, kvPair{key: name, value: value})
			return len(actual) >= len(exp)
		}
		testFunc := func() {
			s.Keeper.IterateParams(s.SdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateParams")
		s.Assert().Equal(exp, actual, "params iterated")
	})
}

func (s *KeeperTestSuite) TestKeeper_GetImmediateSanctionMinDeposit() {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	// Set the defaults to different things to help sus out problems.
	origS := sanction.DefaultImmediateSanctionMinDeposit
	origU := sanction.DefaultImmediateUnsanctionMinDeposit
	sanction.DefaultImmediateSanctionMinDeposit = cz("3dflts")
	sanction.DefaultImmediateUnsanctionMinDeposit = cz("6dfltu")
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origS
		sanction.DefaultImmediateUnsanctionMinDeposit = origU
	}()

	// prep is something that should be done at the start of a test case.
	type prep struct {
		value  string
		set    bool
		delete bool
	}

	store := s.GetStore()
	testFuncSetSanct := func() {
		s.Keeper.OnlyTestsSetParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit, "98unsanct")
	}
	s.Require().NotPanics(testFuncSetSanct, "setParam(%q, %q)", keeper.ParamNameImmediateUnsanctionMinDeposit, "98unsanct")

	tests := []struct {
		name string
		prep []prep
		exp  sdk.Coins
	}{
		{
			name: "not in store",
			prep: []prep{{delete: true}},
			exp:  sanction.DefaultImmediateSanctionMinDeposit,
		},
		{
			name: "empty string in store",
			prep: []prep{{value: "", set: true}},
			exp:  nil,
		},
		{
			name: "3sanct in store",
			prep: []prep{{value: "3sanct", set: true}},
			exp:  cz("3sanct"),
		},
		{
			name: "bad value in store",
			prep: []prep{{value: "how how", set: true}},
			exp:  sanction.DefaultImmediateSanctionMinDeposit,
		},
		{
			name: "not in store again",
			prep: []prep{{delete: true}},
			exp:  sanction.DefaultImmediateSanctionMinDeposit,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, p := range tc.prep {
				if p.set {
					testFuncSet := func() {
						s.Keeper.OnlyTestsSetParam(store, keeper.ParamNameImmediateSanctionMinDeposit, p.value)
					}
					s.Require().NotPanics(testFuncSet, "setParam(%q, %q)", keeper.ParamNameImmediateSanctionMinDeposit, p.value)
				}
				if p.delete {
					testFuncDelete := func() {
						s.Keeper.OnlyTestsDeleteParam(store, keeper.ParamNameImmediateSanctionMinDeposit)
					}
					s.Require().NotPanics(testFuncDelete, "deleteParam(%q)", keeper.ParamNameImmediateSanctionMinDeposit)
				}
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.Keeper.GetImmediateSanctionMinDeposit(s.SdkCtx)
			}
			s.Require().NotPanics(testFunc, "GetImmediateSanctionMinDeposit")
			s.Assert().Equal(tc.exp, actual, "GetImmediateSanctionMinDeposit result")
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_GetImmediateUnsanctionMinDeposit() {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	// Set the defaults to different things to help sus out problems.
	origS := sanction.DefaultImmediateSanctionMinDeposit
	origU := sanction.DefaultImmediateUnsanctionMinDeposit
	sanction.DefaultImmediateSanctionMinDeposit = cz("2dflts")
	sanction.DefaultImmediateUnsanctionMinDeposit = cz("5dfltu")
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origS
		sanction.DefaultImmediateUnsanctionMinDeposit = origU
	}()

	// prep is something that should be done at the start of a test case.
	type prep struct {
		value  string
		set    bool
		delete bool
	}

	store := s.GetStore()
	testFuncSetSanct := func() {
		s.Keeper.OnlyTestsSetParam(store, keeper.ParamNameImmediateSanctionMinDeposit, "99sanct")
	}
	s.Require().NotPanics(testFuncSetSanct, "setParam(%q, %q)", keeper.ParamNameImmediateSanctionMinDeposit, "99sanct")

	tests := []struct {
		name string
		prep []prep
		exp  sdk.Coins
	}{
		{
			name: "not in store",
			prep: []prep{{delete: true}},
			exp:  sanction.DefaultImmediateUnsanctionMinDeposit,
		},
		{
			name: "empty string in store",
			prep: []prep{{value: "", set: true}},
			exp:  nil,
		},
		{
			name: "3unsanct in store",
			prep: []prep{{value: "3unsanct", set: true}},
			exp:  cz("3unsanct"),
		},
		{
			name: "bad value in store",
			prep: []prep{{value: "what what", set: true}},
			exp:  sanction.DefaultImmediateUnsanctionMinDeposit,
		},
		{
			name: "not in store again",
			prep: []prep{{delete: true}},
			exp:  sanction.DefaultImmediateUnsanctionMinDeposit,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, p := range tc.prep {
				if p.set {
					testFuncSet := func() {
						s.Keeper.OnlyTestsSetParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit, p.value)
					}
					s.Require().NotPanics(testFuncSet, "setParam(%q, %q)", keeper.ParamNameImmediateUnsanctionMinDeposit, p.value)
				}
				if p.delete {
					testFuncDelete := func() {
						s.Keeper.OnlyTestsDeleteParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit)
					}
					s.Require().NotPanics(testFuncDelete, "deleteParam(%q)", keeper.ParamNameImmediateUnsanctionMinDeposit)
				}
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.Keeper.GetImmediateUnsanctionMinDeposit(s.SdkCtx)
			}
			s.Require().NotPanics(testFunc, "GetImmediateUnsanctionMinDeposit")
			s.Assert().Equal(tc.exp, actual, "GetImmediateUnsanctionMinDeposit result")
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_getSetDeleteParam() {
	store := s.GetStore()
	var toDelete []string

	newParamName := "new param"
	s.Run("get param that does not exist", func() {
		var actual string
		var ok bool
		testFuncGet := func() {
			actual, ok = s.Keeper.OnlyTestsGetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet, "getParam(%q)", newParamName)
		s.Assert().Equal("", actual, "getParam(%q) result string", newParamName)
		s.Assert().False(ok, "getParam(%q) result bool", newParamName)
	})

	newParamValue := "new param value"
	s.Run("set param new param", func() {
		var alreadyExists bool
		testFuncGet := func() {
			_, alreadyExists = s.Keeper.OnlyTestsGetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet, "getParam(%q) on param that should not exist yet", newParamName)
		s.Require().False(alreadyExists, "getParam(%q) result bool on param that should not exist yet", newParamName)
		testFuncSet := func() {
			s.Keeper.OnlyTestsSetParam(store, newParamName, newParamValue)
		}
		s.Require().NotPanics(testFuncSet, "setParam(%q, %q)", newParamName, newParamValue)
		toDelete = append(toDelete, newParamName)
	})

	s.Run("get param new param", func() {
		var actual string
		var ok bool
		testFuncGet := func() {
			actual, ok = s.Keeper.OnlyTestsGetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet, "getParam(%q)", newParamName)
		s.Require().True(ok, "getParam(%q) result bool", newParamName)
		s.Require().Equal(newParamValue, actual, "getParam(%q) result string", newParamName)
	})

	s.Run("set and get fruits", func() {
		name := "fruits"
		value := "bananas, apples, pears, papaya, pineapple, pomegranate"
		testFuncSet := func() {
			s.Keeper.OnlyTestsSetParam(store, name, value)
		}
		s.Require().NotPanics(testFuncSet, "setParam(%q, %q)", name, value)
		toDelete = append(toDelete, name)
		var actual string
		var ok bool
		testFuncGet := func() {
			actual, ok = s.Keeper.OnlyTestsGetParam(store, name)
		}
		s.Require().NotPanics(testFuncGet, "getParam(%q)", name)
		s.Assert().True(ok, "getParam(%q) result bool", name)
		s.Assert().Equal(value, actual, "getParam(%q) result string", name)
	})

	s.Run("get new param again", func() {
		var actual string
		var ok bool
		testFuncGet := func() {
			actual, ok = s.Keeper.OnlyTestsGetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet, "getParam(%q)", newParamName)
		s.Require().True(ok, "getParam(%q) result bool", newParamName)
		s.Require().Equal(newParamValue, actual, "getParam(%q) result string", newParamName)
	})

	s.Run("update and get first param", func() {
		var alreadyExists bool
		testFuncGet1 := func() {
			_, alreadyExists = s.Keeper.OnlyTestsGetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet1, "getParam(%q) on param that should not exist yet", newParamName)
		s.Require().True(alreadyExists, "getParam(%q) result bool on param that should not exist yet", newParamName)
		newParamValue = "this is an updated new param value"
		testFuncSet := func() {
			s.Keeper.OnlyTestsSetParam(store, newParamName, newParamValue)
		}
		s.Require().NotPanics(testFuncSet, "setParam(%q, %q)", newParamName, newParamValue)

		var actual string
		var ok bool
		testFuncGet2 := func() {
			actual, ok = s.Keeper.OnlyTestsGetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet2, "getParam(%q)", newParamName)
		s.Require().True(ok, "getParam(%q) result bool", newParamName)
		s.Require().Equal(newParamValue, actual, "getParam(%q) result string", newParamName)
	})

	for _, name := range toDelete {
		s.Run("delete "+name, func() {
			testDeleteFunc := func() {
				s.Keeper.OnlyTestsDeleteParam(store, name)
			}
			s.Require().NotPanics(testDeleteFunc, "deleteParam(%q)", name)
			var actual string
			var ok bool
			testGetFunc := func() {
				actual, ok = s.Keeper.OnlyTestsGetParam(store, name)
			}
			s.Require().NotPanics(testGetFunc, "getParam(%q)", name)
			s.Assert().False(ok, "getParam(%q) result bool", name)
			s.Assert().Equal("", actual, "getParam(%q) result string", name)
		})
	}

	s.Run("delete new param again", func() {
		testDeleteFunc := func() {
			s.Keeper.OnlyTestsDeleteParam(store, newParamName)
		}
		s.Require().NotPanics(testDeleteFunc, "deleteParam(%q)", newParamName)
	})
}

func (s *KeeperTestSuite) TestKeeper_getParamAsCoinsOrDefault() {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name     string
		setFirst bool
		setTo    string
		param    string
		dflt     sdk.Coins
		exp      sdk.Coins
	}{
		{
			name:     "unknown name",
			setFirst: false,
			param:    "unknown",
			dflt:     cz("1default"),
			exp:      cz("1default"),
		},
		{
			name:     "param not a coin",
			setFirst: true,
			setTo:    "not a coin",
			param:    "not-a-coin",
			dflt:     cz("1default"),
			exp:      cz("1default"),
		},
		{
			name:     "empty string",
			setFirst: true,
			setTo:    "",
			param:    "empty-string",
			dflt:     cz("1default"),
			exp:      nil,
		},
		{
			name:     "coin string one denom",
			setFirst: true,
			setTo:    "5acoin",
			param:    "one-denom",
			dflt:     cz("1default"),
			exp:      cz("5acoin"),
		},
		{
			name:     "coin string two denoms",
			setFirst: true,
			setTo:    "4acoin,10walnut",
			param:    "two-denom",
			dflt:     cz("1default"),
			exp:      cz("4acoin,10walnut"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setFirst {
				store := s.GetStore()
				s.Keeper.OnlyTestsSetParam(store, tc.param, tc.setTo)
				defer func() {
					s.Keeper.OnlyTestsDeleteParam(store, tc.param)
				}()
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.Keeper.OnlyTestsGetParamAsCoinsOrDefault(s.SdkCtx, tc.param, tc.dflt)
			}
			s.Require().NotPanics(testFunc, "getParamAsCoinsOrDefault")
			s.Assert().Equal(tc.exp, actual, "getParamAsCoinsOrDefault result")
		})
	}
}

func (s *KeeperTestSuite) Test_toCoinsOrDefault() {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}
	tests := []struct {
		name  string
		coins string
		dflt  sdk.Coins
		exp   sdk.Coins
	}{
		{
			name:  "empty string",
			coins: "",
			dflt:  cz("1defaultcoin,2banana"),
			exp:   nil,
		},
		{
			name:  "bad string",
			coins: "bad",
			dflt:  cz("1goodcoin,8defaults"),
			exp:   cz("1goodcoin,8defaults"),
		},
		{
			name:  "one denom",
			coins: "1particle",
			dflt:  cz("8quark"),
			exp:   cz("1particle"),
		},
		{
			name:  "two denoms",
			coins: "50handcoin,99gloves",
			dflt:  cz("42towels"),
			exp:   cz("50handcoin,99gloves"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual sdk.Coins
			testFunc := func() {
				actual = keeper.OnlyTestsToCoinsOrDefault(tc.coins, tc.dflt)
			}
			s.Require().NotPanics(testFunc, "toCoinsOrDefault")
			s.Assert().Equal(tc.exp, actual, "toCoinsOrDefault result")
		})
	}
}

func (s *KeeperTestSuite) Test_toAccAddrs() {
	tests := []struct {
		name   string
		addrs  []string
		exp    []sdk.AccAddress
		expErr []string
	}{
		{
			name:  "nil list",
			addrs: nil,
			exp:   []sdk.AccAddress{},
		},
		{
			name:  "empty list",
			addrs: []string{},
			exp:   []sdk.AccAddress{},
		},
		{
			name:  "one good address",
			addrs: []string{sdk.AccAddress("one good address").String()},
			exp:   []sdk.AccAddress{sdk.AccAddress("one good address")},
		},
		{
			name:   "one bad address",
			addrs:  []string{"one bad address"},
			expErr: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "five addresses all good",
			addrs: []string{
				sdk.AccAddress("good address 0").String(),
				sdk.AccAddress("good address 1").String(),
				sdk.AccAddress("good address 2").String(),
				sdk.AccAddress("good address 3").String(),
				sdk.AccAddress("good address 4").String(),
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("good address 0"),
				sdk.AccAddress("good address 1"),
				sdk.AccAddress("good address 2"),
				sdk.AccAddress("good address 3"),
				sdk.AccAddress("good address 4"),
			},
		},
		{
			name: "five addresses first bad",
			addrs: []string{
				"bad address 0",
				sdk.AccAddress("good address 1").String(),
				sdk.AccAddress("good address 2").String(),
				sdk.AccAddress("good address 3").String(),
				sdk.AccAddress("good address 4").String(),
			},
			expErr: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "five addresses third bad",
			addrs: []string{
				sdk.AccAddress("good address 0").String(),
				sdk.AccAddress("good address 1").String(),
				"bad address 2",
				sdk.AccAddress("good address 3").String(),
				sdk.AccAddress("good address 4").String(),
			},
			expErr: []string{"invalid address[2]", "decoding bech32 failed"},
		},
		{
			name: "five addresses fifth bad",
			addrs: []string{
				sdk.AccAddress("good address 0").String(),
				sdk.AccAddress("good address 1").String(),
				sdk.AccAddress("good address 2").String(),
				sdk.AccAddress("good address 3").String(),
				"bad address 4",
			},
			expErr: []string{"invalid address[4]", "decoding bech32 failed"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual []sdk.AccAddress
			var err error
			testFunc := func() {
				actual, err = keeper.OnlyTestsToAccAddrs(tc.addrs)
			}
			s.Require().NotPanics(testFunc, "toAccAddrs")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "toAccAddrs error")
			s.Assert().Equal(tc.exp, actual, "toAccAddrs result")
		})
	}
}
