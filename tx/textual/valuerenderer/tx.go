package valuerenderer

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/tx/signing"
)

type txValueRenderer struct {
	signerData signing.SignerData
	t          *Textual
}

// NewTimestampValueRenderer returns a ValueRenderer for the protobuf Tx type,
// as called the transaction envelope. It follows the specification defined
// in ADR-050.
func NewTxValueRenderer(t *Textual, signerData signing.SignerData) ValueRenderer {
	return txValueRenderer{
		t:          t,
		signerData: signerData,
	}
}

// Format implements the ValueRenderer interface.
func (vr txValueRenderer) Format(ctx context.Context, v protoreflect.Value) ([]Screen, error) {
	// Reify the reflected message as a proto Timestamp
	msg := v.Message().Interface()
	protoTx, ok := msg.(*txv1beta1.Tx)
	if !ok {
		return nil, fmt.Errorf("expected Tx, got %T", msg)
	}

	screens := make([]Screen, 3)
	screens[0].Text = fmt.Sprintf("Chain ID: %s", vr.signerData.ChainID)
	screens[1].Text = fmt.Sprintf("Account number: %d", vr.signerData.AccountNumber)
	pkMsgType, err := protoregistry.GlobalTypes.FindMessageByURL(vr.signerData.PubKey.TypeUrl)
	if err != nil {
		return nil, err
	}
	pk := pkMsgType.New()
	err = proto.Unmarshal(vr.signerData.PubKey.GetValue(), pk.Interface())
	if err != nil {
		return nil, err
	}
	screens[2].Text = fmt.Sprintf("Public key: %s", pk)
	screens[2].Expert = true

	// Get sdk.Msgs screens, from Tx.Body.Messages (field number 1).
	// msgVr, err := vr.t.GetValueRenderer(protoTx.Body.ProtoReflect().Descriptor().Fields().ByNumber(1))
	// if err != nil {
	// 	return nil, err
	// }
	// msgScreens, err := msgVr.Format(ctx, protoreflect.ValueOf(protoTx.Body.Messages)) // TODO not sure this works...
	// if err != nil {
	// 	return nil, err
	// }
	// screens = append(screens, msgScreens...)

	// Fees
	vr.Format(ctx, protoTx.AuthInfo.Fee.ProtoReflect().Get(protoTx.AuthInfo.Fee.ProtoReflect().Descriptor().Fields().ByNumber(1)))

	return screens, nil
}

// Parse implements the ValueRenderer interface.
func (vr txValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	panic("TODO")
}
