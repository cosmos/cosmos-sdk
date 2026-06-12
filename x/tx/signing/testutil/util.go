package testutil

import (
	"github.com/cosmos/cosmos-proto/anyutil"
	gogoproto "github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"cosmossdk.io/api/cosmos/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/x/tx/signing"
)

// HandlerArgumentOptions are options for MakeHandlerArguments.
type HandlerArgumentOptions struct {
	ChainID          string
	Memo             string
	Msg              proto.Message
	AccNum           uint64
	AccSeq           uint64
	Unordered        bool
	Timeouttimestamp *timestamppb.Timestamp
	Fee              *signing.TxFeeData
	SignerAddress    string
}

// MakeHandlerArguments creates signing.SignerData and signing.TxData for use in tests.
func MakeHandlerArguments(options HandlerArgumentOptions) (signing.SignerData, signing.TxData, error) {
	pk := &secp256k1.PubKey{
		Key: make([]byte, 256),
	}
	anyPk, err := anyutil.New(pk)
	if err != nil {
		return signing.SignerData{}, signing.TxData{}, err
	}

	anyMsg, err := anyutil.New(options.Msg)
	if err != nil {
		return signing.SignerData{}, signing.TxData{}, err
	}

	txBodyData := &signing.TxBodyData{
		Messages:         []signing.RawMsg{{TypeUrl: anyMsg.TypeUrl, Value: anyMsg.Value}},
		Memo:             options.Memo,
		Unordered:        options.Unordered,
		TimeoutTimestamp: options.Timeouttimestamp,
	}

	var feeData signing.TxFeeData
	if options.Fee != nil {
		feeData = *options.Fee
	}

	txAuthInfoData := &signing.TxAuthInfoData{Fee: feeData}

	// Build raw bytes using minimal gogoproto-compatible structs.
	// This avoids importing types/tx which creates an import cycle in tests.
	bodyBz, err := gogoproto.Marshal(&rawTxBody{
		Messages: []*gogotypes.Any{{TypeUrl: anyMsg.TypeUrl, Value: anyMsg.Value}},
		Memo:     options.Memo,
	})
	if err != nil {
		return signing.SignerData{}, signing.TxData{}, err
	}

	authInfoBz, err := gogoproto.Marshal(&rawAuthInfo{
		SignerInfos: []*rawSignerInfo{
			{
				PublicKey: &gogotypes.Any{TypeUrl: anyPk.TypeUrl, Value: anyPk.Value},
				ModeInfo:  &rawModeInfo{Single: &rawModeInfoSingle{Mode: 3}},
				Sequence:  options.AccSeq,
			},
		},
	})
	if err != nil {
		return signing.SignerData{}, signing.TxData{}, err
	}

	txData := signing.TxData{
		Body:          txBodyData,
		AuthInfo:      txAuthInfoData,
		AuthInfoBytes: authInfoBz,
		BodyBytes:     bodyBz,
	}

	signerData := signing.SignerData{
		ChainID:       options.ChainID,
		AccountNumber: options.AccNum,
		Sequence:      options.AccSeq,
		Address:       options.SignerAddress,
		PubKey:        anyPk,
	}

	return signerData, txData, nil
}

// minimal gogoproto-compatible structs for marshaling TxBody and AuthInfo bytes
// without importing github.com/cosmos/cosmos-sdk/types/tx (which creates cycles).

type rawTxBody struct {
	Messages []*gogotypes.Any `protobuf:"bytes,1,rep,name=messages,proto3"`
	Memo     string           `protobuf:"bytes,2,opt,name=memo,proto3"`
}

func (*rawTxBody) Reset()         {}
func (*rawTxBody) String() string { return "" }
func (*rawTxBody) ProtoMessage()  {}

type rawAuthInfo struct {
	SignerInfos []*rawSignerInfo `protobuf:"bytes,1,rep,name=signer_infos,json=signerInfos,proto3"`
}

func (*rawAuthInfo) Reset()         {}
func (*rawAuthInfo) String() string { return "" }
func (*rawAuthInfo) ProtoMessage()  {}

type rawSignerInfo struct {
	PublicKey *gogotypes.Any `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3"`
	ModeInfo  *rawModeInfo   `protobuf:"bytes,2,opt,name=mode_info,json=modeInfo,proto3"`
	Sequence  uint64         `protobuf:"varint,3,opt,name=sequence,proto3"`
}

func (*rawSignerInfo) Reset()         {}
func (*rawSignerInfo) String() string { return "" }
func (*rawSignerInfo) ProtoMessage()  {}

type rawModeInfo struct {
	Single *rawModeInfoSingle `protobuf:"bytes,1,opt,name=single,proto3,oneof"`
}

func (*rawModeInfo) Reset()         {}
func (*rawModeInfo) String() string { return "" }
func (*rawModeInfo) ProtoMessage()  {}

type rawModeInfoSingle struct {
	Mode int32 `protobuf:"varint,1,opt,name=mode,proto3,enum=cosmos.tx.signing.v1beta1.SignMode"`
}

func (*rawModeInfoSingle) Reset()         {}
func (*rawModeInfoSingle) String() string { return "" }
func (*rawModeInfoSingle) ProtoMessage()  {}
