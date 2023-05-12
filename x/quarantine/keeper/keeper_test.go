package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

// updateQR updates the AccAddresses using the provided addrs.
// Any AccAddress that is 1 byte long and can be an index in addrs,
// is replaced by the addrs entry using that byte as the index.
// E.g. if UnacceptedFromAddresses is []sdk.AccAddress{{1}}, then it will be replaced with addrs[1].
func updateQR(addrs []sdk.AccAddress, record *quarantine.QuarantineRecord) {
	if record != nil {
		for i, addr := range record.UnacceptedFromAddresses {
			if len(addr) == 1 && int(addr[0]) < len(addrs) {
				record.UnacceptedFromAddresses[i] = addrs[addr[0]]
			}
		}
		for i, addr := range record.AcceptedFromAddresses {
			if len(addr) == 1 && int(addr[0]) < len(addrs) {
				record.AcceptedFromAddresses[i] = addrs[addr[0]]
			}
		}
	}
}

// qrsi is just a shorter way to create a *quarantine.QuarantineRecordSuffixIndex
func qrsi(suffixes ...[]byte) *quarantine.QuarantineRecordSuffixIndex {
	rv := &quarantine.QuarantineRecordSuffixIndex{}
	if len(suffixes) > 0 {
		rv.RecordSuffixes = suffixes
	}
	return rv
}

// qrsis is just a shorter way to create []*quarantine.QuarantineRecordSuffixIndex
// Combine with qrsi for true power.
func qrsis(vals ...*quarantine.QuarantineRecordSuffixIndex) []*quarantine.QuarantineRecordSuffixIndex {
	return vals
}

// accs is just a shorter way to create an []sdk.AccAddress
func accs(accz ...sdk.AccAddress) []sdk.AccAddress {
	return accz
}

type TestSuite struct {
	suite.Suite

	app        *simapp.SimApp
	sdkCtx     sdk.Context
	stdlibCtx  context.Context
	keeper     keeper.Keeper
	bankKeeper bankkeeper.Keeper

	blockTime time.Time
	addr1     sdk.AccAddress
	addr2     sdk.AccAddress
	addr3     sdk.AccAddress
	addr4     sdk.AccAddress
	addr5     sdk.AccAddress
}

func (s *TestSuite) SetupTest() {
	s.blockTime = tmtime.Now()
	s.app = simapp.Setup(s.T(), false)
	s.sdkCtx = s.app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockHeader(tmproto.Header{Time: s.blockTime})
	s.stdlibCtx = sdk.WrapSDKContext(s.sdkCtx)
	s.keeper = s.app.QuarantineKeeper
	s.bankKeeper = s.app.BankKeeper

	addrs := simapp.AddTestAddrsIncremental(s.app, s.sdkCtx, 5, sdk.NewInt(1_000_000_000))
	s.addr1 = addrs[0]
	s.addr2 = addrs[1]
	s.addr3 = addrs[2]
	s.addr4 = addrs[3]
	s.addr5 = addrs[4]
}

func (s *TestSuite) cz(coins string) sdk.Coins {
	s.T().Helper()
	rv, err := sdk.ParseCoinsNormalized(coins)
	s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
	return rv
}

func (s *TestSuite) AssertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestGetFundsHolder() {
	s.Run("initial value", func() {
		expected := authtypes.NewModuleAddress(quarantine.ModuleName)

		actual := s.keeper.GetFundsHolder()
		s.Assert().Equal(expected, actual, "funds holder")
	})

	s.Run("set to nil", func() {
		k := s.keeper.WithFundsHolder(nil)

		actual := k.GetFundsHolder()
		s.Assert().Nil(actual, "funds holder")
	})

	s.Run("set to something else", func() {
		k := s.keeper.WithFundsHolder(s.addr1)

		actual := k.GetFundsHolder()
		s.Assert().Equal(s.addr1, actual, "funds holder")
	})
}

func (s *TestSuite) TestQuarantineOptInOut() {
	s.Run("is quarantined before opting in", func() {
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().False(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt in and check", func() {
		err := s.keeper.SetOptIn(s.sdkCtx, s.addr2)
		s.Require().NoError(err, "SetOptIn addr2")
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().True(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt in again and check", func() {
		err := s.keeper.SetOptIn(s.sdkCtx, s.addr2)
		s.Require().NoError(err, "SetOptIn addr2")
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().True(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt out and check", func() {
		err := s.keeper.SetOptOut(s.sdkCtx, s.addr2)
		s.Require().NoError(err, "SetOptOut addr2")
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().False(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt out again and check", func() {
		err := s.keeper.SetOptOut(s.sdkCtx, s.addr2)
		s.Require().NoError(err, "SetOptOut addr2")
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().False(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt in event", func() {
		ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
		err := s.keeper.SetOptIn(ctx, s.addr3)
		s.Require().NoError(err, "SetOptIn addr3")

		expected := sdk.Events{
			{
				Type: "cosmos.quarantine.v1beta1.EventOptIn",
				Attributes: []abci.EventAttribute{
					{
						Key:   []byte("to_address"),
						Value: []byte(fmt.Sprintf(`"%s"`, s.addr3.String())),
					},
				},
			},
		}
		actual := ctx.EventManager().Events()
		s.Assert().Equal(expected, actual, "emitted events")
	})

	s.Run("opt out event", func() {
		ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
		err := s.keeper.SetOptOut(ctx, s.addr3)
		s.Require().NoError(err, "SetOptOut addr3")

		expected := sdk.Events{
			{
				Type: "cosmos.quarantine.v1beta1.EventOptOut",
				Attributes: []abci.EventAttribute{
					{
						Key:   []byte("to_address"),
						Value: []byte(fmt.Sprintf(`"%s"`, s.addr3.String())),
					},
				},
			},
		}
		actual := ctx.EventManager().Events()
		s.Assert().Equal(expected, actual, "emitted events")
	})
}

func (s *TestSuite) TestQuarantinedAccountsIterateAndGetAll() {
	// Opt in all of them except addr4.
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr1), "SetOptIn addr1")
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr2), "SetOptIn addr2")
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr3), "SetOptIn addr3")
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr5), "SetOptIn addr5")

	// Now opt out addr2.
	s.Require().NoError(s.keeper.SetOptOut(s.sdkCtx, s.addr2), "SetOptOut addr2")

	allAddrs := accs(s.addr1, s.addr3, s.addr5)
	sort.Slice(allAddrs, func(i, j int) bool {
		return bytes.Compare(allAddrs[i], allAddrs[j]) < 0
	})

	s.Run("IterateQuarantinedAccounts", func() {
		expected := allAddrs
		addrs := make([]sdk.AccAddress, 0, len(expected))
		callback := func(toAddr sdk.AccAddress) bool {
			addrs = append(addrs, toAddr)
			return false
		}

		testFunc := func() {
			s.keeper.IterateQuarantinedAccounts(s.sdkCtx, callback)
		}
		s.Require().NotPanics(testFunc, "IterateQuarantinedAccounts")
		s.Assert().Equal(expected, addrs, "iterated addrs")
	})

	s.Run("IterateQuarantinedAccounts early stop", func() {
		stopLen := 2
		expected := allAddrs[:stopLen]
		addrs := make([]sdk.AccAddress, 0, stopLen)
		callback := func(toAddr sdk.AccAddress) bool {
			addrs = append(addrs, toAddr)
			return len(addrs) >= stopLen
		}

		testFunc := func() {
			s.keeper.IterateQuarantinedAccounts(s.sdkCtx, callback)
		}
		s.Require().NotPanics(testFunc, "IterateQuarantinedAccounts")
		s.Assert().Equal(expected, addrs, "iterated addrs")
	})

	s.Run("GetAllQuarantinedAccounts", func() {
		expected := make([]string, len(allAddrs))
		for i, addr := range allAddrs {
			expected[i] = addr.String()
		}

		var actual []string
		testFunc := func() {
			actual = s.keeper.GetAllQuarantinedAccounts(s.sdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllQuarantinedAccounts")
		s.Assert().Equal(expected, actual, "GetAllQuarantinedAccounts")
	})
}

func (s *TestSuite) TestAutoResponseGetSet() {
	allAddrs := accs(s.addr1, s.addr2, s.addr3, s.addr4, s.addr5)
	allResps := []quarantine.AutoResponse{
		quarantine.AUTO_RESPONSE_ACCEPT,
		quarantine.AUTO_RESPONSE_DECLINE,
		quarantine.AUTO_RESPONSE_UNSPECIFIED,
	}

	s.Run("GetAutoResponse on unset addrs", func() {
		expected := quarantine.AUTO_RESPONSE_UNSPECIFIED
		for i, addrI := range allAddrs {
			for j, addrJ := range allAddrs {
				if i == j {
					continue
				}
				actual := s.keeper.GetAutoResponse(s.sdkCtx, addrI, addrJ)
				s.Assert().Equal(expected, actual, "GetAutoResponse addr%d addr%d", i+1, j+1)
			}
		}
	})

	s.Run("GetAutoResponse on same addr", func() {
		expected := quarantine.AUTO_RESPONSE_ACCEPT
		for i, addr := range allAddrs {
			actual := s.keeper.GetAutoResponse(s.sdkCtx, addr, addr)
			s.Assert().Equal(expected, actual, "GetAutoResponse addr%d addr%d", i+1, i+1)
		}
	})

	for _, expected := range allResps {
		s.Run(fmt.Sprintf("set %s", expected), func() {
			testFunc := func() {
				s.keeper.SetAutoResponse(s.sdkCtx, s.addr3, s.addr1, expected)
			}
			s.Require().NotPanics(testFunc, "SetAutoResponse addr3 addr1 %s", expected)
			actual := s.keeper.GetAutoResponse(s.sdkCtx, s.addr3, s.addr1)
			s.Assert().Equal(expected, actual, "GetAutoResponse after set %s", expected)
		})
	}

	s.Run("IsAutoAccept", func() {
		testFunc := func() {
			s.keeper.SetAutoResponse(s.sdkCtx, s.addr4, s.addr2, quarantine.AUTO_RESPONSE_ACCEPT)
		}
		s.Require().NotPanics(testFunc, "SetAutoResponse")

		actual42 := s.keeper.IsAutoAccept(s.sdkCtx, s.addr4, s.addr2)
		s.Assert().True(actual42, "IsAutoAccept addr4 addr2")
		actual43 := s.keeper.IsAutoAccept(s.sdkCtx, s.addr4, s.addr3)
		s.Assert().False(actual43, "IsAutoAccept addr4 addr3")
		actual44 := s.keeper.IsAutoAccept(s.sdkCtx, s.addr4, s.addr4)
		s.Assert().True(actual44, "IsAutoAccept self")
	})

	s.Run("IsAutoDecline", func() {
		testFunc := func() {
			s.keeper.SetAutoResponse(s.sdkCtx, s.addr5, s.addr2, quarantine.AUTO_RESPONSE_DECLINE)
		}
		s.Require().NotPanics(testFunc, "SetAutoResponse")

		actual52 := s.keeper.IsAutoDecline(s.sdkCtx, s.addr5, s.addr2)
		s.Assert().True(actual52, "IsAutoDecline addr5 addr2")
		actual53 := s.keeper.IsAutoDecline(s.sdkCtx, s.addr5, s.addr3)
		s.Assert().False(actual53, "IsAutoDecline addr5 addr3")
		actual55 := s.keeper.IsAutoDecline(s.sdkCtx, s.addr5, s.addr5)
		s.Assert().False(actual55, "IsAutoDecline self")
	})
}

func (s *TestSuite) TestAutoResponsesItateAndGetAll() {
	setAutoTestFunc := func(addrA, addrB sdk.AccAddress, response quarantine.AutoResponse) func() {
		return func() {
			s.keeper.SetAutoResponse(s.sdkCtx, addrA, addrB, response)
		}
	}
	// Shorten up the names a bit.
	arAccept := quarantine.AUTO_RESPONSE_ACCEPT
	arDecline := quarantine.AUTO_RESPONSE_DECLINE
	arUnspecified := quarantine.AUTO_RESPONSE_UNSPECIFIED

	// Set up some auto-responses.
	// This is purposely done in a random order.

	// Set account 1 to auto-accept from all.
	// Set account 2 to auto-accept from all.
	// Set 3 to auto-decline from all
	// Set 4 to auto-accept from 2 and 3 and auto-decline from 5
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr3, arAccept), "4 <- 3 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr2, arDecline), "3 <- 2 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr5, arAccept), "2 <- 5 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr5, arDecline), "3 <- 5 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr4, arAccept), "2 <- 4 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr1, arAccept), "2 <- 1 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr5, arDecline), "4 <- 5 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr4, arDecline), "3 <- 4 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr1, arDecline), "3 <- 1 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr1, s.addr5, arAccept), "1 <- 5 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr1, s.addr2, arAccept), "1 <- 2 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr1, s.addr4, arAccept), "1 <- 4 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr3, arAccept), "2 <- 3 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr2, arAccept), "4 <- 2 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr1, s.addr3, arAccept), "1 <- 3 accept")

	// Now undo/change a few of those.
	// Set 2 to unspecified from 3 and 4
	// Set 3 to unspecified from 5
	// Set 4 to auto-decline from 3 and auto-accept from 5
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr5, arAccept), "4 <- 5 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr3, arUnspecified), "2 <- 3 unspecified")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr5, arUnspecified), "3 <- 5 unspecified")
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr3, arDecline), "4 <- 3 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr4, arUnspecified), "2 <- 4 unspecified")

	// Setup result:
	// 1 <- 2 = accept   3 <- 1 = decline
	// 1 <- 3 = accept   3 <- 2 = decline
	// 1 <- 4 = accept   3 <- 4 = decline
	// 1 <- 5 = accept   4 <- 2 = accept
	// 2 <- 1 = accept   4 <- 3 = decline
	// 2 <- 5 = accept   4 <- 5 = accept

	// Let's hope the addresses are actually incremental or else this gets a lot tougher to define.
	type callbackArgs struct {
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		response quarantine.AutoResponse
	}

	allArgs := []callbackArgs{
		{toAddr: s.addr1, fromAddr: s.addr2, response: arAccept},
		{toAddr: s.addr1, fromAddr: s.addr3, response: arAccept},
		{toAddr: s.addr1, fromAddr: s.addr4, response: arAccept},
		{toAddr: s.addr1, fromAddr: s.addr5, response: arAccept},
		{toAddr: s.addr2, fromAddr: s.addr1, response: arAccept},
		{toAddr: s.addr2, fromAddr: s.addr5, response: arAccept},
		{toAddr: s.addr3, fromAddr: s.addr1, response: arDecline},
		{toAddr: s.addr3, fromAddr: s.addr2, response: arDecline},
		{toAddr: s.addr3, fromAddr: s.addr4, response: arDecline},
		{toAddr: s.addr4, fromAddr: s.addr2, response: arAccept},
		{toAddr: s.addr4, fromAddr: s.addr3, response: arDecline},
		{toAddr: s.addr4, fromAddr: s.addr5, response: arAccept},
	}

	s.Run("IterateAutoResponses all", func() {
		expected := allArgs
		actualAllArgs := make([]callbackArgs, 0, len(allArgs))
		callback := func(toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) bool {
			actualAllArgs = append(actualAllArgs, callbackArgs{toAddr: toAddr, fromAddr: fromAddr, response: response})
			return false
		}
		testFunc := func() {
			s.keeper.IterateAutoResponses(s.sdkCtx, nil, callback)
		}
		s.Require().NotPanics(testFunc, "IterateAutoResponses")
		s.Assert().Equal(expected, actualAllArgs, "iterated args")
	})

	for i, addr := range accs(s.addr1, s.addr2, s.addr3, s.addr4, s.addr5) {
		s.Run(fmt.Sprintf("IterateAutoResponses addr%d", i+1), func() {
			var expected []callbackArgs
			for _, args := range allArgs {
				if addr.Equals(args.toAddr) {
					expected = append(expected, args)
				}
			}
			var actual []callbackArgs
			callback := func(toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) bool {
				actual = append(actual, callbackArgs{toAddr: toAddr, fromAddr: fromAddr, response: response})
				return false
			}
			testFunc := func() {
				s.keeper.IterateAutoResponses(s.sdkCtx, addr, callback)
			}
			s.Require().NotPanics(testFunc, "IterateAutoResponses")
			s.Assert().Equal(expected, actual, "iterated args")
		})
	}

	s.Run("IterateAutoResponses stop early", func() {
		stopLen := 4
		expected := allArgs[:stopLen]
		actual := make([]callbackArgs, 0, stopLen)
		callback := func(toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) bool {
			actual = append(actual, callbackArgs{toAddr: toAddr, fromAddr: fromAddr, response: response})
			return len(actual) >= stopLen
		}
		testFunc := func() {
			s.keeper.IterateAutoResponses(s.sdkCtx, nil, callback)
		}
		s.Require().NotPanics(testFunc, "IterateAutoResponses")
		s.Assert().Equal(expected, actual, "iterated args")
	})

	s.Run("GetAllAutoResponseEntries", func() {
		expected := make([]*quarantine.AutoResponseEntry, len(allArgs))
		for i, args := range allArgs {
			expected[i] = &quarantine.AutoResponseEntry{
				ToAddress:   args.toAddr.String(),
				FromAddress: args.fromAddr.String(),
				Response:    args.response,
			}
		}

		var actual []*quarantine.AutoResponseEntry
		testFunc := func() {
			actual = s.keeper.GetAllAutoResponseEntries(s.sdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllAutoResponseEntries")
		s.Assert().Equal(expected, actual, "GetAllAutoResponseEntries results")
	})
}

func (s *TestSuite) TestBzToQuarantineRecord() {
	cdc := s.keeper.GetCodec()

	tests := []struct {
		name     string
		bz       []byte
		expected *quarantine.QuarantineRecord
		expErr   string
	}{
		{
			name: "control",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(s.addr1),
				AcceptedFromAddresses:   accs(s.addr2),
				Coins:                   s.cz("9000bar,888foo"),
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(s.addr1),
				AcceptedFromAddresses:   accs(s.addr2),
				Coins:                   s.cz("9000bar,888foo"),
				Declined:                false,
			},
		},
		{
			name: "nil bz",
			bz:   nil,
			expected: &quarantine.QuarantineRecord{
				Coins: sdk.Coins{},
			},
		},
		{
			name: "empty bz",
			bz:   nil,
			expected: &quarantine.QuarantineRecord{
				Coins: sdk.Coins{},
			},
		},
		{
			name: "not a quarantine record",
			bz: cdc.MustMarshal(&quarantine.AutoResponseEntry{
				ToAddress:   s.addr4.String(),
				FromAddress: s.addr3.String(),
				Response:    quarantine.AUTO_RESPONSE_ACCEPT,
			}),
			expErr: "proto: wrong wireType = 0 for field Coins",
		},
		{
			name:   "unknown bytes",
			bz:     []byte{0x75, 110, 0153, 0x6e, 0157, 119, 0156, 0xff, 0142, 0x79, 116, 0x65, 0163},
			expErr: "proto: illegal wireType 7",
		},
		{
			name: "declined",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(s.addr1),
				AcceptedFromAddresses:   accs(s.addr2),
				Coins:                   s.cz("9001bar,889foo"),
				Declined:                true,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(s.addr1),
				AcceptedFromAddresses:   accs(s.addr2),
				Coins:                   s.cz("9001bar,889foo"),
				Declined:                true,
			},
		},
		{
			name: "no unaccepted",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   accs(s.addr2, s.addr1, s.addr3),
				Coins:                   s.cz("9002bar,890foo"),
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   accs(s.addr2, s.addr1, s.addr3),
				Coins:                   s.cz("9002bar,890foo"),
				Declined:                false,
			},
		},
		{
			name: "no accepted",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(s.addr4, s.addr2, s.addr5),
				AcceptedFromAddresses:   []sdk.AccAddress{},
				Coins:                   s.cz("9003bar,891foo"),
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(s.addr4, s.addr2, s.addr5),
				AcceptedFromAddresses:   nil,
				Coins:                   s.cz("9003bar,891foo"),
				Declined:                false,
			},
		},
		{
			name: "no coins",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(s.addr1),
				AcceptedFromAddresses:   accs(s.addr2),
				Coins:                   nil,
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(s.addr1),
				AcceptedFromAddresses:   accs(s.addr2),
				Coins:                   sdk.Coins{},
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual *quarantine.QuarantineRecord
			var err error
			testFunc := func() {
				actual, err = s.keeper.BzToQuarantineRecord(tc.bz)
			}
			s.Require().NotPanics(testFunc, "bzToQuarantineRecord")
			if len(tc.expErr) > 0 {
				s.Require().EqualError(err, tc.expErr, "bzToQuarantineRecord error: %v", actual)
			} else {
				s.Require().NoError(err, "bzToQuarantineRecord error")
				s.Assert().Equal(tc.expected, actual, "bzToQuarantineRecord record")
			}
		})

		s.Run("must "+tc.name, func() {
			var actual *quarantine.QuarantineRecord
			testFunc := func() {
				actual = s.keeper.MustBzToQuarantineRecord(tc.bz)
			}
			if len(tc.expErr) > 0 {
				s.Require().PanicsWithError(tc.expErr, testFunc, "mustBzToQuarantineRecord: %v", actual)
			} else {
				s.Require().NotPanics(testFunc, "mustBzToQuarantineRecord")
				s.Assert().Equal(tc.expected, actual, "mustBzToQuarantineRecord record")
			}
		})
	}
}

