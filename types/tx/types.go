package tx

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MaxGasWanted defines the max gas allowed.
const MaxGasWanted = uint64((1 << 63) - 1)

// Interface implementation checks.
var _, _, _, _ codectypes.UnpackInterfacesMessage = &Tx{}, &TxBody{}, &AuthInfo{}, &SignerInfo{}
var _ sdk.Tx = &Tx{}

// GetMsgs implements the GetMsgs method on sdk.Tx.
func (t *Tx) GetMsgs() []sdk.Msg {
	if t == nil || t.Body == nil {
		return nil
	}

	anys := t.Body.Messages
	res := make([]sdk.Msg, len(anys))
	for i, any := range anys {
		res[i] = any.GetCachedValue().(sdk.Msg)
	}
	return res
}

// ValidateBasic implements the ValidateBasic method on sdk.Tx.
func (t *Tx) ValidateBasic() error {
	if t == nil {
		return fmt.Errorf("bad Tx")
	}

	body := t.Body
	if body == nil {
		return fmt.Errorf("missing TxBody")
	}

	authInfo := t.AuthInfo
	if authInfo == nil {
		return fmt.Errorf("missing AuthInfo")
	}

	fee := authInfo.Fee
	if fee == nil {
		return fmt.Errorf("missing fee")
	}

	if fee.GasLimit > MaxGasWanted {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid gas supplied; %d > %d", fee.GasLimit, MaxGasWanted,
		)
	}

	if fee.Amount.IsAnyNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFee,
			"invalid fee provided: %s", fee.Amount,
		)
	}

	if fee.Payer != "" {
		_, err := sdk.AccAddressFromBech32(fee.Payer)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid fee payer address (%s)", err)
		}
	}

	sigs := t.Signatures

	if len(sigs) == 0 {
		return sdkerrors.ErrNoSignatures
	}

	if len(sigs) != len(t.GetSigners()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"wrong number of signers; expected %d, got %d", len(t.GetSigners()), len(sigs),
		)
	}

	return nil
}

// GetSigners retrieves all the signers of a tx.
// This includes all unique signers of the messages (in order),
// as well as the FeePayer (if specified and not already included).
func (t *Tx) GetSigners() []sdk.AccAddress {
	var signers []sdk.AccAddress
	seen := map[string]bool{}

	for _, msg := range t.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			if !seen[addr.String()] {
				signers = append(signers, addr)
				seen[addr.String()] = true
			}
		}
	}

	// ensure any specified fee payer is included in the required signers (at the end)
	feePayer := t.AuthInfo.Fee.Payer
	if feePayer != "" && !seen[feePayer] {
		payerAddr, err := sdk.AccAddressFromBech32(feePayer)
		if err != nil {
			panic(err)
		}
		signers = append(signers, payerAddr)
		seen[feePayer] = true
	}

	return signers
}

func (t *Tx) GetGas() uint64 {
	return t.AuthInfo.Fee.GasLimit
}
func (t *Tx) GetFee() sdk.Coins {
	return t.AuthInfo.Fee.Amount
}
func (t *Tx) FeePayer() sdk.AccAddress {
	feePayer := t.AuthInfo.Fee.Payer
	if feePayer != "" {
		payerAddr, err := sdk.AccAddressFromBech32(feePayer)
		if err != nil {
			panic(err)
		}
		return payerAddr
	}
	// use first signer as default if no payer specified
	return t.GetSigners()[0]
}

func (t *Tx) FeeGranter() sdk.AccAddress {
	feePayer := t.AuthInfo.Fee.Granter
	if feePayer != "" {
		granterAddr, err := sdk.AccAddressFromBech32(feePayer)
		if err != nil {
			panic(err)
		}
		return granterAddr
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

// isServiceMsg checks if a type URL corresponds to a service method name,
// i.e. /cosmos.bank.Msg/Send vs /cosmos.bank.MsgSend
func isServiceMsg(typeURL string) bool {
	return strings.Count(typeURL, "/") >= 2
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (m *TxBody) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, any := range m.Messages {
		// We gracefully keep support for ServiceMsgs, even though they are
		// deprecated.
		// TODO Remove ServiceMsgs in https://github.com/cosmos/cosmos-sdk/issues/9172.
		if isServiceMsg(any.TypeUrl) {
			// Recall that a ServiceMsg is packed inside an `Any`, its Any.Value
			// being the encoded bytes of another `Any` packing the request
			// parameter (a skd.Msg). So there's an `Any` packed inside another
			// `Any`.
			// Our strategy here is to:
			// - create an `Any` called `requestAny`, will represents the request
			//   parameter,
			// - unmarshal the ServiceMsg `Any`'s `Any.Value` into `requestAny`,
			// - then use the interface registry to unpack `requestAny` into a
			//   sdk.Msg.
			//
			// The initial ServiceMsg `Any` is replaced by the request parameter
			// `sdk.Msg` `Any`.
			var requestAny codectypes.Any
			err := proto.Unmarshal(any.Value, &requestAny)
			if err != nil {
				return err
			}

			var msg sdk.Msg
			err = unpacker.UnpackAny(&requestAny, &msg)
			if err != nil {
				return err
			}

			*any = requestAny
		} else {
			var msg sdk.Msg
			err := unpacker.UnpackAny(any, &msg)
			if err != nil {
				return err
			}
		}

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

// RegisterInterfaces registers the sdk.Tx interface.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface("cosmos.tx.v1beta1.Tx", (*sdk.Tx)(nil))
	registry.RegisterImplementations((*sdk.Tx)(nil), &Tx{})
}
