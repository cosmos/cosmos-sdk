package common

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestPrettyParams_MarshalYAML(t *testing.T) {
	communityTax := types.NewDecWithPrec(2, 2)
	baseProposalReward := types.NewDecWithPrec(1, 2)
	bonusProposalReward := types.NewDecWithPrec(4, 2)
	withdrawEnabled := false

	communityTaxBz, err := json.Marshal(communityTax)
	require.Nil(t, err)
	baseProposalRewardBz, err := json.Marshal(baseProposalReward)
	require.Nil(t, err)
	bonusProposalRewardBz, err := json.Marshal(bonusProposalReward)
	require.Nil(t, err)
	withdrawEnabledBz, err := json.Marshal(withdrawEnabled)
	require.Nil(t, err)

	pp := PrettyParams{
		CommunityTax:        communityTaxBz,
		BaseProposerReward:  baseProposalRewardBz,
		BonusProposerReward: bonusProposalRewardBz,
		WithdrawAddrEnabled: withdrawEnabledBz,
	}

	bz, err := yaml.Marshal(pp)
	require.Nil(t, err)
	expected := fmt.Sprintf(`|
  communitytax: "0.020000000000000000"
  baseproposerreward: "0.010000000000000000"
  bonusproposerreward: "0.040000000000000000"
  withdrawaddrenabled: false
`)
	require.Equal(t, expected, string(bz))

}
