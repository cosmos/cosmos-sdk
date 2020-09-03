package tx

import codectypes "github.com/cosmos/cosmos-sdk/codec/types"

// IsAnyTx represents a tx that can be wrapped into an Any.
type IsAnyTx interface {
	GetAnyTx() *codectypes.Any
}
