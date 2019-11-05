package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cmn "github.com/tendermint/tendermint/libs/common"
)

// Evidence defines the contract which concrete evidence types of misbehavior
// must implement.
type Evidence interface {
	Route() string
	Type() string
	String() string
	Hash() cmn.HexBytes
	ValidateBasic() error

	// The consensus address of the malicious validator at time of infraction
	GetConsensusAddress() sdk.ConsAddress

	// Height at which the infraction occurred
	GetHeight() int64

	// The total power of the malicious validator at time of infraction
	GetValidatorPower() int64

	// The total validator set power at time of infraction
	GetTotalPower() int64
}
