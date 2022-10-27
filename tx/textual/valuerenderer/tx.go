package valuerenderer

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	textualv1 "cosmossdk.io/api/cosmos/msg/textual/v1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
)

type txValueRenderer struct {
	tr *Textual
}

// NewTxValueRenderer returns a ValueRenderer for the protobuf
// TextualData type. It follows the specification defined in ADR-050.
// The reason we create a renderer for TextualData (and not directly Tx)
// is that TextualData is a single place that contains all data needed
// to create the `[]Screen` SignDoc.
func NewTxValueRenderer(tr *Textual) ValueRenderer {
	return txValueRenderer{
		tr: tr,
	}
}

// Format implements the ValueRenderer interface.
func (vr txValueRenderer) Format(ctx context.Context, v protoreflect.Value) ([]Screen, error) {
	// Reify the reflected message as a proto Tx
	msg := v.Message().Interface()
	textualData, ok := msg.(*textualv1.TextualData)
	if !ok {
		return nil, fmt.Errorf("expected Tx, got %T", msg)
	}

	txBody := &txv1beta1.TxBody{}
	txAuthInfo := &txv1beta1.AuthInfo{}
	err := proto.Unmarshal(textualData.BodyBytes, txBody)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(textualData.AuthInfoBytes, txAuthInfo)
	if err != nil {
		return nil, err
	}
	signerData := textualData.SignerData

	screens := make([]Screen, 3)
	screens[0].Text = fmt.Sprintf("Chain ID: %s", signerData.ChainId)
	screens[1].Text = fmt.Sprintf("Account number: %d", signerData.AccountNumber)
	screens[2].Text = fmt.Sprintf("Sequence: %d", signerData.Sequence)

	// TODO Public key: needs Any
	// TODO Body messages: needs repeated

	screens, err = vr.appendScreen(screens, ctx, txAuthInfo.Fee.ProtoReflect(), "amount", "Fees", false)
	if err != nil {
		return nil, err
	}
	screens, err = vr.appendScreen(screens, ctx, txAuthInfo.Fee.ProtoReflect(), "payer", "Fee payer", true)
	if err != nil {
		return nil, err
	}
	screens, err = vr.appendScreen(screens, ctx, txAuthInfo.Fee.ProtoReflect(), "granter", "Fee granter", true)
	if err != nil {
		return nil, err
	}
	screens, err = vr.appendScreen(screens, ctx, txBody.ProtoReflect(), "memo", "Memo", false)
	if err != nil {
		return nil, err
	}
	screens, err = vr.appendScreen(screens, ctx, txAuthInfo.Fee.ProtoReflect(), "gas_limit", "Gas limit", true)
	if err != nil {
		return nil, err
	}
	screens, err = vr.appendScreen(screens, ctx, txBody.ProtoReflect(), "timeout_height", "Timeout height", true)
	if err != nil {
		return nil, err
	}
	screens, err = vr.appendScreen(screens, ctx, txAuthInfo.Tip.ProtoReflect(), "tipper", "Tipper", true)
	if err != nil {
		return nil, err
	}
	screens, err = vr.appendScreen(screens, ctx, txAuthInfo.Tip.ProtoReflect(), "amount", "Tip", true)
	if err != nil {
		return nil, err
	}
	// screens, err = vr.appendScreen(screens, ctx, txBody.ProtoReflect(), "extension_options", "Body extensions", true)
	// if err != nil {
	// 	return nil, err
	// }
	// screens, err = vr.appendScreen(screens, ctx, txBody.ProtoReflect(), "non_critical_extension_options options", "Non-critical body extensions", true)
	// if err != nil {
	// 	return nil, err
	// }
	if len(txAuthInfo.SignerInfos) > 1 {
		// Get all signer infos except the signer's one.
		otherSigners := make([]*txv1beta1.SignerInfo, 0, len(txAuthInfo.SignerInfos)-1)
		for _, si := range txAuthInfo.SignerInfos {
			if !proto.Equal(si.PublicKey, signerData.PubKey) {
				otherSigners = append(otherSigners, si)
			}
		}

		// Create the sametxAuthInfo.SignerInfos message, but without the signer
		fd := txAuthInfo.ProtoReflect().Descriptor().Fields().ByName("signer_infos")
		newField := txAuthInfo.ProtoReflect().NewField(fd)
		r, err := vr.tr.GetValueRenderer(fd)
		if err != nil {
			return nil, err
		}
		newScreens, err := r.Format(ctx, newField)
		if err != nil {
			return nil, err
		}

		// Manually replace the first screen, to add the "other" word
		newScreens[0].Text = fmt.Sprintf("This transaction has %d other signers:", len(otherSigners))

		screens = append(screens, newScreens...)
	}

	screens = append(screens, Screen{
		Text:   fmt.Sprintf("Hash of raw bytes: %s", getHash(textualData.BodyBytes, textualData.AuthInfoBytes)),
		Expert: true,
	})

	return screens, nil
}

func (vr txValueRenderer) appendScreen(
	screens []Screen, ctx context.Context,
	msg protoreflect.Message, fieldName,
	label string, expert bool,
) ([]Screen, error) {
	// Skip if the message is empty
	if !msg.IsValid() {
		return screens, nil
	}
	fd := msg.Descriptor().Fields().ByName(protoreflect.Name(fieldName))
	value := msg.Get(fd)
	// Skip if the value is empty
	if !value.IsValid() || isValueEmpty(msg, fd, value) {
		return screens, nil
	}

	// Get the value-rendered text of the inner Message
	r, err := vr.tr.GetValueRenderer(fd)
	if err != nil {
		return nil, err
	}
	new, err := r.Format(ctx, value)
	if err != nil {
		return nil, err
	}

	// Add label and expert fields as needed
	new[0].Text = fmt.Sprintf("%s: %s", label, new[0].Text)
	if expert {
		for i := range new {
			new[i].Expert = true
		}
	}

	screens = append(screens, new...)

	return screens, nil
}

// isValueEmpty checks if the protoreflect.Value is equal to its empty
// (default) value.
// Protobuf only exposes the `Equal` method on messages, so here we create
// two messages: one empty, and the other empty except for the field with the
// given Value. We then do proto.Equal on these two messages.
func isValueEmpty(msg protoreflect.Message, fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
	msgWithValue := msg.New()
	msgWithValue.Set(fd, v)
	emptyMsg := msg.New()

	return proto.Equal(emptyMsg.Interface(), msgWithValue.Interface())
}

// getHash gets the hash of raw bytes to be signed over.
func getHash(bodyBz, authInfoBz []byte) string {
	bodyLen, authInfoLen := make([]byte, 8), make([]byte, 8)
	binary.BigEndian.PutUint64(bodyLen, uint64(len(bodyBz)))
	binary.BigEndian.PutUint64(authInfoLen, uint64(len(authInfoBz)))

	b := make([]byte, 16+len(bodyBz)+len(authInfoBz))
	copy(b[:8], bodyLen)
	copy(b[8:8+len(bodyBz)], bodyBz)
	copy(b[8+len(bodyBz):16+len(bodyBz)], authInfoLen)
	copy(b[16+len(bodyBz):], authInfoBz)

	h := sha256.Sum256(b)

	return hex.EncodeToString(h[:])
}

// Parse implements the ValueRenderer interface.
func (vr txValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	panic("TODO")
}
