package utils

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	defaultPage  = 1
	defaultLimit = 30 // should be consistent with https://github.com/cometbft/cometbft/tree/v0.37.x/rpc/core#pagination
)

// Proposer contains metadata of a governance proposal used for querying a
// proposer.
type Proposer struct {
	ProposalID uint64 `json:"proposal_id" yaml:"proposal_id"`
	Proposer   string `json:"proposer" yaml:"proposer"`
}

// NewProposer returns a new Proposer given id and proposer
func NewProposer(proposalID uint64, proposer string) Proposer {
	return Proposer{proposalID, proposer}
}

// String implements the fmt.Stringer interface.
func (p Proposer) String() string {
	return fmt.Sprintf("Proposal with ID %d was proposed by %s", p.ProposalID, p.Proposer)
}

// QueryDepositsByTxQuery will query for deposits via a direct txs tags query. It
// will fetch and build deposits directly from the returned txs and returns a
// JSON marshalled result or any error that occurred.
//
// NOTE: SearchTxs is used to facilitate the txs query which does not currently
// support configurable pagination.
func QueryDepositsByTxQuery(clientCtx client.Context, params v1.QueryProposalParams) ([]byte, error) {
	var deposits []v1.Deposit

	// initial deposit was submitted with proposal, so must be queried separately
	initialDeposit, err := queryInitialDepositByTxQuery(clientCtx, params.ProposalID)
	if err != nil {
		return nil, err
	}

	if !sdk.Coins(initialDeposit.Amount).IsZero() {
		deposits = append(deposits, initialDeposit)
	}

	q := fmt.Sprintf("%s.%s='%d'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, params.ProposalID)
	searchResult, err := authtx.QueryTxsByEvents(clientCtx, defaultPage, defaultLimit, q, "")
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			if depMsg, ok := msg.(*v1beta1.MsgDeposit); ok {
				deposits = append(deposits, v1.Deposit{
					Depositor:  depMsg.Depositor,
					ProposalId: params.ProposalID,
					Amount:     depMsg.Amount,
				})
			}

			if depMsg, ok := msg.(*v1.MsgDeposit); ok {
				deposits = append(deposits, v1.Deposit{
					Depositor:  depMsg.Depositor,
					ProposalId: params.ProposalID,
					Amount:     depMsg.Amount,
				})
			}
		}
	}

	bz, err := clientCtx.LegacyAmino.MarshalJSON(deposits)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// QueryVotesByTxQuery will query for votes via a direct txs tags query. It
