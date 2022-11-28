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
	"cosmossdk.io/tx/textual/internal/enveloppe"
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

	p1 := &enveloppe.Part1{
		ChainId:       textualData.SignerData.ChainId,
		AccountNumber: textualData.SignerData.AccountNumber,
		Sequence:      textualData.SignerData.Sequence,
	}
	p2 := &enveloppe.Part2{
		PublicKey: textualData.SignerData.PubKey,
	}
	p3 := &enveloppe.Part3{
		Memo: txBody.Memo,
		Fees: txAuthInfo.Fee.Amount,
	}
	p4 := &enveloppe.Part4{
		FeePayer:   txAuthInfo.Fee.Payer,
		FeeGranter: txAuthInfo.Fee.Granter,
	}
	p5 := &enveloppe.Part5{}
	if txAuthInfo.Tip != nil {
		p5.Tip = txAuthInfo.Tip.Amount
		p5.Tipper = txAuthInfo.Tip.Tipper
	}
	p6 := &enveloppe.Part6{
		GasLimit:       txAuthInfo.Fee.GasLimit,
		TimeoutHeight:  txBody.TimeoutHeight,
		HashOfRawBytes: getHash(textualData.BodyBytes, textualData.AuthInfoBytes),
	}

	screens1, err := vr.formatPart(ctx, p1, false)
	if err != nil {
		return nil, err
	}
	screens2, err := vr.formatPart(ctx, p2, true)
	if err != nil {
		return nil, err
	}
	screens3, err := vr.formatPart(ctx, p3, false)
	if err != nil {
		return nil, err
	}
	screens4, err := vr.formatPart(ctx, p4, true)
	if err != nil {
		return nil, err
	}
	screens5, err := vr.formatPart(ctx, p5, false)
	if err != nil {
		return nil, err
	}
	screens6, err := vr.formatPart(ctx, p6, true)
	if err != nil {
		return nil, err
	}

	screens := append(screens1, append(screens2, append(screens3, append(screens4, append(screens5, screens6...)...)...)...)...)

	return screens, nil
}

func (vr txValueRenderer) formatPart(ctx context.Context, m proto.Message, expert bool) ([]Screen, error) {
	messageVR := NewMessageValueRenderer(vr.tr, m.ProtoReflect().Descriptor())
	screens, err := messageVR.Format(ctx, protoreflect.ValueOf(m.ProtoReflect()))
	if err != nil {
		return nil, err
	}

	// Remove 1st screen which is the message name
	screens = screens[1:]

	// Remove indentations on all subscreens
	for i := range screens {
		screens[i].Indent--
		if expert {
			screens[i].Expert = true
		}
	}

	return screens, nil
}

// getHash gets the hash of raw bytes to be signed over:
// HEX(sha256(len(body_bytes) ++ body_bytes ++ len(auth_info_bytes) ++ auth_info_bytes))
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
func (vr txValueRenderer) Parse(ctx context.Context, screens []Screen) (protoreflect.Value, error) {
	res := &textualv1.TextualData{}

	_, _, err := vr.parsePart(ctx, screens, &enveloppe.Part1{})
	if err != nil {
		return nilValue, err
	}

	return protoreflect.ValueOfMessage(res.ProtoReflect()), nil
}

func (vr txValueRenderer) parsePart(ctx context.Context, screens []Screen, m proto.Message) ([]Screen, proto.Message, error) {
	messageVR := NewMessageValueRenderer(vr.tr, m.ProtoReflect().Descriptor())

	// Manually add the "<message_name> object" header screen, and indent correctly
	for i := range screens {
		screens[i].Indent++
	}
	screens = append([]Screen{{Text: "Part1 object"}}, screens...)

	v, err := messageVR.Parse(ctx, screens)
	if err != nil {
		return nil, nil, err
	}

	return screens, v.Message().Interface(), err
}
