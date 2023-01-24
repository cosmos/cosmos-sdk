package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

type GovHooksTestSuite struct {
	BaseTestSuite
}

func (s *GovHooksTestSuite) SetupTest() {
	s.BaseSetup()
}

func TestGovHooksTestSuite(t *testing.T) {
	suite.Run(t, new(GovHooksTestSuite))
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalSubmission() {
	// Since this just calls proposalGovHook, all we should test in here is
	// that the proposalGovHook function was called for the given gov prop id.
	// So just mock up the gov keeper to return a proposal of interest, but with a bad status.
	// Hopefully the panic message that causes is unique to the proposalGovHook function.
	// We test that the call panics with the expected message.
	// We also test that GetProposal was called as expected.

	govPropID := uint64(3982)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 5555,
	}

	expPanic := "invalid governance proposal status: [5555]"
	testFunc := func() {
		s.Keeper.AfterProposalSubmission(s.SdkCtx, govPropID)
	}
	s.GovKeeper.GetProposalCalls = nil
	testutil.RequirePanicsWithMessage(s.T(), expPanic, testFunc, "AfterProposalSubmission")
	actualCalls := s.GovKeeper.GetProposalCalls
	if s.Assert().Len(actualCalls, 1, "number of calls made to GetProposal") {
		s.Assert().Equal(int(govPropID), int(actualCalls[0]), "the proposal requested to GetProposal")
	}
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalDeposit() {
	// Since this just calls proposalGovHook, all we should test in here is
	// that the proposalGovHook function was called for the given gov prop id.
	// So just mock up the gov keeper to return a proposal of interest, but with a bad status.
	// Hopefully the panic message that causes is unique to the proposalGovHook function.
	// We test that the call panics with the expected message.
	// We also test that GetProposal was called as expected.

	govPropID := uint64(5994)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 4434,
	}

	expPanic := "invalid governance proposal status: [4434]"
	testFunc := func() {
		s.Keeper.AfterProposalDeposit(s.SdkCtx, govPropID, sdk.AccAddress("this doesn't matter"))
	}
	s.GovKeeper.GetProposalCalls = nil
	testutil.RequirePanicsWithMessage(s.T(), expPanic, testFunc, "AfterProposalDeposit")
	actualCalls := s.GovKeeper.GetProposalCalls
	if s.Assert().Len(actualCalls, 1, "number of calls made to GetProposal") {
		s.Assert().Equal(int(govPropID), int(actualCalls[0]), "the proposal requested to GetProposal")
	}
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalVote() {
	// This one shouldn't do anything. So again, set it up to panic like the others,
	// but make sure it doesn't panic and that no calls were made to GetProposal

	govPropID := uint64(6370)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 2411,
	}

	testFunc := func() {
		s.Keeper.AfterProposalVote(s.SdkCtx, govPropID, sdk.AccAddress("this doesn't matter either"))
	}
	s.GovKeeper.GetProposalCalls = nil
	s.Require().NotPanics(testFunc, "AfterProposalVote")
	actualCalls := s.GovKeeper.GetProposalCalls
	s.Require().Nil(actualCalls, "calls made to GetProposal")
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalFailedMinDeposit() {
	// Since this just calls proposalGovHook, all we should test in here is
	// that the proposalGovHook function was called for the given gov prop id.
	// So just mock up the gov keeper to return a proposal of interest, but with a bad status.
	// Hopefully the panic message that causes is unique to the proposalGovHook function.
	// We test that the call panics with the expected message.
	// We also test that GetProposal was called as expected.

	govPropID := uint64(2111)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 3275,
	}

	expPanic := "invalid governance proposal status: [3275]"
	testFunc := func() {
		s.Keeper.AfterProposalFailedMinDeposit(s.SdkCtx, govPropID)
	}
	s.GovKeeper.GetProposalCalls = nil
	testutil.RequirePanicsWithMessage(s.T(), expPanic, testFunc, "AfterProposalFailedMinDeposit")
	actualCalls := s.GovKeeper.GetProposalCalls
	if s.Assert().Len(actualCalls, 1, "number of calls made to GetProposal") {
		s.Assert().Equal(int(govPropID), int(actualCalls[0]), "the proposal requested to GetProposal")
	}
}

func (s *GovHooksTestSuite) TestKeeper_AfterProposalVotingPeriodEnded() {
	// Since this just calls proposalGovHook, all we should test in here is
	// that the proposalGovHook function was called for the given gov prop id.
	// So just mock up the gov keeper to return a proposal of interest, but with a bad status.
	// Hopefully the panic message that causes is unique to the proposalGovHook function.
	// We test that the call panics with the expected message.
	// We also test that GetProposal was called as expected.

	govPropID := uint64(4041)
	s.GovKeeper.GetProposalReturns[govPropID] = govv1.Proposal{
		Id: govPropID,
		Messages: []*codectypes.Any{
			s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this addr doesn't matter"},
				Authority: "neither does this authority",
			}),
		},
		Status: 99,
	}

	expPanic := "invalid governance proposal status: [99]"
	testFunc := func() {
		s.Keeper.AfterProposalVotingPeriodEnded(s.SdkCtx, govPropID)
	}
	s.GovKeeper.GetProposalCalls = nil
	testutil.RequirePanicsWithMessage(s.T(), expPanic, testFunc, "AfterProposalVotingPeriodEnded")
	actualCalls := s.GovKeeper.GetProposalCalls
	if s.Assert().Len(actualCalls, 1, "number of calls made to GetProposal") {
		s.Assert().Equal(int(govPropID), int(actualCalls[0]), "the proposal requested to GetProposal")
	}
}