// will fetch and build votes directly from the returned txs and returns a JSON
// marshalled result or any error that occurred.
func QueryVotesByTxQuery(clientCtx client.Context, params v1.QueryProposalVotesParams) ([]byte, error) {
	var (
		votes      []*v1.Vote
		nextTxPage = defaultPage
		totalLimit = params.Limit * params.Page
	)

	// query interrupted either if we collected enough votes or tx indexer run out of relevant txs
	for len(votes) < totalLimit {
		// Search for both (legacy) votes and weighted votes.
		q := fmt.Sprintf("%s.%s='%d'", types.EventTypeProposalVote, types.AttributeKeyProposalID, params.ProposalID)
		searchResult, err := authtx.QueryTxsByEvents(clientCtx, defaultPage, defaultLimit, q, "")
		if err != nil {
			return nil, err
		}

		for _, info := range searchResult.Txs {
			for _, msg := range info.GetTx().GetMsgs() {
				if voteMsg, ok := msg.(*v1beta1.MsgVote); ok {
					votes = append(votes, &v1.Vote{
						Voter:      voteMsg.Voter,
						ProposalId: params.ProposalID,
						Options:    v1.NewNonSplitVoteOption(v1.VoteOption(voteMsg.Option)),
					})
				}

				if voteMsg, ok := msg.(*v1.MsgVote); ok {
					votes = append(votes, &v1.Vote{
						Voter:      voteMsg.Voter,
						ProposalId: params.ProposalID,
						Options:    v1.NewNonSplitVoteOption(voteMsg.Option),
					})
				}

				if voteWeightedMsg, ok := msg.(*v1beta1.MsgVoteWeighted); ok {
					votes = append(votes, convertVote(voteWeightedMsg))
				}

				if voteWeightedMsg, ok := msg.(*v1.MsgVoteWeighted); ok {
					votes = append(votes, &v1.Vote{
						Voter:      voteWeightedMsg.Voter,
						ProposalId: params.ProposalID,
						Options:    voteWeightedMsg.Options,
					})
				}
			}
		}
		if len(searchResult.Txs) != defaultLimit {
			break
		}

		nextTxPage++
	}
	start, end := client.Paginate(len(votes), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		votes = []*v1.Vote{}
	} else {
		votes = votes[start:end]
	}

	bz, err := clientCtx.LegacyAmino.MarshalJSON(votes)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// QueryVoteByTxQuery will query for a single vote via a direct txs tags query.
func QueryVoteByTxQuery(clientCtx client.Context, params v1.QueryVoteParams) ([]byte, error) {
	q1 := fmt.Sprintf("%s.%s='%d'", types.EventTypeProposalVote, types.AttributeKeyProposalID, params.ProposalID)
	q2 := fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, params.Voter.String())
	q3 := fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, params.Voter)
	searchResult, err := authtx.QueryTxsByEvents(clientCtx, defaultPage, defaultLimit, fmt.Sprintf("%s AND (%s OR %s)", q1, q2, q3), "")
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single vote under the given conditions
			var vote *v1.Vote
			if voteMsg, ok := msg.(*v1beta1.MsgVote); ok {
				vote = &v1.Vote{
					Voter:      voteMsg.Voter,
					ProposalId: params.ProposalID,
					Options:    v1.NewNonSplitVoteOption(v1.VoteOption(voteMsg.Option)),
				}
			}

			if voteMsg, ok := msg.(*v1.MsgVote); ok {
				vote = &v1.Vote{
					Voter:      voteMsg.Voter,
					ProposalId: params.ProposalID,
					Options:    v1.NewNonSplitVoteOption(voteMsg.Option),
				}
			}

			if voteWeightedMsg, ok := msg.(*v1beta1.MsgVoteWeighted); ok {
				vote = convertVote(voteWeightedMsg)
			}

			if voteWeightedMsg, ok := msg.(*v1.MsgVoteWeighted); ok {
				vote = &v1.Vote{
					Voter:      voteWeightedMsg.Voter,
					ProposalId: params.ProposalID,
					Options:    voteWeightedMsg.Options,
				}
			}

			if vote != nil {
				bz, err := clientCtx.Codec.MarshalJSON(vote)
				if err != nil {
					return nil, err
				}

				return bz, nil
			}
		}
	}

	return nil, fmt.Errorf("address '%s' did not vote on proposalID %d", params.Voter, params.ProposalID)
}

// QueryDepositByTxQuery will query for a single deposit via a direct txs tags
// query.
func QueryDepositByTxQuery(clientCtx client.Context, params v1.QueryDepositParams) ([]byte, error) {
	// initial deposit was submitted with proposal, so must be queried separately
	initialDeposit, err := queryInitialDepositByTxQuery(clientCtx, params.ProposalID)
	if err != nil {
		return nil, err
	}

	if !sdk.Coins(initialDeposit.Amount).IsZero() {
		bz, err := clientCtx.Codec.MarshalJSON(&initialDeposit)
		if err != nil {
			return nil, err
		}

		return bz, nil
	}

	q1 := fmt.Sprintf("%s.%s='%d'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, params.ProposalID)
	q2 := fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, params.Depositor.String())
	searchResult, err := authtx.QueryTxsByEvents(clientCtx, defaultPage, defaultLimit, fmt.Sprintf("%s AND %s", q1, q2), "")
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single deposit under the given conditions
			if depMsg, ok := msg.(*v1beta1.MsgDeposit); ok {
				deposit := v1.Deposit{
					Depositor:  depMsg.Depositor,
					ProposalId: params.ProposalID,
					Amount:     depMsg.Amount,
				}

				bz, err := clientCtx.Codec.MarshalJSON(&deposit)
				if err != nil {
					return nil, err
				}

				return bz, nil
			}

			if depMsg, ok := msg.(*v1.MsgDeposit); ok {
				deposit := v1.Deposit{
					Depositor:  depMsg.Depositor,
					ProposalId: params.ProposalID,
					Amount:     depMsg.Amount,
				}

				bz, err := clientCtx.Codec.MarshalJSON(&deposit)
				if err != nil {
					return nil, err
				}

				return bz, nil
			}
		}
	}

	return nil, fmt.Errorf("address '%s' did not deposit to proposalID %d", params.Depositor, params.ProposalID)
}