func (s *TestSuite) TestQuarantineRecordGetSet() {
	s.Run("get does not exist", func() {
		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, s.addr1, s.addr2)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Nil(actual, "GetQuarantineRecord")
	})

	s.Run("get multiple froms does not exist", func() {
		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, s.addr1, s.addr2, s.addr3, s.addr5)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Nil(actual, "GetQuarantineRecord")
	})

	s.Run("get no froms", func() {
		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, s.addr5)
		}
		s.Assert().Panics(testFunc, "GetQuarantineRecord")
		s.Assert().Nil(actual, "GetQuarantineRecord")
	})

	s.Run("set get one unaccepted no accepted", func() {
		toAddr := MakeTestAddr("sgouna", 0)
		uFromAddr := MakeTestAddr("sgouna", 1)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: accs(uFromAddr),
			AcceptedFromAddresses:   nil,
			Coins:                   s.cz("456bar,1233foo"),
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actual *quarantine.QuarantineRecord
		testFuncGet := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr)
		}
		var actualBackwards *quarantine.QuarantineRecord
		testFuncGetBackwards := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, uFromAddr, toAddr)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecord") {
			s.Assert().Equal(expected, actual, "GetQuarantineRecord")
		}
		if s.Assert().NotPanics(testFuncGetBackwards, "GetQuarantineRecord wrong to/from order") {
			s.Assert().Nil(actualBackwards, "GetQuarantineRecord wrong to/from order")
		}
	})

	s.Run("set get one unaccepted one accepted", func() {
		toAddr := MakeTestAddr("sgouoa", 0)
		uFromAddr := MakeTestAddr("sgouoa", 1)
		aFromAddr := MakeTestAddr("sgouoa", 2)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: accs(uFromAddr),
			AcceptedFromAddresses:   accs(aFromAddr),
			Coins:                   sdk.Coins{},
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actualUA *quarantine.QuarantineRecord
		testFuncGetOrderUA := func() {
			actualUA = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr, aFromAddr)
		}
		var actualAU *quarantine.QuarantineRecord
		testFuncGetOrderAU := func() {
			actualAU = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, aFromAddr, uFromAddr)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		if s.Assert().NotPanics(testFuncGetOrderUA, "GetQuarantineRecord order: ua") {
			s.Assert().Equal(expected, actualUA, "GetQuarantineRecord order: ua")
		}
		if s.Assert().NotPanics(testFuncGetOrderAU, "GetQuarantineRecord order: au") {
			s.Assert().Equal(expected, actualAU, "GetQuarantineRecord order: au")
		}
	})

	s.Run("set get two unaccepted no accepted", func() {
		toAddr := MakeTestAddr("sgtuna", 0)
		uFromAddr1 := MakeTestAddr("sgtuna", 1)
		uFromAddr2 := MakeTestAddr("sgtuna", 2)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: accs(uFromAddr1, uFromAddr2),
			AcceptedFromAddresses:   nil,
			Coins:                   sdk.Coins{},
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actual12 *quarantine.QuarantineRecord
		testFuncGet12 := func() {
			actual12 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr1, uFromAddr2)
		}
		var actual21 *quarantine.QuarantineRecord
		testFuncGet21 := func() {
			actual21 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr2, uFromAddr1)
		}
		var actualJust1 *quarantine.QuarantineRecord
		testFuncGetJust1 := func() {
			actualJust1 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr1)
		}
		var actualJust2 *quarantine.QuarantineRecord
		testFuncGetJust2 := func() {
			actualJust2 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr2)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		if s.Assert().NotPanics(testFuncGet12, "GetQuarantineRecord order: 1 2") {
			s.Assert().Equal(expected, actual12, "GetQuarantineRecord result order: 1 2")
		}
		if s.Assert().NotPanics(testFuncGet21, "GetQuarantineRecord order: 2 1") {
			s.Assert().Equal(expected, actual21, "GetQuarantineRecord result order: 2 1")
		}
		if s.Assert().NotPanics(testFuncGetJust1, "GetQuarantineRecord just 1") {
			s.Assert().Nil(actualJust1, "GetQuarantineRecord just 1")
		}
		if s.Assert().NotPanics(testFuncGetJust2, "GetQuarantineRecord just 2") {
			s.Assert().Nil(actualJust2, "GetQuarantineRecord just 2")
		}
	})

	s.Run("set get no unaccepted one accepted", func() {
		toAddr := MakeTestAddr("sgnuoa", 0)
		aFromAddr := MakeTestAddr("sgnuoa", 1)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: nil,
			AcceptedFromAddresses:   accs(aFromAddr),
			Coins:                   sdk.Coins{},
			Declined:                false,
		}

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actual *quarantine.QuarantineRecord
		testFuncGet := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, aFromAddr)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		s.Require().NotPanics(testFuncGet, "GetQuarantineRecord")
		s.Assert().Nil(actual, "GetQuarantineRecord")
	})

	s.Run("set get no unaccepted two accepted", func() {
		toAddr := MakeTestAddr("sgnuta", 0)
		aFromAddr1 := MakeTestAddr("sgnuta", 1)
		aFromAddr2 := MakeTestAddr("sgnuta", 2)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: nil,
			AcceptedFromAddresses:   accs(aFromAddr1, aFromAddr2),
			Coins:                   sdk.Coins{},
			Declined:                false,
		}

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actual12 *quarantine.QuarantineRecord
		testFuncGet12 := func() {
			actual12 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, aFromAddr1, aFromAddr2)
		}
		var actual21 *quarantine.QuarantineRecord
		testFuncGet21 := func() {
			actual21 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, aFromAddr2, aFromAddr1)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		if s.Assert().NotPanics(testFuncGet12, "GetQuarantineRecord order: 1 2") {
			s.Assert().Nil(actual12, "GetQuarantineRecord order: 1 2")
		}
		if s.Assert().NotPanics(testFuncGet21, "GetQuarantineRecord order: 2 1") {
			s.Assert().Nil(actual21, "GetQuarantineRecord order: 2 1")
		}
	})

	s.Run("set get two unaccepted one accepted", func() {
		toAddr := MakeTestAddr("sgtuoa", 0)
		uFromAddr1 := MakeTestAddr("sgtuoa", 1)
		uFromAddr2 := MakeTestAddr("sgtuoa", 2)
		aFromAddr := MakeTestAddr("sgtuoa", 3)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: accs(uFromAddr1, uFromAddr2),
			AcceptedFromAddresses:   accs(aFromAddr),
			Coins:                   sdk.Coins{},
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")

		positiveTests := []struct {
			name      string
			fromAddrs []sdk.AccAddress
		}{
			{"1 2 a", accs(uFromAddr1, uFromAddr2, aFromAddr)},
			{"1 a 2", accs(uFromAddr1, aFromAddr, uFromAddr2)},
			{"2 1 a", accs(uFromAddr2, uFromAddr1, aFromAddr)},
			{"2 a 1", accs(uFromAddr2, aFromAddr, uFromAddr1)},
			{"a 1 2", accs(aFromAddr, uFromAddr1, uFromAddr2)},
			{"a 2 1", accs(aFromAddr, uFromAddr2, uFromAddr1)},
		}
		for _, tc := range positiveTests {
			s.Run("GetQuarantineRecord order "+tc.name, func() {
				var actual *quarantine.QuarantineRecord
				testFunc := func() {
					actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, tc.fromAddrs...)
				}
				if s.Assert().NotPanics(testFunc, "GetQuarantineRecord") {
					s.Assert().Equal(expected, actual, "GetQuarantineRecord")
				}
			})
		}

		negativeTests := []struct {
			name      string
			fromAddrs []sdk.AccAddress
		}{
			{"1", accs(uFromAddr1)},
			{"2", accs(uFromAddr2)},
			{"a", accs(aFromAddr)},
			{"1 1", accs(uFromAddr1, uFromAddr1)},
			{"1 2", accs(uFromAddr1, uFromAddr2)},
			{"1 a", accs(uFromAddr1, aFromAddr)},
			{"2 1", accs(uFromAddr2, uFromAddr1)},
			{"2 2", accs(uFromAddr2, uFromAddr2)},
			{"2 a", accs(uFromAddr2, aFromAddr)},
			{"a 1", accs(aFromAddr, uFromAddr1)},
			{"a 2", accs(aFromAddr, uFromAddr2)},
			{"a a", accs(aFromAddr, aFromAddr)},
			{"1 1 2", accs(uFromAddr1, uFromAddr1, uFromAddr2)},
			{"2 2 a", accs(uFromAddr2, uFromAddr2, aFromAddr)},
			{"1 2 a 1", accs(uFromAddr1, uFromAddr2, aFromAddr, uFromAddr1)},
			{"1 2 2 a", accs(uFromAddr1, uFromAddr2, uFromAddr2, aFromAddr)},
			{"a 1 2 a", accs(aFromAddr, uFromAddr1, uFromAddr2, aFromAddr)},
		}
		for _, tc := range negativeTests {
			s.Run("GetQuarantineRecord order "+tc.name, func() {
				var actual *quarantine.QuarantineRecord
				testFunc := func() {
					actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, tc.fromAddrs...)
				}
				if s.Assert().NotPanics(testFunc, "GetQuarantineRecord") {
					s.Assert().Nil(actual, "GetQuarantineRecord")
				}
			})
		}
	})

	s.Run("set get one unaccepted two accepted", func() {
		toAddr := MakeTestAddr("sgouta", 0)
		uFromAddr := MakeTestAddr("sgouta", 1)
		aFromAddr1 := MakeTestAddr("sgouta", 2)
		aFromAddr2 := MakeTestAddr("sgouta", 3)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: accs(uFromAddr),
			AcceptedFromAddresses:   accs(aFromAddr1, aFromAddr2),
			Coins:                   sdk.Coins{},
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")

		positiveTests := []struct {
			name      string
			fromAddrs []sdk.AccAddress
		}{
			{"1 2 u", accs(aFromAddr1, aFromAddr2, uFromAddr)},
			{"1 u 2", accs(aFromAddr1, uFromAddr, aFromAddr2)},
			{"2 1 u", accs(aFromAddr2, aFromAddr1, uFromAddr)},
			{"2 u 1", accs(aFromAddr2, uFromAddr, aFromAddr1)},
			{"u 1 2", accs(uFromAddr, aFromAddr1, aFromAddr2)},
			{"u 2 1", accs(uFromAddr, aFromAddr2, aFromAddr1)},
		}
		for _, tc := range positiveTests {
			s.Run("GetQuarantineRecord order "+tc.name, func() {
				var actual *quarantine.QuarantineRecord
				testFunc := func() {
					actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, tc.fromAddrs...)
				}
				if s.Assert().NotPanics(testFunc, "GetQuarantineRecord") {
					s.Assert().Equal(expected, actual, "GetQuarantineRecord")
				}
			})
		}

		negativeTests := []struct {
			name      string
			fromAddrs []sdk.AccAddress
		}{
			{"1", accs(aFromAddr1)},
			{"2", accs(aFromAddr2)},
			{"u", accs(uFromAddr)},
			{"1 1", accs(aFromAddr1, aFromAddr1)},
			{"1 2", accs(aFromAddr1, aFromAddr2)},
			{"1 u", accs(aFromAddr1, uFromAddr)},
			{"2 1", accs(aFromAddr2, aFromAddr1)},
			{"2 2", accs(aFromAddr2, aFromAddr2)},
			{"2 u", accs(aFromAddr2, uFromAddr)},
			{"u 1", accs(uFromAddr, aFromAddr1)},
			{"u 2", accs(uFromAddr, aFromAddr2)},
			{"u u", accs(uFromAddr, uFromAddr)},
			{"1 1 2", accs(aFromAddr1, aFromAddr1, aFromAddr2)},
			{"2 2 u", accs(aFromAddr2, aFromAddr2, uFromAddr)},
			{"1 2 u 1", accs(aFromAddr1, aFromAddr2, uFromAddr, aFromAddr1)},
			{"1 2 2 u", accs(aFromAddr1, aFromAddr2, aFromAddr2, uFromAddr)},
			{"u 1 2 u", accs(uFromAddr, aFromAddr1, uFromAddr, uFromAddr)},
		}
		for _, tc := range negativeTests {
			s.Run("GetQuarantineRecord order "+tc.name, func() {
				var actual *quarantine.QuarantineRecord
				testFunc := func() {
					actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, tc.fromAddrs...)
				}
				if s.Assert().NotPanics(testFunc, "GetQuarantineRecord") {
					s.Assert().Nil(actual, "GetQuarantineRecord")
				}
			})
		}
	})
}

