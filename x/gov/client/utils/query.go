package utils

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	defaultPage  = 1
	defaultLimit = 30 // should be consistent with tendermint/tendermint/rpc/core/pipe.go:19
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

func (p Proposer) String() string {
	return fmt.Sprintf("Proposal with ID %d was proposed by %s", p.ProposalID, p.Proposer)
}

// QueryDepositsByTxQuery will query for deposits via a direct txs tags query. It
// will fetch and build deposits directly from the returned txs and return a
// JSON marshaled result or any error that occurred.
//
// NOTE: SearchTxs is used to facilitate the txs query which does not currently
// support configurable pagination.
func QueryDepositsByTxQuery(clientCtx client.Context, params types.QueryProposalParams) ([]byte, error) {
	var deposits []types.Deposit

	// initial deposit was submitted with proposal, so must be queried separately
	initialDeposit, err := queryInitialDepositByTxQuery(clientCtx, params.ProposalID)
	if err != nil {
		return nil, err
	}

	if !initialDeposit.Amount.IsZero() {
		deposits = append(deposits, initialDeposit)
	}

	searchResult, err := combineEvents(
		clientCtx, defaultPage,
		// Query legacy Msgs event action
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgDeposit),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		},
		// Query proto Msgs event action
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, sdk.MsgTypeURL(&types.MsgDeposit{})),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		},
	)
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			if depMsg, ok := msg.(*types.MsgDeposit); ok {
				deposits = append(deposits, types.Deposit{
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
// will fetch and build votes directly from the returned txs and return a JSON
// marshaled result or any error that occurred.
func QueryVotesByTxQuery(clientCtx client.Context, params types.QueryProposalVotesParams) ([]byte, error) {
	var (
		votes      []types.Vote
		nextTxPage = defaultPage
		totalLimit = params.Limit * params.Page
	)

	// query interrupted either if we collected enough votes or tx indexer run out of relevant txs
	for len(votes) < totalLimit {
		// Search for both (legacy) votes and weighted votes.
		searchResult, err := combineEvents(
			clientCtx, nextTxPage,
			// Query legacy Vote Msgs
			[]string{
				fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVote),
				fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			},
			// Query Vote proto Msgs
			[]string{
				fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, sdk.MsgTypeURL(&types.MsgVote{})),
				fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			},
			// Query legacy VoteWeighted Msgs
			[]string{
				fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVoteWeighted),
				fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			},
			// Query VoteWeighted proto Msgs
			[]string{
				fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, sdk.MsgTypeURL(&types.MsgVoteWeighted{})),
				fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			},
		)
		if err != nil {
			return nil, err
		}

		for _, info := range searchResult.Txs {
			for _, msg := range info.GetTx().GetMsgs() {
				if voteMsg, ok := msg.(*types.MsgVote); ok {
					votes = append(votes, types.Vote{
						Voter:      voteMsg.Voter,
						ProposalId: params.ProposalID,
						Options:    types.NewNonSplitVoteOption(voteMsg.Option),
					})
				}

				if voteWeightedMsg, ok := msg.(*types.MsgVoteWeighted); ok {
					votes = append(votes, types.Vote{
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
		votes = []types.Vote{}
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
func QueryVoteByTxQuery(clientCtx client.Context, params types.QueryVoteParams) ([]byte, error) {
	searchResult, err := combineEvents(
		clientCtx, defaultPage,
		// Query legacy Vote Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVote),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Voter.String())),
		},
		// Query Vote proto Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, sdk.MsgTypeURL(&types.MsgVote{})),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Voter.String())),
		},
		// Query legacy VoteWeighted Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVoteWeighted),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Voter.String())),
		},
		// Query VoteWeighted proto Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, sdk.MsgTypeURL(&types.MsgVoteWeighted{})),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Voter.String())),
		},
	)
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single vote under the given conditions
			var vote *types.Vote
			if voteMsg, ok := msg.(*types.MsgVote); ok {
				vote = &types.Vote{
					Voter:      voteMsg.Voter,
					ProposalId: params.ProposalID,
					Options:    types.NewNonSplitVoteOption(voteMsg.Option),
				}
			}

			if voteWeightedMsg, ok := msg.(*types.MsgVoteWeighted); ok {
				vote = &types.Vote{
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
func QueryDepositByTxQuery(clientCtx client.Context, params types.QueryDepositParams) ([]byte, error) {
	// initial deposit was submitted with proposal, so must be queried separately
	initialDeposit, err := queryInitialDepositByTxQuery(clientCtx, params.ProposalID)
	if err != nil {
		return nil, err
	}

	if !initialDeposit.Amount.IsZero() {
		bz, err := clientCtx.Codec.MarshalJSON(&initialDeposit)
		if err != nil {
			return nil, err
		}

		return bz, nil
	}

	searchResult, err := combineEvents(
		clientCtx, defaultPage,
		// Query legacy Msgs event action
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgDeposit),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Depositor.String())),
		},
		// Query proto Msgs event action
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, sdk.MsgTypeURL(&types.MsgDeposit{})),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Depositor.String())),
		},
	)
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single deposit under the given conditions
			if depMsg, ok := msg.(*types.MsgDeposit); ok {
				deposit := types.Deposit{
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

// QueryProposerByTxQuery will query for a proposer of a governance proposal by
// ID.
func QueryProposerByTxQuery(clientCtx client.Context, proposalID uint64) (Proposer, error) {
	searchResult, err := combineEvents(
		clientCtx,
		defaultPage,
		// Query legacy Msgs event action
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgSubmitProposal),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
		},
		// Query proto Msgs event action
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, sdk.MsgTypeURL(&types.MsgSubmitProposal{})),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
		},
	)
	if err != nil {
		return Proposer{}, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single proposal under the given conditions
			if subMsg, ok := msg.(*types.MsgSubmitProposal); ok {
				return NewProposer(proposalID, subMsg.Proposer), nil
			}
		}
	}

	return Proposer{}, fmt.Errorf("failed to find the proposer for proposalID %d", proposalID)
}

