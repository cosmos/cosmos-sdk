// DONTCOVER

package v036

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v034"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
)

// ----------------------------------------------------------------------------
// Types and Constants
// ----------------------------------------------------------------------------

const (
	ModuleName = "distribution"

	// RouterKey is the message route for distribution
	RouterKey = ModuleName

	// ProposalTypeCommunityPoolSpend defines the type for a CommunityPoolSpendProposal
	ProposalTypeCommunityPoolSpend = "CommunityPoolSpend"
)

type (
	ValidatorAccumulatedCommission = sdk.DecCoins

	ValidatorSlashEventRecord struct {
		ValidatorAddress sdk.ValAddress                `json:"validator_address"`
		Height           uint64                        `json:"height"`
		Period           uint64                        `json:"period"`
		Event            v034distr.ValidatorSlashEvent `json:"validator_slash_event"`
	}

	GenesisState struct {
		FeePool                         v034distr.FeePool                                `json:"fee_pool"`
		CommunityTax                    sdk.Dec                                          `json:"community_tax"`
		BaseProposerReward              sdk.Dec                                          `json:"base_proposer_reward"`
		BonusProposerReward             sdk.Dec                                          `json:"bonus_proposer_reward"`
		WithdrawAddrEnabled             bool                                             `json:"withdraw_addr_enabled"`
		DelegatorWithdrawInfos          []v034distr.DelegatorWithdrawInfo                `json:"delegator_withdraw_infos"`
		PreviousProposer                sdk.ConsAddress                                  `json:"previous_proposer"`
		OutstandingRewards              []v034distr.ValidatorOutstandingRewardsRecord    `json:"outstanding_rewards"`
		ValidatorAccumulatedCommissions []v034distr.ValidatorAccumulatedCommissionRecord `json:"validator_accumulated_commissions"`
		ValidatorHistoricalRewards      []v034distr.ValidatorHistoricalRewardsRecord     `json:"validator_historical_rewards"`
		ValidatorCurrentRewards         []v034distr.ValidatorCurrentRewardsRecord        `json:"validator_current_rewards"`
		DelegatorStartingInfos          []v034distr.DelegatorStartingInfoRecord          `json:"delegator_starting_infos"`
		ValidatorSlashEvents            []ValidatorSlashEventRecord                      `json:"validator_slash_events"`
	}

	// CommunityPoolSpendProposal spends from the community pool
	CommunityPoolSpendProposal struct {
		Title       string         `json:"title" yaml:"title"`
		Description string         `json:"description" yaml:"description"`
		Recipient   sdk.AccAddress `json:"recipient" yaml:"recipient"`
		Amount      sdk.Coins      `json:"amount" yaml:"amount"`
	}
)

func NewGenesisState(
	feePool v034distr.FeePool, communityTax, baseProposerReward, bonusProposerReward sdk.Dec,
	withdrawAddrEnabled bool, dwis []v034distr.DelegatorWithdrawInfo, pp sdk.ConsAddress,
	r []v034distr.ValidatorOutstandingRewardsRecord, acc []v034distr.ValidatorAccumulatedCommissionRecord,
	historical []v034distr.ValidatorHistoricalRewardsRecord, cur []v034distr.ValidatorCurrentRewardsRecord,
	dels []v034distr.DelegatorStartingInfoRecord, slashes []ValidatorSlashEventRecord,
) GenesisState {

	return GenesisState{
		FeePool:                         feePool,
		CommunityTax:                    communityTax,
		BaseProposerReward:              baseProposerReward,
		BonusProposerReward:             bonusProposerReward,
		WithdrawAddrEnabled:             withdrawAddrEnabled,
		DelegatorWithdrawInfos:          dwis,
		PreviousProposer:                pp,
		OutstandingRewards:              r,
		ValidatorAccumulatedCommissions: acc,
		ValidatorHistoricalRewards:      historical,
		ValidatorCurrentRewards:         cur,
		DelegatorStartingInfos:          dels,
		ValidatorSlashEvents:            slashes,
	}
}

var _ v036gov.Content = CommunityPoolSpendProposal{}

// GetTitle returns the title of a community pool spend proposal.
func (csp CommunityPoolSpendProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp CommunityPoolSpendProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp CommunityPoolSpendProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp CommunityPoolSpendProposal) ProposalType() string { return ProposalTypeCommunityPoolSpend }

// ValidateBasic runs basic stateless validity checks
func (csp CommunityPoolSpendProposal) ValidateBasic() error {
	err := v036gov.ValidateAbstract(csp)
	if err != nil {
		return err
	}
	if !csp.Amount.IsValid() {
		return types.ErrInvalidProposalAmount
	}
	if csp.Recipient.Empty() {
		return types.ErrEmptyProposalRecipient
	}

	return nil
}

// String implements the Stringer interface.
func (csp CommunityPoolSpendProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Spend Proposal:
  Title:       %s
  Description: %s
  Recipient:   %s
  Amount:      %s
`, csp.Title, csp.Description, csp.Recipient, csp.Amount))
	return b.String()
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(CommunityPoolSpendProposal{}, "cosmos-sdk/CommunityPoolSpendProposal", nil)
}
