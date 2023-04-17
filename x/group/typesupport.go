package group

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/x/group/errors"
)

// MemberRequests defines a repeated slice of MemberRequest objects.
type MemberRequests struct {
	Members []MemberRequest
}

// ValidateBasic performs stateless validation on an array of members. On top
// of validating each member individually, it also makes sure there are no
// duplicate addresses.
func (ms MemberRequests) ValidateBasic() error {
	index := make(map[string]struct{}, len(ms.Members))
	for i := range ms.Members {
		member := ms.Members[i]
		if err := member.Validate(); err != nil {
			return err
		}
		addr := member.Address
		if _, exists := index[addr]; exists {
			return errorsmod.Wrapf(errors.ErrDuplicate, "address: %s", addr)
		}
		index[addr] = struct{}{}
	}
	return nil
}
