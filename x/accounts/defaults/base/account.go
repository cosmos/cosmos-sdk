package base

import (
	"context"
	"errors"
	"fmt"

	gogotypes "github.com/cosmos/gogoproto/types/any"
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

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	PubKeyPrefix     = collections.NewPrefix(0)
	PubKeyTypePrefix = collections.NewPrefix(1)
	SequencePrefix   = collections.NewPrefix(2)
)

type Option func(a *Account)

func NewAccount(name string, handlerMap *signing.HandlerMap, options ...Option) accountstd.AccountCreatorFunc {
	return func(deps accountstd.Dependencies) (string, accountstd.Interface, error) {
		acc := Account{
			PubKey:           collections.NewItem(deps.SchemaBuilder, PubKeyPrefix, "pub_key_bytes", collections.BytesValue),
			PubKeyType:       collections.NewItem(deps.SchemaBuilder, PubKeyTypePrefix, "pub_key_type", collections.StringValue),
			Sequence:         collections.NewSequence(deps.SchemaBuilder, SequencePrefix, "sequence"),
			addrCodec:        deps.AddressCodec,
			hs:               deps.Environment.HeaderService,
			supportedPubKeys: map[string]pubKeyImpl{},
			signingHandlers:  handlerMap,
		}
		for _, option := range options {
			option(&acc)
		}
		if len(acc.supportedPubKeys) == 0 {
			return "", nil, fmt.Errorf("no public keys plugged for account type %s", name)
		}
		return name, acc, nil
	}
}

// Account implements a base account.
type Account struct {
	PubKey     collections.Item[[]byte]
	PubKeyType collections.Item[string]

	Sequence collections.Sequence

	addrCodec address.Codec
	hs        header.Service

	supportedPubKeys map[string]pubKeyImpl

	signingHandlers *signing.HandlerMap
}

func (a Account) Init(ctx context.Context, msg *v1.MsgInit) (*v1.MsgInitResponse, error) {
	if msg.InitSequence != 0 {
		err := a.Sequence.Set(ctx, msg.InitSequence)
		if err != nil {
			return nil, err
		}
	}
	return &v1.MsgInitResponse{}, a.savePubKey(ctx, msg.PubKey)
}

func (a Account) SwapPubKey(ctx context.Context, msg *v1.MsgSwapPubKey) (*v1.MsgSwapPubKeyResponse, error) {
	if !accountstd.SenderIsSelf(ctx) {
		return nil, errors.New("unauthorized")
	}

	return &v1.MsgSwapPubKeyResponse{}, a.savePubKey(ctx, msg.NewPubKey)
}

