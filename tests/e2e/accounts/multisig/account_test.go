package multisig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	multisigaccount "cosmossdk.io/x/accounts/defaults/multisig"
	v1 "cosmossdk.io/x/accounts/defaults/multisig/v1"
	accountsv1 "cosmossdk.io/x/accounts/v1"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, NewE2ETestSuite())
}

// TestFullMultisig creates a multisig account with 1 member, sends a tx, votes and executes it
// then adds 2 more members and changes the config to require 2/3 majority (also through a proposal).
// Finally it creates a proposal that won't pass.
func (s *E2ETestSuite) TestFullMultisig() {
	t := s.T()
	app := setupApp(t)
	currentTime := time.Now()
	ctx := sdk.NewContext(app.CommitMultiStore(), false, app.Logger()).WithHeaderInfo(header.Info{
		Time: currentTime,
	})
	member0, err := app.AuthKeeper.AddressCodec().BytesToString(members[0])
	require.NoError(t, err)
	member1, err := app.AuthKeeper.AddressCodec().BytesToString(members[1])
	require.NoError(t, err)
	member2, err := app.AuthKeeper.AddressCodec().BytesToString(members[2])
	require.NoError(t, err)

	s.fundAccount(app, ctx, members[0], sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000000))})
	randAcc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	_, accountAddr, err := app.AccountsKeeper.Init(ctx, multisigaccount.MULTISIG_ACCOUNT, members[0], &v1.MsgInit{
		Members: []*v1.Member{{Address: member0, Weight: 100}},
		Config: &v1.Config{
			Threshold:      100,
			Quorum:         100,
			VotingPeriod:   120,
			Revote:         false,
			EarlyExecution: true,
		},
	}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))})
	require.NoError(t, err)

	addr, err := app.AuthKeeper.AddressCodec().BytesToString(randAcc)
	require.NoError(t, err)

	accAddrStr := sdk.AccAddress(accountAddr).String()

	balance := app.BankKeeper.GetBalance(ctx, randAcc, "stake")
	require.True(t, balance.Amount.Equal(math.NewInt(0)))

	// do a simple bank send
	msg := &bankv1beta1.MsgSend{
		FromAddress: accAddrStr,
		ToAddress:   addr,
		Amount: []*basev1beta1.Coin{
			{
				Denom:  "stake",
				Amount: "100",
			},
		},
	}
	anyMsg, err := codectypes.NewAnyWithValue(msg)
	require.NoError(t, err)

	propReq := &v1.MsgCreateProposal{
		Proposal: &v1.Proposal{
			Title:    "test",
			Summary:  "test",
			Messages: []*codectypes.Any{anyMsg},
		},
	}
	err = s.executeTx(ctx, propReq, app, accountAddr, members[0])
	require.NoError(t, err)

	// now we vote for it
	voteReq := &v1.MsgVote{
		ProposalId: 0,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	}

	err = s.executeTx(ctx, voteReq, app, accountAddr, members[0])
	require.NoError(t, err)

	// now we execute it
	execReq := &v1.MsgExecuteProposal{
		ProposalId: 0,
	}

	err = s.executeTx(ctx, execReq, app, accountAddr, members[0])
	require.NoError(t, err)

	for _, v := range ctx.EventManager().Events() {
		if v.Type == "proposal_tally" {
			status, found := v.GetAttribute("status")
			require.True(t, found)
			require.Equal(t, v1.ProposalStatus_PROPOSAL_STATUS_PASSED.String(), status.Value)

			yesVotes, found := v.GetAttribute("yes_votes")
			require.True(t, found)
			require.Equal(t, "100", yesVotes.Value)

			noVotes, found := v.GetAttribute("no_votes")
			require.True(t, found)
			require.Equal(t, "0", noVotes.Value)

			propID, found := v.GetAttribute("proposal_id")
			require.True(t, found)
			require.Equal(t, "0", propID.Value)

			execErr, found := v.GetAttribute("exec_err")
			require.True(t, found)
			require.Equal(t, "<nil>", execErr.Value)

			rejectErr, found := v.GetAttribute("reject_err")
			require.True(t, found)
			require.Equal(t, "<nil>", rejectErr.Value)
		}
	}
	balance = app.BankKeeper.GetBalance(ctx, randAcc, "stake")
	require.Equal(t, int64(100), balance.Amount.Int64())

	err = s.executeTx(ctx, execReq, app, accountAddr, members[0])
	require.Error(t, err)

	// Add 2 members and pass the proposal
	// create proposal
	updateMsg := &v1.MsgUpdateConfig{
		UpdateMembers: []*v1.Member{
			{
				Address: member1,
				Weight:  100,
			},
			{
				Address: member2,
				Weight:  100,
			},
		},
		Config: &v1.Config{
			Threshold:      200, // 3 members with 100 power each, 2/3 majority
			Quorum:         200,
			VotingPeriod:   120,
			Revote:         false,
			EarlyExecution: false,
		},
	}

	msgExec := &accountsv1.MsgExecute{
		Sender:  accAddrStr,
		Target:  accAddrStr,
		Message: codectypes.UnsafePackAny(updateMsg),
		Funds:   []sdk.Coin{},
	}

	proposal := &v1.MsgCreateProposal{
		Proposal: &v1.Proposal{
			Title:    "Change config",
			Summary:  "Change config",
			Messages: []*codectypes.Any{codectypes.UnsafePackAny(msgExec)},
		},
	}

	err = s.executeTx(ctx, proposal, app, accountAddr, members[0])
	require.NoError(t, err)

	// vote
	voteReq = &v1.MsgVote{
		ProposalId: 1,
		Vote:       v1.VoteOption_VOTE_OPTION_YES,
	}

	err = s.executeTx(ctx, voteReq, app, accountAddr, members[0])
	require.NoError(t, err)

	execReq = &v1.MsgExecuteProposal{
		ProposalId: 1,
	}
	err = s.executeTx(ctx, execReq, app, accountAddr, members[0])
	require.NoError(t, err)

	// get members
	res, err := s.queryAcc(ctx, &v1.QueryConfig{}, app, accountAddr)
	require.NoError(t, err)
	resp := res.(*v1.QueryConfigResponse)
	require.Len(t, resp.Members, 3)
	require.Equal(t, int64(200), resp.Config.Threshold)

	// Try to remove a member, but it doesn't reach passing threshold
	// create proposal
	updateMsg = &v1.MsgUpdateConfig{
		UpdateMembers: []*v1.Member{
			{
				Address: member1,
				Weight:  0,
			},
		},
		Config: &v1.Config{
			Threshold:      200, // 3 members with 100 power each, 2/3 majority
			Quorum:         200,
			VotingPeriod:   120,
			Revote:         false,
			EarlyExecution: false,
		},
	}

	msgExec = &accountsv1.MsgExecute{
		Sender:  accAddrStr,
		Target:  accAddrStr,
		Message: codectypes.UnsafePackAny(updateMsg),
		Funds:   []sdk.Coin{},
	}

	proposal = &v1.MsgCreateProposal{
		Proposal: &v1.Proposal{
			Title:    "Change config",
			Summary:  "Change config",
			Messages: []*codectypes.Any{codectypes.UnsafePackAny(msgExec)},
		},
	}

	err = s.executeTx(ctx, proposal, app, accountAddr, members[0])
	require.NoError(t, err)

	// vote
	voteReq = &v1.MsgVote{
		ProposalId: 2,
		Vote:       v1.VoteOption_VOTE_OPTION_NO,
	}

	err = s.executeTx(ctx, voteReq, app, accountAddr, members[0])
	require.NoError(t, err)

	execReq = &v1.MsgExecuteProposal{
		ProposalId: 2,
	}

	// need to wait until voting period is over because we disabled early execution on the last
	// config update
	err = s.executeTx(ctx, execReq, app, accountAddr, members[0])
	require.ErrorContains(t, err, "voting period has not ended yet, and early execution is not enabled")

	// vote with member 1
	voteReq = &v1.MsgVote{
		ProposalId: 2,
		Vote:       v1.VoteOption_VOTE_OPTION_NO,
	}

	err = s.executeTx(ctx, voteReq, app, accountAddr, members[1])
	require.NoError(t, err)

	// need to wait until voting period is over because we disabled early execution on the last
	// config update
	err = s.executeTx(ctx, execReq, app, accountAddr, members[0])
	require.ErrorContains(t, err, "voting period has not ended yet, and early execution is not enabled")

	headerInfo := ctx.HeaderInfo()
	headerInfo.Time = headerInfo.Time.Add(time.Second * 121)
	ctx = ctx.WithHeaderInfo(headerInfo)

	// now it should work, but the proposal will fail
	err = s.executeTx(ctx, execReq, app, accountAddr, members[0])
	require.NoError(t, err)

	for _, v := range ctx.EventManager().Events() {
		if v.Type == "proposal_tally" {
			propID, found := v.GetAttribute("proposal_id")
			require.True(t, found)

			if propID.Value == "2" {
				status, found := v.GetAttribute("status")
				require.True(t, found)
				require.Equal(t, v1.ProposalStatus_PROPOSAL_STATUS_REJECTED.String(), status.Value)

				// exec_err is nil because the proposal didn't execute
				execErr, found := v.GetAttribute("exec_err")
				require.True(t, found)
				require.Equal(t, "<nil>", execErr.Value)

				rejectErr, found := v.GetAttribute("reject_err")
				require.True(t, found)
				require.Equal(t, "threshold not reached", rejectErr.Value)
			}
		}
	}

	// get members
	res, err = s.queryAcc(ctx, &v1.QueryConfig{}, app, accountAddr)
	require.NoError(t, err)
	resp = res.(*v1.QueryConfigResponse)
	require.Len(t, resp.Members, 3)
	require.Equal(t, int64(200), resp.Config.Threshold)
}
