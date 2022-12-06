package valuerenderer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	msg "cosmossdk.io/api/cosmos/msg/v1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/tx/textual/internal/textualpb"
)

var (
	// msgRe is a regex matching the beginning of the TxBody msgs in the enveloppe.
	msgRe = regexp.MustCompile("Message: ([0-9]+) Any")
	// inverseMsgRe is a regex matching the textual output of the TxBody msgs
	// header.
	inverseMsgRe = regexp.MustCompile("This transaction has ([0-9]+) Messages?")
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
	textualData, ok := msg.(*textualpb.TextualData)
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

	enveloppe := &textualpb.Enveloppe{
		ChainId:                     textualData.SignerData.ChainId,
		AccountNumber:               textualData.SignerData.AccountNumber,
		Sequence:                    textualData.SignerData.Sequence,
		Address:                     textualData.SignerData.Address,
		PublicKey:                   textualData.SignerData.PubKey,
		Message:                     txBody.Messages,
		Memo:                        txBody.Memo,
		Fees:                        txAuthInfo.Fee.Amount,
		FeePayer:                    txAuthInfo.Fee.Payer,
		FeeGranter:                  txAuthInfo.Fee.Granter,
		GasLimit:                    txAuthInfo.Fee.GasLimit,
		TimeoutHeight:               txBody.TimeoutHeight,
		ExtensionOptions:            txBody.ExtensionOptions,
		NonCriticalExtensionOptions: txBody.NonCriticalExtensionOptions,
		HashOfRawBytes:              getHash(textualData.BodyBytes, textualData.AuthInfoBytes),
	}
	if txAuthInfo.Tip != nil {
		enveloppe.Tip = txAuthInfo.Tip.Amount
		enveloppe.Tipper = txAuthInfo.Tip.Tipper
	}
	// Find all other tx signers than the current signer. In the case where our
	// Textual signer is one key of a multisig, then otherSigners will include
	// the multisig pubkey.
	otherSigners := []*txv1beta1.SignerInfo{}
	for _, si := range txAuthInfo.SignerInfos {
		if bytes.Equal(si.PublicKey.Value, textualData.SignerData.PubKey.Value) {
			continue
		}

		otherSigners = append(otherSigners, si)
	}
	enveloppe.OtherSigner = otherSigners

	mvr, err := vr.tr.GetMessageValueRenderer(enveloppe.ProtoReflect().Descriptor())
	if err != nil {
		return nil, err
	}

	screens, err := mvr.Format(ctx, protoreflect.ValueOf(enveloppe.ProtoReflect()))
	if err != nil {
		return nil, err
	}

	// Since we're value-rendering the (internal) Enveloppe message, we do some
	// postprocessing. First, we remove first enveloppe header screen, and
	// unindent 1 level.

	// Remove 1st screen
	screens = screens[1:]
	for i := range screens {
		screens[i].Indent--
	}

	// Expert fields.
	expert := map[string]struct{}{
		"Public key":                     {},
		"Fee payer":                      {},
		"Fee granter":                    {},
		"Gas limit":                      {},
		"Timeout height":                 {},
		"Other signer":                   {},
		"Extension options":              {},
		"Non critical extension options": {},
		"Hash of raw bytes":              {},
	}

	for i := range screens {
		if screens[i].Indent == 0 {
			// Do expert fields.
			screenKV := strings.Split(screens[i].Text, ": ")
			_, ok := expert[screenKV[0]]
			if ok {
				expertify(screens, i, screenKV[0])
			}

			// Replace:
			// "Message: <N> Any"
			// with:
			// "This transaction has <N> Message"
			matches := msgRe.FindStringSubmatch(screens[i].Text)
			if len(matches) > 0 {
				screens[i].Text = fmt.Sprintf("This transaction has %s Message", matches[1])
				if matches[1] != "1" {
					screens[i].Text += "s"
				}
			}
		}
	}

	return screens, nil
}