func (s *TestSuite) TestGetQuarantineRecords() {
	addr0 := MakeTestAddr("gqr", 0)
	addr1 := MakeTestAddr("gqr", 1)
	addr2 := MakeTestAddr("gqr", 2)
	addr3 := MakeTestAddr("gqr", 3)

	mustCoins := func(amt string) sdk.Coins {
		coins, err := sdk.ParseCoinsNormalized(amt)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", amt)
		return coins
	}

	recordA := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: accs(addr1),
		Coins:                   mustCoins("1acoin"),
		Declined:                true,
	}
	recordB := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: accs(addr1),
		AcceptedFromAddresses:   accs(addr2),
		Coins:                   mustCoins("10bcoin"),
	}
	recordC := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: accs(addr2),
		Coins:                   mustCoins("100ccoin"),
	}
	recordD := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: accs(addr1),
		AcceptedFromAddresses:   accs(addr0, addr2),
		Coins:                   mustCoins("1000dcoin"),
	}
	recordE := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: accs(addr3),
		Coins:                   mustCoins("100000ecoin"),
	}

	testFunc := func(toAddr sdk.AccAddress, record *quarantine.QuarantineRecord) func() {
		return func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
	}

	s.Require().NotPanics(testFunc(addr0, recordA), "SetQuarantineRecord recordA")
	s.Require().NotPanics(testFunc(addr0, recordB), "SetQuarantineRecord recordB")
	s.Require().NotPanics(testFunc(addr0, recordC), "SetQuarantineRecord recordC")
	s.Require().NotPanics(testFunc(addr0, recordD), "SetQuarantineRecord recordD")
	s.Require().NotPanics(testFunc(addr0, recordE), "SetQuarantineRecord recordE")

	// Setup:
	// 0 <- 1:  1acoin declined
	// 0 <- 1 2: 10bcoin
	// 0 <- 2: 100ccoin
	// 0 <- 0 1 2: 1000dcoin
	// 0 <- 3: 10000ecoin

	qrs := func(qrz ...*quarantine.QuarantineRecord) []*quarantine.QuarantineRecord {
		return qrz
	}

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []sdk.AccAddress
		expected  []*quarantine.QuarantineRecord
	}{
		{
			name:      "to 0 from none",
			toAddr:    addr0,
			fromAddrs: accs(),
			expected:  nil,
		},
		{
			name:      "to 1 from none",
			toAddr:    addr0,
			fromAddrs: accs(),
			expected:  nil,
		},
		{
			name:      "to 2 from none",
			toAddr:    addr0,
			fromAddrs: accs(),
			expected:  nil,
		},
		{
			name:      "to 3 from none",
			toAddr:    addr0,
			fromAddrs: accs(),
			expected:  nil,
		},
		{
			name:      "to 1 from 0",
			toAddr:    addr1,
			fromAddrs: accs(addr0),
			expected:  nil,
		},
		{
			name:      "to 1 from 1",
			toAddr:    addr1,
			fromAddrs: accs(addr1),
			expected:  nil,
		},
		{
			name:      "to 1 from 2",
			toAddr:    addr1,
			fromAddrs: accs(addr2),
			expected:  nil,
		},
		{
			name:      "to 1 from 3",
			toAddr:    addr1,
			fromAddrs: accs(addr3),
			expected:  nil,
		},
		{
			name:      "to 2 from 0",
			toAddr:    addr2,
			fromAddrs: accs(addr0),
			expected:  nil,
		},
		{
			name:      "to 2 from 1",
			toAddr:    addr2,
			fromAddrs: accs(addr1),
			expected:  nil,
		},
		{
			name:      "to 2 from 2",
			toAddr:    addr2,
			fromAddrs: accs(addr2),
			expected:  nil,
		},
		{
			name:      "to 2 from 3",
			toAddr:    addr2,
			fromAddrs: accs(addr3),
			expected:  nil,
		},
		{
			name:      "to 3 from 0",
			toAddr:    addr3,
			fromAddrs: accs(addr0),
			expected:  nil,
		},
		{
			name:      "to 3 from 1",
			toAddr:    addr3,
			fromAddrs: accs(addr1),
			expected:  nil,
		},
		{
			name:      "to 3 from 2",
			toAddr:    addr3,
			fromAddrs: accs(addr2),
			expected:  nil,
		},
		{
			name:      "to 3 from 3",
			toAddr:    addr3,
			fromAddrs: accs(addr3),
			expected:  nil,
		},
		{
			name:      "to 3 from 0 1 2 3",
			toAddr:    addr3,
			fromAddrs: accs(addr0, addr1, addr2, addr3),
			expected:  nil,
		},
		{
			name:      "to 0 from 0 finds 1: d",
			toAddr:    addr0,
			fromAddrs: accs(addr0),
			expected:  qrs(recordD),
		},
		{
			name:      "to 0 from 1 finds 3: abd",
			toAddr:    addr0,
			fromAddrs: accs(addr1),
			expected:  qrs(recordA, recordB, recordD),
		},
		{
			name:      "to 0 from 2 finds 3: bcd",
			toAddr:    addr0,
			fromAddrs: accs(addr2),
			expected:  qrs(recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 3 finds 1: e",
			toAddr:    addr0,
			fromAddrs: accs(addr3),
			expected:  qrs(recordE),
		},
		{
			name:      "to 0 from 0 0 finds 1: d",
			toAddr:    addr0,
			fromAddrs: accs(addr0, addr0),
			expected:  qrs(recordD),
		},
		{
			name:      "to 0 from 0 1 finds 3: abd",
			toAddr:    addr0,
			fromAddrs: accs(addr0, addr1),
			expected:  qrs(recordA, recordB, recordD),
		},
		{
			name:      "to 0 from 0 2 finds 3: bcd",
			toAddr:    addr0,
			fromAddrs: accs(addr0, addr2),
			expected:  qrs(recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 0 3 finds 2: de",
			toAddr:    addr0,
			fromAddrs: accs(addr0, addr3),
			expected:  qrs(recordD, recordE),
		},
		{
			name:      "to 0 from 1 0 finds 3: abd",
			toAddr:    addr0,
			fromAddrs: accs(addr1, addr0),
			expected:  qrs(recordA, recordB, recordD),
		},
		{
			name:      "to 0 from 1 1 finds 3: abd",
			toAddr:    addr0,
			fromAddrs: accs(addr1, addr1),
			expected:  qrs(recordA, recordB, recordD),
		},
		{
			name:      "to 0 from 1 2 finds 4: abcd",
			toAddr:    addr0,
			fromAddrs: accs(addr1, addr2),
			expected:  qrs(recordA, recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 1 3 finds 4: abde",
			toAddr:    addr0,
			fromAddrs: accs(addr1, addr3),
			expected:  qrs(recordA, recordB, recordD, recordE),
		},
		{
			name:      "to 0 from 2 0 finds 3: bcd",
			toAddr:    addr0,
			fromAddrs: accs(addr2, addr0),
			expected:  qrs(recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 2 1 finds 4: abcd",
			toAddr:    addr0,
			fromAddrs: accs(addr2, addr1),
			expected:  qrs(recordA, recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 2 2 finds 3: bcd",
			toAddr:    addr0,
			fromAddrs: accs(addr2, addr2),
			expected:  qrs(recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 2 3 finds 4: bcde",
			toAddr:    addr0,
			fromAddrs: accs(addr2, addr3),
			expected:  qrs(recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 3 0 finds 2: de",
			toAddr:    addr0,
			fromAddrs: accs(addr3, addr0),
			expected:  qrs(recordD, recordE),
		},
		{
			name:      "to 0 from 3 1 finds 4: abde",
			toAddr:    addr0,
			fromAddrs: accs(addr3, addr1),
			expected:  qrs(recordA, recordB, recordD, recordE),
		},
		{
			name:      "to 0 from 3 2 finds 4: bcde",
			toAddr:    addr0,
			fromAddrs: accs(addr3, addr2),
			expected:  qrs(recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 3 3 finds 1: e",
			toAddr:    addr0,
			fromAddrs: accs(addr3, addr3),
			expected:  qrs(recordE),
		},
		{
			name:      "to 0 from 0 1 2 finds 4: abcd",
			toAddr:    addr0,
			fromAddrs: accs(addr0, addr1, addr2),
			expected:  qrs(recordA, recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 0 1 3 finds 4: abde",
			toAddr:    addr0,
			fromAddrs: accs(addr0, addr1, addr3),
			expected:  qrs(recordA, recordB, recordD, recordE),
		},
		{
			name:      "to 0 from 0 2 3 finds 4: bcde",
			toAddr:    addr0,
			fromAddrs: accs(addr0, addr2, addr3),
			expected:  qrs(recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 1 2 3 finds 5: abcde",
			toAddr:    addr0,
			fromAddrs: accs(addr1, addr2, addr3),
			expected:  qrs(recordA, recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 0 1 2 3 finds 5: abcde",
			toAddr:    addr0,
			fromAddrs: accs(addr0, addr1, addr2, addr3),
			expected:  qrs(recordA, recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 1 3 0 2 finds 5: abcde",
			toAddr:    addr0,
			fromAddrs: accs(addr1, addr3, addr0, addr2),
			expected:  qrs(recordA, recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 1 3 1 2 finds 5: abcde",
			toAddr:    addr0,
			fromAddrs: accs(addr1, addr3, addr1, addr2),
			expected:  qrs(recordA, recordB, recordC, recordD, recordE),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual := s.keeper.GetQuarantineRecords(s.sdkCtx, tc.toAddr, tc.fromAddrs...)
			s.Assert().ElementsMatch(tc.expected, actual, "GetQuarantineRecords A = expected vs B = actual")
		})
	}
}

func (s *TestSuite) TestAddQuarantinedCoins() {
	// Getting a little tricky here because I want different addresses for each test.
	// The addrBase is used to generate addrCount addresses.
	// Then, the autoAccept, autoDecline, toAddr and fromAddrs are address indexes to use.
	// The tricky part is that both the existing and expected Quarantine Records will have their
	// AccAddress slices updated before doing anything. For any AccAddress in them that's 1 byte long, and that byte
	// is less than addrCount, it's used as an index and the entry is updated to be that address.
	tests := []struct {
		name        string
		addrBase    string
		addrCount   uint8
		existing    *quarantine.QuarantineRecord
		autoAccept  []int
		autoDecline []int
		coins       sdk.Coins
		toAddr      int
		fromAddrs   []int
		expected    *quarantine.QuarantineRecord
		expErrFmt   string
		expErrAddrs []int
	}{
		{
			name:      "new record is created",
			addrBase:  "nr",
			addrCount: 2,
			coins:     s.cz("99bananas"),
			toAddr:    0,
			fromAddrs: []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("99bananas"),
			},
		},
		{
			name:      "new record 2 froms is created",
			addrBase:  "nr2f",
			addrCount: 3,
			coins:     s.cz("88crazy"),
			toAddr:    0,
			fromAddrs: []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("88crazy"),
			},
		},
		{
			name:      "existing record same denom is updated",
			addrBase:  "ersd",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("11pants"),
			},
			coins:     s.cz("200pants"),
			toAddr:    0,
			fromAddrs: []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("211pants"),
			},
		},
		{
			name:      "existing record new denom is updated",
			addrBase:  "ernd",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("102tower"),
			},
			coins:     s.cz("5pit"),
			toAddr:    0,
			fromAddrs: []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("5pit,102tower"),
			},
		},
		{
			name:      "existing record 2 froms is updated",
			addrBase:  "er2f",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   s.cz("53pcoin"),
			},
			coins:     s.cz("9000pcoin"),
			toAddr:    0,
			fromAddrs: []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   s.cz("9053pcoin"),
			},
		},
		{
			name:      "existing record 2 froms other order is updated",
			addrBase:  "er2foo",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   s.cz("35pcoin"),
			},
			coins:     s.cz("800pcoin"),
			toAddr:    0,
			fromAddrs: []int{2, 1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   s.cz("835pcoin"),
			},
		},
		{
			name:      "existing record unaccepted now auto-accept is still unaccepted",
			addrBase:  "eruna",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("543interstellar"),
			},
			autoAccept: []int{1},
			coins:      s.cz("5012interstellar"),
			toAddr:     0,
			fromAddrs:  []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("5555interstellar"), // One more time!
			},
		},
		{
			name:        "new record from is auto-accept nothing stored",
			addrBase:    "nrfa",
			addrCount:   2,
			autoAccept:  []int{1},
			coins:       s.cz("76trombones"),
			toAddr:      0,
			fromAddrs:   []int{1},
			expected:    nil,
			expErrFmt:   `cannot add quarantined funds "76trombones" to %s from %s: already fully accepted`,
			expErrAddrs: []int{0, 1},
		},
		{
			name:       "new record two froms first is auto-accept is marked as such",
			addrBase:   "nr2fa",
			addrCount:  3,
			autoAccept: []int{1},
			coins:      s.cz("52pinata"),
			toAddr:     0,
			fromAddrs:  []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{2}},
				AcceptedFromAddresses:   []sdk.AccAddress{{1}},
				Coins:                   s.cz("52pinata"),
			},
		},
		{
			name:       "new record two froms second is auto-accept is marked as such",
			addrBase:   "nr2sa",
			addrCount:  3,
			autoAccept: []int{2},
			coins:      s.cz("3fiddy"), // Loch Ness Monster, is that you?
			toAddr:     0,
			fromAddrs:  []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   s.cz("3fiddy"),
			},
		},
		{
			name:        "new record two froms both auto-accept nothing stored",
			addrBase:    "nr2ba",
			addrCount:   3,
			autoAccept:  []int{1, 2},
			coins:       s.cz("4moo"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected:    nil,
			expErrFmt:   `cannot add quarantined funds "4moo" to %s from %s, %s: already fully accepted`,
			expErrAddrs: []int{0, 1, 2},
		},
		{
			name:      "existing record not declined not auto-decline result is not declined",
			addrBase:  "erndna",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("8nodeca"),
				Declined:                false,
			},
			coins:     s.cz("50nodeca"),
			toAddr:    0,
			fromAddrs: []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("58nodeca"),
				Declined:                false,
			},
		},
		{
			name:      "existing record not declined is auto-decline result is declined",
			addrBase:  "erndad",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("20deca"),
				Declined:                false,
			},
			autoDecline: []int{1},
			coins:       s.cz("406deca"),
			toAddr:      0,
			fromAddrs:   []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("426deca"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined is auto-declined result is declined",
			addrBase:  "erdad",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("3000yarp"),
				Declined:                true,
			},
			autoDecline: []int{1},
			coins:       s.cz("3yarp"),
			toAddr:      0,
			fromAddrs:   []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("3003yarp"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined not auto-declined result is not declined",
			addrBase:  "erdna",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("14dalmatian"),
				Declined:                true,
			},
			autoDecline: nil,
			coins:       s.cz("87dalmatian"),
			toAddr:      0,
			fromAddrs:   []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   s.cz("101dalmatian"),
				Declined:                false,
			},
		},
		{
			name:      "existing record not declined 2 froms neither are auto-decline result is not declined",
			addrBase:  "ernd2fna",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("3bill"),
				Declined:                false,
			},
			autoDecline: nil,
			coins:       s.cz("4bill"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("7bill"),
				Declined:                false,
			},
		},
		{
			name:      "existing record not declined 2 froms first is auto-decline result is declined",
			addrBase:  "ernd2ffa",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("20123zela"),
				Declined:                false,
			},
			autoDecline: []int{1},
			coins:       s.cz("5000000000zela"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("5000020123zela"),
				Declined:                true,
			},
		},
		{
			name:      "existing record not declined 2 froms second is auto-decline result is declined",
			addrBase:  "ernd2fsd",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("456789vids"),
				Declined:                false,
			},
			autoDecline: []int{2},
			coins:       s.cz("123000000vids"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("123456789vids"),
				Declined:                true,
			},
		},
		{
			name:      "existing record not declined 2 froms both are auto-decline result is declined",
			addrBase:  "ernd2fba",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("5green"),
				Declined:                false,
			},
			autoDecline: []int{1, 2},
			coins:       s.cz("333333333333333green"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("333333333333338green"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined 2 froms neither are auto-decline result is not declined",
			addrBase:  "erd2fna",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("4frank"),
				Declined:                true,
			},
			autoDecline: nil,
			coins:       s.cz("3frank"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("7frank"),
				Declined:                false,
			},
		},
		{
			name:      "existing record declined 2 froms first is auto-decline result is declined",
			addrBase:  "erd2ffa",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("10zulu"),
				Declined:                true,
			},
			autoDecline: []int{1},
			coins:       s.cz("11zulu"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("21zulu"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined 2 froms second is auto-decline result is declined",
			addrBase:  "erd2fsd",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("11stars"),
				Declined:                true,
			},
			autoDecline: []int{2},
			coins:       s.cz("99stars"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("110stars"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined 2 froms both are auto-decline result is declined",
			addrBase:  "erd2fba",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("44blue"),
				Declined:                true,
			},
			autoDecline: []int{1, 2},
			coins:       s.cz("360blue"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   s.cz("404blue"),
				Declined:                true,
			},
		},
	}

	seenAddrBases := map[string]bool{}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Make sure the address base isn't used by an earlier test.
			s.Require().NotEqual(tc.addrBase, "no AddrBase defined")
			s.Require().False(seenAddrBases[tc.addrBase], "an earlier test already used the address base %q", tc.addrBase)
			seenAddrBases[tc.addrBase] = true
			s.Require().GreaterOrEqual(int(tc.addrCount), 1, "addrCount")

			// Set up all the address stuff.
			addrs := make([]sdk.AccAddress, tc.addrCount)
			for i := range addrs {
				addrs[i] = MakeTestAddr(tc.addrBase, uint8(i))
			}
			toAddr := addrs[tc.toAddr]
			fromAddrs := make([]sdk.AccAddress, len(tc.fromAddrs))
			for i, fi := range tc.fromAddrs {
				fromAddrs[i] = addrs[fi]
			}
			autoAccept := make([]sdk.AccAddress, len(tc.autoAccept))
			for i, ai := range tc.autoAccept {
				autoAccept[i] = addrs[ai]
			}
			autoDecline := make([]sdk.AccAddress, len(tc.autoDecline))
			for i, ai := range tc.autoDecline {
				autoDecline[i] = addrs[ai]
			}
			updateQR(addrs, tc.existing)
			updateQR(addrs, tc.expected)

			expErr := ""
			if len(tc.expErrFmt) > 0 {
				fmtArgs := make([]any, len(tc.expErrAddrs))
				for i, addrsI := range tc.expErrAddrs {
					fmtArgs[i] = addrs[addrsI]
				}
				expErr = fmt.Sprintf(tc.expErrFmt, fmtArgs...)
			}

			// Set the existing value
			if tc.existing != nil {
				testFuncSet := func() {
					s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, tc.existing)
				}
				s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
			}

			// Set up auto-accept and auto-decline
			testFuncAuto := func(fromAddr sdk.AccAddress, response quarantine.AutoResponse) func() {
				return func() {
					s.keeper.SetAutoResponse(s.sdkCtx, toAddr, fromAddr, response)
				}
			}
			for i, fromAddr := range autoAccept {
				s.Require().NotPanics(testFuncAuto(fromAddr, quarantine.AUTO_RESPONSE_ACCEPT), "SetAutoResponse %d accept", i+1)
			}
			for i, fromAddr := range autoDecline {
				s.Require().NotPanics(testFuncAuto(fromAddr, quarantine.AUTO_RESPONSE_DECLINE), "SetAutoResponse %d decline", i+1)
			}

			expectedEvents := sdk.Events{}
			if len(expErr) == 0 {
				// Create events expected to be emitted by AddQuarantinedCoins.
				event, err := sdk.TypedEventToEvent(&quarantine.EventFundsQuarantined{
					ToAddress: toAddr.String(),
					Coins:     tc.coins,
				})
				s.Require().NoError(err, "TypedEventToEvent EventFundsQuarantined")
				expectedEvents = append(expectedEvents, event)
			}

			// Get a context with a fresh event manager and call AddQuarantinedCoins.
			// Make sure it doesn't panic and make sure it doesn't return an error.
			// Note: As of writing, the only error it could return is from emitting the events,
			// and who knows how to actually trigger/test that.
			ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
			var err error
			testFuncAdd := func() {
				err = s.keeper.AddQuarantinedCoins(ctx, tc.coins, toAddr, fromAddrs...)
			}
			s.Require().NotPanics(testFuncAdd, "AddQuarantinedCoins")
			if len(expErr) > 0 {
				s.Require().EqualError(err, expErr, "AddQuarantinedCoins")
			} else {
				s.Require().NoError(err, "AddQuarantinedCoins")
			}
			actualEvents := ctx.EventManager().Events()
			s.Assert().Equal(expectedEvents, actualEvents)

			// Now look up the record and make sure it's as expected.
			var actual *quarantine.QuarantineRecord
			testFuncGet := func() {
				actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, fromAddrs...)
			}
			s.Require().NotPanics(testFuncGet, "GetQuarantineRecord")
			s.Assert().Equal(tc.expected, actual, "resulting quarantine record")
		})
	}
}

