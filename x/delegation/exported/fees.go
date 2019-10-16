package exported

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FeeAllowance implementations are tied to a given delegator and delegatee,
// and are used to enforce limits on this payment.
type FeeAllowance interface {
	// Accept can use fee payment requested as well as timestamp/height of the current block
	// to determine whether or not to process this.
	// If it returns an error, the fee payment is rejected, otherwise it is accepted.
	// The FeeAllowance implementation is expected to update it's internal state
	// and will be saved again after an acceptance.
	// If remove is true (regardless of the error), the FeeAllowance will be deleted from storage
	// (eg. when it expires)
	Accept(fee sdk.Coins, blockTime time.Time, blockHeight int64) (remove bool, err error)

	// ValidateBasic should evaluate this FeeAllowance for internal consistency.
	// Don't allow negative amounts, or negative periods for example.
	ValidateBasic() error
}
