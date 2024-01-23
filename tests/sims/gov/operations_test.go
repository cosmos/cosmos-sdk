package simulation_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	_ "cosmossdk.io/x/auth"
	authkeeper "cosmossdk.io/x/auth/keeper"
	_ "cosmossdk.io/x/auth/tx/config"
	_ "cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/testutil"
	_ "cosmossdk.io/x/gov"
	"cosmossdk.io/x/gov/keeper"
	"cosmossdk.io/x/gov/simulation"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"
	_ "cosmossdk.io/x/protocolpool"
	_ "cosmossdk.io/x/staking"
	stakingkeeper "cosmossdk.io/x/staking/keeper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
)

var (
	_ simtypes.WeightedProposalMsg     = MockWeightedProposals{}
	_ simtypes.WeightedProposalContent = MockWeightedProposals{} //nolint:staticcheck // testing legacy code path
)

type MockWeightedProposals struct {
	n int
}

func (m MockWeightedProposals) AppParamsKey() string {
	return fmt.Sprintf("AppParamsKey-%d", m.n)
}

func (m MockWeightedProposals) DefaultWeight() int {
	return m.n
}

func (m MockWeightedProposals) MsgSimulatorFn() simtypes.MsgSimulatorFn {
	return func(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
		return nil
	}
}

func (m MockWeightedProposals) ContentSimulatorFn() simtypes.ContentSimulatorFn { //nolint:staticcheck // testing legacy code path
	return func(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) simtypes.Content { //nolint:staticcheck // testing legacy code path
		return v1beta1.NewTextProposal(
			fmt.Sprintf("title-%d: %s", m.n, simtypes.RandStringOfLength(r, 100)),
			fmt.Sprintf("description-%d: %s", m.n, simtypes.RandStringOfLength(r, 4000)),
		)
	}
}

func mockWeightedProposalMsg(n int) []simtypes.WeightedProposalMsg {
	wpc := make([]simtypes.WeightedProposalMsg, n)
	for i := 0; i < n; i++ {
		wpc[i] = MockWeightedProposals{i}
	}
	return wpc
}

