package tx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

var (
	_ transaction.Tx = wrappedTx{}
	_ Tx             = wrappedTx{}
)

// wrappedTx wraps a transaction and provides a codec for binary encoding/decoding.
type wrappedTx struct {
	*decode.DecodedTx

	cdc codec.BinaryCodec
}

func newWrapperTx(cdc codec.BinaryCodec, decodedTx *decode.DecodedTx) *wrappedTx {
	return &wrappedTx{
		DecodedTx: decodedTx,
		cdc:       cdc,
	}
}

// GetSigners fetches the addresses of the signers of the transaction.
func (w wrappedTx) GetSigners() ([][]byte, error) {
	return w.Signers, nil
}

// GetPubKeys retrieves the public keys of the signers from the transaction's SignerInfos.
func (w wrappedTx) GetPubKeys() ([]cryptotypes.PubKey, error) {
	signerInfos := w.Tx.AuthInfo.SignerInfos
	pks := make([]cryptotypes.PubKey, len(signerInfos))

	for i, si := range signerInfos {
		// NOTE: it is okay to leave this nil if there is no PubKey in the SignerInfo.
		// PubKey's can be left unset in SignerInfo.
		if si.PublicKey == nil {
			continue
		}
		maybePk, err := w.decodeAny(si.PublicKey)
		if err != nil {
			return nil, err
		}
		pk, ok := maybePk.(cryptotypes.PubKey)
		if !ok {
			return nil, fmt.Errorf("invalid public key type: %T", maybePk)
		}
		pks[i] = pk
	}

	return pks, nil
}

// GetSignatures fetches the signatures attached to the transaction.
func (w wrappedTx) GetSignatures() ([]Signature, error) {
	signerInfos := w.Tx.AuthInfo.SignerInfos
	sigs := w.Tx.Signatures

	pubKeys, err := w.GetPubKeys()
	if err != nil {
		return nil, err
	}
	signatures := make([]Signature, len(sigs))

	for i, si := range signerInfos {
		if si.ModeInfo == nil || si.ModeInfo.Sum == nil {
			signatures[i] = Signature{
				PubKey: pubKeys[i],
			}
		} else {
			sigData, err := modeInfoAndSigToSignatureData(si.ModeInfo, sigs[i])
			if err != nil {
				return nil, err
			}
			signatures[i] = Signature{
				PubKey:   pubKeys[i],
				Data:     sigData,
				Sequence: si.GetSequence(),
			}
		}
	}

	return signatures, nil
}

func (w wrappedTx) GetSigningTxData() (signing.TxData, error) {
	return signing.TxData{
		Body:                       w.Tx.Body,
		AuthInfo:                   w.Tx.AuthInfo,
		BodyBytes:                  w.TxRaw.BodyBytes,
		AuthInfoBytes:              w.TxRaw.AuthInfoBytes,
		BodyHasUnknownNonCriticals: w.TxBodyHasUnknownNonCriticals,
	}, nil
}

// decodeAny decodes a protobuf Any message into a concrete proto.Message.
func (w wrappedTx) decodeAny(anyPb *anypb.Any) (proto.Message, error) {
	name := anyPb.GetTypeUrl()
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		name = name[i+len("/"):]
	}
	typ := proto.MessageType(name)
	if typ == nil {
		return nil, fmt.Errorf("unknown type: %s", name)
	}
	v1 := reflect.New(typ.Elem()).Interface().(proto.Message)
	err := w.cdc.Unmarshal(anyPb.GetValue(), v1)
	if err != nil {
		return nil, err
	}
	return v1, nil
}