// Authenticate implements the authentication flow of an abstracted base account.
func (a Account) Authenticate(ctx context.Context, msg *aa_interface_v1.MsgAuthenticate) (*aa_interface_v1.MsgAuthenticateResponse, error) {
	if !accountstd.SenderIsAccountsModule(ctx) {
		return nil, errors.New("unauthorized: only accounts module is allowed to call this")
	}

	pubKey, signerData, err := a.computeSignerData(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to compute signer data: %w", err)
	}

	txData, err := a.getTxData(msg)
	if err != nil {
		return nil, fmt.Errorf("unable to get tx data: %w", err)
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
		return nil, fmt.Errorf("unable to get sign bytes: %w", err)
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
func (a Account) computeSignerData(ctx context.Context) (PubKey, signing.SignerData, error) {
	addrStr, err := a.addrCodec.BytesToString(accountstd.Whoami(ctx))
	if err != nil {
		return nil, signing.SignerData{}, err
	}
	chainID := a.hs.HeaderInfo(ctx).ChainID

	wantSequence, err := a.Sequence.Next(ctx)
	if err != nil {
		return nil, signing.SignerData{}, err
	}

	pk, err := a.loadPubKey(ctx)
	if err != nil {
		return nil, signing.SignerData{}, err
	}

	pkAny, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return nil, signing.SignerData{}, err
	}

	accNum, err := a.getNumber(ctx, addrStr)
	if err != nil {
		return nil, signing.SignerData{}, err
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
	accNum, err := accountstd.QueryModule[*accountsv1.AccountNumberResponse](ctx, &accountsv1.AccountNumberRequest{Address: addrStr})
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

func (a Account) loadPubKey(ctx context.Context) (PubKey, error) {
	pkType, err := a.PubKeyType.Get(ctx)
	if err != nil {
		return nil, err
	}

	publicKey, exists := a.supportedPubKeys[pkType]
	// this means that the chain developer suddenly started using a key type.
	if !exists {
		return nil, fmt.Errorf("pubkey type %s is not supported by the chain anymore", pkType)
	}

	pkBytes, err := a.PubKey.Get(ctx)
	if err != nil {
		return nil, err
	}

	pubKey, err := publicKey.decode(pkBytes)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func (a Account) savePubKey(ctx context.Context, anyPk *codectypes.Any) error {
	// check if known
	name := nameFromTypeURL(anyPk.TypeUrl)
	impl, exists := a.supportedPubKeys[name]
	if !exists {
		return fmt.Errorf("unknown pubkey type %s", name)
	}
	pk, err := impl.decode(anyPk.Value)
	if err != nil {
		return fmt.Errorf("unable to decode pubkey: %w", err)
	}
	err = impl.validate(pk)
	if err != nil {
		return fmt.Errorf("unable to validate pubkey: %w", err)
	}

	// save into state
	err = a.PubKey.Set(ctx, anyPk.Value)
	if err != nil {
		return fmt.Errorf("unable to save pubkey: %w", err)
	}
	return a.PubKeyType.Set(ctx, name)
}

func (a Account) QuerySequence(ctx context.Context, _ *v1.QuerySequence) (*v1.QuerySequenceResponse, error) {
	seq, err := a.Sequence.Peek(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.QuerySequenceResponse{Sequence: seq}, nil
}

func (a Account) QueryPubKey(ctx context.Context, _ *v1.QueryPubKey) (*v1.QueryPubKeyResponse, error) {
	pubKey, err := a.loadPubKey(ctx)
	if err != nil {
		return nil, err
	}
	anyPubKey, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return nil, err
	}
	return &v1.QueryPubKeyResponse{PubKey: anyPubKey}, nil
}

func (a Account) AuthRetroCompatibility(ctx context.Context, _ *authtypes.QueryLegacyAccount) (*authtypes.QueryLegacyAccountResponse, error) {
	addr, err := a.addrCodec.BytesToString(accountstd.Whoami(ctx))
	if err != nil {
		return nil, err
	}

	accNumber, err := accountstd.QueryModule[*accountsv1.AccountNumberResponse](ctx, &accountsv1.AccountNumberRequest{Address: addr})
	if err != nil {
		return nil, err
	}
	pk, err := a.loadPubKey(ctx)
	if err != nil {
		return nil, err
	}
	anyPk, err := gogotypes.NewAnyWithCacheWithValue(pk)
	if err != nil {
		return nil, err
	}

	seq, err := a.Sequence.Peek(ctx)
	if err != nil {
		return nil, err
	}

	baseAccount := &authtypes.BaseAccount{
		Address:       addr,
		PubKey:        anyPk,
		AccountNumber: accNumber.Number,
		Sequence:      seq,
	}

	baseAccountAny, err := gogotypes.NewAnyWithCacheWithValue(baseAccount)
	if err != nil {
		return nil, err
	}

	return &authtypes.QueryLegacyAccountResponse{
		Account: baseAccountAny,
		Base:    baseAccount,
	}, nil
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
	accountstd.RegisterQueryHandler(builder, a.QueryPubKey)
	accountstd.RegisterQueryHandler(builder, a.AuthRetroCompatibility)
}
