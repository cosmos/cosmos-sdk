package v1

import (
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type GovernorI interface {
	GetMoniker() string                  // moniker of the governor
	GetStatus() GovernorStatus           // status of the governor
	IsActive() bool                      // check if has a active status
	IsInactive() bool                    // check if has status inactive
	GetAddress() types.GovernorAddress   // governor address to receive/return governors delegations
	GetDescription() GovernorDescription // description of the governor
}
