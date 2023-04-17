package tx

import (
	protov2 "google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxGasWanted defines the max gas allowed.
const MaxGasWanted = uint64((1 << 63) - 1)

// Interface implementation checks.
var (
	_, _, _, _ codectypes.UnpackInterfacesMessage = &Tx{}, &TxBody{}, &AuthInfo{}, &SignerInfo{}
)

// GetMsgs implements the GetMsgs method on sdk.Tx.
func (t *Tx) GetMsgs() []sdk.Msg {
	if t == nil || t.Body == nil {
		return nil
	}

	anys := t.Body.Messages
	res, err := GetMsgs(anys, "transaction")
	if err != nil {
		panic(err)
	}
	return res
}

// GetSigners retrieves all the signers of a tx.
// This includes all unique signers of the messages (in order),
// as well as the FeePayer (if specified and not already included).
func (t *Tx) GetSigners(cdc codec.Codec) ([][]byte, []protov2.Message, error) {
	var signers [][]byte
	seen := map[string]bool{}

	var msgsv2 []protov2.Message
	for _, msg := range t.Body.Messages {
		xs, msgv2, err := cdc.GetMsgAnySigners(msg)
		if err != nil {
			return nil, nil, err
		}

		msgsv2 = append(msgsv2, msgv2)

		for _, signer := range xs {
			if !seen[string(signer)] {
				signers = append(signers, signer)
				seen[string(signer)] = true
			}
		}
	}

	// ensure any specified fee payer is included in the required signers (at the end)
	feePayer := t.AuthInfo.Fee.Payer
	var feePayerAddr []byte
	if feePayer != "" {
		var err error
		feePayerAddr, err = cdc.InterfaceRegistry().SigningContext().AddressCodec().StringToBytes(feePayer)
		if err != nil {
			return nil, nil, err
		}
	}
	if feePayerAddr != nil && !seen[string(feePayerAddr)] {
		signers = append(signers, feePayerAddr)
		seen[string(feePayerAddr)] = true
	}

	return signers, msgsv2, nil
}

func (t *Tx) GetGas() uint64 {
	return t.AuthInfo.Fee.GasLimit
}

func (t *Tx) GetFee() sdk.Coins {
	return t.AuthInfo.Fee.Amount
}

func (t *Tx) FeePayer(cdc codec.Codec) []byte {
	feePayer := t.AuthInfo.Fee.Payer
	if feePayer != "" {
		feePayerAddr, err := cdc.InterfaceRegistry().SigningContext().AddressCodec().StringToBytes(feePayer)
		if err != nil {
			panic(err)
		}
		return feePayerAddr
	}
	// use first signer as default if no payer specified
	signers, _, err := t.GetSigners(cdc)
	if err != nil {
		panic(err)
	}

	return signers[0]
}

func (t *Tx) FeeGranter() sdk.AccAddress {
	feePayer := t.AuthInfo.Fee.Granter
	if feePayer != "" {
		return sdk.MustAccAddressFromBech32(feePayer)
	}
	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (t *Tx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if t.Body != nil {
		if err := t.Body.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}

	if t.AuthInfo != nil {
		return t.AuthInfo.UnpackInterfaces(unpacker)
	}

	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (m *TxBody) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if err := UnpackInterfaces(unpacker, m.Messages); err != nil {
		return err
	}

	if err := unpackTxExtensionOptionsI(unpacker, m.ExtensionOptions); err != nil {
		return err
	}

	if err := unpackTxExtensionOptionsI(unpacker, m.NonCriticalExtensionOptions); err != nil {
		return err
	}

	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (m *AuthInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, signerInfo := range m.SignerInfos {
		err := signerInfo.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (m *SignerInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(m.PublicKey, new(cryptotypes.PubKey))
}

// RegisterInterfaces registers the sdk.Tx and MsgResponse interfaces.
// Note: the registration of sdk.Msg is done in sdk.RegisterInterfaces, but it
// could be moved inside this function.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(msgResponseInterfaceProtoName, (*MsgResponse)(nil))

	registry.RegisterInterface("cosmos.tx.v1beta1.Tx", (*sdk.Tx)(nil))

	registry.RegisterInterface("cosmos.tx.v1beta1.TxExtensionOptionI", (*ExtensionOptionI)(nil))
}