func (s *TestSuite) TestAcceptQuarantinedFunds() {
	// makeEvent creates a funds-released event.
	makeEvent := func(t *testing.T, addr sdk.AccAddress, amt sdk.Coins) sdk.Event {
		event, err := sdk.TypedEventToEvent(&quarantine.EventFundsReleased{
			ToAddress: addr.String(),
			Coins:     amt,
		})
		require.NoError(t, err, "TypedEventToEvent EventFundsReleased")
		return event
	}

	// An event maker knows the coins, and takes in the address to output an
	// event with the (presently unknown) ToAddress and the (known) coins.
	type eventMaker func(t *testing.T, addr sdk.AccAddress) sdk.Event

	// makes the event maker functions, one for each string provided.
	makeEventMakers := func(coins ...string) []eventMaker {
		rv := make([]eventMaker, len(coins))
		for i, amtStr := range coins {
			// doing this now so that an invalid coin string fails the test before it gets started.
			// Really, I didn't want to have to update cz to also take in a *testing.T.
			amt := s.cz(amtStr)
			rv[i] = func(t *testing.T, addr sdk.AccAddress) sdk.Event {
				return makeEvent(t, addr, amt)
			}
		}
		return rv
	}

	// Getting a little tricky here because I want different addresses for each test.
	// The addrBase is used to generate addrCount addresses.
	// Then, addrs[0] becomes the toAddr. The fromAddrs are indexes of the addrs to use.
	// The tricky part is that the existing and expected Quarantine Records will have their
	// AccAddresses updated before doing anything. For any AccAddress in them that's 1 byte long, and that byte
	// is less than addrCount, it's used as an index and the entry is updated to be that address.
	// Also, the provided []eventMaker is used to create all expected events receiving the toAddr.
	tests := []struct {
		name            string
		addrBase        string
		addrCount       uint8
		records         []*quarantine.QuarantineRecord
		autoDecline     []int
		fromAddrs       []int
		expectedRecords []*quarantine.QuarantineRecord
		expectedSent    []sdk.Coins
		expectedEvents  []eventMaker
	}{
		{
			name:            "one from zero records",
			addrBase:        "ofzr",
			addrCount:       2,
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    nil,
			expectedEvents:  nil,
		},
		{
			name:      "one from one record fully",
			addrBase:  "oforf",
			addrCount: 2,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("17lemon"),
				},
			},
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{s.cz("17lemon")},
			expectedEvents:  makeEventMakers("17lemon"),
		},
		{
			name:      "one from one record finally fully",
			addrBase:  "foforf",
			addrCount: 4,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}, {3}},
					Coins:                   s.cz("8878pillow"),
				},
			},
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{s.cz("8878pillow")},
			expectedEvents:  makeEventMakers("8878pillow"),
		},
		{
			name:      "one from one record fully previously declined",
			addrBase:  "oforfpd",
			addrCount: 2,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("5rings,4birds,3hens"),
					Declined:                true,
				},
			},
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{s.cz("5rings,4birds,3hens")},
			expectedEvents:  makeEventMakers("5rings,4birds,3hens"),
		},
		{
			name:      "one from one record not fully",
			addrBase:  "ofornf",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("1snow"),
					Declined:                false,
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("1snow"),
					Declined:                false,
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from one record not fully previously declined",
			addrBase:  "ofornfpd",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("55orchid"),
					Declined:                true,
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("55orchid"),
					Declined:                false,
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from one record remaining unaccepted is auto-decline",
			addrBase:  "oforruad",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("99redballoons"),
					Declined:                true,
				},
			},
			autoDecline: []int{2},
			fromAddrs:   []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("99redballoons"),
					Declined:                true,
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from one record accepted was auto-decline",
			addrBase:  "oforawad",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("7777frog"),
					Declined:                true,
				},
			},
			autoDecline: []int{1},
			fromAddrs:   []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("7777frog"),
					Declined:                false,
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from two records neither fully",
			addrBase:  "oftrnf",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("20533lamp"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   s.cz("45sun"),
					Declined:                true,
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("20533lamp"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("45sun"),
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from two records first fully",
			addrBase:  "oftrff",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = 0264500F71512C3B111D2D2EAA7322F018DA16B13CBB5D516BD4B51C4F1A94EC
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}},
					Coins:                   s.cz("43bulb"),
				},
				// key suffix = 47F604CA662719863E40CF215D4DE088C22B7FF217236D887A99AF63A8F124E9
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   s.cz("5005shade"),
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("5005shade"),
				},
			},
			expectedSent:   []sdk.Coins{s.cz("43bulb")},
			expectedEvents: makeEventMakers("43bulb"),
		},
		{
			name:      "one from two records second fully",
			addrBase:  "ofttrsf",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = EFC545E02C1785EEAAE9004385C6106E75AC42E8096556376097037A0C122E41
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   s.cz("346awning"),
				},
				// key suffix = F898B0EAF64B4D67BC2C285E541D381FA422D85B05C69D697C099B1968003955
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}},
					Coins:                   s.cz("9444sprout"),
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("346awning"),
				},
			},
			expectedSent:   []sdk.Coins{s.cz("9444sprout")},
			expectedEvents: makeEventMakers("9444sprout"),
		},
		{
			name:      "one from two records both fully",
			addrBase:  "oftrbf",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("4312stand"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}, {3}},
					Coins:                   s.cz("9867sit"),
				},
			},
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{s.cz("4312stand"), s.cz("9867sit")},
			expectedEvents:  makeEventMakers("4312stand", "9867sit"),
		},
		{
			name:            "two froms zero records",
			addrBase:        "tfzr",
			addrCount:       3,
			fromAddrs:       []int{1, 2},
			expectedRecords: nil,
			expectedSent:    nil,
			expectedEvents:  nil,
		},
		{
			name:      "two froms one record fully",
			addrBase:  "tforf",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("838hibiscus"),
				},
			},
			fromAddrs:       []int{1, 2},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{s.cz("838hibiscus")},
			expectedEvents:  makeEventMakers("838hibiscus"),
		},
		{
			name:      "two froms other order one record fully",
			addrBase:  "tfooorf",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("10downing"),
				},
			},
			fromAddrs:       []int{2, 1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{s.cz("10downing")},
			expectedEvents:  makeEventMakers("10downing"),
		},
		{
			name:      "two froms one record not fully",
			addrBase:  "tfornf",
			addrCount: 4,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}, {3}},
					Coins:                   s.cz("1060waddison"),
				},
			},
			fromAddrs: []int{1, 2},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("1060waddison"),
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "two froms other order one record not fully",
			addrBase:  "tfooornf",
			addrCount: 4,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}, {3}},
					Coins:                   s.cz("1060waddison"),
				},
			},
			fromAddrs: []int{2, 1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("1060waddison"),
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "two froms two records neither fully",
			addrBase:  "tftrnf",
			addrCount: 5,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = 70705D4547681D550CF0D2A5B0996B6C2B42E181FF3F84A71CF6DAD8527C8C9C
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}, {4}},
					Coins:                   s.cz("12drummers"),
				},
				// key suffix = 83A580037E196C7BB4B36FDB5531BA715DF24F86681A61FE7D72D77BE2ABA4E8
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}, {3}},
					Coins:                   s.cz("11pipers"),
				},
			},
			fromAddrs: []int{1, 2},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{4}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("12drummers"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("11pipers"),
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "two froms two records first fully",
			addrBase:  "tftrff",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = 72536EA1F5EB0C1FF2897309892EF28553E7A6C2508AB1751D363B8C3A31A56F
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   s.cz("8maids,7swans"),
				},
				// key suffix = BDA18A04E7AC80DDA290C262CBEF7C2928B95F9DBFE8F392BA82EC0186DBA0CC
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("10lords,9ladies"),
				},
			},
			fromAddrs: []int{1, 3},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("10lords,9ladies"),
				},
			},
			expectedSent:   []sdk.Coins{s.cz("8maids,7swans")},
			expectedEvents: makeEventMakers("8maids,7swans"),
		},
		{
			name:      "two froms two records second fully",
			addrBase:  "tftrsf",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = 00E641E0BF6DF9F97E61B94BBBA58B78F74198BB72681C9A24C12D2BF1DDC371
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("6geese"),
				},
				// key suffix = D052411A78E6208D482F600692C7382C814C35FB75B49430E5CF895B4FE5EEFF
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   s.cz("2doves,1peartree"),
				},
			},
			fromAddrs: []int{1, 3},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("6geese"),
				},
			},
			expectedSent:   []sdk.Coins{s.cz("2doves,1peartree")},
			expectedEvents: makeEventMakers("2doves,1peartree"),
		},
		{
			name:      "two froms two records both fully",
			addrBase:  "tftrbf",
			addrCount: 3,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("3amigos"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					Coins:                   s.cz("8amigos"),
				},
			},
			fromAddrs:       []int{1, 2},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{s.cz("3amigos"), s.cz("8amigos")},
			expectedEvents:  makeEventMakers("3amigos", "8amigos"),
		},
	}

	seenAddrBases := map[string]bool{}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Make sure the address base isn't used by an earlier test.
			s.Require().NotEqual(tc.addrBase, "", "no AddrBase defined")
			s.Require().False(seenAddrBases[tc.addrBase], "an earlier test already used the address base %q", tc.addrBase)
			seenAddrBases[tc.addrBase] = true
			s.Require().GreaterOrEqual(int(tc.addrCount), 1, "addrCount")

			// Set up all the address stuff.
			addrs := make([]sdk.AccAddress, tc.addrCount)
			for i := range addrs {
				addrs[i] = MakeTestAddr(tc.addrBase, uint8(i))
			}

			toAddr := addrs[0]
			fromAddrs := make([]sdk.AccAddress, len(tc.fromAddrs))
			for i, fi := range tc.fromAddrs {
				fromAddrs[i] = addrs[fi]
			}

			autoDecline := make([]sdk.AccAddress, len(tc.autoDecline))
			for i, addr := range tc.autoDecline {
				autoDecline[i] = addrs[addr]
			}

			for _, record := range tc.records {
				updateQR(addrs, record)
			}

			for _, record := range tc.expectedRecords {
				updateQR(addrs, record)
			}

			var expectedSends []*SentCoins
			if len(tc.expectedSent) > 0 {
				expectedSends = make([]*SentCoins, len(tc.expectedSent))
				for i, sent := range tc.expectedSent {
					expectedSends[i] = &SentCoins{
						FromAddr: s.keeper.GetFundsHolder(),
						ToAddr:   toAddr,
						Amt:      sent,
					}
				}
			}

			expectedEvents := make(sdk.Events, len(tc.expectedEvents))
			for i, ev := range tc.expectedEvents {
				expectedEvents[i] = ev(s.T(), toAddr)
			}

			// Now that we have all the expected stuff defined, let's get things set up.

			// mock the bank keeper and use that, so we don't have to fund stuff,
			// and we get a record of the sends made.
			bKeeper := NewMockBankKeeper() // bzzzzzzzzzz
			qKeeper := s.keeper.WithBankKeeper(bKeeper)

			// Set the existing records
			for i, existing := range tc.records {
				if existing != nil {
					testFuncSet := func() {
						qKeeper.SetQuarantineRecord(s.sdkCtx, toAddr, existing)
					}
					s.Require().NotPanics(testFuncSet, "SetQuarantineRecord[%d]", i)
					recordKey := quarantine.CreateRecordKey(toAddr, existing.GetAllFromAddrs()...)
					_, suffix := quarantine.ParseRecordIndexKey(recordKey)
					s.T().Logf("existing[%d] suffix: %v", i, suffix)
				}
			}

			// Set existing auto-declines
			for i, addr := range autoDecline {
				testFuncAuto := func() {
					qKeeper.SetAutoResponse(s.sdkCtx, toAddr, addr, quarantine.AUTO_RESPONSE_DECLINE)
				}
				s.Require().NotPanics(testFuncAuto, "SetAutoResponse[%d]", i)
			}

			expectedFundsReleased := sdk.Coins{}
			for _, coins := range tc.expectedSent {
				expectedFundsReleased = expectedFundsReleased.Add(coins...)
			}

			// Setup done. Let's do this.
			var err error
			var fundsReleased sdk.Coins
			ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
			testFuncAccept := func() {
				fundsReleased, err = qKeeper.AcceptQuarantinedFunds(ctx, toAddr, fromAddrs...)
			}
			s.Require().NotPanics(testFuncAccept, "AcceptQuarantinedFunds")
			s.Require().NoError(err, "AcceptQuarantinedFunds")
			s.Assert().Equal(expectedFundsReleased, fundsReleased, "AcceptQuarantinedFunds funds released")

			// And check the expected.
			var actualRecords []*quarantine.QuarantineRecord
			testFuncGet := func() {
				actualRecords = qKeeper.GetQuarantineRecords(s.sdkCtx, toAddr, fromAddrs...)
			}
			if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecords") {
				s.Assert().Equal(tc.expectedRecords, actualRecords, "resulting QuarantineRecords")
			}

			actualSends := bKeeper.SentCoins
			s.Assert().Equal(expectedSends, actualSends, "sends made")

			actualEvents := ctx.EventManager().Events()
			s.Assert().Equal(expectedEvents, actualEvents, "events emitted during accept")
		})
	}

	s.Run("send returns an error", func() {
		// Setup: There will be 4 records to send, the 3rd will return an error.
		// Check that:
		// 1. The error is returned by AcceptQuarantinedFunds
		// 2. The 1st and 2nd records are removed but the 3rd and 4th remain.
		// 3. SendCoins was called for the 1st and 2nd records.
		// 4. Events were emitted for the 1st and 2nd records.

		// Setup address stuff.
		addrBase := "sre"
		s.Require().False(seenAddrBases[addrBase], "an earlier test already used the address base %q", addrBase)
		seenAddrBases[addrBase] = true

		toAddr := MakeTestAddr(addrBase, 0)
		fromAddr1 := MakeTestAddr(addrBase, 1)
		fromAddr2 := MakeTestAddr(addrBase, 2)
		fromAddr3 := MakeTestAddr(addrBase, 3)
		fromAddr4 := MakeTestAddr(addrBase, 4)
		fromAddrs := accs(fromAddr1, fromAddr2, fromAddr3, fromAddr4)

		// Define the existing records and expected stuff.
		existingRecords := []*quarantine.QuarantineRecord{
			{
				UnacceptedFromAddresses: accs(fromAddr1),
				Coins:                   s.cz("1addra"),
			},
			{
				UnacceptedFromAddresses: accs(fromAddr2),
				Coins:                   s.cz("2addrb"),
			},
			{
				UnacceptedFromAddresses: accs(fromAddr3),
				Coins:                   s.cz("3addrc"),
			},
			{
				UnacceptedFromAddresses: accs(fromAddr4),
				Coins:                   s.cz("4addrd"),
			},
		}

		expectedErr := "this is a test error"

		expectedRecords := []*quarantine.QuarantineRecord{
			{
				UnacceptedFromAddresses: accs(fromAddr3),
				Coins:                   s.cz("3addrc"),
			},
			{
				UnacceptedFromAddresses: accs(fromAddr4),
				Coins:                   s.cz("4addrd"),
			},
		}

		expectedSends := []*SentCoins{
			{
				FromAddr: s.keeper.GetFundsHolder(),
				ToAddr:   toAddr,
				Amt:      s.cz("1addra"),
			},
			{
				FromAddr: s.keeper.GetFundsHolder(),
				ToAddr:   toAddr,
				Amt:      s.cz("2addrb"),
			},
		}

		// Since an error is being returned, funds released should be nil.
		expectedFundsReleased := sdk.Coins(nil)

		expectedEvents := sdk.Events{
			makeEvent(s.T(), toAddr, s.cz("1addra")),
			makeEvent(s.T(), toAddr, s.cz("2addrb")),
		}

		// mock the bank keeper and set it to return an error on the 3rd send.
		bKeeper := NewMockBankKeeper() // bzzzzzzzzzz
		bKeeper.QueuedSendCoinsErrors = []error{
			nil,
			nil,
			fmt.Errorf(expectedErr),
		}
		qKeeper := s.keeper.WithBankKeeper(bKeeper)

		// Store the existing records.
		for i, record := range existingRecords {
			testFuncSet := func() {
				qKeeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
			}
			s.Require().NotPanics(testFuncSet, "SetQuarantineRecord[%d]", i)
		}

		// Do the thing.
		var actualErr error
		var fundsReleased sdk.Coins
		ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
		testFuncAccept := func() {
			fundsReleased, actualErr = qKeeper.AcceptQuarantinedFunds(ctx, toAddr, fromAddrs...)
		}
		s.Require().NotPanics(testFuncAccept, "AcceptQuarantinedFunds")

		// Check that: 1. The error is returned by AcceptQuarantinedFunds
		s.Assert().EqualError(actualErr, expectedErr, "AcceptQuarantinedFunds error")

		// Check that: 2. The expected funds released was returned.
		s.Assert().Equal(expectedFundsReleased, fundsReleased, "AcceptQuarantinedFunds funds released")

		// Check that: 3. The 1st and 2nd records are removed but the 3rd and 4th remain.
		var actualRecords []*quarantine.QuarantineRecord
		testFuncGet := func() {
			actualRecords = qKeeper.GetQuarantineRecords(ctx, toAddr, fromAddrs...)
		}
		if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecords") {
			s.Assert().Equal(expectedRecords, actualRecords)
		}

		// Check that: 4. SendCoins was called for the 1st and 2nd records.
		actualSends := bKeeper.SentCoins
		s.Assert().Equal(expectedSends, actualSends, "sends made")

		// Check that: 5. Events were emitted for the 1st and 2nd records.
		actualEvents := ctx.EventManager().Events()
		s.Assert().Equal(expectedEvents, actualEvents, "events emitted")
	})
}

