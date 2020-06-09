package types

import (
	"fmt"
)

func DecodeMultisignatures(bz []byte) ([][]byte, error) {
	multisig := MultiSignature{}
	err := multisig.Unmarshal(bz)
	if err != nil {
		return nil, err
	}
	// unrecognized fields must be discarded because otherwise this would present a transaction malleability issue
	// which could allow transactions to be bloated with arbitrary bytes
	if len(multisig.XXX_unrecognized) > 0 {
		return nil, fmt.Errorf("rejecting unrecognized fields found in MultiSignature")
	}
	return multisig.Signatures, nil
}