func mockWeightedLegacyProposalContent(n int) []simtypes.WeightedProposalContent { //nolint:staticcheck // testing legacy code path
	wpc := make([]simtypes.WeightedProposalContent, n) //nolint:staticcheck // testing legacy code path
	for i := 0; i < n; i++ {
		wpc[i] = MockWeightedProposals{i}
	}
	return wpc
}

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	suite, ctx := createTestSuite(t, false)
	app := suite.App
	ctx.WithChainID("test-chain")
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(appParams, suite.TxConfig, suite.AccountKeeper,
		suite.BankKeeper, suite.GovKeeper, mockWeightedProposalMsg(3), mockWeightedLegacyProposalContent(1),
	)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, suite.AccountKeeper, suite.BankKeeper, suite.StakingKeeper, ctx, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.DefaultWeightMsgDeposit, types.ModuleName, simulation.TypeMsgDeposit},
		{simulation.DefaultWeightMsgVote, types.ModuleName, simulation.TypeMsgVote},
		{simulation.DefaultWeightMsgVoteWeighted, types.ModuleName, simulation.TypeMsgVoteWeighted},
		{simulation.DefaultWeightMsgCancelProposal, types.ModuleName, simulation.TypeMsgCancelProposal},
		{0, types.ModuleName, simulation.TypeMsgSubmitProposal},
		{1, types.ModuleName, simulation.TypeMsgSubmitProposal},
		{2, types.ModuleName, simulation.TypeMsgSubmitProposal},
		{0, types.ModuleName, simulation.TypeMsgSubmitProposal},
	}

	require.Equal(t, len(weightesOps), len(expected), "number of operations should be the same")
	for i, w := range weightesOps {
		operationMsg, _, err := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
		require.NoError(t, err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(t, expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(t, expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(t, expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgSubmitProposal tests the normal scenario of a valid message of type TypeMsgSubmitProposal.
// Abnormal scenarios, where errors occur, are not tested here.
func TestSimulateMsgSubmitProposal(t *testing.T) {
	suite, ctx := createTestSuite(t, false)
	app := suite.App

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, suite.AccountKeeper, suite.BankKeeper, suite.StakingKeeper, ctx, 3)

	// execute operation
	op := simulation.SimulateMsgSubmitProposal(suite.TxConfig, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper, MockWeightedProposals{3}.MsgSimulatorFn())
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgSubmitProposal
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(t, err)
	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Proposer)
	require.NotEqual(t, len(msg.InitialDeposit), 0)
	require.Equal(t, "47841094stake", msg.InitialDeposit[0].String())
	require.Equal(t, simulation.TypeMsgSubmitProposal, sdk.MsgTypeURL(&msg))
}

// TestSimulateMsgSubmitProposal tests the normal scenario of a valid message of type TypeMsgSubmitProposal.
// Abnormal scenarios, where errors occur, are not tested here.
func TestSimulateMsgSubmitLegacyProposal(t *testing.T) {
	suite, ctx := createTestSuite(t, false)
	app := suite.App

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, suite.AccountKeeper, suite.BankKeeper, suite.StakingKeeper, ctx, 3)

	// execute operation
	op := simulation.SimulateMsgSubmitLegacyProposal(suite.TxConfig, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper, MockWeightedProposals{3}.ContentSimulatorFn())
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgSubmitProposal
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(t, err)
	var msgLegacyContent v1.MsgExecLegacyContent
	err = proto.Unmarshal(msg.Messages[0].Value, &msgLegacyContent)
	require.NoError(t, err)
	var textProposal v1beta1.TextProposal
	err = proto.Unmarshal(msgLegacyContent.Content.Value, &textProposal)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", msg.Proposer)
	require.NotEqual(t, len(msg.InitialDeposit), 0)
	require.Equal(t, "25166256stake", msg.InitialDeposit[0].String())
	require.Equal(t, "title-3: ZBSpYuLyYggwexjxusrBqDOTtGTOWeLrQKjLxzIivHSlcxgdXhhuTSkuxKGLwQvuyNhYFmBZHeAerqyNEUzXPFGkqEGqiQWIXnku",
		textProposal.GetTitle())
	require.Equal(t, "description-3: NJWzHdBNpAXKJPHWQdrGYcAHSctgVlqwqHoLfHsXUdStwfefwzqLuKEhmMyYLdbZrcPgYqjNHxPexsruwEGStAneKbWkQDDIlCWBLSiAASNhZqNFlPtfqPJoxKsgMdzjWqLWdqKQuJqWPMvwPQWZUtVMOTMYKJbfdlZsjdsomuScvDmbDkgRualsxDvRJuCAmPOXitIbcyWsKGSdrEunFAOdmXnsuyFVgJqEjbklvmwrUlsxjRSfKZxGcpayDdgoFcnVSutxjRgOSFzPwidAjubMncNweqpbxhXGchpZUxuFDOtpnhNUycJICRYqsPhPSCjPTWZFLkstHWJxvdPEAyEIxXgLwbNOjrgzmaujiBABBIXvcXpLrbcEWNNQsbjvgJFgJkflpRohHUutvnaUqoopuKjTDaemDeSdqbnOzcfJpcTuAQtZoiLZOoAIlboFDAeGmSNwkvObPRvRWQgWkGkxwtPauYgdkmypLjbqhlHJIQTntgWjXwZdOyYEdQRRLfMSdnxqppqUofqLbLQDUjwKVKfZJUJQPsWIPwIVaSTrmKskoAhvmZyJgeRpkaTfGgrJzAigcxtfshmiDCFkuiluqtMOkidknnTBtumyJYlIsWLnCQclqdVmikUoMOPdPWwYbJxXyqUVicNxFxyqJTenNblyyKSdlCbiXxUiYUiMwXZASYfvMDPFgxniSjWaZTjHkqlJvtBsXqwPpyVxnJVGFWhfSxgOcduoxkiopJvFjMmFabrGYeVtTXLhxVUEiGwYUvndjFGzDVntUvibiyZhfMQdMhgsiuysLMiePBNXifRLMsSmXPkwlPloUbJveCvUlaalhZHuvdkCnkSHbMbmOnrfEGPwQiACiPlnihiaOdbjPqPiTXaHDoJXjSlZmltGqNHHNrcKdlFSCdmVOuvDcBLdSklyGJmcLTbSFtALdGlPkqqecJrpLCXNPWefoTJNgEJlyMEPneVaxxduAAEqQpHWZodWyRkDAxzyMnFMcjSVqeRXLqsNyNtQBbuRvunZflWSbbvXXdkyLikYqutQhLPONXbvhcQZJPSWnOulqQaXmbfFxAkqfYeseSHOQidHwbcsOaMnSrrmGjjRmEMQNuknupMxJiIeVjmgZvbmjPIQTEhQFULQLBMPrxcFPvBinaOPYWGvYGRKxLZdwamfRQQFngcdSlvwjfaPbURasIsGJVHtcEAxnIIrhSriiXLOlbEBLXFElXJFGxHJczRBIxAuPKtBisjKBwfzZFagdNmjdwIRvwzLkFKWRTDPxJCmpzHUcrPiiXXHnOIlqNVoGSXZewdnCRhuxeYGPVTfrNTQNOxZmxInOazUYNTNDgzsxlgiVEHPKMfbesvPHUqpNkUqbzeuzfdrsuLDpKHMUbBMKczKKWOdYoIXoPYtEjfOnlQLoGnbQUCuERdEFaptwnsHzTJDsuZkKtzMpFaZobynZdzNydEeJJHDYaQcwUxcqvwfWwNUsCiLvkZQiSfzAHftYgAmVsXgtmcYgTqJIawstRYJrZdSxlfRiqTufgEQVambeZZmaAyRQbcmdjVUZZCgqDrSeltJGXPMgZnGDZqISrGDOClxXCxMjmKqEPwKHoOfOeyGmqWqihqjINXLqnyTesZePQRqaWDQNqpLgNrAUKulklmckTijUltQKuWQDwpLmDyxLppPVMwsmBIpOwQttYFMjgJQZLYFPmxWFLIeZihkRNnkzoypBICIxgEuYsVWGIGRbbxqVasYnstWomJnHwmtOhAFSpttRYYzBmyEtZXiCthvKvWszTXDbiJbGXMcrYpKAgvUVFtdKUfvdMfhAryctklUCEdjetjuGNfJjajZtvzdYaqInKtFPPLYmRaXPdQzxdSQfmZDEVHlHGEGNSPRFJuIfKLLfUmnHxHnRjmzQPNlqrXgifUdzAGKVabYqvcDeYoTYgPsBUqehrBhmQUgTvDnsdpuhUoxskDdppTsYMcnDIPSwKIqhXDCIxOuXrywahvVavvHkPuaenjLmEbMgrkrQLHEAwrhHkPRNvonNQKqprqOFVZKAtpRSpvQUxMoXCMZLSSbnLEFsjVfANdQNQVwTmGxqVjVqRuxREAhuaDrFgEZpYKhwWPEKBevBfsOIcaZKyykQafzmGPLRAKDtTcJxJVgiiuUkmyMYuDUNEUhBEdoBLJnamtLmMJQgmLiUELIhLpiEvpOXOvXCPUeldLFqkKOwfacqIaRcnnZvERKRMCKUkMABbDHytQqQblrvoxOZkwzosQfDKGtIdfcXRJNqlBNwOCWoQBcEWyqrMlYZIAXYJmLfnjoJepgSFvrgajaBAIksoyeHqgqbGvpAstMIGmIhRYGGNPRIfOQKsGoKgxtsidhTaAePRCBFqZgPDWCIkqOJezGVkjfYUCZTlInbxBXwUAVRsxHTQtJFnnpmMvXDYCVlEmnZBKhmmxQOIQzxFWpJQkQoSAYzTEiDWEOsVLNrbfzeHFRyeYATakQQWmFDLPbVMCJcWjFGJjfqCoVzlbNNEsqxdSmNPjTjHYOkuEMFLkXYGaoJlraLqayMeCsTjWNRDPBywBJLAPVkGQqTwApVVwYAetlwSbzsdHWsTwSIcctkyKDuRWYDQikRqsKTMJchrliONJeaZIzwPQrNbTwxsGdwuduvibtYndRwpdsvyCktRHFalvUuEKMqXbItfGcNGWsGzubdPMYayOUOINjpcFBeESdwpdlTYmrPsLsVDhpTzoMegKrytNVZkfJRPuDCUXxSlSthOohmsuxmIZUedzxKmowKOdXTMcEtdpHaPWgIsIjrViKrQOCONlSuazmLuCUjLltOGXeNgJKedTVrrVCpWYWHyVrdXpKgNaMJVjbXxnVMSChdWKuZdqpisvrkBJPoURDYxWOtpjzZoOpWzyUuYNhCzRoHsMjmmWDcXzQiHIyjwdhPNwiPqFxeUfMVFQGImhykFgMIlQEoZCaRoqSBXTSWAeDumdbsOGtATwEdZlLfoBKiTvodQBGOEcuATWXfiinSjPmJKcWgQrTVYVrwlyMWhxqNbCMpIQNoSMGTiWfPTCezUjYcdWppnsYJihLQCqbNLRGgqrwHuIvsazapTpoPZIyZyeeSueJuTIhpHMEJfJpScshJubJGfkusuVBgfTWQoywSSliQQSfbvaHKiLnyjdSbpMkdBgXepoSsHnCQaYuHQqZsoEOmJCiuQUpJkmfyfbIShzlZpHFmLCsbknEAkKXKfRTRnuwdBeuOGgFbJLbDksHVapaRayWzwoYBEpmrlAxrUxYMUekKbpjPNfjUCjhbdMAnJmYQVZBQZkFVweHDAlaqJjRqoQPoOMLhyvYCzqEuQsAFoxWrzRnTVjStPadhsESlERnKhpEPsfDxNvxqcOyIulaCkmPdambLHvGhTZzysvqFauEgkFRItPfvisehFmoBhQqmkfbHVsgfHXDPJVyhwPllQpuYLRYvGodxKjkarnSNgsXoKEMlaSKxKdcVgvOkuLcfLFfdtXGTclqfPOfeoVLbqcjcXCUEBgAGplrkgsmIEhWRZLlGPGCwKWRaCKMkBHTAcypUrYjWwCLtOPVygMwMANGoQwFnCqFrUGMCRZUGJKTZIGPyldsifauoMnJPLTcDHmilcmahlqOELaAUYDBuzsVywnDQfwRLGIWozYaOAilMBcObErwgTDNGWnwQMUgFFSKtPDMEoEQCTKVREqrXZSGLqwTMcxHfWotDllNkIJPMbXzjDVjPOOjCFuIvTyhXKLyhUScOXvYthRXpPfKwMhptXaxIxgqBoUqzrWbaoLTVpQoottZyPFfNOoMioXHRuFwMRYUiKvcWPkrayyTLOCFJlAyslDameIuqVAuxErqFPEWIScKpBORIuZqoXlZuTvAjEdlEWDODFRregDTqGNoFBIHxvimmIZwLfFyKUfEWAnNBdtdzDmTPXtpHRGdIbuucfTjOygZsTxPjfweXhSUkMhPjMaxKlMIJMOXcnQfyzeOcbWwNbeH",
		textProposal.GetDescription())
	require.Equal(t, simulation.TypeMsgSubmitProposal, sdk.MsgTypeURL(&msg))
}

// TestSimulateMsgCancelProposal tests the normal scenario of a valid message of type TypeMsgCancelProposal.
// Abnormal scenarios, where errors occur, are not tested here.
func TestSimulateMsgCancelProposal(t *testing.T) {
	suite, ctx := createTestSuite(t, false)
	app := suite.App
	blockTime := time.Now().UTC()
	ctx = ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, suite.AccountKeeper, suite.BankKeeper, suite.StakingKeeper, ctx, 3)
	// setup a proposal
	proposer := accounts[0].Address
	content := v1beta1.NewTextProposal("Test", "description")
	contentMsg, err := v1.NewLegacyContent(content, suite.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String())
	require.NoError(t, err)

	submitTime := ctx.HeaderInfo().Time
	params, _ := suite.GovKeeper.Params.Get(ctx)
	depositPeriod := params.MaxDepositPeriod

	proposal, err := v1.NewProposal([]sdk.Msg{contentMsg}, 1, submitTime, submitTime.Add(*depositPeriod), "", "title", "summary", proposer, v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	require.NoError(t, err)

	err = suite.GovKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// execute operation
	op := simulation.SimulateMsgCancelProposal(suite.TxConfig, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgCancelProposal
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(t, err)
	require.NoError(t, err)
	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, proposer.String(), msg.Proposer)
	require.Equal(t, simulation.TypeMsgCancelProposal, sdk.MsgTypeURL(&msg))
}

// TestSimulateMsgDeposit tests the normal scenario of a valid message of type TypeMsgDeposit.
// Abnormal scenarios, where errors occur, are not tested here.
func TestSimulateMsgDeposit(t *testing.T) {
	suite, ctx := createTestSuite(t, false)
	app := suite.App
	blockTime := time.Now().UTC()
	ctx = ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, suite.AccountKeeper, suite.BankKeeper, suite.StakingKeeper, ctx, 3)

	// setup a proposal
	content := v1beta1.NewTextProposal("Test", "description")
	contentMsg, err := v1.NewLegacyContent(content, suite.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String())
	require.NoError(t, err)

	submitTime := ctx.HeaderInfo().Time
	params, _ := suite.GovKeeper.Params.Get(ctx)
	depositPeriod := params.MaxDepositPeriod

	proposal, err := v1.NewProposal([]sdk.Msg{contentMsg}, 1, submitTime, submitTime.Add(*depositPeriod), "", "text proposal", "description", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	require.NoError(t, err)

	err = suite.GovKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// execute operation
	op := simulation.SimulateMsgDeposit(suite.TxConfig, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgDeposit
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(t, err)
	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Depositor)
	require.NotEqual(t, len(msg.Amount), 0)
	require.Equal(t, "560969stake", msg.Amount[0].String())
	require.Equal(t, simulation.TypeMsgDeposit, sdk.MsgTypeURL(&msg))
}

// TestSimulateMsgVote tests the normal scenario of a valid message of type TypeMsgVote.
// Abnormal scenarios, where errors occur, are not tested here.
func TestSimulateMsgVote(t *testing.T) {
	suite, ctx := createTestSuite(t, false)
	app := suite.App
	blockTime := time.Now().UTC()
	ctx = ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, suite.AccountKeeper, suite.BankKeeper, suite.StakingKeeper, ctx, 3)

	// setup a proposal
	govAcc := suite.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String()
	contentMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Test", "description"), govAcc)
	require.NoError(t, err)

	submitTime := ctx.HeaderInfo().Time
	params, _ := suite.GovKeeper.Params.Get(ctx)
	depositPeriod := params.MaxDepositPeriod

	proposal, err := v1.NewProposal([]sdk.Msg{contentMsg}, 1, submitTime, submitTime.Add(*depositPeriod), "", "text proposal", "description", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	require.NoError(t, err)

	err = suite.GovKeeper.ActivateVotingPeriod(ctx, proposal)
	require.NoError(t, err)

	// execute operation
	op := simulation.SimulateMsgVote(suite.TxConfig, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgVote
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(t, err)
	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Voter)
	require.Equal(t, v1.OptionYes, msg.Option)
	require.Equal(t, simulation.TypeMsgVote, sdk.MsgTypeURL(&msg))
}

// TestSimulateMsgVoteWeighted tests the normal scenario of a valid message of type TypeMsgVoteWeighted.
// Abnormal scenarios, where errors occur, are not tested here.
func TestSimulateMsgVoteWeighted(t *testing.T) {
	suite, ctx := createTestSuite(t, false)
	app := suite.App
	blockTime := time.Now().UTC()
	ctx = ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, suite.AccountKeeper, suite.BankKeeper, suite.StakingKeeper, ctx, 3)

	// setup a proposal
	govAcc := suite.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String()
	contentMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Test", "description"), govAcc)
	require.NoError(t, err)
	submitTime := ctx.HeaderInfo().Time
	params, _ := suite.GovKeeper.Params.Get(ctx)
	depositPeriod := params.MaxDepositPeriod

	proposal, err := v1.NewProposal([]sdk.Msg{contentMsg}, 1, submitTime, submitTime.Add(*depositPeriod), "", "text proposal", "test", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	require.NoError(t, err)

	err = suite.GovKeeper.ActivateVotingPeriod(ctx, proposal)
	require.NoError(t, err)

	// execute operation
	op := simulation.SimulateMsgVoteWeighted(suite.TxConfig, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgVoteWeighted
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(t, err)
	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Voter)
	require.True(t, len(msg.Options) >= 1)
	require.Equal(t, simulation.TypeMsgVoteWeighted, sdk.MsgTypeURL(&msg))
}

type suite struct {
	TxConfig      client.TxConfig
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	GovKeeper     *keeper.Keeper
	StakingKeeper *stakingkeeper.Keeper
	App           *runtime.App
}

// returns context and an app with updated mint keeper
func createTestSuite(t *testing.T, isCheckTx bool) (suite, sdk.Context) {
	t.Helper()
	res := suite{}

	app, err := simtestutil.Setup(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.TxModule(),
				configurator.BankModule(),
				configurator.StakingModule(),
				configurator.ConsensusModule(),
				configurator.GovModule(),
				configurator.ProtocolPoolModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&res.TxConfig, &res.AccountKeeper, &res.BankKeeper, &res.GovKeeper, &res.StakingKeeper)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(isCheckTx)

	res.App = app
	return res, ctx
}

func getTestingAccounts(
	t *testing.T, r *rand.Rand,
	accountKeeper authkeeper.AccountKeeper, bankKeeper bankkeeper.Keeper, stakingKeeper *stakingkeeper.Keeper,
	ctx sdk.Context, n int,
) []simtypes.Account {
	t.Helper()
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := stakingKeeper.TokensFromConsensusPower(ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := accountKeeper.NewAccountWithAddress(ctx, account.Address)
		accountKeeper.SetAccount(ctx, acc)
		require.NoError(t, testutil.FundAccount(ctx, bankKeeper, account.Address, initCoins))
	}

	return accounts
}
