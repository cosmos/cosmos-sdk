package types

import types "github.com/cosmos/cosmos-sdk/types/tx"

type DecodedMultisignature struct {
	ModeInfo   *types.ModeInfo_Multi
	Signatures [][]byte
}

type GetSignBytesFunc func(single *types.ModeInfo_Single) ([]byte, error)

type MultisigPubKey interface {
	VerifyMultisignature(getSignBytes GetSignBytesFunc, sig DecodedMultisignature) bool
}