func (s *TestSuite) TestDeclineQuarantinedFunds() {
	tests := []struct {
		name      string
		addrBase  string
		addrCount uint8
		fromAddrs []int
		existing  []*quarantine.QuarantineRecord
		expected  []*quarantine.QuarantineRecord
	}{
		{
			name:      "one from zero records",
			addrBase:  "ofzr",
			addrCount: 2,
			fromAddrs: []int{1},
			existing:  nil,
			expected:  nil,
		},
		{
			name:      "one from one record",
			addrBase:  "ofor",
			addrCount: 2,
			fromAddrs: []int{1},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("13ofor"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("13ofor"),
					Declined:                true,
				},
			},
		},
		{
			name:      "one from one record previously accepted",
			addrBase:  "oforpa",
			addrCount: 3,
			fromAddrs: []int{1},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("8139oforpa"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}, {1}},
					Coins:                   s.cz("8139oforpa"),
					Declined:                true,
				},
			},
		},
		{
			name:      "one from two records",
			addrBase:  "oftr",
			addrCount: 4,
			fromAddrs: []int{1},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("85oftr"),
					Declined:                false,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   s.cz("190oftr"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("85oftr"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   s.cz("190oftr"),
					Declined:                true,
				},
			},
		},
		{
			name:      "two froms zero records",
			addrBase:  "tfzr",
			addrCount: 3,
			fromAddrs: []int{1, 2},
			existing:  nil,
			expected:  nil,
		},
		{
			name:      "two froms one record from first",
			addrBase:  "tforff",
			addrCount: 4,
			fromAddrs: []int{1, 2},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   s.cz("321tforff"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}, {1}},
					Coins:                   s.cz("321tforff"),
					Declined:                true,
				},
			},
		},
		{
			name:      "two froms one record from second",
			addrBase:  "tforfs",
			addrCount: 4,
			fromAddrs: []int{1, 2},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}},
					Coins:                   s.cz("321tforfs"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}, {2}},
					Coins:                   s.cz("321tforfs"),
					Declined:                true,
				},
			},
		},
		{
			name:      "two froms one record from both",
			addrBase:  "tforfb",
			addrCount: 4,
			fromAddrs: []int{1, 2},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("4tforfb"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}, {1}, {2}},
					AcceptedFromAddresses:   nil,
					Coins:                   s.cz("4tforfb"),
					Declined:                true,
				},
			},
		},
		{
			name:      "two froms two records from first",
			addrBase:  "tftrff",
			addrCount: 5,
			fromAddrs: []int{1, 2},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{3}},
					Coins:                   s.cz("13tftrff"),
					Declined:                false,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {4}},
					Coins:                   s.cz("14tftrff"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{3}},
					Coins:                   s.cz("13tftrff"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {4}},
					Coins:                   s.cz("14tftrff"),
					Declined:                true,
				},
			},
		},
		{
			name:      "two froms two records from second",
			addrBase:  "tftrfs",
			addrCount: 5,
			fromAddrs: []int{1, 2},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{3}},
					Coins:                   s.cz("13tftrfs"),
					Declined:                false,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}, {4}},
					Coins:                   s.cz("14tftrfs"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{3}},
					Coins:                   s.cz("13tftrfs"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}, {4}},
					Coins:                   s.cz("14tftrfs"),
					Declined:                true,
				},
			},
		},
		{
			name:      "two froms two records one from each",
			addrBase:  "tftrofe",
			addrCount: 3,
			fromAddrs: []int{1, 2},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("1tftrofe"),
					Declined:                false,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					Coins:                   s.cz("2tftrofe"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("1tftrofe"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					Coins:                   s.cz("2tftrofe"),
					Declined:                true,
				},
			},
		},
		{
			name:      "two froms two records one from one other from both",
			addrBase:  "tftrofb",
			addrCount: 3,
			fromAddrs: []int{1, 2},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("1tftrofb"),
					Declined:                false,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("12tftrofb"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("1tftrofb"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("12tftrofb"),
					Declined:                true,
				},
			},
		},
		{
			name: "two froms five records",
			// (1st, 2nd, 1st & 2nd, 1st & other, 2nd & other)
			addrBase:  "tffr",
			addrCount: 4,
			fromAddrs: []int{1, 2},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("1tffr"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					Coins:                   s.cz("2tffr"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}},
					Coins:                   s.cz("12tffr"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{3}},
					Coins:                   s.cz("13tffr"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}, {3}},
					Coins:                   s.cz("23tffr"),
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   s.cz("1tffr"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					Coins:                   s.cz("2tffr"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   s.cz("12tffr"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{3}},
					Coins:                   s.cz("13tffr"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}, {3}},
					Coins:                   s.cz("23tffr"),
					Declined:                true,
				},
			},
		},
	}

	seenAddrBases := map[string]bool{}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Make sure the address base isn't used by an earlier test.
			s.Require().NotEqual(tc.addrBase, "", "no AddrBase defined")
			s.Require().False(seenAddrBases[tc.addrBase], "an earlier test already used the address base %q", tc.addrBase)
			seenAddrBases[tc.addrBase] = true
			s.Require().GreaterOrEqual(int(tc.addrCount), 1, "addrCount")

			// Set up all the address stuff.
			addrs := make([]sdk.AccAddress, tc.addrCount)
			for i := range addrs {
				addrs[i] = MakeTestAddr(tc.addrBase, uint8(i))
			}

			toAddr := addrs[0]
			fromAddrs := make([]sdk.AccAddress, len(tc.fromAddrs))
			for i, fi := range tc.fromAddrs {
				fromAddrs[i] = addrs[fi]
			}

			for _, record := range tc.existing {
				updateQR(addrs, record)
			}
			for _, record := range tc.expected {
				updateQR(addrs, record)
			}

			// Set the existing records.
			for i, record := range tc.existing {
				testFuncSet := func() {
					s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
				}
				s.Require().NotPanics(testFuncSet, "SetQuarantineRecord[%d]", i)
				recordKey := quarantine.CreateRecordKey(toAddr, record.GetAllFromAddrs()...)
				_, suffix := quarantine.ParseRecordKey(recordKey)
				s.T().Logf("record[%d] suffix: %v", i, suffix)
			}

			// Do the thing.
			testFuncDecline := func() {
				s.keeper.DeclineQuarantinedFunds(s.sdkCtx, toAddr, fromAddrs...)
			}
			s.Require().NotPanics(testFuncDecline, "DeclineQuarantinedFunds")

			var actual []*quarantine.QuarantineRecord
			testFuncGet := func() {
				actual = s.keeper.GetQuarantineRecords(s.sdkCtx, toAddr, fromAddrs...)
			}
			if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecords") {
				s.Assert().ElementsMatch(tc.expected, actual, "resulting quarantine records")
			}
		})
	}
}

