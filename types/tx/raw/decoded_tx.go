package raw

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	types "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type DecodedTx struct {
	*types.Tx
	Raw           *TxRaw
	Msgs          []sdk.Msg
	PubKeys       []crypto.PubKey
	Signers       []sdk.AccAddress
	SignerInfoMap map[string]*types.SignerInfo
}

var _ sdk.Tx = DecodedTx{}
var _ ante.FeeTx = DecodedTx{}
var _ ante.TxWithMemo = DecodedTx{}

func DefaultTxDecoder(cdc codec.Marshaler) sdk.TxDecoder {
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

		nSigners := len(signers)
		signerInfos := tx.AuthInfo.SignerInfos
		signerInfoMap := map[string]*types.SignerInfo{}
		pubKeys := make([]crypto.PubKey, len(signerInfos))
		for i, si := range signerInfos {
			any := si.PublicKey
			pubKey, ok := any.GetCachedValue().(crypto.PubKey)
			if !ok {
				return nil, fmt.Errorf("can't decode PublicKey from %+v", any)
			}
			pubKeys[i] = pubKey

			if i < nSigners {
				signerInfoMap[signers[i].String()] = si
			}
		}

		return DecodedTx{
			Tx:            &tx,
			Raw:           &raw,
			Msgs:          msgs,
			PubKeys:       pubKeys,
			Signers:       signers,
			SignerInfoMap: signerInfoMap,
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

func (d DecodedTx) GetSignBytes(ctx sdk.Context, acc authexported.Account) ([]byte, error) {
	address := acc.GetAddress()
	signerInfo, ok := d.SignerInfoMap[address.String()]
	if !ok {
		return nil, fmt.Errorf("missing SignerInfo for address %s", address.String())
	}

	genesis := ctx.BlockHeight() == 0
	chainID := ctx.ChainID()
	var accNum uint64
	if !genesis {
		accNum = acc.GetAccountNumber()
	}

	switch modeInfo := signerInfo.ModeInfo.Sum.(type) {
	case *types.ModeInfo_Single_:
		switch modeInfo.Single.Mode {
		case types.SignMode_SIGN_MODE_UNSPECIFIED:
			return nil, fmt.Errorf("unspecified sign mode")
		case types.SignMode_SIGN_MODE_DIRECT:
			return DirectSignBytes(d.Raw.BodyBytes, d.Raw.AuthInfoBytes, chainID, accNum, acc.GetSequence())
		case types.SignMode_SIGN_MODE_TEXTUAL:
			return nil, fmt.Errorf("SIGN_MODE_TEXTUAL is not supported yet")
		case types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
			return auth.StdSignBytes(
				chainID, accNum, acc.GetSequence(), auth.StdFee{Amount: d.GetFee(), Gas: d.GetGas()}, d.Msgs, d.Body.Memo,
			), nil
		}
	case *types.ModeInfo_Multi_:
		return nil, fmt.Errorf("multisig mode info is not supported by GetSignBytes")
	default:
		return nil, fmt.Errorf("unexpected ModeInfo")
	}
	return nil, fmt.Errorf("unexpected")
}

func DirectSignBytes(bodyBz, authInfoBz []byte, chainID string, accnum, sequence uint64) ([]byte, error) {
	signDoc := SignDocRaw{
		BodyBytes:       bodyBz,
		AuthInfoBytes:   authInfoBz,
		ChainId:         chainID,
		AccountNumber:   accnum,
		AccountSequence: sequence,
	}
	return signDoc.Marshal()
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
