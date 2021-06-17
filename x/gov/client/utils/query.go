package utils

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
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
	var deposits []types.Deposit

	// initial deposit was submitted with proposal, so must be queried separately
	initialDeposit, err := queryInitialDepositByTxQuery(clientCtx, params.ProposalID)
	if err != nil {
		return nil, err
	}

	if !initialDeposit.Amount.IsZero() {
		deposits = append(deposits, initialDeposit)
	}

	events := []string{
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgDeposit),
		fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	searchResult, err := authclient.QueryTxsByEvents(clientCtx, events, defaultPage, defaultLimit, "")
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			if msg.Type() == types.TypeMsgDeposit {
				depMsg := msg.(*types.MsgDeposit)

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
		events = []string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVote),
			fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		}
		votes      []types.Vote
		nextTxPage = defaultPage
		totalLimit = params.Limit * params.Page
	)
	// query interrupted either if we collected enough votes or tx indexer run out of relevant txs
	for len(votes) < totalLimit {
		searchResult, err := authclient.QueryTxsByEvents(clientCtx, events, nextTxPage, defaultLimit, "")
		if err != nil {
			return nil, err
		}
		nextTxPage++
		for _, info := range searchResult.Txs {
			for _, msg := range info.GetTx().GetMsgs() {
				if msg.Type() == types.TypeMsgVote {
					voteMsg := msg.(*types.MsgVote)

					votes = append(votes, types.Vote{
						Voter:      voteMsg.Voter,
						ProposalId: params.ProposalID,
						Option:     voteMsg.Option,
					})
				}
			}
		}
		if len(searchResult.Txs) != defaultLimit {
			break
		}
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
	events := []string{
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgVote),
		fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalVote, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Voter.String())),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	searchResult, err := authclient.QueryTxsByEvents(clientCtx, events, defaultPage, defaultLimit, "")
	if err != nil {
		return nil, err
	}
	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single vote under the given conditions
			if msg.Type() == types.TypeMsgVote {
				voteMsg := msg.(*types.MsgVote)

				vote := types.Vote{
					Voter:      voteMsg.Voter,
					ProposalId: params.ProposalID,
					Option:     voteMsg.Option,
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
	// initial deposit was submitted with proposal, so must be queried separately
	initialDeposit, err := queryInitialDepositByTxQuery(clientCtx, params.ProposalID)
	if err != nil {
		return nil, err
	}

	if !initialDeposit.Amount.IsZero() {
		bz, err := clientCtx.JSONMarshaler.MarshalJSON(&initialDeposit)
		if err != nil {
			return nil, err
		}

		return bz, nil
	}

	events := []string{
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgDeposit),
		fmt.Sprintf("%s.%s='%s'", types.EventTypeProposalDeposit, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeySender, []byte(params.Depositor.String())),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	searchResult, err := authclient.QueryTxsByEvents(clientCtx, events, defaultPage, defaultLimit, "")
	if err != nil {
		return nil, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single deposit under the given conditions
			if msg.Type() == types.TypeMsgDeposit {
				depMsg := msg.(*types.MsgDeposit)

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
	events := []string{
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgSubmitProposal),
		fmt.Sprintf("%s.%s='%s'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	searchResult, err := authclient.QueryTxsByEvents(clientCtx, events, defaultPage, defaultLimit, "")
	if err != nil {
		return Proposer{}, err
	}

	for _, info := range searchResult.Txs {
		for _, msg := range info.GetTx().GetMsgs() {
			// there should only be a single proposal under the given conditions
			if msg.Type() == types.TypeMsgSubmitProposal {
				subMsg := msg.(*types.MsgSubmitProposal)
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

// queryInitialDepositByTxQuery will query for a initial deposit of a governance proposal by
// ID.
func queryInitialDepositByTxQuery(clientCtx client.Context, proposalID uint64) (types.Deposit, error) {
	// Query legacy Msgs event action
	events := []string{
		fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, types.TypeMsgSubmitProposal),
		fmt.Sprintf("%s.%s='%s'", types.EventTypeSubmitProposal, types.AttributeKeyProposalID, []byte(fmt.Sprintf("%d", proposalID))),
	}
	searchResult, err := authclient.QueryTxsByEvents(clientCtx, events, defaultPage, defaultLimit, "")

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

	return types.Deposit{}, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "failed to find the initial deposit for proposalID %d", proposalID)
}