func (s *GovHooksTestSuite) TestKeeper_proposalGovHook() {
	// Make it easy to use a different proposal id for each test.
	lastPropID := uint64(0)
	nextPropID := func() uint64 {
		lastPropID += 1
		return lastPropID
	}
	// When using lastPropID directly in test definition, things got out of sync.
	// Basically, nextPropID was being called for all the tests before lastPropID was being used.
	// Having it returned by a func like this fixed that though.
	curPropID := func() uint64 {
		return lastPropID
	}

	addr1 := sdk.AccAddress("1st_hooks_test_addr")
	addr2 := sdk.AccAddress("2nd_hooks_test_addr")
	addr3 := sdk.AccAddress("3rd_hooks_test_addr")
	addr4 := sdk.AccAddress("4th_hooks_test_addr")
	addr5 := sdk.AccAddress("5th_hooks_test_addr")
	addr6 := sdk.AccAddress("6th_hooks_test_addr")

	// nonEmptyState creates a new GenesisState with a few things in it.
	nonEmptyState := func(govPropID uint64) *sanction.GenesisState {
		return &sanction.GenesisState{
			Params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("nesanct", 17)),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("neusanct", 23)),
			},
			SanctionedAddresses: []string{addr3.String(), addr4.String()},
			TemporaryEntries: []*sanction.TemporaryEntry{
				newTempEntry(addr4, govPropID+1, true),
				newTempEntry(addr5, govPropID, true),
				newTempEntry(addr6, govPropID, false),
				newTempEntry(addr6, govPropID+2, true),
			},
		}
	}
	// cleanupStateIni creates a new GenesisState with entries that are expected to be deleted during a test.
	cleanupStateIni := func(govPropID uint64) *sanction.GenesisState {
		return &sanction.GenesisState{
			Params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("nesanct", int64(govPropID-1))),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("neusanct", int64(govPropID+1))),
			},
			SanctionedAddresses: []string{addr3.String(), addr4.String()},
			TemporaryEntries: []*sanction.TemporaryEntry{
				newTempEntry(addr1, 1, true),
				newTempEntry(addr1, 50000, false),
				newTempEntry(addr3, govPropID, false),
				newTempEntry(addr4, govPropID, false),
				newTempEntry(addr5, govPropID-1, false),
				newTempEntry(addr5, govPropID, true),
				newTempEntry(addr5, govPropID+1, true),
				newTempEntry(addr6, govPropID-1, true),
				newTempEntry(addr6, govPropID, true),
				newTempEntry(addr6, govPropID+1, false),
			},
		}
	}
	// cleanupStateExp creates a new GenesisState that is a cleanupStateIni, but without the entries expected to be deleted.
	cleanupStateExp := func(govPropID uint64) *sanction.GenesisState {
		return &sanction.GenesisState{
			Params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("nesanct", int64(govPropID-1))),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("neusanct", int64(govPropID+1))),
			},
			SanctionedAddresses: []string{addr3.String(), addr4.String()},
			TemporaryEntries: []*sanction.TemporaryEntry{
				newTempEntry(addr1, 1, true),
				newTempEntry(addr1, 50000, false),
				newTempEntry(addr5, govPropID-1, false),
				newTempEntry(addr5, govPropID+1, true),
				newTempEntry(addr6, govPropID-1, true),
				newTempEntry(addr6, govPropID+1, false),
			},
		}
	}

	// Create some Any wrapped messages for easier proposal definitions.
	otherAny := s.NewAny(&govv1.MsgVote{
		ProposalId: 99887766,
		Voter:      "voter addr that should not matter",
		Option:     govv1.OptionAbstain,
	})
	updateParamsAny := s.NewAny(&sanction.MsgUpdateParams{
		Params: &sanction.Params{
			ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("ismdcoin", 59)),
			ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("iumdcoin", 86)),
		},
		Authority: "yadda yadda",
	})
	sanctionAny := s.NewAny(&sanction.MsgSanction{
		Addresses: []string{addr1.String(), addr2.String()},
		Authority: "nospendy",
	})
	unsanctionAny := s.NewAny(&sanction.MsgUnsanction{
		Addresses: []string{addr1.String(), addr3.String()},
		Authority: "spendyagainy",
	})
	// Note: In normal operation, it shouldn't be possible to end up with a governance proposal
	// containing a MsgExecLegacyContent sanction or unsanction; they don't implement Content.
	legWrapSanctAny := s.NewAny(&govv1.MsgExecLegacyContent{
		Content: s.NewAny(&sanction.MsgSanction{
			Addresses: []string{addr5.String(), addr6.String()},
			Authority: "notgonnadoit",
		}),
		Authority: "legacywrappingsanct",
	})
	legWrapUnsanctAny := s.NewAny(&govv1.MsgExecLegacyContent{
		Content: s.NewAny(&sanction.MsgUnsanction{
			Addresses: []string{addr4.String(), addr6.String()},
			Authority: "justignoreme",
		}),
		Authority: "legacywrappingunsanct",
	})

	// Define several proposal message sets to commonly use in tests.
	messageSets := []struct {
		name   string
		msgs   []*codectypes.Any
		boring bool // boring = true means none of the msgs should cause any changes.
	}{
		{name: "sanction", msgs: []*codectypes.Any{sanctionAny}},
		{name: "unsanction", msgs: []*codectypes.Any{unsanctionAny}},
		{name: "other", msgs: []*codectypes.Any{otherAny}, boring: true},
		{name: "update params", msgs: []*codectypes.Any{updateParamsAny}, boring: true},
		{name: "legacy wrapped sanction", msgs: []*codectypes.Any{legWrapSanctAny}, boring: true},
		{name: "legacy wrapped unsanction", msgs: []*codectypes.Any{legWrapUnsanctAny}, boring: true},
		{name: "sanction unsanction", msgs: []*codectypes.Any{sanctionAny, unsanctionAny}},
		{name: "unsanction sanction", msgs: []*codectypes.Any{unsanctionAny, sanctionAny}},
		{name: "three ignorable messages", msgs: []*codectypes.Any{legWrapSanctAny, updateParamsAny, legWrapSanctAny}, boring: true},
		{name: "three messages sanction x x", msgs: []*codectypes.Any{sanctionAny, updateParamsAny, otherAny}},
		{name: "three messages unsanction x x", msgs: []*codectypes.Any{unsanctionAny, updateParamsAny, otherAny}},
		{name: "three messages x sanction x", msgs: []*codectypes.Any{otherAny, sanctionAny, legWrapUnsanctAny}},
		{name: "three messages x unsanction x", msgs: []*codectypes.Any{otherAny, unsanctionAny, legWrapUnsanctAny}},
		{name: "three message x x sanction", msgs: []*codectypes.Any{updateParamsAny, legWrapUnsanctAny, sanctionAny}},
		{name: "three message x x unsanction", msgs: []*codectypes.Any{updateParamsAny, legWrapSanctAny, unsanctionAny}},
	}

	type testCase struct {
		name       string
		proposalID uint64
		iniState   *sanction.GenesisState
		proposal   *govv1.Proposal
		expState   *sanction.GenesisState
		expPanic   []string
	}

	tests := []testCase{}

	// prop status unknown -> panic
	for _, msgs := range messageSets {
		tests = append(tests, testCase{
			name:       "unknown prop status on " + msgs.name,
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: msgs.msgs,
				Status:   govv1.ProposalStatus(curPropID() + 5000),
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{fmt.Sprintf("invalid governance proposal status: [%d]", curPropID()+5000)},
		})
	}

	// prop status unspecified -> panic
	for _, msgs := range messageSets {
		tests = append(tests, testCase{
			name:       "unspecified prop status on " + msgs.name,
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: msgs.msgs,
				Status:   govv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			},
			expState: nonEmptyState(curPropID()),
			expPanic: []string{"invalid governance proposal status: [PROPOSAL_STATUS_UNSPECIFIED]"},
		})
	}

	// prop passed -> nothing happens in any case.
	for _, msgs := range messageSets {
		tests = append(tests, testCase{
			name:       "passed on " + msgs.name,
			proposalID: nextPropID(),
			iniState:   nonEmptyState(curPropID()),
			proposal: &govv1.Proposal{
				Id:       curPropID(),
				Messages: msgs.msgs,
				Status:   govv1.StatusPassed,
			},
			expState: nonEmptyState(curPropID()),
		})
	}

	// not found -> it's temp entries are deleted. No need to do this for each message set since there's no proposal.
	tests = append(tests, testCase{
		name:       "not found",
		proposalID: nextPropID(),
		iniState:   cleanupStateIni(curPropID()),
		proposal:   nil,
		expState:   cleanupStateExp(curPropID()),
	})

	// prop rejected, failed -> it's temp entries are deleted.
	for _, status := range []govv1.ProposalStatus{govv1.StatusRejected, govv1.StatusFailed} {
		var baseName string
		switch status {
		case govv1.StatusRejected:
			baseName = "rejected"
		case govv1.StatusFailed:
			baseName = "failed"
		default:
			s.FailNow("unhandled status in setup: %s", status)
		}
		for _, msgs := range messageSets {
			tests = append(tests, testCase{
				name:       baseName + " on " + msgs.name,
				proposalID: nextPropID(),
				iniState:   cleanupStateIni(curPropID()),
				proposal: &govv1.Proposal{
					Id:       curPropID(),
					Messages: msgs.msgs,
					Status:   status,
				},
				expState: cleanupStateExp(curPropID()),
			})
		}
	}

	// prop status deposit period or voting period -> depends on the deposit
	for _, status := range []govv1.ProposalStatus{govv1.StatusDepositPeriod, govv1.StatusVotingPeriod} {
		var baseName string
		switch status {
		case govv1.StatusDepositPeriod:
			baseName = "deposit period"
		case govv1.StatusVotingPeriod:
			baseName = "voting period"
		default:
			s.FailNow("unhandled status in setup: %s", status)
		}

		// min deposit = 0 -> nothing happens
		for _, msgs := range messageSets {
			state := nonEmptyState(nextPropID())
			dep := sdk.Coins{}
			for _, c := range state.Params.ImmediateSanctionMinDeposit.Add(state.Params.ImmediateUnsanctionMinDeposit...) {
				dep = dep.Add(sdk.NewCoin(c.Denom, c.Amount.AddRaw(100)))
			}
			state.Params.ImmediateSanctionMinDeposit = nil
			state.Params.ImmediateUnsanctionMinDeposit = nil
			tests = append(tests, testCase{
				name:       baseName + " min dep = 0 on " + msgs.name,
				proposalID: curPropID(),
				iniState:   state,
				proposal: &govv1.Proposal{
					Id:           curPropID(),
					Messages:     msgs.msgs,
					Status:       status,
					TotalDeposit: dep,
				},
				expState: state,
			})
		}

		// deposit < min deposit -> nothing happens
		for _, msgs := range messageSets {
			state := nonEmptyState(nextPropID())
			dep := sdk.Coins{}
			for _, c := range state.Params.ImmediateSanctionMinDeposit.Add(state.Params.ImmediateUnsanctionMinDeposit...) {
				if c.Amount.GT(sdk.NewInt(1)) {
					dep = dep.Add(sdk.NewCoin(c.Denom, c.Amount.SubRaw(1)))
				}
			}
			tests = append(tests, testCase{
				name:       baseName + " dep < min dep on " + msgs.name,
				proposalID: curPropID(),
				iniState:   state,
				proposal: &govv1.Proposal{
					Id:           curPropID(),
					Messages:     msgs.msgs,
					Status:       status,
					TotalDeposit: dep,
				},
				expState: state,
			})
		}

		// deposit == min deposit or depoist > min deposit -> temp entries added for non-boring msgs.
		mods := []struct {
			name   string
			amount sdk.Int
		}{
			{name: " deposit = min", amount: sdk.NewInt(0)},
			{name: " deposit > min", amount: sdk.NewInt(1)},
		}
		for _, mod := range mods {
			// Using all the messageSets for this would be to complex (just duplicating the code I'm trying to test).
			// Plus, there's a couple extra cases to account for here.
			// So define them verbosely here.
			// The deposit will be added to each proposal before being added to the tests.
			newTests := []testCase{
				{
					name:       baseName + mod.name + " on sanction",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("jsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: nil,
						},
						SanctionedAddresses: []string{addr1.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), false),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{s.NewAny(&sanction.MsgSanction{
							Addresses: []string{addr1.String(), addr4.String()},
							Authority: "whatever",
						})},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("jsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: nil,
						},
						SanctionedAddresses: []string{addr1.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr1, curPropID(), true),
							newTempEntry(addr4, curPropID(), true),
							newTempEntry(addr6, curPropID(), false),
						},
					},
				},
				{
					name:       baseName + mod.name + " on unsanction",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   nil,
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("junsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), true),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{s.NewAny(&sanction.MsgUnsanction{
							Addresses: []string{addr2.String(), addr5.String()},
							Authority: "whatever",
						})},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   nil,
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("junsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr2, curPropID(), false),
							newTempEntry(addr5, curPropID(), false),
							newTempEntry(addr6, curPropID(), true),
						},
					},
				},
				{
					name:       baseName + mod.name + " on sanction sanction",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("jsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: nil,
						},
						SanctionedAddresses: []string{addr1.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), false),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{
							s.NewAny(&sanction.MsgSanction{
								Addresses: []string{addr1.String(), addr4.String()},
								Authority: "whatever",
							}),
							s.NewAny(&sanction.MsgSanction{
								Addresses: []string{addr3.String()},
								Authority: "whatever",
							}),
						},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("jsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: nil,
						},
						SanctionedAddresses: []string{addr1.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr1, curPropID(), true),
							newTempEntry(addr3, curPropID(), true),
							newTempEntry(addr4, curPropID(), true),
							newTempEntry(addr6, curPropID(), false),
						},
					},
				},
				{
					name:       baseName + mod.name + " on unsanction unsanction",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   nil,
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("junsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), true),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{
							s.NewAny(&sanction.MsgUnsanction{
								Addresses: []string{addr2.String(), addr5.String()},
								Authority: "whatever",
							}),
							s.NewAny(&sanction.MsgUnsanction{
								Addresses: []string{addr3.String()},
								Authority: "whatever",
							}),
						},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   nil,
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("junsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr2, curPropID(), false),
							newTempEntry(addr3, curPropID(), false),
							newTempEntry(addr5, curPropID(), false),
							newTempEntry(addr6, curPropID(), true),
						},
					},
				},
				{
					name:       baseName + mod.name + " on sanction unsanction",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("bsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bunsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr1.String(), addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), true),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{
							s.NewAny(&sanction.MsgSanction{
								Addresses: []string{addr1.String(), addr4.String()},
								Authority: "whatever1",
							}),
							s.NewAny(&sanction.MsgUnsanction{
								Addresses: []string{addr2.String(), addr5.String()},
								Authority: "whatever2",
							}),
						},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("bsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bunsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr1.String(), addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr1, curPropID(), true),
							newTempEntry(addr2, curPropID(), false),
							newTempEntry(addr4, curPropID(), true),
							newTempEntry(addr5, curPropID(), false),
							newTempEntry(addr6, curPropID(), true),
						},
					},
				},
				{
					name:       baseName + mod.name + " on unsanction sanction",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("bsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bunsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr1.String(), addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), true),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{
							s.NewAny(&sanction.MsgUnsanction{
								Addresses: []string{addr2.String(), addr5.String()},
								Authority: "whatever2",
							}),
							s.NewAny(&sanction.MsgSanction{
								Addresses: []string{addr1.String(), addr4.String()},
								Authority: "whatever1",
							}),
						},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("bsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bunsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr1.String(), addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr1, curPropID(), true),
							newTempEntry(addr2, curPropID(), false),
							newTempEntry(addr4, curPropID(), true),
							newTempEntry(addr5, curPropID(), false),
							newTempEntry(addr6, curPropID(), true),
						},
					},
				},
				{
					name:       baseName + mod.name + " on sanction x",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("jsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: nil,
						},
						SanctionedAddresses: []string{addr1.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), false),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{
							s.NewAny(&sanction.MsgSanction{
								Addresses: []string{addr1.String(), addr4.String()},
								Authority: "whatever",
							}),
							updateParamsAny,
						},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("jsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: nil,
						},
						SanctionedAddresses: []string{addr1.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr1, curPropID(), true),
							newTempEntry(addr4, curPropID(), true),
							newTempEntry(addr6, curPropID(), false),
						},
					},
				},
				{
					name:       baseName + mod.name + " on x sanction",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("jsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: nil,
						},
						SanctionedAddresses: []string{addr1.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), false),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{
							updateParamsAny,
							s.NewAny(&sanction.MsgSanction{
								Addresses: []string{addr1.String(), addr4.String()},
								Authority: "whatever",
							}),
						},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("jsanct", int64(curPropID()))),
							ImmediateUnsanctionMinDeposit: nil,
						},
						SanctionedAddresses: []string{addr1.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr1, curPropID(), true),
							newTempEntry(addr4, curPropID(), true),
							newTempEntry(addr6, curPropID(), false),
						},
					},
				},
				{
					name:       baseName + mod.name + " on unsanction x",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   nil,
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("junsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), true),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{
							s.NewAny(&sanction.MsgUnsanction{
								Addresses: []string{addr2.String(), addr5.String()},
								Authority: "whatever",
							}),
							legWrapSanctAny,
						},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   nil,
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("junsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr2, curPropID(), false),
							newTempEntry(addr5, curPropID(), false),
							newTempEntry(addr6, curPropID(), true),
						},
					},
				},
				{
					name:       baseName + mod.name + " on x unsanction",
					proposalID: nextPropID(),
					iniState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   nil,
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("junsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr6, curPropID(), true),
						},
					},
					proposal: &govv1.Proposal{
						Id: curPropID(),
						Messages: []*codectypes.Any{
							otherAny,
							s.NewAny(&sanction.MsgUnsanction{
								Addresses: []string{addr2.String(), addr5.String()},
								Authority: "whatever",
							}),
						},
						Status: status,
					},
					expState: &sanction.GenesisState{
						Params: &sanction.Params{
							ImmediateSanctionMinDeposit:   nil,
							ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("junsanct", int64(curPropID()))),
						},
						SanctionedAddresses: []string{addr2.String()},
						TemporaryEntries: []*sanction.TemporaryEntry{
							newTempEntry(addr2, curPropID(), false),
							newTempEntry(addr5, curPropID(), false),
							newTempEntry(addr6, curPropID(), true),
						},
					},
				},
			}

			for _, tc := range newTests {
				dep := sdk.Coins{}
				for _, c := range tc.iniState.Params.ImmediateSanctionMinDeposit.Add(tc.iniState.Params.ImmediateUnsanctionMinDeposit...) {
					dep = dep.Add(sdk.NewCoin(c.Denom, c.Amount.Add(mod.amount)))
				}
				tc.proposal.TotalDeposit = dep
				tests = append(tests, tc)
			}

			// Now add all the boring message sets since they shouldn't do anything.
			for _, msgs := range messageSets {
				if msgs.boring {
					state := nonEmptyState(nextPropID())
					dep := sdk.Coins{}
					for _, c := range state.Params.ImmediateSanctionMinDeposit.Add(state.Params.ImmediateUnsanctionMinDeposit...) {
						dep = dep.Add(sdk.NewCoin(c.Denom, c.Amount.Add(mod.amount)))
					}
					tests = append(tests, testCase{
						name:       baseName + mod.name + " on " + msgs.name,
						proposalID: curPropID(),
						iniState:   state,
						proposal: &govv1.Proposal{
							Id:           curPropID(),
							Messages:     msgs.msgs,
							Status:       status,
							TotalDeposit: dep,
						},
						expState: state,
					})
				}
			}
		}

		// A few cases where the deposit is enough for one message but not another.
		// only enough deposit for sanction, sanction unsanction
		tests = append(tests, testCase{
			name:       baseName + " sanct < dep < unsanct on sanction unsanction ",
			proposalID: nextPropID(),
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("bothcoin", 5)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bothcoin", 6)),
				},
				SanctionedAddresses: []string{addr1.String(), addr2.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr6, curPropID(), false),
				},
			},
			proposal: &govv1.Proposal{
				Id: curPropID(),
				Messages: []*codectypes.Any{
					s.NewAny(&sanction.MsgSanction{
						Addresses: []string{addr1.String(), addr4.String()},
						Authority: "whatever",
					}),
					s.NewAny(&sanction.MsgUnsanction{
						Addresses: []string{addr2.String(), addr5.String()},
						Authority: "whatever",
					}),
				},
				Status:       status,
				TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("bothcoin", 5)),
			},
			expState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("bothcoin", 5)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bothcoin", 6)),
				},
				SanctionedAddresses: []string{addr1.String(), addr2.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, curPropID(), true),
					newTempEntry(addr4, curPropID(), true),
					newTempEntry(addr6, curPropID(), false),
				},
			},
		})
		// only enough deposit for sanction, unsanction sanction
		tests = append(tests, testCase{
			name:       baseName + " sanct < dep < unsanct on unsanction sanction ",
			proposalID: nextPropID(),
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("twocoin", 5)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("twocoin", 6)),
				},
				SanctionedAddresses: []string{addr1.String(), addr2.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr6, curPropID(), false),
				},
			},
			proposal: &govv1.Proposal{
				Id: curPropID(),
				Messages: []*codectypes.Any{
					s.NewAny(&sanction.MsgUnsanction{
						Addresses: []string{addr2.String(), addr5.String()},
						Authority: "whatever",
					}),
					s.NewAny(&sanction.MsgSanction{
						Addresses: []string{addr1.String(), addr4.String()},
						Authority: "whatever",
					}),
				},
				Status:       status,
				TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("twocoin", 5)),
			},
			expState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("twocoin", 5)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("twocoin", 6)),
				},
				SanctionedAddresses: []string{addr1.String(), addr2.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr1, curPropID(), true),
					newTempEntry(addr4, curPropID(), true),
					newTempEntry(addr6, curPropID(), false),
				},
			},
		})
		// only enough deposit for unsanction, sanction unsanction
		tests = append(tests, testCase{
			name:       baseName + " unsanct < dep < sanct on sanction unsanction ",
			proposalID: nextPropID(),
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("twincoin", 6)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("twincoin", 5)),
				},
				SanctionedAddresses: []string{addr1.String(), addr2.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr6, curPropID(), false),
				},
			},
			proposal: &govv1.Proposal{
				Id: curPropID(),
				Messages: []*codectypes.Any{
					s.NewAny(&sanction.MsgSanction{
						Addresses: []string{addr1.String(), addr4.String()},
						Authority: "whatever",
					}),
					s.NewAny(&sanction.MsgUnsanction{
						Addresses: []string{addr2.String(), addr5.String()},
						Authority: "whatever",
					}),
				},
				Status:       status,
				TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("twincoin", 5)),
			},
			expState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("twincoin", 6)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("twincoin", 5)),
				},
				SanctionedAddresses: []string{addr1.String(), addr2.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr2, curPropID(), false),
					newTempEntry(addr5, curPropID(), false),
					newTempEntry(addr6, curPropID(), false),
				},
			},
		})
		// only enough deposit for unsanction, unsanction sanction
		tests = append(tests, testCase{
			name:       baseName + " unsanct < dep < sanct on unsanction sanction ",
			proposalID: nextPropID(),
			iniState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("duocoin", 6)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("duocoin", 5)),
				},
				SanctionedAddresses: []string{addr1.String(), addr2.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr6, curPropID(), false),
				},
			},
			proposal: &govv1.Proposal{
				Id: curPropID(),
				Messages: []*codectypes.Any{
					s.NewAny(&sanction.MsgUnsanction{
						Addresses: []string{addr2.String(), addr5.String()},
						Authority: "whatever",
					}),
					s.NewAny(&sanction.MsgSanction{
						Addresses: []string{addr1.String(), addr4.String()},
						Authority: "whatever",
					}),
				},
				Status:       status,
				TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("duocoin", 5)),
			},
			expState: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("duocoin", 6)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("duocoin", 5)),
				},
				SanctionedAddresses: []string{addr1.String(), addr2.String()},
				TemporaryEntries: []*sanction.TemporaryEntry{
					newTempEntry(addr2, curPropID(), false),
					newTempEntry(addr5, curPropID(), false),
					newTempEntry(addr6, curPropID(), false),
				},
			},
		})
	}

	// sanity check on tests names since there's probably a lot of copy/pasting being done in here.
	names := map[string]bool{}
	for _, tc := range tests {
		if names[tc.name] {
			s.FailNow("test name duplicated: " + tc.name)
		}
		names[tc.name] = true
	}

	// Finally.... run the tests.
	for i, tc := range tests {
		// Including the index in the name since a) most names are from concatenation, and
		// b) so it's easier to change the loop to tests[5:6] to isolate and troubleshoot a single test of interest.
		s.Run(fmt.Sprintf("%03d %s", i, tc.name), func() {
			s.ClearState()
			if tc.iniState != nil {
				s.Require().NotPanics(func() {
					s.Keeper.InitGenesis(s.SdkCtx, tc.iniState)
				}, "InitGenesis")
			}

			s.GovKeeper.GetProposalCalls = nil
			if tc.proposal != nil {
				s.GovKeeper.GetProposalReturns[tc.proposal.Id] = *tc.proposal
			}

			testFunc := func() {
				s.Keeper.OnlyTestsProposalGovHook(s.SdkCtx, tc.proposalID)
			}
			testutil.RequirePanicContents(s.T(), tc.expPanic, testFunc, "proposalGovHook(%d)", tc.proposalID)

			getPropCalls := s.GovKeeper.GetProposalCalls
			if s.Assert().Len(getPropCalls, 1, "number of calls made to GetProposal") {
				// doing it this way because a failure message from an .Equal on two []uint64 slices shows
				// the values in hex. Since they're decimal in test definition, this is just easier.
				s.Assert().Equal(int(tc.proposalID), int(getPropCalls[0]), "gov prop id provided to GetProposal")
			}

			if tc.expState != nil {
				s.ExportAndCheck(tc.expState)
			}
		})
	}
}

