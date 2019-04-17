package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParamChange defines a parameter change.
type ParamChange struct {
	Subspace string `json:"subspace"`
	Key      string `json:"key"`
	Subkey   []byte `json:"subkey"`
	Value    []byte `json:"value"`
}

func NewParamChange(space, key string, subkey, value []byte) ParamChange {
	return ParamChange{space, key, subkey, value}
}

// String implements the Stringer interface.
func (pc ParamChange) String() string {
	var subkey string
	if len(pc.Subkey) != 0 {
		subkey = fmt.Sprintf("(%s)", pc.Subkey)
	}

	return fmt.Sprintf("{%s/%s := %X} (%s)", pc.Key, subkey, pc.Value, pc.Subspace)
}

// ValidateChange performs basic validation checks over a set of ParamChange. It
// returns an error if any ParamChange is invalid.
func ValidateChanges(changes []ParamChange) sdk.Error {
	if len(changes) == 0 {
		return ErrEmptyChanges(DefaultCodespace)
	}

	for _, pc := range changes {
		if len(pc.Subspace) == 0 {
			return ErrEmptySubspace(DefaultCodespace)
		}
		if len(pc.Key) == 0 {
			return ErrEmptyKey(DefaultCodespace)
		}
		if len(pc.Value) == 0 {
			return ErrEmptyValue(DefaultCodespace)
		}
	}

	return nil
}
