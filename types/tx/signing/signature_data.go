package signing

import (
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

type SignatureData interface {
	isSignatureData()
}

type SingleSignatureData struct {
	SignMode  SignMode
	Signature []byte
}

type MultiSignatureData struct {
	BitArray   *types.CompactBitArray
	Signatures []SignatureData
}

var _, _ SignatureData = &SingleSignatureData{}, &MultiSignatureData{}

func (m *SingleSignatureData) isSignatureData() {}
func (m *MultiSignatureData) isSignatureData()  {}