func (s *GovHooksTestSuite) TestKeeper_isModuleGovHooksMsgURL() {
	tests := []struct {
		url string
		exp bool
	}{
		{exp: true, url: sdk.MsgTypeURL(&sanction.MsgSanction{})},
		{exp: true, url: sdk.MsgTypeURL(&sanction.MsgUnsanction{})},
		{exp: false, url: ""},
		{exp: false, url: "     "},
		{exp: false, url: "something random"},
		{exp: false, url: sdk.MsgTypeURL(&sanction.MsgUpdateParams{})},
		{exp: false, url: sdk.MsgTypeURL(&govv1.MsgExecLegacyContent{})},
		{exp: false, url: "cosmos.sanction.v1beta1.MsgSanction"},
		{exp: false, url: "/cosmos.sanction.v1beta1.MsgSanctio"},
		{exp: false, url: "/cosmos.sanction.v1beta1.MsgSanction "},
		{exp: false, url: " /cosmos.sanction.v1beta1.MsgSanction"},
		{exp: false, url: "/cosmos.sanction.v1beta1.MsgSanction2"},
	}

	for _, tc := range tests {
		name := tc.url
		if name == "" {
			name = "empty"
		}
		if strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("spaces x %d", len(name))
		}
		s.Run(name, func() {
			var actual bool
			testFunc := func() {
				actual = s.Keeper.OnlyTestsIsModuleGovHooksMsgURL(tc.url)
			}
			s.Require().NotPanics(testFunc, "isModuleGovHooksMsgURL(%q)", tc.url)
			s.Assert().Equal(tc.exp, actual, "isModuleGovHooksMsgURL(%q) result", tc.url)
		})
	}
}

