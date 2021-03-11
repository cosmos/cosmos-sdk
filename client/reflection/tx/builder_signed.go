package tx

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type SignedBuilder struct {
	tx                     *tx.Tx
	signatureBytesProvider SignatureBytesProvider
	expectedSigners        map[string]SignerInfo
	signatures             map[string]struct{}
	chainID                string
}

func NewSignedBuilder(rawTx *tx.Tx, chainID string, expectedSigners map[string]SignerInfo) (*SignedBuilder, error) {
	// do raw tx verification for correctness
	return &SignedBuilder{
		tx:                     rawTx,
		signatureBytesProvider: defaultSigBytesProvider{},
		expectedSigners:        expectedSigners,
		signatures:             make(map[string]struct{}),
		chainID:                chainID,
	}, nil
}

func (s *SignedBuilder) SetSignature(signer cryptotypes.PubKey, signature []byte) error {
	if _, exists := s.expectedSigners[signer.String()]; !exists {
		return fmt.Errorf("unexpected signer provided")
	}

	if _, exists := s.signatures[signer.String()]; exists {
		return fmt.Errorf("signer already set")
	}

	s.tx.Signatures = append(s.tx.Signatures, signature)
	s.signatures[signer.String()] = struct{}{}

	return nil
}

func (s *SignedBuilder) BytesToSign(signer cryptotypes.PubKey) ([]byte, error) {
	sigInfo, exists := s.expectedSigners[signer.String()]
	if !exists {
		return nil, fmt.Errorf("unexpected signer provided")
	}

	bytesToSign, err := s.signatureBytesProvider.GetSignBytes(sigInfo.SignMode, s.tx, sigInfo.AccountNumber, sigInfo.Sequence, s.chainID)
	if err != nil {
		return nil, err
	}

	return bytesToSign, nil
}

func (s *SignedBuilder) Bytes() ([]byte, error) {
	return s.tx.Marshal()
}
