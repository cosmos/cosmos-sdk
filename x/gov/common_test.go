package gov_test

import (
	"cosmossdk.io/math"
	"cosmossdk.io/x/gov/types/v1beta1"
	stakingtypes "cosmossdk.io/x/staking/types"

	_ "github.com/cosmos/cosmos-sdk/x/consensus"
)

var (
	TestProposal        = v1beta1.NewTextProposal("Test", "description")
	TestDescription     = stakingtypes.NewDescription("T", "E", "S", "T", "Z")
	TestCommissionRates = stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
)
