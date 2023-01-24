package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	OpWeightSanction            = "op_weight_sanction"             //nolint:gosec
	OpWeightSanctionImmediate   = "op_weight_sanction_immediate"   //nolint:gosec
	OpWeightUnsanction          = "op_weight_unsanction"           //nolint:gosec
	OpWeightUnsanctionImmediate = "op_weight_unsanction_immediate" //nolint:gosec
	OpWeightUpdateParams        = "op_weight_update_params"        //nolint:gosec

	DefaultWeightSanction            = 10
	DefaultWeightSanctionImmediate   = 10
	DefaultWeightUnsanction          = 10
	DefaultWeightUnsanctionImmediate = 10
	DefaultWeightUpdateParams        = 10
)

// WeightedOpsArgs holds all the args provided to WeightedOperations so that they can be passed on later more easily.
type WeightedOpsArgs struct {
	AppParams  simtypes.AppParams
	JSONCodec  codec.JSONCodec
	ProtoCodec *codec.ProtoCodec
	AK         sanction.AccountKeeper
	BK         sanction.BankKeeper
	GK         sanction.GovKeeper
	SK         *keeper.Keeper
}

// SendGovMsgArgs holds all the args available and needed for sending a gov msg.
type SendGovMsgArgs struct {
	WeightedOpsArgs

	R       *rand.Rand
	App     *baseapp.BaseApp
	Ctx     sdk.Context
	Accs    []simtypes.Account
	ChainID string

	Sender  simtypes.Account
	Msg     sdk.Msg
	Deposit sdk.Coins
	Comment string
}

func WeightedOperations(
	appParams simtypes.AppParams, jsonCodec codec.JSONCodec, protoCodec *codec.ProtoCodec,
	ak sanction.AccountKeeper, bk sanction.BankKeeper, gk sanction.GovKeeper, sk keeper.Keeper,
) simulation.WeightedOperations {
	args := &WeightedOpsArgs{
		AppParams:  appParams,
		JSONCodec:  jsonCodec,
		ProtoCodec: protoCodec,
		AK:         ak,
		BK:         bk,
		GK:         gk,
		SK:         &sk,
	}

	var (
		weightSanction            int
		weightSanctionImmediate   int
		weightUnsanction          int
		weightUnsanctionImmediate int
		weightUpdateParams        int
	)

	appParams.GetOrGenerate(jsonCodec, OpWeightSanction, &weightSanction, nil,
		func(_ *rand.Rand) { weightSanction = DefaultWeightSanction })
	appParams.GetOrGenerate(jsonCodec, OpWeightSanctionImmediate, &weightSanctionImmediate, nil,
		func(_ *rand.Rand) { weightSanctionImmediate = DefaultWeightSanctionImmediate })
	appParams.GetOrGenerate(jsonCodec, OpWeightUnsanction, &weightUnsanction, nil,
		func(_ *rand.Rand) { weightUnsanction = DefaultWeightUnsanction })
	appParams.GetOrGenerate(jsonCodec, OpWeightUnsanctionImmediate, &weightUnsanctionImmediate, nil,
		func(_ *rand.Rand) { weightUnsanctionImmediate = DefaultWeightUnsanctionImmediate })
	appParams.GetOrGenerate(jsonCodec, OpWeightUpdateParams, &weightUpdateParams, nil,
		func(_ *rand.Rand) { weightUpdateParams = DefaultWeightUpdateParams })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(weightSanction, SimulateGovMsgSanction(args)),
		simulation.NewWeightedOperation(weightSanctionImmediate, SimulateGovMsgSanctionImmediate(args)),
		simulation.NewWeightedOperation(weightUnsanction, SimulateGovMsgUnsanction(args)),
		simulation.NewWeightedOperation(weightUnsanctionImmediate, SimulateGovMsgUnsanctionImmediate(args)),
		simulation.NewWeightedOperation(weightUpdateParams, SimulateGovMsgUpdateParams(args)),
	}
}

