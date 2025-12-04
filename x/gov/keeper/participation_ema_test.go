package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestGetSetParticipationEma(t *testing.T) {
	k, _, _, _, _, _, ctx := setupGovKeeper(t)
	assert := assert.New(t)

	participationEMA, _ := k.ParticipationEMA.Get(ctx)
	constitutionParticipationEMA, _ := k.ConstitutionAmendmentParticipationEMA.Get(ctx)
	lawParticipationEMA, _ := k.LawParticipationEMA.Get(ctx)

	assert.Equal(v1.DefaultParticipationEma, participationEMA.String())
	assert.Equal(v1.DefaultParticipationEma, constitutionParticipationEMA.String())
	assert.Equal(v1.DefaultParticipationEma, lawParticipationEMA.String())

	assert.NoError(k.ParticipationEMA.Set(ctx, math.LegacyNewDecWithPrec(1, 2)))
	assert.NoError(k.ConstitutionAmendmentParticipationEMA.Set(ctx, math.LegacyNewDecWithPrec(2, 2)))
	assert.NoError(k.LawParticipationEMA.Set(ctx, math.LegacyNewDecWithPrec(3, 2)))

	participationEMA, _ = k.ParticipationEMA.Get(ctx)
	constitutionParticipationEMA, _ = k.ConstitutionAmendmentParticipationEMA.Get(ctx)
	lawParticipationEMA, _ = k.LawParticipationEMA.Get(ctx)

	assert.Equal(math.LegacyNewDecWithPrec(1, 2).String(), participationEMA.String())
	assert.Equal(math.LegacyNewDecWithPrec(2, 2).String(), constitutionParticipationEMA.String())
	assert.Equal(math.LegacyNewDecWithPrec(3, 2).String(), lawParticipationEMA.String())

	assert.Equal(math.LegacyNewDecWithPrec(104, 3).String(), k.GetQuorum(ctx).String())
	assert.Equal(math.LegacyNewDecWithPrec(108, 3).String(), k.GetConstitutionAmendmentQuorum(ctx).String())
	assert.Equal(math.LegacyNewDecWithPrec(112, 3).String(), k.GetLawQuorum(ctx).String())
}

func TestUpdateParticipationEma(t *testing.T) {
	tests := []struct {
		name                                        string
		proposal                                    v1.Proposal
		expectedParticipationEma                    string
		expectedConstitutionAmdmentParticipationEma string
		expectedLawParticipationEma                 string
	}{
		{
			name:                     "proposal w/o message",
			proposal:                 v1.Proposal{},
			expectedParticipationEma: math.LegacyNewDecWithPrec(41, 2).String(),
			expectedConstitutionAmdmentParticipationEma: v1.DefaultParticipationEma,
			expectedLawParticipationEma:                 v1.DefaultParticipationEma,
		},
		{
			name:                     "proposal with propose law message",
			proposal:                 v1.Proposal{Messages: setMsgs(t, []sdk.Msg{&v1.MsgProposeLaw{}})},
			expectedParticipationEma: v1.DefaultParticipationEma,
			expectedConstitutionAmdmentParticipationEma: v1.DefaultParticipationEma,
			expectedLawParticipationEma:                 math.LegacyNewDecWithPrec(41, 2).String(),
		},
		{
			name:                     "proposal with propose constitution amendment message",
			proposal:                 v1.Proposal{Messages: setMsgs(t, []sdk.Msg{&v1.MsgProposeConstitutionAmendment{}})},
			expectedParticipationEma: v1.DefaultParticipationEma,
			expectedConstitutionAmdmentParticipationEma: math.LegacyNewDecWithPrec(41, 2).String(),
			expectedLawParticipationEma:                 v1.DefaultParticipationEma,
		},
		{
			name: "proposal with all kinds of messages",
			proposal: v1.Proposal{Messages: setMsgs(t, []sdk.Msg{
				&v1.MsgProposeConstitutionAmendment{},
				&v1.MsgProposeLaw{},
				&banktypes.MsgSend{},
			})},
			expectedParticipationEma:                    math.LegacyNewDecWithPrec(41, 2).String(),
			expectedConstitutionAmdmentParticipationEma: math.LegacyNewDecWithPrec(41, 2).String(),
			expectedLawParticipationEma:                 math.LegacyNewDecWithPrec(41, 2).String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			k, _, _, _, _, _, ctx := setupGovKeeper(t)

			participationEMA, _ := k.ParticipationEMA.Get(ctx)
			constitutionParticipationEMA, _ := k.ConstitutionAmendmentParticipationEMA.Get(ctx)
			lawParticipationEMA, _ := k.LawParticipationEMA.Get(ctx)

			assert.Equal(v1.DefaultParticipationEma, participationEMA.String())
			assert.Equal(v1.DefaultParticipationEma, constitutionParticipationEMA.String())
			assert.Equal(v1.DefaultParticipationEma, lawParticipationEMA.String())
			newParticipation := math.LegacyNewDecWithPrec(5, 2) // 5% participation

			k.UpdateParticipationEMA(ctx, tt.proposal, newParticipation)

			participationEMA, _ = k.ParticipationEMA.Get(ctx)
			constitutionParticipationEMA, _ = k.ConstitutionAmendmentParticipationEMA.Get(ctx)
			lawParticipationEMA, _ = k.LawParticipationEMA.Get(ctx)

			assert.Equal(tt.expectedParticipationEma, participationEMA.String())
			assert.Equal(tt.expectedConstitutionAmdmentParticipationEma, constitutionParticipationEMA.String())
			assert.Equal(tt.expectedLawParticipationEma, lawParticipationEMA.String())
		})
	}
}
