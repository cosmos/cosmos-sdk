package utils

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

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
		fmt.Sprintf("%s='%s'", tags.Action, tags.ActionProposalDeposit),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
	}

	infos, err := tx.SearchTxs(cliCtx, cdc, tags)
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
		fmt.Sprintf("%s='%s'", tags.Action, tags.ActionProposalVote),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
	}

	infos, err := tx.SearchTxs(cliCtx, cdc, tags)
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
		fmt.Sprintf("%s='%s'", tags.Action, tags.ActionProposalVote),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		fmt.Sprintf("%s='%s'", tags.Voter, []byte(params.Voter.String())),
	}

	infos, err := tx.SearchTxs(cliCtx, cdc, tags)
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

	err = fmt.Errorf("address '%s' did not vote on proposalID %d", params.Voter, params.ProposalID)
	return nil, err
}

// QueryDepositByTxQuery will query for a single deposit via a direct txs tags
// query.
func QueryDepositByTxQuery(
	cdc *codec.Codec, cliCtx context.CLIContext, params gov.QueryDepositParams,
) ([]byte, error) {

	tags := []string{
		fmt.Sprintf("%s='%s'", tags.Action, tags.ActionProposalDeposit),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		fmt.Sprintf("%s='%s'", tags.Depositor, []byte(params.Depositor.String())),
	}

	infos, err := tx.SearchTxs(cliCtx, cdc, tags)
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

	err = fmt.Errorf("address '%s' did not deposit to proposalID %d", params.Depositor, params.ProposalID)
	return nil, err
}
