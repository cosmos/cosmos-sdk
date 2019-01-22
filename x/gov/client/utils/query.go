package utils

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

const (
	defaultPage  = 1
	defaultLimit = 30 // should be consistent with tendermint/tendermint/rpc/core/pipe.go:19
)

// Proposer contains metadata of a governance proposal used for querying a
// proposer.
type Proposer struct {
	ProposalID uint64 `json:"proposal_id"`
	Proposer   string `json:"proposer"`
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
func QueryDepositsByTxQuery(
	cdc *codec.Codec, cliCtx context.CLIContext, params gov.QueryProposalParams,
) ([]byte, error) {

	tags := []string{
		fmt.Sprintf("%s='%s'", tags.Action, gov.MsgDeposit{}.Type()),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	infos, err := tx.SearchTxs(cliCtx, cdc, tags, defaultPage, defaultLimit)
	if err != nil {
		return nil, err
	}

	var deposits []gov.Deposit

	for _, info := range infos {
		for _, msg := range info.Tx.GetMsgs() {
			if msg.Type() == gov.TypeMsgDeposit {
				depMsg := msg.(gov.MsgDeposit)

				deposits = append(deposits, gov.Deposit{
					Depositor:  depMsg.Depositor,
					ProposalID: params.ProposalID,
					Amount:     depMsg.Amount,
				})
			}
		}
	}

	if cliCtx.Indent {
		return cdc.MarshalJSONIndent(deposits, "", "  ")
	}

	return cdc.MarshalJSON(deposits)
}

// QueryVotesByTxQuery will query for votes via a direct txs tags query. It
// will fetch and build votes directly from the returned txs and return a JSON
// marshalled result or any error that occurred.
//
// NOTE: SearchTxs is used to facilitate the txs query which does not currently
// support configurable pagination.
func QueryVotesByTxQuery(
	cdc *codec.Codec, cliCtx context.CLIContext, params gov.QueryProposalParams,
) ([]byte, error) {

	tags := []string{
		fmt.Sprintf("%s='%s'", tags.Action, gov.MsgVote{}.Type()),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	infos, err := tx.SearchTxs(cliCtx, cdc, tags, defaultPage, defaultLimit)
	if err != nil {
		return nil, err
	}

	var votes []gov.Vote

	for _, info := range infos {
		for _, msg := range info.Tx.GetMsgs() {
			if msg.Type() == gov.TypeMsgVote {
				voteMsg := msg.(gov.MsgVote)

				votes = append(votes, gov.Vote{
					Voter:      voteMsg.Voter,
					ProposalID: params.ProposalID,
					Option:     voteMsg.Option,
				})
			}
		}
	}

	if cliCtx.Indent {
		return cdc.MarshalJSONIndent(votes, "", "  ")
	}

	return cdc.MarshalJSON(votes)
}

// QueryVoteByTxQuery will query for a single vote via a direct txs tags query.
func QueryVoteByTxQuery(
	cdc *codec.Codec, cliCtx context.CLIContext, params gov.QueryVoteParams,
) ([]byte, error) {

	tags := []string{
		fmt.Sprintf("%s='%s'", tags.Action, gov.MsgVote{}.Type()),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		fmt.Sprintf("%s='%s'", tags.Voter, []byte(params.Voter.String())),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	infos, err := tx.SearchTxs(cliCtx, cdc, tags, defaultPage, defaultLimit)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		for _, msg := range info.Tx.GetMsgs() {
			// there should only be a single vote under the given conditions
			if msg.Type() == gov.TypeMsgVote {
				voteMsg := msg.(gov.MsgVote)

				vote := gov.Vote{
					Voter:      voteMsg.Voter,
					ProposalID: params.ProposalID,
					Option:     voteMsg.Option,
				}

				if cliCtx.Indent {
					return cdc.MarshalJSONIndent(vote, "", "  ")
				}

				return cdc.MarshalJSON(vote)
			}
		}
	}

	return nil, fmt.Errorf("address '%s' did not vote on proposalID %d", params.Voter, params.ProposalID)
}

// QueryDepositByTxQuery will query for a single deposit via a direct txs tags
// query.
func QueryDepositByTxQuery(
	cdc *codec.Codec, cliCtx context.CLIContext, params gov.QueryDepositParams,
) ([]byte, error) {

	tags := []string{
		fmt.Sprintf("%s='%s'", tags.Action, gov.MsgDeposit{}.Type()),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		fmt.Sprintf("%s='%s'", tags.Depositor, []byte(params.Depositor.String())),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	infos, err := tx.SearchTxs(cliCtx, cdc, tags, defaultPage, defaultLimit)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		for _, msg := range info.Tx.GetMsgs() {
			// there should only be a single deposit under the given conditions
			if msg.Type() == gov.TypeMsgDeposit {
				depMsg := msg.(gov.MsgDeposit)

				deposit := gov.Deposit{
					Depositor:  depMsg.Depositor,
					ProposalID: params.ProposalID,
					Amount:     depMsg.Amount,
				}

				if cliCtx.Indent {
					return cdc.MarshalJSONIndent(deposit, "", "  ")
				}

				return cdc.MarshalJSON(deposit)
			}
		}
	}

	return nil, fmt.Errorf("address '%s' did not deposit to proposalID %d", params.Depositor, params.ProposalID)
}

// QueryProposerByTxQuery will query for a proposer of a governance proposal by
// ID.
func QueryProposerByTxQuery(
	cdc *codec.Codec, cliCtx context.CLIContext, proposalID uint64,
) (Proposer, error) {

	tags := []string{
		fmt.Sprintf("%s='%s'", tags.Action, gov.MsgSubmitProposal{}.Type()),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", proposalID))),
	}

	// NOTE: SearchTxs is used to facilitate the txs query which does not currently
	// support configurable pagination.
	infos, err := tx.SearchTxs(cliCtx, cdc, tags, defaultPage, defaultLimit)
	if err != nil {
		return Proposer{}, err
	}

	for _, info := range infos {
		for _, msg := range info.Tx.GetMsgs() {
			// there should only be a single proposal under the given conditions
			if msg.Type() == gov.TypeMsgSubmitProposal {
				subMsg := msg.(gov.MsgSubmitProposal)
				return NewProposer(proposalID, subMsg.Proposer.String()), nil
			}
		}
	}
	return Proposer{}, fmt.Errorf("failed to find the proposer for proposalID %d", proposalID)
}

// QueryProposalByID takes a proposalID and returns a proposal
func QueryProposalByID(proposalID uint64, cliCtx context.CLIContext, cdc *codec.Codec, queryRoute string) ([]byte, error) {
	params := gov.NewQueryProposalParams(proposalID)
	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return nil, err
	}

	res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/proposal", queryRoute), bz)
	if err != nil {
		return nil, err
	}
	return res, err
}
