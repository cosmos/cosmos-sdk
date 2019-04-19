package legacy

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

func TestProposalMigration(t *testing.T) {
	ti := time.Unix(10000, 100)
	lp := Proposal{
		ProposalID:       1,
		Title:            "title",
		Description:      "description",
		ProposalType:     ProposalTypeText,
		Status:           gov.StatusVotingPeriod,
		FinalTallyResult: gov.TallyResult{Yes: sdk.NewInt(3), No: sdk.NewInt(5)},
		SubmitTime:       ti,
		DepositEndTime:   ti,
		VotingStartTime:  ti,
		VotingEndTime:    ti,
		TotalDeposit:     sdk.NewCoins(sdk.NewCoin("mycoin", sdk.NewInt(333))),
	}

	bz := cdc.MustMarshalBinaryLengthPrefixed(lp)
	expbz, err := hex.DecodeString(
		"57ACCBA2DE080112057469746C651A0B6465736372697074696F6E20012802320C0A01331201301A01352201303A0508904E1064420508904E10644A0D0A066D79636F696E1203333333520508904E10645A0508904E1064",
	)
	require.NoError(t, err)
	require.Equal(t, expbz, bz)

	str := string(cdc.MustMarshalJSON(lp))
	expstr := `{"type":"gov/TextProposal","value":{"proposal_id":"1","title":"title","description":"description","proposal_type":"Text","proposal_status":"VotingPeriod","final_tally_result":{"yes":"3","abstain":"0","no":"5","no_with_veto":"0"},"submit_time":"1970-01-01T02:46:40.0000001Z","deposit_end_time":"1970-01-01T02:46:40.0000001Z","total_deposit":[{"denom":"mycoin","amount":"333"}],"voting_start_time":"1970-01-01T02:46:40.0000001Z","voting_end_time":"1970-01-01T02:46:40.0000001Z"}}`
	require.Equal(t, expstr, str)

	p := gov.Proposal{
		Content:          gov.NewTextProposal("title", "description"),
		ProposalID:       1,
		Status:           gov.StatusVotingPeriod,
		FinalTallyResult: gov.TallyResult{Yes: sdk.NewInt(3), No: sdk.NewInt(5)},
		SubmitTime:       ti,
		DepositEndTime:   ti,
		VotingStartTime:  ti,
		VotingEndTime:    ti,
		TotalDeposit:     sdk.NewCoins(sdk.NewCoin("mycoin", sdk.NewInt(333))),
	}

	mp, err := lp.Migrate()
	require.NoError(t, err)
	require.True(t, gov.ProposalEqual(p, mp))
}