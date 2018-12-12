package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

// queryDepositsByTxQuery will query for deposits via a direct txs tags query. It
// will fetch and build deposits directly from the returned txs and write the
// JSON response to the provided ResponseWriter.
//
// NOTE: SearchTxs is used to facilitate the txs query which does not currently
// support configurable pagination.
func queryDepositsByTxQuery(
	cdc *codec.Codec, cliCtx context.CLIContext, w http.ResponseWriter, params gov.QueryProposalParams,
) {

	tags := []string{
		fmt.Sprintf("%s='%s'", tags.Action, tags.ActionProposalDeposit),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
	}

	infos, err := tx.SearchTxs(cliCtx, cdc, tags)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
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

	utils.PostProcessResponse(w, cdc, deposits, cliCtx.Indent)
}

// queryVotesByTxQuery will query for votes via a direct txs tags query. It
// will fetch and build votes directly from the returned txs and write the
// JSON response to the provided ResponseWriter.
//
// NOTE: SearchTxs is used to facilitate the txs query which does not currently
// support configurable pagination.
func queryVotesByTxQuery(
	cdc *codec.Codec, cliCtx context.CLIContext, w http.ResponseWriter, params gov.QueryProposalParams,
) {
	tags := []string{
		fmt.Sprintf("%s='%s'", tags.Action, tags.ActionProposalVote),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
	}

	infos, err := tx.SearchTxs(cliCtx, cdc, tags)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
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

	utils.PostProcessResponse(w, cdc, votes, cliCtx.Indent)
}

// queryVoteByTxQuery will query for a single vote via a direct txs tags query.
//
// NOTE: SearchTxs is used to facilitate the txs query which does not currently
// support configurable pagination.
func queryVoteByTxQuery(
	cdc *codec.Codec, cliCtx context.CLIContext, w http.ResponseWriter, params gov.QueryVoteParams,
) {

	tags := []string{
		fmt.Sprintf("%s='%s'", tags.Action, tags.ActionProposalVote),
		fmt.Sprintf("%s='%s'", tags.ProposalID, []byte(fmt.Sprintf("%d", params.ProposalID))),
		fmt.Sprintf("%s='%s'", tags.Voter, []byte(params.Voter.String())),
	}

	infos, err := tx.SearchTxs(cliCtx, cdc, tags)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, info := range infos {
		for _, msg := range info.Tx.GetMsgs() {
			if msg.Type() == gov.TypeMsgVote {
				voteMsg := msg.(gov.MsgVote)

				// there should only be a single vote under the given condition
				vote := gov.Vote{
					Voter:      voteMsg.Voter,
					ProposalID: params.ProposalID,
					Option:     voteMsg.Option,
				}

				utils.PostProcessResponse(w, cdc, vote, cliCtx.Indent)
				return
			}
		}
	}

	err = fmt.Errorf("voter '%s' did not vote on proposalID %d", params.Voter, params.ProposalID)
	utils.WriteErrorResponse(w, http.StatusNotFound, err.Error())
}
