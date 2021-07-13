package offchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// interface implementation assertions
var _ msg = &MsgSignData{}

const (
	// ExpectedChainID defines the chain id an off-chain message must have
	ExpectedChainID = ""
	// ExpectedAccountNumber defines the account number an off-chain message must have
	ExpectedAccountNumber = 0
	// ExpectedSequence defines the sequence number an off-chain message must have
	ExpectedSequence = 0
	// ExpectedRoute defines the route to use for sdk.Msg.ExpectedRoute() implementation for offchain messages
	ExpectedRoute = "offchain"
)

// msg defines an off-chain msg this exists so that offchain verification
// procedures are only applied to transactions lying in this package.
// TODO: making this exported would allow external types to use the SignatureVerifier and Signer
type msg interface {
	sdk.Msg
	offchain()
}

// NewMsgSignData is MsgSignData's constructor
func NewMsgSignData(signer sdk.AccAddress, data []byte) *MsgSignData {
	return &MsgSignData{
		Signer: signer.String(),
		Data:   data,
	}
}

// sdk.Msg implementation

func (m *MsgSignData) Route() string {
	return ExpectedRoute
}

func (m *MsgSignData) Type() string {
	return "MsgSignData"
}

func (m *MsgSignData) ValidateBasic() error {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid signer: %s", err.Error())
	}
	if signer.Empty() {
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
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (m *MsgSignData) offchain() {}
