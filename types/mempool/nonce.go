package mempool

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ChooseNonce gets the nonce from a transaction. If the transaction is unordered,
// it uses the timeout timestamp as the nonce. Sequence values must be zero in this case.
// If the transaction is ordered, it uses the sequence number as the nonce.
func ChooseNonce(seq uint64, tx sdk.Tx) (uint64, error) {
	// if it's an unordered tx, we use the timeout timestamp instead of the nonce
	if unordered, ok := tx.(sdk.TxWithUnordered); ok && unordered.GetUnordered() {
		if seq > 0 {
			return 0, errors.New("unordered txs must not have sequence set")
		}
		timestamp := unordered.GetTimeoutTimeStamp().UnixNano()
		if timestamp < 0 {
			return 0, errors.New("invalid timestamp value")
		}
		return uint64(timestamp), nil
	}
	// otherwise, use the sequence as normal.
	return seq, nil
}