// SendGovMsg sends a msg as a gov prop.
// It returns whether to skip the rest, an operation message, and any error encountered.
func SendGovMsg(args *SendGovMsgArgs) (bool, simtypes.OperationMsg, error) {
	msgType := sdk.MsgTypeURL(args.Msg)

	spendableCoins := args.BK.SpendableCoins(args.Ctx, args.Sender.Address)
	if spendableCoins.Empty() {
		return true, simtypes.NoOpMsg(sanction.ModuleName, msgType, "sender has no spendable coins"), nil
	}

	_, hasNeg := spendableCoins.SafeSub(args.Deposit...)
	if hasNeg {
		return true, simtypes.NoOpMsg(sanction.ModuleName, msgType, "sender has insufficient balance to cover deposit"), nil
	}

	msgAny, err := codectypes.NewAnyWithValue(args.Msg)
	if err != nil {
		return true, simtypes.NoOpMsg(sanction.ModuleName, msgType, "wrapping MsgSanction as Any"), err
	}

	govMsg := &govv1.MsgSubmitProposal{
		Messages:       []*codectypes.Any{msgAny},
		InitialDeposit: args.Deposit,
		Proposer:       args.Sender.Address.String(),
		Metadata:       "",
	}

	txCtx := simulation.OperationInput{
		R:               args.R,
		App:             args.App,
		TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
		Cdc:             args.ProtoCodec,
		Msg:             govMsg,
		MsgType:         govMsg.Type(),
		CoinsSpentInMsg: govMsg.InitialDeposit,
		Context:         args.Ctx,
		SimAccount:      args.Sender,
		AccountKeeper:   args.AK,
		Bankkeeper:      args.BK,
		ModuleName:      sanction.ModuleName,
	}

	opMsg, _, err := simulation.GenAndDeliverTxWithRandFees(txCtx)
	if opMsg.Comment == "" {
		opMsg.Comment = args.Comment
	}

	return err != nil, opMsg, err
}

// OperationMsgVote returns an operation that casts a yes vote on a gov prop from an account.
func OperationMsgVote(args *WeightedOpsArgs, voter simtypes.Account, govPropID uint64, vote govv1.VoteOption, comment string) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := govv1.NewMsgVote(voter.Address, govPropID, vote, "")

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             args.ProtoCodec,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: sdk.Coins{},
			Context:         ctx,
			SimAccount:      voter,
			AccountKeeper:   args.AK,
			Bankkeeper:      args.BK,
			ModuleName:      sanction.ModuleName,
		}

		opMsg, fops, err := simulation.GenAndDeliverTxWithRandFees(txCtx)
		if opMsg.Comment == "" {
			opMsg.Comment = comment
		}

		return opMsg, fops, err
	}
}

// MaxCoins combines a and b taking the max of each denom.
// The result will have all the denoms from a and all the denoms from b.
// The amount of each denom is the max between a and b for that denom.
func MaxCoins(a, b sdk.Coins) sdk.Coins {
	allDenomsMap := map[string]bool{}
	for _, c := range a {
		allDenomsMap[c.Denom] = true
	}
	for _, c := range b {
		allDenomsMap[c.Denom] = true
	}
	rv := make([]sdk.Coin, 0, len(allDenomsMap))
	for denom := range allDenomsMap {
		cA := a.AmountOf(denom)
		cB := b.AmountOf(denom)
		if cA.GT(cB) {
			rv = append(rv, sdk.NewCoin(denom, cA))
		} else {
			rv = append(rv, sdk.NewCoin(denom, cB))
		}
	}
	return sdk.NewCoins(rv...)
}

func SimulateGovMsgSanction(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &sanction.MsgSanction{
			Authority: args.SK.GetAuthority(),
		}
		msgType := sdk.MsgTypeURL(msg)

		// First, get the governance min deposit needed and immediate sanction min deposit needed.
		govMinDep := sdk.NewCoins(args.GK.GetDepositParams(ctx).MinDeposit...)
		imMinDep := args.SK.GetImmediateSanctionMinDeposit(ctx)
		if !imMinDep.IsZero() && govMinDep.IsAllGTE(imMinDep) {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "cannot sanction without it being immediate"), nil, nil
		}

		// Create 1-10 new accounts to sanction.
		// Sanctioning known accounts breaks other sim ops.
		for _, acct := range simtypes.RandomAccounts(r, r.Intn(10)+1) {
			msg.Addresses = append(msg.Addresses, acct.Address.String())
		}

		sender, _ := simtypes.RandomAcc(r, accs)

		msgArgs := &SendGovMsgArgs{
			WeightedOpsArgs: *args,
			R:               r,
			App:             app,
			Ctx:             ctx,
			Accs:            accs,
			ChainID:         chainID,
			Sender:          sender,
			Msg:             msg,
			Deposit:         govMinDep,
			Comment:         "sanction",
		}

		skip, opMsg, err := SendGovMsg(msgArgs)

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.GK.GetProposalID(ctx)
		if err != nil {
			return opMsg, nil, err
		}

		votingPeriod := args.GK.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        OperationMsgVote(args, acct, proposalID, govv1.OptionYes, msgArgs.Comment),
			}
		}

		return opMsg, fops, nil
	}
}

