package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// MaxEvidenceAge - Max age for evidence - 21 days (3 weeks)
	// TODO Should this be a governance parameter or just modifiable with SoftwareUpgradeProposals?
	// MaxEvidenceAge = 60 * 60 * 24 * 7 * 3
	// TODO Temporarily set to 2 minutes for testnets.
	MaxEvidenceAge int64 = 60 * 2

	// SignedBlocksWindow - sliding window for downtime slashing
	// TODO Governance parameter?
	// TODO Temporarily set to 40000 blocks for testnets
	SignedBlocksWindow int64 = 10000

	// Downtime slashing threshold - 50%
	// TODO Governance parameter?
	MinSignedPerWindow = SignedBlocksWindow / 2

	// Downtime unbond duration
	// TODO Governance parameter?
	// TODO Temporarily set to five minutes for testnets
	DowntimeUnbondDuration int64 = 60 * 5

	// Double-sign unbond duration
	// TODO Governance parameter?
	// TODO Temporarily set to five minutes for testnets
	DoubleSignUnbondDuration int64 = 60 * 5
)

var (
	// SlashFractionDoubleSign - currently 5%
	// TODO Governance parameter?
	SlashFractionDoubleSign = sdk.NewRat(1).Quo(sdk.NewRat(20))

	// SlashFractionDowntime - currently 1%
	// TODO Governance parameter?
	SlashFractionDowntime = sdk.NewRat(10).Quo(sdk.NewRat(100))
)
