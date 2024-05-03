package gov_test

import (
	"cosmossdk.io/math"
	_ "cosmossdk.io/x/consensus"
	"cosmossdk.io/x/gov/types/v1beta1"
	stakingtypes "cosmossdk.io/x/staking/types"
)

var (
	TestProposal        = v1beta1.NewTextProposal("Test", "description")
	TestDescription     = stakingtypes.NewDescription("T", "E", "S", "T", "Z")
	TestCommissionRates = stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
)