func (s *TestSuite) TestQuarantineRecordsIterateAndGetAll() {
	addrBase := "qriga"
	addr0 := MakeTestAddr(addrBase, 0)
	addr1 := MakeTestAddr(addrBase, 1)
	addr2 := MakeTestAddr(addrBase, 2)
	addr3 := MakeTestAddr(addrBase, 3)
	addr4 := MakeTestAddr(addrBase, 4)
	addr5 := MakeTestAddr(addrBase, 5)
	addr6 := MakeTestAddr(addrBase, 6)
	addr7 := MakeTestAddr(addrBase, 7)

	// Create 7 records
	initialRecords := []*struct {
		to     sdk.AccAddress
		record *quarantine.QuarantineRecord
	}{
		{
			to: addr0,
			record: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr1),
				Coins:                   s.cz("1boom"),
			},
		},
		{
			to: addr0,
			record: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr2),
				Coins:                   s.cz("5boom"),
			},
		},
		{
			to: addr3,
			record: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr1),
				Coins:                   s.cz("23boom"),
				Declined:                true,
			},
		},
		{
			to: addr5,
			record: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr6),
				AcceptedFromAddresses:   accs(addr7),
				Coins:                   s.cz("79boom"),
			},
		},
		{
			to: addr0,
			record: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr3),
				Coins:                   s.cz("163boom"),
			},
		},
		{
			to: addr3,
			record: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr4),
				Coins:                   s.cz("331boom"),
			},
		},
		{
			to: addr0,
			record: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: accs(addr7),
				Coins:                   s.cz("673boom"),
			},
		},
	}

	for i, rec := range initialRecords {
		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, rec.to, rec.record)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord[%d]", i)
	}

	// Remove the 2nd one by setting it as fully accepted.
	secondTo := MakeCopyOfAccAddress(initialRecords[1].to)
	secondRec := MakeCopyOfQuarantineRecord(initialRecords[1].record)
	secondRec.AcceptFrom(secondRec.UnacceptedFromAddresses)
	testFuncUnset := func() {
		s.keeper.SetQuarantineRecord(s.sdkCtx, secondTo, secondRec)
	}
	s.Require().NotPanics(testFuncUnset, "SetQuarantineRecord to remove second")

	// Final setup:
	// They should be ordered by key and the keys should start with their indexes.
	// [0] 0 <- 1 1boom
	// [4] 0 <- 3 163boom
	// [6] 0 <- 7 673boom
	// [2] 3 <- 1 23boom (declined)
	// [5] 3 <- 4 331boom
	// [3] 5 <- 6,7 79boom (7 accepted)

	allOrder := []int{0, 4, 6, 2, 5, 3}

	tests := []struct {
		name          string
		toAddr        sdk.AccAddress
		expectedOrder []int
	}{
		{
			name:          "IterateQuarantineRecords all",
			toAddr:        nil,
			expectedOrder: allOrder,
		},
		{
			name:          "IterateQuarantineRecords addr0",
			toAddr:        addr0,
			expectedOrder: []int{0, 4, 6},
		},
		{
			name:          "IterateQuarantineRecords addr1",
			toAddr:        addr1,
			expectedOrder: []int{},
		},
		{
			name:          "IterateQuarantineRecords addr2",
			toAddr:        addr2,
			expectedOrder: []int{},
		},
		{
			name:          "IterateQuarantineRecords addr3",
			toAddr:        addr3,
			expectedOrder: []int{2, 5},
		},
		{
			name:          "IterateQuarantineRecords addr4",
			toAddr:        addr4,
			expectedOrder: []int{},
		},
		{
			name:          "IterateQuarantineRecords addr5",
			toAddr:        addr5,
			expectedOrder: []int{3},
		},
		{
			name:          "IterateQuarantineRecords addr6",
			toAddr:        addr6,
			expectedOrder: []int{},
		},
		{
			name:          "IterateQuarantineRecords addr7",
			toAddr:        addr7,
			expectedOrder: []int{},
		},
	}

	// cbArgs are the arguments provided to the callback of IterateQuarantineRecords
	type cbArgs struct {
		toAddr sdk.AccAddress
		suffix sdk.AccAddress
		record *quarantine.QuarantineRecord
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			expected := make([]*cbArgs, len(tc.expectedOrder))
			for i, iri := range tc.expectedOrder {
				tr := initialRecords[iri]
				key := quarantine.CreateRecordKey(tr.to, tr.record.GetAllFromAddrs()...)
				_, suffix := quarantine.ParseRecordKey(key)
				expected[i] = &cbArgs{
					toAddr: MakeCopyOfAccAddress(tr.to),
					suffix: MakeCopyOfAccAddress(suffix),
					record: MakeCopyOfQuarantineRecord(tr.record),
				}
			}

			actual := make([]*cbArgs, 0, len(expected))
			callback := func(toAddr, recordSuffix sdk.AccAddress, record *quarantine.QuarantineRecord) bool {
				actual = append(actual, &cbArgs{
					toAddr: toAddr,
					suffix: recordSuffix,
					record: record,
				})
				return false
			}
			testFunc := func() {
				s.keeper.IterateQuarantineRecords(s.sdkCtx, tc.toAddr, callback)
			}
			s.Require().NotPanics(testFunc, "IterateQuarantineRecords")
			s.Assert().Equal(expected, actual, "callback args provided to IterateQuarantineRecords")
		})
	}

	s.Run("IterateQuarantineRecords stop early", func() {
		stopLen := 4
		expected := make([]*cbArgs, stopLen)
		for i, iri := range allOrder[:stopLen] {
			tr := initialRecords[iri]
			key := quarantine.CreateRecordKey(tr.to, tr.record.GetAllFromAddrs()...)
			_, suffix := quarantine.ParseRecordKey(key)
			expected[i] = &cbArgs{
				toAddr: MakeCopyOfAccAddress(tr.to),
				suffix: MakeCopyOfAccAddress(suffix),
				record: MakeCopyOfQuarantineRecord(tr.record),
			}
		}

		actual := make([]*cbArgs, 0, len(expected))
		callback := func(toAddr, recordSuffix sdk.AccAddress, record *quarantine.QuarantineRecord) bool {
			actual = append(actual, &cbArgs{
				toAddr: toAddr,
				suffix: recordSuffix,
				record: record,
			})
			return len(actual) >= stopLen
		}
		testFunc := func() {
			s.keeper.IterateQuarantineRecords(s.sdkCtx, nil, callback)
		}
		s.Require().NotPanics(testFunc, "IterateQuarantineRecords")
		s.Assert().Equal(expected, actual, "callback args provided to IterateQuarantineRecords")
	})

	s.Run("GetAllQuarantinedFunds", func() {
		expected := make([]*quarantine.QuarantinedFunds, len(allOrder))
		for i, iri := range allOrder {
			tr := initialRecords[iri]
			expected[i] = tr.record.AsQuarantinedFunds(tr.to)
		}
		var actual []*quarantine.QuarantinedFunds
		testFuncGetAll := func() {
			actual = s.keeper.GetAllQuarantinedFunds(s.sdkCtx)
		}
		s.Require().NotPanics(testFuncGetAll, "GetAllQuarantinedFunds")
		s.Assert().Equal(expected, actual, "GetAllQuarantinedFunds results")
	})
}

