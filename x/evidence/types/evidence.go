package types

import (
	"fmt"
	"time"

	"cosmossdk.io/x/evidence/exported"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Evidence type constants
const RouteEquivocation = "equivocation"

var _ exported.Evidence = &Equivocation{}

// Route returns the Evidence Handler route for an Equivocation type.
func (e *Equivocation) Route() string { return RouteEquivocation }

// Hash returns the hash of an Equivocation object.
func (e *Equivocation) Hash() cmtbytes.HexBytes {
	bz, err := e.Marshal()
	if err != nil {
		panic(err)
	}
	return tmhash.Sum(bz)
}

// ValidateBasic performs basic stateless validation checks on an Equivocation object.
func (e *Equivocation) ValidateBasic() error {
	if e.Time.Unix() <= 0 {
		return fmt.Errorf("invalid equivocation time: %s", e.Time)
	}
	if e.Height < 1 {
		return fmt.Errorf("invalid equivocation height: %d", e.Height)
	}
	if e.Power < 1 {
		return fmt.Errorf("invalid equivocation validator power: %d", e.Power)
	}
	if e.ConsensusAddress == "" {
		return fmt.Errorf("invalid equivocation validator consensus address: %s", e.ConsensusAddress)
	}

	return nil
}

// GetConsensusAddress returns the validator's consensus address at time of the
// Equivocation infraction.
func (e Equivocation) GetConsensusAddress() sdk.ConsAddress {
	addr, _ := sdk.ConsAddressFromBech32(e.ConsensusAddress)
	return addr
}

// GetHeight returns the height at time of the Equivocation infraction.
func (e Equivocation) GetHeight() int64 {
	return e.Height
}

// GetTime returns the time at time of the Equivocation infraction.
func (e Equivocation) GetTime() time.Time {
	return e.Time
}

// GetValidatorPower returns the validator's power at time of the Equivocation
// infraction.
func (e Equivocation) GetValidatorPower() int64 {
	return e.Power
}

// GetTotalPower is a no-op for the Equivocation type.
func (e Equivocation) GetTotalPower() int64 { return 0 }

// FromABCIEvidence converts a Tendermint concrete Evidence type to
// SDK Evidence using Equivocation as the concrete type.
func FromABCIEvidence(e abci.Misbehavior) exported.Evidence {
	bech32PrefixConsAddr := sdk.GetConfig().GetBech32ConsensusAddrPrefix()
	consAddr, err := sdk.Bech32ifyAddressBytes(bech32PrefixConsAddr, e.Validator.Address)
	if err != nil {
		panic(err)
	}

	return &Equivocation{
		Height:           e.Height,
		Power:            e.Validator.Power,
		ConsensusAddress: consAddr,
		Time:             e.Time,
	}
}
