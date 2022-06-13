package simulation_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type MockWeightedProposalContent struct {
	n int
}

func (m MockWeightedProposalContent) AppParamsKey() string {
	return fmt.Sprintf("AppParamsKey-%d", m.n)
}

func (m MockWeightedProposalContent) DefaultWeight() int {
	return m.n
}

func (m MockWeightedProposalContent) ContentSimulatorFn() simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) simtypes.Content {
		return v1beta1.NewTextProposal(
			fmt.Sprintf("title-%d: %s", m.n, simtypes.RandStringOfLength(r, 100)),
			fmt.Sprintf("description-%d: %s", m.n, simtypes.RandStringOfLength(r, 4000)),
		)
	}
}

// make sure the MockWeightedProposalContent satisfied the WeightedProposalContent interface
var _ simtypes.WeightedProposalContent = MockWeightedProposalContent{}

func mockWeightedProposalContent(n int) []simtypes.WeightedProposalContent {
	wpc := make([]simtypes.WeightedProposalContent, n)
	for i := 0; i < n; i++ {
		wpc[i] = MockWeightedProposalContent{i}
	}
	return wpc
}

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	app, ctx := createTestApp(t, false)
	ctx.WithChainID("test-chain")

	cdc := app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(appParams, cdc, app.AccountKeeper,
		app.BankKeeper, app.GovKeeper, mockWeightedProposalContent(3),
	)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, app, ctx, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{0, types.ModuleName, simulation.TypeMsgSubmitProposal},
		{1, types.ModuleName, simulation.TypeMsgSubmitProposal},
		{2, types.ModuleName, simulation.TypeMsgSubmitProposal},
		{simappparams.DefaultWeightMsgDeposit, types.ModuleName, simulation.TypeMsgDeposit},
		{simappparams.DefaultWeightMsgVote, types.ModuleName, simulation.TypeMsgVote},
		{simappparams.DefaultWeightMsgVoteWeighted, types.ModuleName, simulation.TypeMsgVoteWeighted},
	}

	for i, w := range weightesOps {
		operationMsg, _, _ := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
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
	app, ctx := createTestApp(t, false)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgSubmitProposal(app.AccountKeeper, app.BankKeeper, app.GovKeeper, MockWeightedProposalContent{3}.ContentSimulatorFn())
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgSubmitProposal
	err = v1.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", msg.Proposer)
	require.NotEqual(t, len(msg.InitialDeposit), 0)
	require.Equal(t, "2686011stake", msg.InitialDeposit[0].String())
	require.Equal(t, "title-3: ZBSpYuLyYggwexjxusrBqDOTtGTOWeLrQKjLxzIivHSlcxgdXhhuTSkuxKGLwQvuyNhYFmBZHeAerqyNEUzXPFGkqEGqiQWIXnku", msg.Messages[0].GetCachedValue().(*v1.MsgExecLegacyContent).Content.GetCachedValue().(v1beta1.Content).GetTitle())
	require.Equal(t, "description-3: NJWzHdBNpAXKJPHWQdrGYcAHSctgVlqwqHoLfHsXUdStwfefwzqLuKEhmMyYLdbZrcPgYqjNHxPexsruwEGStAneKbWkQDDIlCWBLSiAASNhZqNFlPtfqPJoxKsgMdzjWqLWdqKQuJqWPMvwPQWZUtVMOTMYKJbfdlZsjdsomuScvDmbDkgRualsxDvRJuCAmPOXitIbcyWsKGSdrEunFAOdmXnsuyFVgJqEjbklvmwrUlsxjRSfKZxGcpayDdgoFcnVSutxjRgOSFzPwidAjubMncNweqpbxhXGchpZUxuFDOtpnhNUycJICRYqsPhPSCjPTWZFLkstHWJxvdPEAyEIxXgLwbNOjrgzmaujiBABBIXvcXpLrbcEWNNQsbjvgJFgJkflpRohHUutvnaUqoopuKjTDaemDeSdqbnOzcfJpcTuAQtZoiLZOoAIlboFDAeGmSNwkvObPRvRWQgWkGkxwtPauYgdkmypLjbqhlHJIQTntgWjXwZdOyYEdQRRLfMSdnxqppqUofqLbLQDUjwKVKfZJUJQPsWIPwIVaSTrmKskoAhvmZyJgeRpkaTfGgrJzAigcxtfshmiDCFkuiluqtMOkidknnTBtumyJYlIsWLnCQclqdVmikUoMOPdPWwYbJxXyqUVicNxFxyqJTenNblyyKSdlCbiXxUiYUiMwXZASYfvMDPFgxniSjWaZTjHkqlJvtBsXqwPpyVxnJVGFWhfSxgOcduoxkiopJvFjMmFabrGYeVtTXLhxVUEiGwYUvndjFGzDVntUvibiyZhfMQdMhgsiuysLMiePBNXifRLMsSmXPkwlPloUbJveCvUlaalhZHuvdkCnkSHbMbmOnrfEGPwQiACiPlnihiaOdbjPqPiTXaHDoJXjSlZmltGqNHHNrcKdlFSCdmVOuvDcBLdSklyGJmcLTbSFtALdGlPkqqecJrpLCXNPWefoTJNgEJlyMEPneVaxxduAAEqQpHWZodWyRkDAxzyMnFMcjSVqeRXLqsNyNtQBbuRvunZflWSbbvXXdkyLikYqutQhLPONXbvhcQZJPSWnOulqQaXmbfFxAkqfYeseSHOQidHwbcsOaMnSrrmGjjRmEMQNuknupMxJiIeVjmgZvbmjPIQTEhQFULQLBMPrxcFPvBinaOPYWGvYGRKxLZdwamfRQQFngcdSlvwjfaPbURasIsGJVHtcEAxnIIrhSriiXLOlbEBLXFElXJFGxHJczRBIxAuPKtBisjKBwfzZFagdNmjdwIRvwzLkFKWRTDPxJCmpzHUcrPiiXXHnOIlqNVoGSXZewdnCRhuxeYGPVTfrNTQNOxZmxInOazUYNTNDgzsxlgiVEHPKMfbesvPHUqpNkUqbzeuzfdrsuLDpKHMUbBMKczKKWOdYoIXoPYtEjfOnlQLoGnbQUCuERdEFaptwnsHzTJDsuZkKtzMpFaZobynZdzNydEeJJHDYaQcwUxcqvwfWwNUsCiLvkZQiSfzAHftYgAmVsXgtmcYgTqJIawstRYJrZdSxlfRiqTufgEQVambeZZmaAyRQbcmdjVUZZCgqDrSeltJGXPMgZnGDZqISrGDOClxXCxMjmKqEPwKHoOfOeyGmqWqihqjINXLqnyTesZePQRqaWDQNqpLgNrAUKulklmckTijUltQKuWQDwpLmDyxLppPVMwsmBIpOwQttYFMjgJQZLYFPmxWFLIeZihkRNnkzoypBICIxgEuYsVWGIGRbbxqVasYnstWomJnHwmtOhAFSpttRYYzBmyEtZXiCthvKvWszTXDbiJbGXMcrYpKAgvUVFtdKUfvdMfhAryctklUCEdjetjuGNfJjajZtvzdYaqInKtFPPLYmRaXPdQzxdSQfmZDEVHlHGEGNSPRFJuIfKLLfUmnHxHnRjmzQPNlqrXgifUdzAGKVabYqvcDeYoTYgPsBUqehrBhmQUgTvDnsdpuhUoxskDdppTsYMcnDIPSwKIqhXDCIxOuXrywahvVavvHkPuaenjLmEbMgrkrQLHEAwrhHkPRNvonNQKqprqOFVZKAtpRSpvQUxMoXCMZLSSbnLEFsjVfANdQNQVwTmGxqVjVqRuxREAhuaDrFgEZpYKhwWPEKBevBfsOIcaZKyykQafzmGPLRAKDtTcJxJVgiiuUkmyMYuDUNEUhBEdoBLJnamtLmMJQgmLiUELIhLpiEvpOXOvXCPUeldLFqkKOwfacqIaRcnnZvERKRMCKUkMABbDHytQqQblrvoxOZkwzosQfDKGtIdfcXRJNqlBNwOCWoQBcEWyqrMlYZIAXYJmLfnjoJepgSFvrgajaBAIksoyeHqgqbGvpAstMIGmIhRYGGNPRIfOQKsGoKgxtsidhTaAePRCBFqZgPDWCIkqOJezGVkjfYUCZTlInbxBXwUAVRsxHTQtJFnnpmMvXDYCVlEmnZBKhmmxQOIQzxFWpJQkQoSAYzTEiDWEOsVLNrbfzeHFRyeYATakQQWmFDLPbVMCJcWjFGJjfqCoVzlbNNEsqxdSmNPjTjHYOkuEMFLkXYGaoJlraLqayMeCsTjWNRDPBywBJLAPVkGQqTwApVVwYAetlwSbzsdHWsTwSIcctkyKDuRWYDQikRqsKTMJchrliONJeaZIzwPQrNbTwxsGdwuduvibtYndRwpdsvyCktRHFalvUuEKMqXbItfGcNGWsGzubdPMYayOUOINjpcFBeESdwpdlTYmrPsLsVDhpTzoMegKrytNVZkfJRPuDCUXxSlSthOohmsuxmIZUedzxKmowKOdXTMcEtdpHaPWgIsIjrViKrQOCONlSuazmLuCUjLltOGXeNgJKedTVrrVCpWYWHyVrdXpKgNaMJVjbXxnVMSChdWKuZdqpisvrkBJPoURDYxWOtpjzZoOpWzyUuYNhCzRoHsMjmmWDcXzQiHIyjwdhPNwiPqFxeUfMVFQGImhykFgMIlQEoZCaRoqSBXTSWAeDumdbsOGtATwEdZlLfoBKiTvodQBGOEcuATWXfiinSjPmJKcWgQrTVYVrwlyMWhxqNbCMpIQNoSMGTiWfPTCezUjYcdWppnsYJihLQCqbNLRGgqrwHuIvsazapTpoPZIyZyeeSueJuTIhpHMEJfJpScshJubJGfkusuVBgfTWQoywSSliQQSfbvaHKiLnyjdSbpMkdBgXepoSsHnCQaYuHQqZsoEOmJCiuQUpJkmfyfbIShzlZpHFmLCsbknEAkKXKfRTRnuwdBeuOGgFbJLbDksHVapaRayWzwoYBEpmrlAxrUxYMUekKbpjPNfjUCjhbdMAnJmYQVZBQZkFVweHDAlaqJjRqoQPoOMLhyvYCzqEuQsAFoxWrzRnTVjStPadhsESlERnKhpEPsfDxNvxqcOyIulaCkmPdambLHvGhTZzysvqFauEgkFRItPfvisehFmoBhQqmkfbHVsgfHXDPJVyhwPllQpuYLRYvGodxKjkarnSNgsXoKEMlaSKxKdcVgvOkuLcfLFfdtXGTclqfPOfeoVLbqcjcXCUEBgAGplrkgsmIEhWRZLlGPGCwKWRaCKMkBHTAcypUrYjWwCLtOPVygMwMANGoQwFnCqFrUGMCRZUGJKTZIGPyldsifauoMnJPLTcDHmilcmahlqOELaAUYDBuzsVywnDQfwRLGIWozYaOAilMBcObErwgTDNGWnwQMUgFFSKtPDMEoEQCTKVREqrXZSGLqwTMcxHfWotDllNkIJPMbXzjDVjPOOjCFuIvTyhXKLyhUScOXvYthRXpPfKwMhptXaxIxgqBoUqzrWbaoLTVpQoottZyPFfNOoMioXHRuFwMRYUiKvcWPkrayyTLOCFJlAyslDameIuqVAuxErqFPEWIScKpBORIuZqoXlZuTvAjEdlEWDODFRregDTqGNoFBIHxvimmIZwLfFyKUfEWAnNBdtdzDmTPXtpHRGdIbuucfTjOygZsTxPjfweXhSUkMhPjMaxKlMIJMOXcnQfyzeOcbWwNbeH", msg.Messages[0].GetCachedValue().(*v1.MsgExecLegacyContent).Content.GetCachedValue().(v1beta1.Content).GetDescription())
	require.Equal(t, "gov", msg.Route())
	require.Equal(t, simulation.TypeMsgSubmitProposal, msg.Type())
}

