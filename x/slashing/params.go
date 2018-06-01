package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// MaxEvidenceAge - Max age for evidence - 21 days (3 weeks)
	// TODO Should this be a governance parameter or just modifiable with SoftwareUpgradeProposals?
	// MaxEvidenceAge = 60 * 60 * 24 * 7 * 3
	// TODO Temporarily set to 2 minutes for testnets.
	MaxEvidenceAge int64 = 60 * 2

	// SignedBlocksWindow - sliding window for downtime slashing
	// TODO Governance parameter?
	// TODO Temporarily set to 100 blocks for testnets
	SignedBlocksWindow int64 = 100

	// Downtime slashing threshold - 50%
	// TODO Governance parameter?
	MinSignedPerWindow int64 = SignedBlocksWindow / 2

	// Downtime unbond duration
	// TODO Governance parameter?
	// TODO Temporarily set to 10 minutes for testnets
	DowntimeUnbondDuration int64 = 60 * 10
)

var (
	// SlashFractionDoubleSign - currently 5%
	// TODO Governance parameter?
	SlashFractionDoubleSign = sdk.NewRat(1).Quo(sdk.NewRat(20))

	// SlashFractionDowntime - currently 1%
	// TODO Governance parameter?
	SlashFractionDowntime = sdk.NewRat(1).Quo(sdk.NewRat(100))
)