func (s *TestSuite) TestBzToQuarantineRecordSuffixIndex() {
	cdc := s.keeper.GetCodec()

	tests := []struct {
		name     string
		bz       []byte
		expected *quarantine.QuarantineRecordSuffixIndex
		expErr   string
	}{
		{
			name: "control",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecordSuffixIndex{
				RecordSuffixes: [][]byte{{1, 2, 3}, {1, 5, 5}},
			}),
			expected: &quarantine.QuarantineRecordSuffixIndex{
				RecordSuffixes: [][]byte{{1, 2, 3}, {1, 5, 5}},
			},
		},
		{
			name: "nil bz",
			bz:   nil,
			expected: &quarantine.QuarantineRecordSuffixIndex{
				RecordSuffixes: nil,
			},
		},
		{
			name: "empty bz",
			bz:   []byte{},
			expected: &quarantine.QuarantineRecordSuffixIndex{
				RecordSuffixes: nil,
			},
		},
		{
			name:   "unknown bytes",
			bz:     []byte{0x75, 110, 0153, 0x6e, 0157, 119, 0156, 0xff, 0142, 0x79, 116, 0x65, 0163},
			expErr: "proto: illegal wireType 7",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual *quarantine.QuarantineRecordSuffixIndex
			var err error
			testFunc := func() {
				actual, err = s.keeper.BzToQuarantineRecordSuffixIndex(tc.bz)
			}
			s.Require().NotPanics(testFunc, "bzToQuarantineRecordSuffixIndex")
			if len(tc.expErr) > 0 {
				s.Require().EqualError(err, tc.expErr, "bzToQuarantineRecordSuffixIndex error: %v", actual)
			} else {
				s.Require().NoError(err, "bzToQuarantineRecordSuffixIndex error")
				s.Assert().Equal(tc.expected, actual, "bzToQuarantineRecordSuffixIndex index")
			}
		})

		s.Run("must "+tc.name, func() {
			var actual *quarantine.QuarantineRecordSuffixIndex
			testFunc := func() {
				actual = s.keeper.MustBzToQuarantineRecordSuffixIndex(tc.bz)
			}
			if len(tc.expErr) > 0 {
				s.Require().PanicsWithError(tc.expErr, testFunc, "mustBzToQuarantineRecordSuffixIndex: %v", actual)
			} else {
				s.Require().NotPanics(testFunc, "mustBzToQuarantineRecordSuffixIndex")
				s.Assert().Equal(tc.expected, actual, "bzToQuarantineRecordSuffixIndex index")
			}
		})
	}
}

func (s *TestSuite) TestSuffixIndexGetSet() {
	toAddr := MakeTestAddr("sigs", 0)
	fromAddr := MakeTestAddr("sigs", 1)
	suffix0 := []byte(MakeTestAddr("sfxsigs", 0))
	suffix1 := []byte(MakeTestAddr("sfxsigs", 1))
	suffix2 := []byte(MakeTestAddr("sfxsigs", 2))
	suffix3 := []byte(MakeTestAddr("sfxsigs", 3))
	suffix4 := []byte(MakeTestAddr("sfxsigs", 4))
	suffix5 := []byte(MakeTestAddr("sfxsigs", 5))

	s.Run("1 getQuarantineRecordSuffix on unset entry", func() {
		store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
		expectedIndex := &quarantine.QuarantineRecordSuffixIndex{RecordSuffixes: nil}
		expectedKey := quarantine.CreateRecordIndexKey(toAddr, fromAddr)

		var actualIndex *quarantine.QuarantineRecordSuffixIndex
		var actualKey []byte
		testFunc := func() {
			actualIndex, actualKey = s.keeper.GetQuarantineRecordSuffixIndex(store, toAddr, fromAddr)
		}
		s.Require().NotPanics(testFunc, "getQuarantineRecordSuffixIndex")
		s.Assert().Equal(expectedIndex, actualIndex, "returned index")
		s.Assert().Equal(expectedKey, actualKey, "returned key")
	})

	s.Run("2 setQuarantineRecordSuffixIndex on unset entry", func() {
		store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
		index := &quarantine.QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffix0, suffix1}}
		key := quarantine.CreateRecordIndexKey(toAddr, fromAddr)

		testFunc := func() {
			s.keeper.SetQuarantineRecordSuffixIndex(store, key, index)
		}
		s.Require().NotPanics(testFunc, "setQuarantineRecordSuffixIndex")
	})

	s.Run("3 getQuarantineRecordSuffix on previously set entry", func() {
		store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
		expectedIndex := &quarantine.QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffix0, suffix1}}
		expectedKey := quarantine.CreateRecordIndexKey(toAddr, fromAddr)

		var actualIndex *quarantine.QuarantineRecordSuffixIndex
		var actualKey []byte
		testFunc := func() {
			actualIndex, actualKey = s.keeper.GetQuarantineRecordSuffixIndex(store, toAddr, fromAddr)
		}
		s.Require().NotPanics(testFunc, "getQuarantineRecordSuffixIndex")
		s.Assert().Equal(expectedIndex, actualIndex, "returned index")
		s.Assert().Equal(expectedKey, actualKey, "returned key")
	})

	s.Run("4 set get unordered on previously set entry", func() {
		store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
		expectedIndex := &quarantine.QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffix2, suffix3, suffix1, suffix4, suffix5}}
		expectedKey := quarantine.CreateRecordIndexKey(toAddr, fromAddr)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecordSuffixIndex(store, expectedKey, expectedIndex)
		}
		s.Require().NotPanics(testFuncSet, "setQuarantineRecordSuffixIndex")

		var actualIndex *quarantine.QuarantineRecordSuffixIndex
		var actualKey []byte
		testFuncGet := func() {
			actualIndex, actualKey = s.keeper.GetQuarantineRecordSuffixIndex(store, toAddr, fromAddr)
		}
		s.Require().NotPanics(testFuncGet, "getQuarantineRecordSuffixIndex")
		s.Assert().Equal(expectedIndex, actualIndex, "returned index")
		s.Assert().Equal(expectedKey, actualKey, "returned key")
	})
}

func (s *TestSuite) TestAddQuarantineRecordSuffixIndexes() {
	toAddr := MakeTestAddr("sad", 0)
	fromAddr1 := MakeTestAddr("sad", 1)
	fromAddr2 := MakeTestAddr("sad", 2)
	fromAddr3 := MakeTestAddr("sad", 3)
	suffix0 := []byte(MakeTestAddr("sfxsad", 0))
	suffix1 := []byte(MakeTestAddr("sfxsad", 1))
	suffix2 := []byte(MakeTestAddr("sfxsad", 2))
	suffix3 := []byte(MakeTestAddr("sfxsad", 3))

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []sdk.AccAddress
		suffix    []byte
		expected  *quarantine.QuarantineRecordSuffixIndex
	}{
		{
			name:      "add to one nothing",
			toAddr:    toAddr,
			fromAddrs: accs(fromAddr1),
			suffix:    suffix0,
			expected:  qrsi(suffix0),
		},
		{
			name:      "add to two nothings",
			toAddr:    toAddr,
			fromAddrs: accs(fromAddr2, fromAddr3),
			suffix:    suffix3,
			expected:  qrsi(suffix3),
		},
		{
			name:      "add to one existing",
			toAddr:    toAddr,
			fromAddrs: accs(fromAddr1),
			suffix:    suffix1,
			expected:  qrsi(suffix0, suffix1),
		},
		{
			name:      "add to two existing",
			toAddr:    toAddr,
			fromAddrs: accs(fromAddr2, fromAddr3),
			suffix:    suffix2,
			expected:  qrsi(suffix2, suffix3), // Note that this tests ordering too.
		},
		{
			name:      "entry already exists",
			toAddr:    toAddr,
			fromAddrs: accs(fromAddr1),
			suffix:    suffix0,
			expected:  qrsi(suffix0, suffix1),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
			toAddrOrig := MakeCopyOfAccAddress(tc.toAddr)
			fromAddrsOrig := MakeCopyOfAccAddresses(tc.fromAddrs)
			suffixOrig := MakeCopyOfByteSlice(tc.suffix)

			testFuncAdd := func() {
				s.keeper.AddQuarantineRecordSuffixIndexes(store, tc.toAddr, tc.fromAddrs, tc.suffix)
			}
			s.Require().NotPanics(testFuncAdd, "addQuarantineRecordSuffixIndexes")
			s.Assert().Equal(toAddrOrig, tc.toAddr, "toAddr before and after")
			s.Assert().Equal(fromAddrsOrig, tc.fromAddrs, "fromAddrs before and after")
			s.Assert().Equal(suffixOrig, tc.suffix, "suffix before and after")

			for i, fromAddr := range fromAddrsOrig {
				expected := MakeCopyOfQuarantineRecordSuffixIndex(tc.expected)
				var actual *quarantine.QuarantineRecordSuffixIndex
				testFuncGet := func() {
					actual, _ = s.keeper.GetQuarantineRecordSuffixIndex(store, toAddrOrig, fromAddr)
				}
				if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecordSuffixIndex[%d]", i) {
					s.Assert().Equal(expected, actual, "result of GetQuarantineRecordSuffixIndex[%d]", i)
				}
			}
		})
	}
}

