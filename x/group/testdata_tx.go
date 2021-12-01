package group

import (
	"bytes"

	"github.com/gogo/protobuf/jsonpb"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgAuthenticated{}

func (m MsgAuthenticated) Route() string { return "MsgAuthenticated" }

func (m MsgAuthenticated) Type() string { return "MsgAuthenticated" }

// GetSignBytes returns the bytes for the message signer to sign on
func (m MsgAuthenticated) GetSignBytes() []byte {
	var buf bytes.Buffer
	enc := jsonpb.Marshaler{}
	if err := enc.Marshal(&buf, &m); err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(buf.Bytes())
}

// ValidateBasic does a sanity check on the provided data
func (m MsgAuthenticated) ValidateBasic() error {
	return nil
}

// GetSigners returns the expected signers for a MsgAuthenticated.
func (m MsgAuthenticated) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}
