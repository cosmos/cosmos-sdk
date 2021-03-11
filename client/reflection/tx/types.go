package tx

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type SignatureBytesProvider interface {
	GetSignBytes(mode signing.SignMode, rawTx *tx.Tx, accountNumber, sequence uint64, chainID string) ([]byte, error)
}

type SignerInfo struct {
	PubKey        cryptotypes.PubKey
	SignMode      signing.SignMode
	AccountNumber uint64
	Sequence      uint64
}

// defaultSigBytesProvider aliases the direct mode signing spec, it is repeated here in order not to create
// import dependencies on the auth module. The reflection library should not depend on any module but it
// should be able to interact with any of them.
type defaultSigBytesProvider struct{}

// GetSignBytes returns the expected bytes that need to be signed in order for a transaction to be valid
func (s defaultSigBytesProvider) GetSignBytes(mode signing.SignMode, rawTx *tx.Tx, accountNumber, _ uint64, chainID string) ([]byte, error) {
	if mode != signing.SignMode_SIGN_MODE_DIRECT {
		return nil, fmt.Errorf("unsupported mode")
	}

	authInfoBytes, err := proto.Marshal(rawTx.AuthInfo)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := proto.Marshal(rawTx.Body)

	signDoc := tx.SignDoc{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		ChainId:       chainID,
		AccountNumber: accountNumber,
	}

	return signDoc.Marshal()
}
