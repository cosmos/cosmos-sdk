package utils

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
// JSON marshalled result or any error that occurred.
//
// NOTE: SearchTxs is used to facilitate the txs query which does not currently
// support configurable pagination.
func QueryDepositsByTxQuery(clientCtx client.Context, params types.QueryProposalParams) ([]byte, error) {
	searchResult, err := combineEvents(
		clientCtx, defaultPage,
		// Query old Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgDeposit),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		},
		// Query service Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeSvcMsgDeposit),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		},
	)
	if err != nil {
		return nil, err
	}

	var deposits []types.Deposit

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			var depMsg *types.MsgDeposit
			if msg.Type() == types.TypeSvcMsgDeposit {
				depMsg = msg.(sdk.ServiceMsg).Request.(*types.MsgDeposit)
			} else if protoDepMsg, ok := msg.(*types.MsgDeposit); ok {
				depMsg = protoDepMsg
			}

			if depMsg != nil {
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
// marshalled result or any error that occurred.
func QueryVotesByTxQuery(clientCtx client.Context, params types.QueryProposalVotesParams) ([]byte, error) {
	var (
		votes      []types.Vote
		nextTxPage = defaultPage
		totalLimit = params.Limit * params.Page
	)

	// query interrupted either if we collected enough votes or tx indexer run out of relevant txs
	for len(votes) < totalLimit {
		// Search for both (old) votes and weighted votes.
		searchResult, err := combineEvents(
			clientCtx, nextTxPage,
			// Query old Vote Msgs
			[]string{
				fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVote),
				fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			},
			// Query Vote service Msgs
			[]string{
				fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeSvcMsgVote),
				fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			},
			// Query old VoteWeighted Msgs
			[]string{
				fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVoteWeighted),
				fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			},
			// Query VoteWeighted service Msgs
			[]string{
				fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeSvcMsgVoteWeighted),
				fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			},
		)
		if err != nil {
			return nil, err
		}

		for _, info := range searchResult.Txs {
			for _, msg := range info.GetTx().GetMsgs() {
				var voteMsg *types.MsgVote
				if msg.Type() == types.TypeSvcMsgVote {
					voteMsg = msg.(sdk.ServiceMsg).Request.(*types.MsgVote)
				} else if protoVoteMsg, ok := msg.(*types.MsgVote); ok {
					voteMsg = protoVoteMsg
				}

				if voteMsg != nil {
					votes = append(votes, types.Vote{
						Voter:      voteMsg.Voter,
						ProposalId: params.ProposalID,
						Options:    types.NewNonSplitVoteOption(voteMsg.Option),
					})
				}

				var voteWeightedMsg *types.MsgVoteWeighted
				if msg.Type() == types.TypeSvcMsgVoteWeighted {
					voteWeightedMsg = msg.(sdk.ServiceMsg).Request.(*types.MsgVoteWeighted)
				} else if protoVoteWeightedMsg, ok := msg.(*types.MsgVoteWeighted); ok {
					voteWeightedMsg = protoVoteWeightedMsg
				}

				if voteWeightedMsg != nil {
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
		// Query old Vote Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVote),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Voter.String())),
		},
		// Query Vote service Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeSvcMsgVote),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Voter.String())),
		},
		// Query old VoteWeighted Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVoteWeighted),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Voter.String())),
		},
		// Query VoteWeighted service Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeSvcMsgVoteWeighted),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Voter.String())),
		},
	)
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			var voteMsg *types.MsgVote
			// there should only be a single vote under the given conditions
			if msg.Type() == types.TypeSvcMsgVote {
				voteMsg = msg.(sdk.ServiceMsg).Request.(*types.MsgVote)
			} else if protoVoteMsg, ok := msg.(*types.MsgVote); ok {
				voteMsg = protoVoteMsg
			}

			if voteMsg != nil {
				vote := types.Vote{
					Voter:      voteMsg.Voter,
					ProposalId: params.ProposalID,
					Options:    types.NewNonSplitVoteOption(voteMsg.Option),
				}

				bz, err := clientCtx.JSONMarshaler.MarshalJSON(&vote)
				if err != nil {
					return nil, err
				}

				return bz, nil
			} else if msg.Type() == types.TypeMsgVoteWeighted {
				voteMsg := msg.(*types.MsgVoteWeighted)

				vote := types.Vote{
					Voter:      voteMsg.Voter,
					ProposalId: params.ProposalID,
					Options:    voteMsg.Options,
				}

				bz, err := clientCtx.JSONMarshaler.MarshalJSON(&vote)
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
	searchResult, err := combineEvents(
		clientCtx, defaultPage,
		// Query old Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgDeposit),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Depositor.String())),
		},
		// Query service Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeSvcMsgDeposit),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Depositor.String())),
		},
	)
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			var depMsg *types.MsgDeposit
			// there should only be a single deposit under the given conditions
			if msg.Type() == types.TypeSvcMsgDeposit {
				depMsg = msg.(sdk.ServiceMsg).Request.(*types.MsgDeposit)
			} else if protoDepMsg, ok := msg.(*types.MsgDeposit); ok {
				depMsg = protoDepMsg
			}

			if depMsg != nil {
				deposit := types.Deposit{
					Depositor:  depMsg.Depositor,
					ProposalId: params.ProposalID,
					Amount:     depMsg.Amount,
				}

				bz, err := clientCtx.JSONMarshaler.MarshalJSON(&deposit)
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
		// Query old Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgSubmitProposal),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
		},
		// Query service Msgs
		[]string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeSvcMsgSubmitProposal),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
		},
	)
	if err != nil {
		return Proposer{}, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single proposal under the given conditions
			if msg.Type() == types.TypeSvcMsgSubmitProposal {
				subMsg := msg.(sdk.ServiceMsg).Request.(*types.MsgSubmitProposal)

				return NewProposer(proposalID, subMsg.Proposer), nil
			} else if protoSubMsg, ok := msg.(*types.MsgSubmitProposal); ok {
				subMsg := protoSubMsg

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
// - via ADR-031 service msgs, their `Type()` is the protobuf FQ method name.
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
