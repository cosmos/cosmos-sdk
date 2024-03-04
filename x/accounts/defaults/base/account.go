package base

import (
	"context"
	"errors"
	"fmt"

	dcrd_secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/accounts/accountstd"
	v1 "cosmossdk.io/x/accounts/defaults/base/v1"
	aa_interface_v1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

var (
	PubKeyPrefix   = collections.NewPrefix(0)
	SequencePrefix = collections.NewPrefix(1)
)

func NewAccount(name string, handlerMap *signing.HandlerMap) accountstd.AccountCreatorFunc {
	return func(deps accountstd.Dependencies) (string, accountstd.Interface, error) {
		return name, Account{
			PubKey:          collections.NewItem(deps.SchemaBuilder, PubKeyPrefix, "pub_key", codec.CollValue[secp256k1.PubKey](deps.LegacyStateCodec)),
			Sequence:        collections.NewSequence(deps.SchemaBuilder, SequencePrefix, "sequence"),
			addrCodec:       deps.AddressCodec,
			signingHandlers: handlerMap,
		}, nil
	}
}

// Account implements a base account.
type Account struct {
	PubKey   collections.Item[secp256k1.PubKey]
	Sequence collections.Sequence

	addrCodec address.Codec
	hs        header.Service

	signingHandlers *signing.HandlerMap
}

func (a Account) Init(ctx context.Context, msg *v1.MsgInit) (*v1.MsgInitResponse, error) {
	return &v1.MsgInitResponse{}, a.verifyAndSetPubKey(ctx, msg.PubKey)
}

func (a Account) SwapPubKey(ctx context.Context, msg *v1.MsgSwapPubKey) (*v1.MsgSwapPubKeyResponse, error) {
	if !accountstd.SenderIsSelf(ctx) {
		return nil, errors.New("unauthorized")
	}

	return &v1.MsgSwapPubKeyResponse{}, a.verifyAndSetPubKey(ctx, msg.NewPubKey)
}

func (a Account) verifyAndSetPubKey(ctx context.Context, key []byte) error {
	_, err := dcrd_secp256k1.ParsePubKey(key)
	if err != nil {
		return err
	}
	return a.PubKey.Set(ctx, secp256k1.PubKey{Key: key})
}

// Authenticate implements the authentication flow of an abstracted base account.
func (a Account) Authenticate(ctx context.Context, msg *aa_interface_v1.MsgAuthenticate) (*aa_interface_v1.MsgAuthenticateResponse, error) {
	if !accountstd.SenderIsAccountsModule(ctx) {
		return nil, errors.New("unauthorized: only accounts module is allowed to call this")
	}

	pubKey, signerData, err := a.computeSignerData(ctx)
	if err != nil {
		return nil, err
	}

	txData, err := a.getTxData(msg)
	if err != nil {
		return nil, err
	}

	gotSeq := msg.Tx.AuthInfo.SignerInfos[msg.SignerIndex].Sequence
	if gotSeq != signerData.Sequence {
		return nil, fmt.Errorf("unexpected sequence number, wanted: %d, got: %d", signerData.Sequence, gotSeq)
	}

	signMode, err := parseSignMode(msg.Tx.AuthInfo.SignerInfos[msg.SignerIndex].ModeInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to parse sign mode: %w", err)
	}

	signature := msg.Tx.Signatures[msg.SignerIndex]

	signBytes, err := a.signingHandlers.GetSignBytes(ctx, signMode, signerData, txData)
	if err != nil {
		return nil, err
	}

	if !pubKey.VerifySignature(signBytes, signature) {
		return nil, errors.New("signature verification failed")
	}

	return &aa_interface_v1.MsgAuthenticateResponse{}, nil
}

func parseSignMode(info *tx.ModeInfo) (signingv1beta1.SignMode, error) {
	single, ok := info.Sum.(*tx.ModeInfo_Single_)
	if !ok {
		return 0, fmt.Errorf("only sign mode single accepted got: %v", info.Sum)
	}
	return signingv1beta1.SignMode(single.Single.Mode), nil
}

// computeSignerData will populate signer data and also increase the sequence.
func (a Account) computeSignerData(ctx context.Context) (secp256k1.PubKey, signing.SignerData, error) {
	addrStr, err := a.addrCodec.BytesToString(accountstd.Whoami(ctx))
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}
	chainID := a.hs.GetHeaderInfo(ctx).ChainID

	wantSequence, err := a.Sequence.Next(ctx)
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}

	pk, err := a.PubKey.Get(ctx)
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}

	pkAny, err := codectypes.NewAnyWithValue(&pk)
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}

	accNum, err := a.getNumber(ctx, addrStr)
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}

	return pk, signing.SignerData{
		Address:       addrStr,
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      wantSequence,
		PubKey: &anypb.Any{
			TypeUrl: pkAny.TypeUrl,
			Value:   pkAny.Value,
		},
	}, nil
}

func (a Account) getNumber(ctx context.Context, addrStr string) (uint64, error) {
	accNum, err := accountstd.QueryModule[accountsv1.AccountNumberResponse](ctx, &accountsv1.AccountNumberRequest{Address: addrStr})
	if err != nil {
		return 0, err
	}

	return accNum.Number, nil
}

func (a Account) getTxData(msg *aa_interface_v1.MsgAuthenticate) (signing.TxData, error) {
	// TODO: add a faster way to do this, we can avoid unmarshalling but we need
	// to write a function that converts this into the protov2 counterparty.
	txBody := new(txv1beta1.TxBody)
	err := proto.Unmarshal(msg.RawTx.BodyBytes, txBody)
	if err != nil {
		return signing.TxData{}, err
	}

	authInfo := new(txv1beta1.AuthInfo)
	err = proto.Unmarshal(msg.RawTx.AuthInfoBytes, authInfo)
	if err != nil {
		return signing.TxData{}, err
	}

	return signing.TxData{
		Body:                       txBody,
		AuthInfo:                   authInfo,
		BodyBytes:                  msg.RawTx.BodyBytes,
		AuthInfoBytes:              msg.RawTx.AuthInfoBytes,
		BodyHasUnknownNonCriticals: false, // NOTE: amino signing must be disabled.
	}, nil
}

func (a Account) QuerySequence(ctx context.Context, _ *v1.QuerySequence) (*v1.QuerySequenceResponse, error) {
	seq, err := a.Sequence.Peek(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.QuerySequenceResponse{Sequence: seq}, nil
}

func (a Account) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

func (a Account) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.SwapPubKey)
	accountstd.RegisterExecuteHandler(builder, a.Authenticate) // account abstraction
}

func (a Account) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QuerySequence)
}
