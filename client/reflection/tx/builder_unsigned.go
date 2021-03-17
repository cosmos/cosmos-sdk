package tx

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// UnsignedBuilder implements a raw unsigned transaction builder
// it of course cannot offer the full set of functionalities
// of the standard sdk tx builder as it cannot deduct if a
// given message is valid and who are the expected signers.
type UnsignedBuilder struct {
	msgs          []proto.Message
	gasLimit      uint64
	fees          sdktypes.Coins
	chainID       string
	memo          string
	accountsInfo  map[string]SignerInfo
	feePayer      string
	timeoutHeight uint64
}

func NewUnsignedTxBuilder() *UnsignedBuilder {
	return &UnsignedBuilder{
		msgs:         nil,
		gasLimit:     0,
		fees:         nil,
		chainID:      "",
		accountsInfo: make(map[string]SignerInfo),
	}
}

func (t *UnsignedBuilder) SetMemo(memo string) {
	t.memo = memo
}

func (t *UnsignedBuilder) SetFees(fees sdktypes.Coins) {
	t.fees = fees
}

func (t *UnsignedBuilder) SetGasLimit(limit uint64) {
	t.gasLimit = limit
}

func (t *UnsignedBuilder) AddMsg(msg proto.Message) {
	t.msgs = append(t.msgs, msg)
}

// AddSigner is going to add a signer
func (t *UnsignedBuilder) AddSigner(signer SignerInfo) {
	if signer.PubKey == nil {
		panic("nil signer pub key")
	}
	t.accountsInfo[signer.PubKey.String()] = signer
}

func (t *UnsignedBuilder) SignedBuilder() (*SignedBuilder, error) {
	// add prereq checks
	// get fee payer
	if t.feePayer == "" {
		return nil, fmt.Errorf("fee payer not specified")
	}
	// pack msgs as any
	anyMsgs := make([]*codectypes.Any, len(t.msgs))

	for i, msg := range t.msgs {
		msgBytes, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
		typeURL := fmt.Sprintf("/%s", msg.ProtoReflect().Descriptor().FullName())
		anyMsg := &codectypes.Any{
			TypeUrl: typeURL,
			Value:   msgBytes,
		}
		anyMsgs[i] = anyMsg
	}

	signersInfo := make([]*tx.SignerInfo, 0, len(t.accountsInfo))

	for _, signer := range t.accountsInfo {
		signerInfo, err := getSignerInfo(signer)
		if err != nil {
			return nil, err
		}
		signersInfo = append(signersInfo, signerInfo)
	}

	rawTx := &tx.Tx{
		Body: &tx.TxBody{
			Messages:                    anyMsgs,
			Memo:                        t.memo,
			TimeoutHeight:               0,
			ExtensionOptions:            nil,
			NonCriticalExtensionOptions: nil,
		},
		AuthInfo: &tx.AuthInfo{
			SignerInfos: signersInfo,
			Fee: &tx.Fee{
				Amount:   t.fees,
				GasLimit: t.gasLimit,
				Payer:    t.feePayer,
				Granter:  "",
			},
		},
		Signatures: nil,
	}

	return NewSignedBuilder(rawTx, t.chainID, t.accountsInfo)
}

func (t *UnsignedBuilder) SetFeePayer(payer string) {
	t.feePayer = payer
}

func (t *UnsignedBuilder) SetChainID(chainID string) {
	t.chainID = chainID
}

func getSignerInfo(signer SignerInfo) (*tx.SignerInfo, error) {
	anyPubKey, err := codectypes.NewAnyWithValue(signer.PubKey)
	if err != nil {
		return nil, err
	}
	sig := &tx.SignerInfo{
		PublicKey: anyPubKey,
		ModeInfo: &tx.ModeInfo{
			Sum: &tx.ModeInfo_Single_{
				Single: &tx.ModeInfo_Single{Mode: signer.SignMode},
			},
		},
		Sequence: signer.Sequence,
	}

	return sig, nil
}
