package gov

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestLegacyProposal(t *testing.T) {
	mapp, keeper, _, _, _, _ := GetMockApp(t, 0, GenesisState{}, nil)

	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	mapp.InitChainer(ctx, abci.RequestInitChain{})

	ti := time.Unix(10000, 100)
	lp := legacyProposal{
		ProposalID:       1,
		Title:            "title",
		Description:      "description",
		ProposalType:     legacyProposalTypeText,
		Status:           StatusVotingPeriod,
		FinalTallyResult: TallyResult{Yes: sdk.NewInt(3), No: sdk.NewInt(5)},
		SubmitTime:       ti,
		DepositEndTime:   ti,
		VotingStartTime:  ti,
		VotingEndTime:    ti,
		TotalDeposit:     sdk.NewCoins(sdk.NewCoin("mycoin", sdk.NewInt(333))),
	}

	store := ctx.KVStore(keeper.storeKey)
	bz := legacyCdc.MustMarshalBinaryLengthPrefixed(lp)
	expbz, err := hex.DecodeString(
		"57ACCBA2DE080112057469746C651A0B6465736372697074696F6E20012802320C0A01331201301A01352201303A0508904E1064420508904E10644A0D0A066D79636F696E1203333333520508904E10645A0508904E1064",
	)
	require.NoError(t, err)
	require.Equal(t, expbz, bz)

	store.Set(KeyProposal(1), bz)

	proposal, ok := keeper.GetProposal(ctx, 1)
	require.True(t, ok)
	require.True(t, ProposalEqual(proposalFromLegacy(lp), proposal))
}
