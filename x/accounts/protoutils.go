package accounts

import (
	"time"

	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/accounts/internal/implementation"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func protoV2TxToProtoV1(t *txv1beta1.Tx) *tx.Tx {
	return &tx.Tx{
		Body: &tx.TxBody{
			Messages:                    protoV2AnyToV1(t.Body.Messages...),
			Memo:                        t.Body.Memo,
			TimeoutHeight:               t.Body.TimeoutHeight,
			Unordered:                   t.Body.Unordered,
			TimeoutTimestamp:            protoV2TimestampToV1(t.Body.TimeoutTimestamp),
			ExtensionOptions:            protoV2AnyToV1(t.Body.ExtensionOptions...),
			NonCriticalExtensionOptions: protoV2AnyToV1(t.Body.NonCriticalExtensionOptions...),
		},
		AuthInfo: &tx.AuthInfo{
			SignerInfos: protoV2SignerInfoToV1(t.AuthInfo.SignerInfos),
			Fee:         nil, // Fee and Tip are expected
			Tip:         nil, // to be empty.
		},
		Signatures: t.Signatures,
	}
}

func protoV2TimestampToV1(timestamp *timestamppb.Timestamp) *time.Time {
	if timestamp == nil {
		return nil
	}
	ts := timestamp.AsTime()
	return &ts
}

func protov2TxRawToProtoV1(raw *txv1beta1.TxRaw) *tx.TxRaw {
	return &tx.TxRaw{
		BodyBytes:     raw.BodyBytes,
		AuthInfoBytes: raw.AuthInfoBytes,
		Signatures:    raw.Signatures,
	}
}

func protoV2AnyToV1(v2s ...*anypb.Any) []*implementation.Any {
	v1s := make([]*implementation.Any, len(v2s))
	for i, v2 := range v2s {
		v1s[i] = &implementation.Any{
			TypeUrl: v2.TypeUrl,
			Value:   v2.Value,
		}
	}
	return v1s
}

func protoV2SignerInfoToV1(infos []*txv1beta1.SignerInfo) []*tx.SignerInfo {
	v1s := make([]*tx.SignerInfo, len(infos))
	for i, info := range infos {
		v1s[i] = &tx.SignerInfo{
			PublicKey: protoV2AnyToV1(info.PublicKey)[0],
			ModeInfo:  protoV2ModeInfoToV1(info.ModeInfo),
			Sequence:  info.Sequence,
		}
	}
	return v1s
}

func protoV2ModeInfoToV1(info *txv1beta1.ModeInfo) *tx.ModeInfo {
	switch v := info.Sum.(type) {
	case *txv1beta1.ModeInfo_Single_:
		return &tx.ModeInfo{
			Sum: &tx.ModeInfo_Single_{Single: &tx.ModeInfo_Single{Mode: signing.SignMode(v.Single.Mode)}},
		}
	default:
		// NOTE(tip): we have a check that disallows modes different from single
		panic("unexpected mode info")
	}
}