func (s *GovHooksTestSuite) TestKeeper_getMsgAddresses() {
	addr1 := sdk.AccAddress("1_good_addr_for_test")
	addr2 := sdk.AccAddress("2_good_addr_for_test")
	addr3 := sdk.AccAddress("3_good_addr_for_test")
	addr4 := sdk.AccAddress("4_good_addr_for_test")
	addr5 := sdk.AccAddress("5_good_addr_for_test")
	addr6 := sdk.AccAddress("6_good_addr_for_test")

	tests := []struct {
		name     string
		msg      *codectypes.Any
		exp      []sdk.AccAddress
		expPanic []string
	}{
		// Tests for things outside the switch.
		{
			name: "nil",
			msg:  nil,
			exp:  nil,
		},
		{
			name: "type url is empty but content is MsgSanction",
			msg: s.CustomAny(nil, &sanction.MsgSanction{
				Addresses: []string{addr1.String()},
				Authority: "whatever",
			}),
			exp: nil,
		},
		{
			name: "type url is empty but content is MsgUnsanction",
			msg: s.CustomAny(nil, &sanction.MsgUnsanction{
				Addresses: []string{addr1.String()},
				Authority: "whatever",
			}),
			exp: nil,
		},
		{
			name: "MsgUpdateParams",
			msg: s.NewAny(&sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("pcoin", 1)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("pcoin", 2)),
				},
				Authority: "whatever",
			}),
			exp: nil,
		},
		{
			name: "MsgExecLegacyContent with a MsgSanction in it",
			msg: s.NewAny(&govv1.MsgExecLegacyContent{
				Content: s.NewAny(&sanction.MsgSanction{
					Addresses: []string{addr1.String()},
					Authority: "whatever2",
				}),
				Authority: "whatever",
			}),
			exp: nil,
		},
		{
			name: "MsgExecLegacyContent with a MsgUnsanction in it",
			msg: s.NewAny(&govv1.MsgExecLegacyContent{
				Content: s.NewAny(&sanction.MsgUnsanction{
					Addresses: []string{addr1.String()},
					Authority: "whatever2",
				}),
				Authority: "whatever",
			}),
			exp: nil,
		},

		// Tests for the MsgSanction case.
		{
			name: "MsgSanction nil addrs",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: nil,
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{},
		},
		{
			name: "MsgSanction empty addrs",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{},
		},
		{
			name: "MsgSanction one addr good",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{addr1.String()},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{addr1},
		},
		{
			name: "MsgSanction one addr bad",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{"this1isnotgood"},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "MsgSanction six addrs all good",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{addr1, addr2, addr3, addr4, addr5, addr6},
		},
		{
			name: "MsgSanction six addrs bad xxxx",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{
					"this1isalsobad",
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "MsgSanction six addrs bad third",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					"this1isthethirdbadone",
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[2]", "decoding bech32 failed"},
		},
		{
			name: "MsgSanction six addrs bad sixth",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					"another1thatisnotgood",
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[5]", "decoding bech32 failed"},
		},
		{
			name: "type is MsgSanction but content is not",
			msg: s.CustomAny(&sanction.MsgSanction{}, &govv1.MsgVote{
				ProposalId: 5,
				Voter:      addr1.String(),
				Option:     govv1.OptionNo,
				Metadata:   "I do not know what is going on",
			}),
			expPanic: []string{"no registered implementations of type *sanction.MsgSanction"},
		},

		// Tests for the MsgUnsanction case.
		{
			name: "MsgUnsanction nil addrs",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: nil,
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{},
		},
		{
			name: "MsgUnsanction empty addrs",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{},
		},
		{
			name: "MsgUnsanction one addr good",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{addr1.String()},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{addr1},
		},
		{
			name: "MsgUnsanction one addr bad",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{"this1isnotgood"},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "MsgUnsanction six addrs all good",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			exp: []sdk.AccAddress{addr1, addr2, addr3, addr4, addr5, addr6},
		},
		{
			name: "MsgUnsanction six addrs bad xxxx",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{
					"this1isalsobad",
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "MsgUnsanction six addrs bad third",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					"this1isthethirdbadone",
					addr4.String(),
					addr5.String(),
					addr6.String(),
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[2]", "decoding bech32 failed"},
		},
		{
			name: "MsgUnsanction six addrs bad sixth",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: []string{
					addr1.String(),
					addr2.String(),
					addr3.String(),
					addr4.String(),
					addr5.String(),
					"another1thatisnotgood",
				},
				Authority: "whatever",
			}),
			expPanic: []string{"invalid address[5]", "decoding bech32 failed"},
		},
		{
			name: "type is MsgUnsanction but content is not",
			msg: s.CustomAny(&sanction.MsgUnsanction{}, &govv1.MsgVote{
				ProposalId: 5,
				Voter:      addr1.String(),
				Option:     govv1.OptionNo,
				Metadata:   "I do not know what is going on",
			}),
			expPanic: []string{"no registered implementations of type *sanction.MsgUnsanction"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual []sdk.AccAddress
			testFunc := func() {
				actual = s.Keeper.OnlyTestsGetMsgAddresses(tc.msg)
			}
			testutil.AssertPanicContents(s.T(), tc.expPanic, testFunc, "getMsgAddresses")
			s.Assert().Equal(tc.exp, actual, "getMsgAddresses result")
		})
	}
}