// TestSimulateMsgDeposit tests the normal scenario of a valid message of type TypeMsgDeposit.
// Abnormal scenarios, where errors occur, are not tested here.
func TestSimulateMsgDeposit(t *testing.T) {
	app, ctx := createTestApp(t, false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// setup a proposal
	content := v1beta1.NewTextProposal("Test", "description")
	contentMsg, err := v1.NewLegacyContent(content, app.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String())
	require.NoError(t, err)

	submitTime := ctx.BlockHeader().Time
	depositPeriod := app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod

	proposal, err := v1.NewProposal([]sdk.Msg{contentMsg}, 1, "", submitTime, submitTime.Add(*depositPeriod))
	require.NoError(t, err)

	app.GovKeeper.SetProposal(ctx, proposal)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgDeposit(app.AccountKeeper, app.BankKeeper, app.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgDeposit
	err = v1.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Depositor)
	require.NotEqual(t, len(msg.Amount), 0)
	require.Equal(t, "560969stake", msg.Amount[0].String())
	require.Equal(t, "gov", msg.Route())
	require.Equal(t, simulation.TypeMsgDeposit, msg.Type())
}

// TestSimulateMsgVote tests the normal scenario of a valid message of type TypeMsgVote.
// Abnormal scenarios, where errors occur, are not tested here.
func TestSimulateMsgVote(t *testing.T) {
	app, ctx := createTestApp(t, false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// setup a proposal
	govAcc := app.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String()
	contentMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Test", "description"), govAcc)
	require.NoError(t, err)

	submitTime := ctx.BlockHeader().Time
	depositPeriod := app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod

	proposal, err := v1.NewProposal([]sdk.Msg{contentMsg}, 1, "", submitTime, submitTime.Add(*depositPeriod))
	require.NoError(t, err)

	app.GovKeeper.ActivateVotingPeriod(ctx, proposal)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgVote(app.AccountKeeper, app.BankKeeper, app.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgVote
	v1.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Voter)
	require.Equal(t, v1.OptionYes, msg.Option)
	require.Equal(t, "gov", msg.Route())
	require.Equal(t, simulation.TypeMsgVote, msg.Type())
}

// TestSimulateMsgVoteWeighted tests the normal scenario of a valid message of type TypeMsgVoteWeighted.
// Abnormal scenarios, where errors occur, are not tested here.
func TestSimulateMsgVoteWeighted(t *testing.T) {
	app, ctx := createTestApp(t, false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// setup a proposal
	govAcc := app.GovKeeper.GetGovernanceAccount(ctx).GetAddress().String()
	contentMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Test", "description"), govAcc)
	require.NoError(t, err)
	submitTime := ctx.BlockHeader().Time
	depositPeriod := app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod

	proposal, err := v1.NewProposal([]sdk.Msg{contentMsg}, 1, "", submitTime, submitTime.Add(*depositPeriod))
	require.NoError(t, err)

	app.GovKeeper.ActivateVotingPeriod(ctx, proposal)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgVoteWeighted(app.AccountKeeper, app.BankKeeper, app.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg v1.MsgVoteWeighted
	v1.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Voter)
	require.True(t, len(msg.Options) >= 1)
	require.Equal(t, "gov", msg.Route())
	require.Equal(t, simulation.TypeMsgVoteWeighted, msg.Type())
}

// returns context and an app with updated mint keeper
func createTestApp(t *testing.T, isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(t, isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
	app.MintKeeper.SetMinter(ctx, minttypes.DefaultInitialMinter())

	return app, ctx
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *simapp.SimApp, ctx sdk.Context, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, account.Address, initCoins))
	}

	return accounts
}
