package base

import (
	"context"
	"fmt"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/accounts/accountstd"
	v1 "cosmossdk.io/x/accounts/defaults/base/v1"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func (a Account) Authenticate(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error) {
	pubKey, signerData, err := a.getSignerData(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to compute signer data: %w", err)
	}

	authenticationData, err := a.getAuthenticationData(ctx, msg.UserOperation.AuthenticationData)
	if err != nil {
		return nil, fmt.Errorf("unable to parse authentication data: %w", err)
	}

	txData, err := a.getTxData(ctx, msg.UserOperation, authenticationData)
	if err != nil {
		return nil, fmt.Errorf("unable to compute tx data: %w", err)
	}

	signBytes, err := a.signModeHandlers.GetSignBytes(ctx, signingv1beta1.SignMode(authenticationData.SignMode), signerData, txData)
	if err != nil {
		return nil, fmt.Errorf("unable to get required signature bytes: %w", err)
	}
	if err != nil {
		return nil, err
	}

	ok := pubKey.VerifySignature(signBytes, authenticationData.Signature)
	if !ok {
		return nil, fmt.Errorf("signature verification failed please check: chainID(want:%s), sequence(want:%d)", signerData.ChainID, signerData.Sequence)
	}

	return &account_abstractionv1.MsgAuthenticateResponse{}, nil
}

func (a Account) getSignerData(ctx context.Context) (secp256k1.PubKey, signing.SignerData, error) {
	pk, err := a.PubKey.Get(ctx)
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}
	pkAny, err := codectypes.NewAnyWithValue(&pk)
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}

	addr := accountstd.Whoami(ctx)
	addrStr, err := a.addrCodec.BytesToString(addr)
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}

	seq, err := a.Sequence.Next(ctx)
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}

	accNum, err := accountstd.QueryModule[*accountsv1.AccountNumberResponse](ctx, &accountsv1.AccountNumberRequest{Address: addrStr})
	if err != nil {
		return secp256k1.PubKey{}, signing.SignerData{}, err
	}

	chainID := a.hs.GetHeaderInfo(ctx).ChainID

	return pk, signing.SignerData{
		Address:       addrStr,
		ChainID:       chainID,
		AccountNumber: accNum.Number,
		Sequence:      seq,
		PubKey: &anypb.Any{
			TypeUrl: pkAny.TypeUrl,
			Value:   pkAny.Value,
		},
	}, nil
}

func (a Account) getAuthenticationData(_ context.Context, authData *codectypes.Any) (*v1.AuthenticationData, error) {
	v := authData.GetCachedValue()
	if v == nil {
		return nil, fmt.Errorf("authentication data is not populated")
	}
	concrete, ok := v.(*v1.AuthenticationData)
	if !ok {
		return nil, fmt.Errorf("wanted v1.AuthenticationData, got %T", v)
	}
	return concrete, nil
}

func (a Account) getTxData(ctx context.Context, op *accountsv1.UserOperation, authData *v1.AuthenticationData) (signing.TxData, error) {
	// it's coming from a TX, so we can safely return signing.TxData
	if op.TxCompat != nil {
		return getTxDataFromTxCompat(op.TxCompat)
	}
	// if it's coming from an operation, then we need to compute this information ourselves.
	return a.getTxDataFromOperation(ctx, op, authData)
}

// TODO: in a future PR.
func (a Account) getTxDataFromOperation(ctx context.Context, op *accountsv1.UserOperation, authData *v1.AuthenticationData) (signing.TxData, error) {
	panic("impl")
}

// getTxDataFromTxCompat computes the signing.TxData from TxCompat.
func getTxDataFromTxCompat(compat *accountsv1.TxCompat) (signing.TxData, error) {
	// maybe avoid the unmarshal again!!
	body := new(txv1beta1.TxBody)
	authInfo := new(txv1beta1.AuthInfo)
	err := proto.Unmarshal(compat.BodyBytes, body)
	if err != nil {
		return signing.TxData{}, err
	}
	err = proto.Unmarshal(compat.AuthInfoBytes, authInfo)
	if err != nil {
		return signing.TxData{}, err
	}
	return signing.TxData{
		Body:                       body,
		AuthInfo:                   authInfo,
		BodyBytes:                  compat.BodyBytes,
		AuthInfoBytes:              compat.AuthInfoBytes,
		BodyHasUnknownNonCriticals: false, // TODO: this is needed for amino..
	}, nil
}
