package offchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// interface implementation assertions
var _ msg = &MsgSignData{}

// Route defines the route to use for sdk.Msg.Route() implementation for offchain messages
const Route = "offchain"

// msg defines an off-chain msg this exists so that offchain verification
// procedures are only applied to transactions lying in this package.
// TODO: making this exported would allow external types to use the SignatureVerifier
type msg interface {
	sdk.Msg
	offchain()
}

// NewMsgSignData is MsgSignData's constructor
func NewMsgSignData(signer sdk.AccAddress, data []byte) *MsgSignData {
	return &MsgSignData{
		Signer: signer,
		Data:   data,
	}
}

// sdk.Msg implementation

func (m *MsgSignData) Route() string {
	return Route
}

func (m *MsgSignData) Type() string {
	return "MsgSignData"
}

func (m *MsgSignData) ValidateBasic() error {
	if m.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty signer")
	}
	if len(m.Data) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty data")
	}
	return nil
}

func (m *MsgSignData) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m *MsgSignData) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Signer}
}

func (m *MsgSignData) offchain() {}
