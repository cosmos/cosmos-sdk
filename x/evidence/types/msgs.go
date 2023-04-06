package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/evidence/exported"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

var (
	_ sdk.Msg                       = &MsgSubmitEvidence{}
	_ legacytx.LegacyMsg            = &MsgSubmitEvidence{}
	_ types.UnpackInterfacesMessage = MsgSubmitEvidence{}
	_ exported.MsgSubmitEvidenceI   = &MsgSubmitEvidence{}
)

// NewMsgSubmitEvidence returns a new MsgSubmitEvidence with a signer/submitter.
func NewMsgSubmitEvidence(s sdk.AccAddress, evi exported.Evidence) (*MsgSubmitEvidence, error) {
	msg, ok := evi.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", evi)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}
	return &MsgSubmitEvidence{Submitter: s.String(), Evidence: any}, nil
}

// ValidateBasic performs basic (non-state-dependant) validation on a MsgSubmitEvidence.
func (m MsgSubmitEvidence) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Submitter); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid submitter address: %s", err)
	}

	evi := m.GetEvidence()
	if evi == nil {
		return errorsmod.Wrap(ErrInvalidEvidence, "missing evidence")
	}
	if err := evi.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the raw bytes a signer is expected to sign when submitting
// a MsgSubmitEvidence message.
func (m MsgSubmitEvidence) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the single expected signer for a MsgSubmitEvidence.
func (m MsgSubmitEvidence) GetSigners() []sdk.AccAddress {
	submitter, _ := sdk.AccAddressFromBech32(m.Submitter)
	return []sdk.AccAddress{submitter}
}

func (m MsgSubmitEvidence) GetEvidence() exported.Evidence {
	evi, ok := m.Evidence.GetCachedValue().(exported.Evidence)
	if !ok {
		return nil
	}
	return evi
}

func (m MsgSubmitEvidence) GetSubmitter() sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		return nil
	}
	return accAddr
}

func (m MsgSubmitEvidence) UnpackInterfaces(ctx types.AnyUnpacker) error {
	var evi exported.Evidence
	return ctx.UnpackAny(m.Evidence, &evi)
}
