package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ codectypes.UnpackInterfacesMessage = &Tx{}
var _ codectypes.UnpackInterfacesMessage = &TxBody{}
var _ codectypes.UnpackInterfacesMessage = &SignDoc{}

func (m *Tx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Body.UnpackInterfaces(unpacker)
}

func (m *SignDoc) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Body.UnpackInterfaces(unpacker)
}

func (m *TxBody) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, any := range m.Messages {
		var msg sdk.Msg
		err := unpacker.UnpackAny(any, &msg)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ sdk.Tx = &Tx{}

func (tx *Tx) GetMsgs() []sdk.Msg {
	anys := tx.Body.Messages
	res := make([]sdk.Msg, len(anys))
	for i, any := range anys {
		msg := any.GetCachedValue().(sdk.Msg)
		res[i] = msg
	}
	return res
}

func (tx *Tx) ValidateBasic() error {
	// TODO
	//stdSigs := tx.GetSignatures()
	//
	//if tx.Fee.Gas > MaxGasWanted {
	//	return sdkerrors.Wrapf(
	//		sdkerrors.ErrInvalidRequest,
	//		"invalid gas supplied; %d > %d", tx.Fee.Gas, MaxGasWanted,
	//	)
	//}
	//if tx.Fee.Amount.IsAnyNegative() {
	//	return sdkerrors.Wrapf(
	//		sdkerrors.ErrInsufficientFee,
	//		"invalid fee provided: %s", tx.Fee.Amount,
	//	)
	//}
	//if len(stdSigs) == 0 {
	//	return sdkerrors.ErrNoSignatures
	//}
	//if len(stdSigs) != len(tx.GetSigners()) {
	//	return sdkerrors.Wrapf(
	//		sdkerrors.ErrUnauthorized,
	//		"wrong number of signers; expected %d, got %d", tx.GetSigners(), len(stdSigs),
	//	)
	//}
	return nil
}