// QueryProposerByTxQuery will query for a proposer of a governance proposal by ID.
func QueryProposerByTxQuery(clientCtx client.Context, proposalID uint64) (Proposer, error) {
	q := fmt.Sprintf("%s.%s='%d'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, proposalID)
	searchResult, err := authtx.QueryTxsByEvents(clientCtx, defaultPage, defaultLimit, q, "")
	if err != nil {
		return Proposer{}, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single proposal under the given conditions
			if subMsg, ok := msg.(*v1beta1.MsgSubmitProposal); ok {
				return NewProposer(proposalID, subMsg.Proposer), nil
			}
			if subMsg, ok := msg.(*v1.MsgSubmitProposal); ok {
				return NewProposer(proposalID, subMsg.Proposer), nil
			}
		}
	}

	return Proposer{}, fmt.Errorf("failed to find the proposer for proposalID %d", proposalID)
}

// QueryProposalByID takes a proposalID and returns a proposal
func QueryProposalByID(proposalID uint64, clientCtx client.Context, queryRoute string) ([]byte, error) {
	params := v1.NewQueryProposalParams(proposalID)
	bz, err := clientCtx.LegacyAmino.MarshalJSON(params)
	if err != nil {
		return nil, err
	}

	res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/proposal", queryRoute), bz)
	if err != nil {
		return nil, err
	}

	return res, err
}

// queryInitialDepositByTxQuery will query for a initial deposit of a governance proposal by
// ID.
func queryInitialDepositByTxQuery(clientCtx client.Context, proposalID uint64) (v1.Deposit, error) {
	searchResult, err := authtx.QueryTxsByEvents(clientCtx, defaultPage, defaultLimit, fmt.Sprintf("%s.%s='%d'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, proposalID), "")
	if err != nil {
		return v1.Deposit{}, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single proposal under the given conditions
			if subMsg, ok := msg.(*v1beta1.MsgSubmitProposal); ok {
				return v1.Deposit{
					ProposalId: proposalID,
					Depositor:  subMsg.Proposer,
					Amount:     subMsg.InitialDeposit,
				}, nil
			}

			if subMsg, ok := msg.(*v1.MsgSubmitProposal); ok {
				return v1.Deposit{
					ProposalId: proposalID,
					Depositor:  subMsg.Proposer,
					Amount:     subMsg.InitialDeposit,
				}, nil
			}
		}
	}

	return v1.Deposit{}, sdkerrors.ErrNotFound.Wrapf("failed to find the initial deposit for proposalID %d", proposalID)
}

// convertVote converts a MsgVoteWeighted into a *v1.Vote.
func convertVote(v *v1beta1.MsgVoteWeighted) *v1.Vote {
	opts := make([]*v1.WeightedVoteOption, len(v.Options))
	for i, o := range v.Options {
		opts[i] = &v1.WeightedVoteOption{
			Option: v1.VoteOption(o.Option),
			Weight: o.Weight.String(),
		}
	}
	return &v1.Vote{
		Voter:      v.Voter,
		ProposalId: v.ProposalId,
		Options:    opts,
	}
}