func (s *GovHooksTestSuite) TestKeeper_getImmediateMinDeposit() {
	origSanctMin := sanction.DefaultImmediateSanctionMinDeposit
	origUnsanctMin := sanction.DefaultImmediateUnsanctionMinDeposit
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origSanctMin
		sanction.DefaultImmediateUnsanctionMinDeposit = origUnsanctMin
	}()
	sanction.DefaultImmediateSanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("dsanct", 3))
	sanction.DefaultImmediateUnsanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("dunsanct", 7))

	paramSanctMin := sdk.NewCoins(sdk.NewInt64Coin("psanct", 5))
	paramUnsanctMin := sdk.NewCoins(sdk.NewInt64Coin("punsanct", 10))

	tests := []struct {
		name string
		msg  *codectypes.Any
		exp  sdk.Coins // expected for either case.
		expd sdk.Coins // expected when getting from defaults.
		expp sdk.Coins // expected when getting from params.
	}{
		{
			name: "nil",
			msg:  nil,
			exp:  sdk.Coins{},
		},
		{
			name: "MsgSanction",
			msg: s.NewAny(&sanction.MsgSanction{
				Addresses: nil,
				Authority: "whatever",
			}),
			expd: sanction.DefaultImmediateSanctionMinDeposit,
			expp: paramSanctMin,
		},
		{
			name: "MsgUnsanction",
			msg: s.NewAny(&sanction.MsgUnsanction{
				Addresses: nil,
				Authority: "whatever",
			}),
			expd: sanction.DefaultImmediateUnsanctionMinDeposit,
			expp: paramUnsanctMin,
		},
		{
			name: "MsgExecLegacyContent with a MsgSanction",
			msg: s.NewAny(&govv1.MsgExecLegacyContent{
				Content: s.NewAny(&sanction.MsgSanction{
					Addresses: []string{"some dumb addr", "another unsavory addr"},
					Authority: "whatever2",
				}),
				Authority: "whatever",
			}),
			exp: sdk.Coins{},
		},
		{
			name: "MsgExecLegacyContent with a MsgUnsanction",
			msg: s.NewAny(&govv1.MsgExecLegacyContent{
				Content: s.NewAny(&sanction.MsgUnsanction{
					Addresses: []string{"some dumb addr", "another unsavory addr"},
					Authority: "whatever2",
				}),
				Authority: "whatever",
			}),
			exp: sdk.Coins{},
		},
		{
			name: "MsgUpdateParams",
			msg: s.NewAny(&sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("qcoin", 72)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("qcoin", 91)),
				},
				Authority: "whatever",
			}),
			exp: sdk.Coins{},
		},
	}

	// Delete the params so that the defaults are used.
	s.RequireNotPanicsNoError(func() error {
		return s.Keeper.SetParams(s.SdkCtx, nil)
	}, "SetParams(nil)")

	for _, tc := range tests {
		s.Run(tc.name+" from defaults", func() {
			expected := tc.expd
			if expected == nil {
				expected = tc.exp
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.Keeper.OnlyTestsGetImmediateMinDeposit(s.SdkCtx, tc.msg)
			}
			s.Require().NotPanics(testFunc, "getImmediateMinDeposit")
			s.Assert().Equal(expected, actual, "getImmediateMinDeposit result")
		})
	}

	// Now, set the params appropriately.
	s.RequireNotPanicsNoError(func() error {
		return s.Keeper.SetParams(s.SdkCtx, &sanction.Params{
			ImmediateSanctionMinDeposit:   paramSanctMin,
			ImmediateUnsanctionMinDeposit: paramUnsanctMin,
		})
	}, "SetParams with values")

	for _, tc := range tests {
		s.Run(tc.name+" from params", func() {
			expected := tc.expp
			if expected == nil {
				expected = tc.exp
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.Keeper.OnlyTestsGetImmediateMinDeposit(s.SdkCtx, tc.msg)
			}
			s.Require().NotPanics(testFunc, "getImmediateMinDeposit")
			s.Assert().Equal(expected, actual, "getImmediateMinDeposit result")
		})
	}
}
