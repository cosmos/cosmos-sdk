package base

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/accountstd"
	v1 "cosmossdk.io/x/accounts/defaults/base/v1"
	aa_interface_v1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	"cosmossdk.io/x/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

func setupBaseAccount(t *testing.T, ss store.KVStoreService) Account {
	t.Helper()
	deps := makeMockDependencies(ss)
	handler := directHandler{}

	createAccFn := NewAccount("base", signing.NewHandlerMap(handler))
	_, acc, err := createAccFn(deps)
	baseAcc := acc.(Account)
	require.NoError(t, err)

	return baseAcc
}

func TestInit(t *testing.T) {
	ctx, ss := newMockContext(t)
	baseAcc := setupBaseAccount(t, ss)
	_, err := baseAcc.Init(ctx, &v1.MsgInit{
		PubKey: secp256k1.GenPrivKey().PubKey().Bytes(),
	})
	require.NoError(t, err)

	testcases := []struct {
		name     string
		msg      *v1.MsgInit
		isExpErr bool
	}{
		{
			"valid init",
			&v1.MsgInit{
				PubKey: secp256k1.GenPrivKey().PubKey().Bytes(),
			},
			false,
		},
		{
			"invalid pubkey",
			&v1.MsgInit{
				PubKey: []byte("invalid_pk"),
			},
			true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := baseAcc.Init(ctx, tc.msg)
			if tc.isExpErr {
				require.NotNil(t, err, tc.name)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestSwapKey(t *testing.T) {
	ctx, ss := newMockContext(t)
	baseAcc := setupBaseAccount(t, ss)
	_, err := baseAcc.Init(ctx, &v1.MsgInit{
		PubKey: secp256k1.GenPrivKey().PubKey().Bytes(),
	})
	require.NoError(t, err)

	testcases := []struct {
		name     string
		genCtx   func(ctx context.Context) context.Context
		msg      *v1.MsgSwapPubKey
		isExpErr bool
		expErr   error
	}{
		{
			"valid transaction",
			func(ctx context.Context) context.Context {
				return accountstd.SetSender(ctx, []byte("mock_base_account"))
			},
			&v1.MsgSwapPubKey{
				NewPubKey: secp256k1.GenPrivKey().PubKey().Bytes(),
			},
			false,
			nil,
		},
		{
			"invalid transaction, sender is not self",
			func(ctx context.Context) context.Context {
				return accountstd.SetSender(ctx, []byte("sender"))
			},
			&v1.MsgSwapPubKey{
				NewPubKey: secp256k1.GenPrivKey().PubKey().Bytes(),
			},
			true,
			errors.New("unauthorized"),
		},
		{
			"invalid transaction, invalid pubkey",
			func(ctx context.Context) context.Context {
				return accountstd.SetSender(ctx, []byte("mock_base_account"))
			},
			&v1.MsgSwapPubKey{
				NewPubKey: []byte("invalid_pk"),
			},
			true,
			nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.genCtx != nil {
				ctx = tc.genCtx(ctx)
			}
			_, err := baseAcc.SwapPubKey(ctx, tc.msg)
			if tc.isExpErr {
				if tc.expErr != nil {
					require.Equal(t, tc.expErr, err, tc.name)
				} else {
					require.NotNil(t, err, tc.name)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	ctx, ss := newMockContext(t)
	baseAcc := setupBaseAccount(t, ss)
	privKey := secp256k1.GenPrivKey()
	pkAny, err := codectypes.NewAnyWithValue(privKey.PubKey())
	require.NoError(t, err)
	_, err = baseAcc.Init(ctx, &v1.MsgInit{
		PubKey: privKey.PubKey().Bytes(),
	})
	require.NoError(t, err)

	ctx = accountstd.SetSender(ctx, address.Module("accounts"))
	require.NoError(t, err)

	transaction := tx.Tx{
		Body: &tx.TxBody{},
		AuthInfo: &tx.AuthInfo{
			SignerInfos: []*tx.SignerInfo{
				{
					PublicKey: pkAny,
					ModeInfo: &tx.ModeInfo{
						Sum: &tx.ModeInfo_Single_{
							Single: &tx.ModeInfo_Single{
								Mode: 1,
							},
						},
					},
					Sequence: 0,
				},
			},
		},
		Signatures: [][]byte{},
	}

	bodyByte, err := transaction.Body.Marshal()
	require.NoError(t, err)
	authByte, err := transaction.AuthInfo.Marshal()
	require.NoError(t, err)

	txDoc := tx.SignDoc{
		BodyBytes:     bodyByte,
		AuthInfoBytes: authByte,
		ChainId:       "test",
		AccountNumber: 1,
	}
	signBytes, err := txDoc.Marshal()
	require.NoError(t, err)

	sig, err := privKey.Sign(signBytes)
	require.NoError(t, err)

	transaction.Signatures = append(transaction.Signatures, sig)

	rawTx := tx.TxRaw{
		BodyBytes:     bodyByte,
		AuthInfoBytes: authByte,
		Signatures:    transaction.Signatures,
	}

	_, err = baseAcc.Authenticate(ctx, &aa_interface_v1.MsgAuthenticate{
		RawTx:       &rawTx,
		Tx:          &transaction,
		SignerIndex: 0,
	})
	require.NoError(t, err)

	// testing with invalid signature

	// update sequence number
	transaction = tx.Tx{
		Body: &tx.TxBody{},
		AuthInfo: &tx.AuthInfo{
			SignerInfos: []*tx.SignerInfo{
				{
					PublicKey: pkAny,
					ModeInfo: &tx.ModeInfo{
						Sum: &tx.ModeInfo_Single_{
							Single: &tx.ModeInfo_Single{
								Mode: 1,
							},
						},
					},
					Sequence: 1,
				},
			},
		},
		Signatures: [][]byte{},
	}
	authByte, err = transaction.AuthInfo.Marshal()
	require.NoError(t, err)

	txDoc.BodyBytes = []byte("invalid_msg")
	txDoc.AuthInfoBytes = authByte
	signBytes, err = txDoc.Marshal()
	require.NoError(t, err)
	invalidSig, err := privKey.Sign(signBytes)
	require.NoError(t, err)

	transaction.Signatures = append([][]byte{}, invalidSig)

	rawTx.Signatures = transaction.Signatures

	_, err = baseAcc.Authenticate(ctx, &aa_interface_v1.MsgAuthenticate{
		RawTx:       &rawTx,
		Tx:          &transaction,
		SignerIndex: 0,
	})
	require.Equal(t, errors.New("signature verification failed"), err)
}