func (s *TestSuite) TestDeleteQuarantineRecordSuffixIndexes() {
	toAddr := MakeTestAddr("sad", 0)
	fromAddr1 := MakeTestAddr("sad", 1)
	fromAddr2 := MakeTestAddr("sad", 2)
	fromAddr3 := MakeTestAddr("sad", 3)
	suffix0 := []byte(MakeTestAddr("sfxsad", 0))
	suffix1 := []byte(MakeTestAddr("sfxsad", 1))
	suffix2 := []byte(MakeTestAddr("sfxsad", 2))
	suffix3 := []byte(MakeTestAddr("sfxsad", 3))

	// Create some existing entries that can then be altered.
	existing := []struct {
		key   []byte
		value *quarantine.QuarantineRecordSuffixIndex
	}{
		{
			key:   quarantine.CreateRecordIndexKey(toAddr, fromAddr1),
			value: qrsi(suffix0, suffix1, suffix2),
		},
		{
			key:   quarantine.CreateRecordIndexKey(toAddr, fromAddr2),
			value: qrsi(suffix2, suffix3),
		},
	}

	for i, e := range existing {
		store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
		testFuncSet := func() {
			s.keeper.SetQuarantineRecordSuffixIndex(store, e.key, e.value)
		}
		s.Require().NotPanics(testFuncSet, "setQuarantineRecordSuffixIndex[%d] for setup", i)
	}

	// Setup:
	// All will have the same toAddr.
	// <- 1: 0, 1, 2
	// <- 2: 2, 3
	// The above will be altered as the tests progress.

	tests := []struct {
		name      string
		fromAddrs []sdk.AccAddress
		suffix    []byte
		expected  []*quarantine.QuarantineRecordSuffixIndex
	}{
		{
			name:      "index does not exist",
			fromAddrs: accs(fromAddr3),
			suffix:    suffix0,
			expected:  qrsis(qrsi()),
		},
		{
			name:      "index exists without suffix",
			fromAddrs: accs(fromAddr1),
			suffix:    suffix3,
			expected:  qrsis(qrsi(suffix0, suffix1, suffix2)),
		},
		{
			name:      "index has suffix",
			fromAddrs: accs(fromAddr1),
			suffix:    suffix1,
			expected:  qrsis(qrsi(suffix0, suffix2)),
		},
		{
			name:      "three froms two have suffix other no index",
			fromAddrs: accs(fromAddr1, fromAddr2, fromAddr3),
			suffix:    suffix2,
			expected:  qrsis(qrsi(suffix0), qrsi(suffix3), qrsi()),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
			toAddrOrig := MakeCopyOfAccAddress(toAddr)
			fromAddrsOrig := MakeCopyOfAccAddresses(tc.fromAddrs)
			suffixOrig := MakeCopyOfByteSlice(tc.suffix)

			testFuncDelete := func() {
				s.keeper.DeleteQuarantineRecordSuffixIndexes(store, toAddr, tc.fromAddrs, tc.suffix)
			}
			s.Require().NotPanics(testFuncDelete, "deleteQuarantineRecordSuffixIndexes")
			s.Assert().Equal(toAddrOrig, toAddr, "toAddr before and after")
			s.Assert().Equal(fromAddrsOrig, tc.fromAddrs, "fromAddrs before and after")
			s.Assert().Equal(suffixOrig, tc.suffix, "suffix before and after")

			for i, expected := range tc.expected {
				var actual *quarantine.QuarantineRecordSuffixIndex
				testFuncGet := func() {
					actual, _ = s.keeper.GetQuarantineRecordSuffixIndex(store, toAddrOrig, fromAddrsOrig[i])
				}
				if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecordSuffixIndex[%d]", i) {
					s.Assert().Equal(expected, actual, "result of GetQuarantineRecordSuffixIndex[%d]", i)
				}
			}
		})
	}

	s.Run("deleting last entry removes it from the store", func() {
		store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
		key := quarantine.CreateRecordIndexKey(toAddr, fromAddr1)
		s.Require().True(store.Has(key), "store.Has(key) before the test starts")

		testFuncDelete := func() {
			s.keeper.DeleteQuarantineRecordSuffixIndexes(store, toAddr, accs(fromAddr1), suffix0)
		}
		s.Require().NotPanics(testFuncDelete, "deleteQuarantineRecordSuffixIndexes")
		if !s.Assert().False(store.Has(key), "store.Has(key) after last record should have been removed") {
			var actual *quarantine.QuarantineRecordSuffixIndex
			testFuncGet := func() {
				actual, _ = s.keeper.GetQuarantineRecordSuffixIndex(store, toAddr, fromAddr1)
			}
			if s.Assert().NotPanics(testFuncGet, "getQuarantineRecordSuffixIndex") {
				// This should always fail because getQuarantineRecordSuffixIndex never returns nil.
				// This is being called only when toe store still has the entry.
				// So this is just here so that the value in the store is included in the failure messages.
				s.Assert().Nil(actual)
			}
		}
	})
}

func (s *TestSuite) TestGetQuarantineRecordSuffixes() {
	// The effects of getQuarantineRecordSuffixes are well tested elsewhere. Just doing a big one-off here.
	toAddr := MakeTestAddr("gqrs", 0)
	fromAddr1 := MakeTestAddr("gqrs", 1)
	fromAddr2 := MakeTestAddr("gqrs", 2)
	fromAddr3 := MakeTestAddr("gqrs", 3)
	fromAddr4 := MakeTestAddr("gqrs", 4)
	suffix5 := []byte(MakeTestAddr("sfxgqrs", 5))
	suffix6 := []byte(MakeTestAddr("sfxgqrs", 6))
	suffix7 := []byte(MakeTestAddr("sfxgqrs", 7))
	fromAddr8 := MakeTestAddr("gqrs", 8)

	// sfxs is just a shorter way of creating a [][]byte
	sfxs := func(suffixes ...[]byte) [][]byte {
		return suffixes
	}

	// Setup:
	// This may or may not actually make sense, but it's how I'm setting it up.
	// {toAddr} <- {fromAddr} = {suffixes}
	// 0 <- 1 = 5,6
	// 0 <- 2 = 5,6
	// 0 <- 3 = 5
	// 0 <- 4 = (none)
	// 0 <- 8 = 7

	existing := []struct {
		from   sdk.AccAddress
		suffix []byte
	}{
		{from: fromAddr1, suffix: suffix5},
		{from: fromAddr1, suffix: suffix6},
		{from: fromAddr2, suffix: suffix5},
		{from: fromAddr2, suffix: suffix6},
		{from: fromAddr1, suffix: suffix5},
		{from: fromAddr3, suffix: suffix5},
		{from: fromAddr8, suffix: suffix7},
	}

	for i, e := range existing {
		store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
		testFuncAdd := func() {
			s.keeper.AddQuarantineRecordSuffixIndexes(store, toAddr, accs(e.from), e.suffix)
		}
		s.Require().NotPanics(testFuncAdd, "addQuarantineRecordSuffixIndexes[%d] for setup", i)
	}

	tests := []struct {
		name      string
		fromAddrs []sdk.AccAddress
		expected  [][]byte
	}{
		{name: "nil froms", fromAddrs: nil, expected: nil},
		{name: "empty froms", fromAddrs: []sdk.AccAddress{}, expected: nil},
		{name: "one from with zero suffixes", fromAddrs: accs(fromAddr4), expected: sfxs(fromAddr4)},
		{name: "one from with one shared suffix", fromAddrs: accs(fromAddr3), expected: sfxs(fromAddr3, suffix5)},
		{name: "one from with one lone suffix", fromAddrs: accs(fromAddr8), expected: sfxs(suffix7, fromAddr8)},
		{name: "one from with two suffixes", fromAddrs: accs(fromAddr1), expected: sfxs(fromAddr1, suffix5, suffix6)},
		{
			name:      "two froms with two overlapping suffixes",
			fromAddrs: accs(fromAddr1, fromAddr2),
			expected:  sfxs(fromAddr1, fromAddr2, suffix5, suffix6)},
		{
			name:      "two froms with two different suffixes",
			fromAddrs: accs(fromAddr8, fromAddr3),
			expected:  sfxs(fromAddr3, suffix5, suffix7, fromAddr8)},
		{
			name:      "five froms plus a dupe with 3 suffixes",
			fromAddrs: accs(fromAddr4, fromAddr1, fromAddr8, fromAddr2, fromAddr3, fromAddr1),
			expected:  sfxs(fromAddr1, fromAddr2, fromAddr3, fromAddr4, suffix5, suffix6, suffix7, fromAddr8)},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
			toAddrOrig := MakeCopyOfAccAddress(toAddr)
			fromAddrsOrig := MakeCopyOfAccAddresses(tc.fromAddrs)

			var actual [][]byte
			testFuncGet := func() {
				actual = s.keeper.GetQuarantineRecordSuffixes(store, toAddr, tc.fromAddrs)
			}
			s.Require().NotPanics(testFuncGet, "getQuarantineRecordSuffixes")
			s.Assert().Equal(toAddrOrig, toAddr, "toAddr before and after")
			s.Assert().Equal(fromAddrsOrig, tc.fromAddrs, "fromAddrs before and after")
			s.Assert().Equal(tc.expected, actual, "result of getQuarantineRecordSuffixes")
		})
	}
}

func (s *TestSuite) TestInitAndExportGenesis() {
	addr0 := MakeTestAddr("ieg", 0).String()
	addr1 := MakeTestAddr("ieg", 1).String()
	addr2 := MakeTestAddr("ieg", 2).String()
	addr3 := MakeTestAddr("ieg", 3).String()
	addr4 := MakeTestAddr("ieg", 4).String()
	addr5 := MakeTestAddr("ieg", 5).String()
	addr6 := MakeTestAddr("ieg", 6).String()
	addr7 := MakeTestAddr("ieg", 7).String()

	genesisState := &quarantine.GenesisState{
		QuarantinedAddresses: []string{addr0, addr2, addr4, addr6, addr7, addr5, addr1, addr3},
		AutoResponses: []*quarantine.AutoResponseEntry{
			{
				ToAddress:   addr5,
				FromAddress: addr4,
				Response:    quarantine.AUTO_RESPONSE_ACCEPT,
			},
			{
				ToAddress:   addr5,
				FromAddress: addr1,
				Response:    quarantine.AUTO_RESPONSE_ACCEPT,
			},
			{
				ToAddress:   addr5,
				FromAddress: addr2,
				Response:    quarantine.AUTO_RESPONSE_DECLINE,
			},
			{
				ToAddress:   addr2,
				FromAddress: addr5,
				Response:    quarantine.AUTO_RESPONSE_ACCEPT,
			},
			{
				ToAddress:   addr0,
				FromAddress: addr7,
				Response:    quarantine.AUTO_RESPONSE_UNSPECIFIED,
			},
			{
				ToAddress:   addr0,
				FromAddress: addr3,
				Response:    quarantine.AUTO_RESPONSE_DECLINE,
			},
		},
		QuarantinedFunds: []*quarantine.QuarantinedFunds{
			{
				ToAddress:               addr5,
				UnacceptedFromAddresses: []string{addr2},
				Coins:                   s.cz("2dull,5fancy"),
				Declined:                false,
			},
			{
				ToAddress:               addr0,
				UnacceptedFromAddresses: []string{addr1},
				Coins:                   s.cz("8fancy"),
				Declined:                false,
			},
			{
				ToAddress:               addr4,
				UnacceptedFromAddresses: []string{addr6},
				Coins:                   s.cz("200000dolla"),
				Declined:                false,
			},
			{
				ToAddress:               addr0,
				UnacceptedFromAddresses: []string{addr1, addr2},
				Coins:                   s.cz("21fancy"),
				Declined:                false,
			},
		},
	}

	expectedGenesisState := &quarantine.GenesisState{
		QuarantinedAddresses: []string{addr0, addr1, addr2, addr3, addr4, addr5, addr6, addr7},
		AutoResponses: []*quarantine.AutoResponseEntry{
			MakeCopyOfAutoResponseEntry(genesisState.AutoResponses[5]),
			MakeCopyOfAutoResponseEntry(genesisState.AutoResponses[3]),
			MakeCopyOfAutoResponseEntry(genesisState.AutoResponses[1]),
			MakeCopyOfAutoResponseEntry(genesisState.AutoResponses[2]),
			MakeCopyOfAutoResponseEntry(genesisState.AutoResponses[0]),
		},
		QuarantinedFunds: []*quarantine.QuarantinedFunds{
			MakeCopyOfQuarantinedFunds(genesisState.QuarantinedFunds[1]),
			MakeCopyOfQuarantinedFunds(genesisState.QuarantinedFunds[3]),
			MakeCopyOfQuarantinedFunds(genesisState.QuarantinedFunds[2]),
			MakeCopyOfQuarantinedFunds(genesisState.QuarantinedFunds[0]),
		},
	}

	s.Run("export while empty", func() {
		expected := &quarantine.GenesisState{
			QuarantinedAddresses: nil,
			AutoResponses:        nil,
			QuarantinedFunds:     nil,
		}
		var actual *quarantine.GenesisState
		testFuncExport := func() {
			actual = s.keeper.ExportGenesis(s.sdkCtx)
		}
		s.Require().NotPanics(testFuncExport, "ExportGenesis")
		s.Assert().Equal(expected, actual, "exported genesis state")

	})

	s.Run("init not enough funds", func() {
		bKeeper := NewMockBankKeeper()
		bKeeper.AllBalances[string(s.keeper.GetFundsHolder())] = s.cz("199999dolla,1dull,33fancy")
		qKeeper := s.keeper.WithBankKeeper(bKeeper)
		expectedErr := fmt.Sprintf("quarantine fund holder account %q does not have enough funds %q to cover quarantined funds %q",
			s.keeper.GetFundsHolder().String(), "199999dolla,1dull,33fancy", "200000dolla,2dull,34fancy")

		genStateCopy := MakeCopyOfGenesisState(genesisState)
		testFuncInit := func() {
			qKeeper.InitGenesis(s.sdkCtx, genStateCopy)
		}
		s.Require().PanicsWithError(expectedErr, testFuncInit, "InitGenesis")
	})

	s.Run("init with enough funds", func() {
		bKeeper := NewMockBankKeeper()
		bKeeper.AllBalances[string(s.keeper.GetFundsHolder())] = s.cz("200000dolla,2dull,34fancy")
		qKeeper := s.keeper.WithBankKeeper(bKeeper)

		genStateCopy := MakeCopyOfGenesisState(genesisState)
		testFuncInit := func() {
			qKeeper.InitGenesis(s.sdkCtx, genStateCopy)
		}
		s.Require().NotPanics(testFuncInit, "InitGenesis")
	})

	s.Run("export after successful init", func() {
		var actualGenesisState *quarantine.GenesisState
		testFuncExport := func() {
			actualGenesisState = s.keeper.ExportGenesis(s.sdkCtx)
		}
		s.Require().NotPanics(testFuncExport, "ExportGenesis")
		s.Assert().Equal(expectedGenesisState, actualGenesisState, "exported genesis state")
	})
}