func SimulateGovMsgSanctionImmediate(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &sanction.MsgSanction{
			Authority: args.SK.GetAuthority(),
		}
		msgType := sdk.MsgTypeURL(msg)

		// Get the governance and immediate sanction min deposits and make sure immediate is possible.
		govMinDep := sdk.NewCoins(args.GK.GetDepositParams(ctx).MinDeposit...)
		imMinDep := args.SK.GetImmediateSanctionMinDeposit(ctx)
		if imMinDep.IsZero() {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "immediate sanction min deposit is zero"), nil, nil
		}

		// The deposit needs to be >= both the gov min dep and im min dep.
		deposit := MaxCoins(imMinDep, govMinDep)

		// Decide early whether we're going to vote yes or no on this.
		// By doing it early, we use r before anything else, which makes testing easier.
		vote := govv1.OptionYes
		if r.Intn(2) == 0 {
			vote = govv1.OptionNo
		}

		// Create 1-10 new accounts to sanction.
		// Sanctioning known accounts breaks other sim ops.
		for _, acct := range simtypes.RandomAccounts(r, r.Intn(10)+1) {
			msg.Addresses = append(msg.Addresses, acct.Address.String())
		}

		sender, _ := simtypes.RandomAcc(r, accs)

		msgArgs := &SendGovMsgArgs{
			WeightedOpsArgs: *args,
			R:               r,
			App:             app,
			Ctx:             ctx,
			Accs:            accs,
			ChainID:         chainID,
			Sender:          sender,
			Msg:             msg,
			Deposit:         deposit,
			Comment:         "immediate sanction",
		}

		skip, opMsg, err := SendGovMsg(msgArgs)

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.GK.GetProposalID(ctx)
		if err != nil {
			return opMsg, nil, err
		}

		votingPeriod := args.GK.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        OperationMsgVote(args, acct, proposalID, vote, msgArgs.Comment),
			}
		}

		return opMsg, fops, nil
	}
}

func SimulateGovMsgUnsanction(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &sanction.MsgUnsanction{
			Authority: args.SK.GetAuthority(),
		}
		msgType := sdk.MsgTypeURL(msg)

		sanctionedAddrs := args.SK.GetAllSanctionedAddresses(ctx)
		if len(sanctionedAddrs) == 0 {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "no addresses are sanctioned"), nil, nil
		}

		// Get the governance min deposit needed and immediate sanction min deposit needed.
		govMinDep := sdk.NewCoins(args.GK.GetDepositParams(ctx).MinDeposit...)
		imMinDep := args.SK.GetImmediateUnsanctionMinDeposit(ctx)
		if !imMinDep.IsZero() && govMinDep.IsAllGTE(imMinDep) {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "cannot unsanction without it being immediate"), nil, nil
		}

		// Unsanction 1/4 of the sanctioned addresses but at least 4.
		// If there are fewer than 4 sanctioned addresses, unsanction them all.
		count := len(sanctionedAddrs) / 4
		if count < 4 {
			count = 4
		}
		if count > len(sanctionedAddrs) {
			count = len(sanctionedAddrs)
		} else {
			r.Shuffle(count, func(i, j int) {
				sanctionedAddrs[i], sanctionedAddrs[j] = sanctionedAddrs[j], sanctionedAddrs[i]
			})
		}
		msg.Addresses = sanctionedAddrs[:count]

		sender, _ := simtypes.RandomAcc(r, accs)

		msgArgs := &SendGovMsgArgs{
			WeightedOpsArgs: *args,
			R:               r,
			App:             app,
			Ctx:             ctx,
			Accs:            accs,
			ChainID:         chainID,
			Sender:          sender,
			Msg:             msg,
			Deposit:         govMinDep,
			Comment:         "unsanction",
		}

		skip, opMsg, err := SendGovMsg(msgArgs)

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.GK.GetProposalID(ctx)
		if err != nil {
			return opMsg, nil, err
		}

		votingPeriod := args.GK.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        OperationMsgVote(args, acct, proposalID, govv1.OptionYes, msgArgs.Comment),
			}
		}

		return opMsg, fops, nil
	}
}

