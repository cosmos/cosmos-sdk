package signing

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	antetypes "github.com/cosmos/cosmos-sdk/x/auth/ante/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type DecodedTx struct {
	*types.Tx
	Raw     *TxRaw
	Msgs    []sdk.Msg
	PubKeys []crypto.PubKey
	Signers []sdk.AccAddress
}

var _ sdk.Tx = DecodedTx{}
var _ antetypes.FeeTx = DecodedTx{}
var _ antetypes.TxWithMemo = DecodedTx{}
var _ antetypes.HasPubKeysTx = DecodedTx{}

func DefaultTxDecoder(cdc codec.Marshaler, keyCodec cryptotypes.PublicKeyCodec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		var raw TxRaw
		err := cdc.UnmarshalBinaryBare(txBytes, &raw)
		if err != nil {
			return nil, err
		}

		var tx types.Tx
		err = cdc.UnmarshalBinaryBare(txBytes, &tx)
		if err != nil {
			return nil, err
		}

		anyMsgs := tx.Body.Messages
		msgs := make([]sdk.Msg, len(anyMsgs))
		for i, any := range anyMsgs {
			msg, ok := any.GetCachedValue().(sdk.Msg)
			if !ok {
				return nil, fmt.Errorf("can't decode sdk.Msg from %+v", any)
			}
			msgs[i] = msg
		}

		var signers []sdk.AccAddress
		seen := map[string]bool{}

		for _, msg := range msgs {
			for _, addr := range msg.GetSigners() {
				if !seen[addr.String()] {
					signers = append(signers, addr)
					seen[addr.String()] = true
				}
			}
		}

		signerInfos := tx.AuthInfo.SignerInfos
		pubKeys := make([]crypto.PubKey, len(signerInfos))
		for i, si := range signerInfos {
			pubKey, err := keyCodec.Decode(si.PublicKey)
			if err != nil {
				return nil, errors.Wrap(err, "can't decode public key")
			}
			pubKeys[i] = pubKey
		}

		return DecodedTx{
			Tx:      &tx,
			Raw:     &raw,
			Msgs:    msgs,
			PubKeys: pubKeys,
			Signers: signers,
		}, nil
	}
}

func (d DecodedTx) GetMsgs() []sdk.Msg {
	return d.Msgs
}

func (d DecodedTx) ValidateBasic() error {
	sigs := d.Signatures

	if d.GetGas() > auth.MaxGasWanted {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid gas supplied; %d > %d", d.GetGas(), auth.MaxGasWanted,
		)
	}
	if d.GetFee().IsAnyNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFee,
			"invalid fee provided: %s", d.GetFee(),
		)
	}
	if len(sigs) == 0 {
		return sdkerrors.ErrNoSignatures
	}
	if len(sigs) != len(d.GetSigners()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"wrong number of signers; expected %d, got %d", d.GetSigners(), len(sigs),
		)
	}

	return nil
}

func (d DecodedTx) GetSigners() []sdk.AccAddress {
	return d.Signers
}

func (d DecodedTx) GetPubKeys() []crypto.PubKey {
	return d.PubKeys
}

func (d DecodedTx) GetGas() uint64 {
	return d.AuthInfo.Fee.GasLimit
}

func (d DecodedTx) GetFee() sdk.Coins {
	return d.AuthInfo.Fee.Amount
}

func (d DecodedTx) FeePayer() sdk.AccAddress {
	signers := d.GetSigners()
	if signers != nil {
		return signers[0]
	}
	return sdk.AccAddress{}
}

func (d DecodedTx) GetMemo() string {
	return d.Body.Memo
}