// expertify marks all screens starting from `fromIdx` as expert, and stops
// just before it finds the next screen with Indent==0 (unless it's a "End of"
// termination screen). It modifies screens in-place.
func expertify(screens []Screen, fromIdx int, fieldName string) {
	for i := fromIdx; i < len(screens); i++ {
		if i > fromIdx &&
			screens[i].Indent == 0 &&
			screens[i].Text != fmt.Sprintf("End of %s", fieldName) {
			break
		}

		screens[i].Expert = true
	}
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
	// Process the screens to be parsable by a Enveloppe message value renderer
	parsable := make([]Screen, len(screens)+1)
	parsable[0] = Screen{Text: "Enveloppe object"}
	for i := range screens {
		parsable[i+1].Indent = screens[i].Indent + 1

		// Take same text, except that we weplace:
		// "This transaction has <N> Message"
		// with:
		// "Message: <N> Any"
		matches := inverseMsgRe.FindStringSubmatch(screens[i].Text)
		if len(matches) > 0 {
			parsable[i+1].Text = fmt.Sprintf("Message: %s Any", matches[1])
		} else {
			parsable[i+1].Text = screens[i].Text
		}
	}

	mvr, err := vr.tr.GetMessageValueRenderer((&textualpb.Enveloppe{}).ProtoReflect().Descriptor())
	if err != nil {
		return nilValue, err
	}

	enveloppeV, err := mvr.Parse(ctx, parsable)
	if err != nil {
		return nilValue, err
	}
	enveloppe := enveloppeV.Message().Interface().(*textualpb.Enveloppe)

	txBody := &txv1beta1.TxBody{
		Messages:                    enveloppe.Message,
		Memo:                        enveloppe.Memo,
		TimeoutHeight:               enveloppe.TimeoutHeight,
		ExtensionOptions:            enveloppe.ExtensionOptions,
		NonCriticalExtensionOptions: enveloppe.NonCriticalExtensionOptions,
	}
	authInfo := &txv1beta1.AuthInfo{
		Fee: &txv1beta1.Fee{
			Amount:   enveloppe.Fees,
			GasLimit: enveloppe.GasLimit,
			Payer:    enveloppe.FeePayer,
			Granter:  enveloppe.FeeGranter,
		},
	}
	if enveloppe.Tip != nil {
		authInfo.Tip = &txv1beta1.Tip{
			Amount: enveloppe.Tip,
			Tipper: enveloppe.Tipper,
		}
	}

	// Figure out the signers in the correct order.
	signers, err := getSigners(txBody, authInfo)
	if err != nil {
		return nilValue, err
	}

	signerInfos := make([]*txv1beta1.SignerInfo, len(signers))
	for i, s := range signers {
		if s == enveloppe.Address {
			signerInfos[i] = &txv1beta1.SignerInfo{
				PublicKey: enveloppe.PublicKey,
				ModeInfo: &txv1beta1.ModeInfo{
					Sum: &txv1beta1.ModeInfo_Single_{
						Single: &txv1beta1.ModeInfo_Single{
							Mode: signingv1beta1.SignMode_SIGN_MODE_TEXTUAL,
						},
					},
				},
				Sequence: enveloppe.Sequence,
			}
		} else {
			// We know that signerInfos is well ordered, so just pop from it.
			signerInfos[i] = enveloppe.OtherSigner[0]
			enveloppe.OtherSigner = enveloppe.OtherSigner[1:]
		}
	}
	authInfo.SignerInfos = signerInfos

	// Note that we might not always get back the exact bodyBz and authInfoBz
	// that was passed into, because protobuf is not deterministic.
	// In tests, we don't check bytes equality, but protobuf object equality.
	bodyBz, err := proto.Marshal(txBody)
	if err != nil {
		return nilValue, err
	}
	authInfoBz, err := proto.Marshal(authInfo)
	if err != nil {
		return nilValue, err
	}

	tx := &textualpb.TextualData{
		BodyBytes:     bodyBz,
		AuthInfoBytes: authInfoBz,
		SignerData: &textualpb.SignerData{
			Address:       enveloppe.Address,
			AccountNumber: enveloppe.AccountNumber,
			ChainId:       enveloppe.ChainId,
			Sequence:      enveloppe.Sequence,
			PubKey:        enveloppe.PublicKey,
		},
	}

	return protoreflect.ValueOf(tx.ProtoReflect()), nil
}

// getSigners gets the ordered signers of a transaction. It's mostly a
// copy-paste of `types/tx/types.go` GetSigners method, but uses the proto
// annotation `cosmos.msg.v1.signer`, instead of the sdk.Msg#GetSigners method.
func getSigners(body *txv1beta1.TxBody, authInfo *txv1beta1.AuthInfo) ([]string, error) {
	var signers []string
	seen := map[string]bool{}

	for _, msgAny := range body.Messages {
		m, err := anypb.UnmarshalNew(msgAny, proto.UnmarshalOptions{})
		if err != nil {
			return nil, err
		}

		ext := proto.GetExtension(m.ProtoReflect().Descriptor().Options(), msg.E_Signer)
		signerFields, ok := ext.([]string)
		if !ok {
			return nil, fmt.Errorf("expected []string, got %T", ext)
		}

		for _, fieldName := range signerFields {
			fd := m.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name(fieldName))
			addr := m.ProtoReflect().Get(fd).String()
			if !seen[addr] {
				signers = append(signers, addr)
				seen[addr] = true
			}
		}
	}

	// ensure any specified fee payer is included in the required signers (at the end)
	feePayer := authInfo.Fee.Payer
	if feePayer != "" && !seen[feePayer] {
		signers = append(signers, feePayer)
		seen[feePayer] = true
	}

	return signers, nil
}