func SimulateGovMsgUnsanctionImmediate(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &sanction.MsgUnsanction{
			Authority: args.SK.GetAuthority(),
		}
		msgType := sdk.MsgTypeURL(msg)

		sanctionedAddrs := args.SK.GetAllSanctionedAddresses(ctx)
		if len(sanctionedAddrs) == 0 {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "no addresses are sanctioned"), nil, nil
		}

		// Get the governance and immediate sanction min deposits and make sure immediate is possible.
		govMinDep := sdk.NewCoins(args.GK.GetDepositParams(ctx).MinDeposit...)
		imMinDep := args.SK.GetImmediateUnsanctionMinDeposit(ctx)
		if imMinDep.IsZero() {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "immediate unsanction min deposit is zero"), nil, nil
		}

		// The deposit needs to be >= both the gov min dep and im min dep.
		deposit := MaxCoins(imMinDep, govMinDep)

		// Decide early whether we're going to vote yes or no on this.
		// By doing it early, we use r before anything else, which makes testing easier.
		vote := govv1.OptionYes
		if r.Intn(2) == 0 {
			vote = govv1.OptionNo
		}

		// Unsanction 1/4 of the sanctioned addresses but at least 4.
		// If there are fewer than 4 sanctioned addresses, unsanction them all.
		count := len(sanctionedAddrs) / 4
		if count < 4 {
			count = 4
		}
		if count > len(sanctionedAddrs) {
			count = len(sanctionedAddrs)
		} else {
			r.Shuffle(count, func(i, j int) {
				sanctionedAddrs[i], sanctionedAddrs[j] = sanctionedAddrs[j], sanctionedAddrs[i]
			})
		}
		msg.Addresses = sanctionedAddrs[:count]

		sender, _ := simtypes.RandomAcc(r, accs)

		msgArgs := &SendGovMsgArgs{
			WeightedOpsArgs: *args,
			R:               r,
			App:             app,
			Ctx:             ctx,
			Accs:            accs,
			ChainID:         chainID,
			Sender:          sender,
			Msg:             msg,
			Deposit:         deposit,
			Comment:         "immediate unsanction",
		}

		skip, opMsg, err := SendGovMsg(msgArgs)

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.GK.GetProposalID(ctx)
		if err != nil {
			return opMsg, nil, err
		}

		votingPeriod := args.GK.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        OperationMsgVote(args, acct, proposalID, vote, msgArgs.Comment),
			}
		}

		return opMsg, fops, nil
	}
}

func SimulateGovMsgUpdateParams(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get the governance min deposit needed.
		govMinDep := sdk.NewCoins(args.GK.GetDepositParams(ctx).MinDeposit...)

		// Pick the random params first, so R isn't used for anything else before,
		// which makes testing easier.
		msg := &sanction.MsgUpdateParams{
			Params:    RandomParams(r),
			Authority: args.SK.GetAuthority(),
		}

		sender, _ := simtypes.RandomAcc(r, accs)

		msgArgs := &SendGovMsgArgs{
			WeightedOpsArgs: *args,
			R:               r,
			App:             app,
			Ctx:             ctx,
			Accs:            accs,
			ChainID:         chainID,
			Sender:          sender,
			Msg:             msg,
			Deposit:         govMinDep,
			Comment:         "update params",
		}

		skip, opMsg, err := SendGovMsg(msgArgs)

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.GK.GetProposalID(ctx)
		if err != nil {
			return opMsg, nil, err
		}

		votingPeriod := args.GK.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        OperationMsgVote(args, acct, proposalID, govv1.OptionYes, msgArgs.Comment),
			}
		}

		return opMsg, fops, nil
	}
}