// QueryProposalByID takes a proposalID and returns a proposal
func QueryProposalByID(proposalID uint64, clientCtx client.Context, queryRoute string) ([]byte, error) {
	params := types.NewQueryProposalParams(proposalID)
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

// combineEvents queries txs by events with all events from each event group,
// and combines all those events together.
//
// Tx are indexed in tendermint via their Msgs `Type()`, which can be:
// - via legacy Msgs (amino or proto), their `Type()` is a custom string,
// - via ADR-031 proto msgs, their `Type()` is the protobuf FQ method name.
// In searching for events, we search for both `Type()`s, and we use the
// `combineEvents` function here to merge events.
func combineEvents(clientCtx client.Context, page int, eventGroups ...[]string) (*sdk.SearchTxsResult, error) {
	// Only the Txs field will be populated in the final SearchTxsResult.
	allTxs := []*sdk.TxResponse{}
	for _, events := range eventGroups {
		res, err := authtx.QueryTxsByEvents(clientCtx, events, page, defaultLimit, "")
		if err != nil {
			return nil, err
		}
		allTxs = append(allTxs, res.Txs...)
	}

	return &sdk.SearchTxsResult{Txs: allTxs}, nil
}

// queryInitialDepositByTxQuery will query for a initial deposit of a governance proposal by
// ID.
func queryInitialDepositByTxQuery(clientCtx client.Context, proposalID uint64) (types.Deposit, error) {
	searchResult, err := combineEvents(
		clientCtx, defaultPage,
		// Query legacy Msgs event action
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgSubmitProposal),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
		},
		// Query proto Msgs event action
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, sdk.MsgTypeURL(&types.MsgSubmitProposal{})),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
		},
	)
	if err != nil {
		return types.Deposit{}, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single proposal under the given conditions
			if subMsg, ok := msg.(*types.MsgSubmitProposal); ok {
				return types.Deposit{
					ProposalId: proposalID,
					Depositor:  subMsg.Proposer,
					Amount:     subMsg.InitialDeposit,
				}, nil
			}
		}
	}

	return types.Deposit{}, sdkerrors.ErrNotFound.Wrapf("failed to find the initial deposit for proposalID %d", proposalID)
}
