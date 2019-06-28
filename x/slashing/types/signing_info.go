package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Signing info for a validator
type ValidatorSigningInfo struct {
	Address             sdk.ConsAddress `json:"address"`               // validator consensus address
	LastMissHeight      int64           `json:"last_miss_height"`      // height at which the validator last missed a block
	JailedUntil         time.Time       `json:"jailed_until"`          // timestamp validator cannot be unjailed until
	Tombstoned          bool            `json:"tombstoned"`            // whether or not a validator has been tombstoned (killed out of validator set)
	MissedBlocksCounter int64           `json:"missed_blocks_counter"` // missed blocks counter (to avoid scanning the array every time)
}

// Construct a new `ValidatorSigningInfo` struct
func NewValidatorSigningInfo(
	condAddr sdk.ConsAddress, indexOffset int64,
	jailedUntil time.Time, tombstoned bool, missedBlocksCounter int64,
) ValidatorSigningInfo {

	return ValidatorSigningInfo{
		Address:             condAddr,
		JailedUntil:         jailedUntil,
		Tombstoned:          tombstoned,
		MissedBlocksCounter: missedBlocksCounter,
	}
}

// Return human readable signing info
func (i ValidatorSigningInfo) String() string {
	return fmt.Sprintf(`Validator Signing Info:
  Address:               %s
  Jailed Until:          %v
  Tombstoned:            %t
  Missed Blocks Counter: %d`,
		i.Address, i.JailedUntil,
		i.Tombstoned, i.MissedBlocksCounter)
}
