package base

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	types "github.com/cosmos/gogoproto/types/any"
	dcrd_secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/accountstd"
	v1 "cosmossdk.io/x/accounts/defaults/base/v1"
	aa_interface_v1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	"cosmossdk.io/x/tx/signing"

	authn "cosmossdk.io/x/accounts/defaults/base/keys/authn"
	cometcrypto "github.com/cometbft/cometbft/crypto"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type CeremonyType string

type TokenBindingStatus string

type TokenBinding struct {
	Status TokenBindingStatus `json:"status"`
	Id     string             `json:"id"`
}

type CollectedClientData struct {
	// Type the string "webauthn.create" when creating new credentials,
	// and "webauthn.get" when getting an assertion from an existing credential. The
	// purpose of this member is to prevent certain types of signature confusion attacks
	//(where an attacker substitutes one legitimate signature for another).
	Type         CeremonyType  `json:"type"`
	Challenge    string        `json:"challenge"`
	Origin       string        `json:"origin"`
	TokenBinding *TokenBinding `json:"tokenBinding,omitempty"`
	// Chromium (Chrome) returns a hint sometimes about how to handle clientDataJSON in a safe manner
	Hint string `json:"new_keys_may_be_added_here,omitempty"`
}

func GenerateAuthnKey(t *testing.T) (*ecdsa.PrivateKey, authn.AuthnPubKey) {
	t.Helper()
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	require.NoError(t, err)
	pkBytes := elliptic.MarshalCompressed(curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
	pk := authn.AuthnPubKey{
		KeyId: "a099eda0fb05e5783379f73a06acca726673b8e07e436edcd0d71645982af65c",
		Key:   pkBytes,
	}

	return privateKey, pk
}

func GenerateClientData(t *testing.T, msg []byte) []byte {
	t.Helper()
	clientData := CollectedClientData{
		// purpose of this member is to prevent certain types of signature confusion attacks
		//(where an attacker substitutes one legitimate signature for another).
		Type:         "webauthn.create",
		Challenge:    base64.RawURLEncoding.EncodeToString(msg),
		Origin:       "https://blue.kujira.network",
		TokenBinding: nil,
		Hint:         "",
	}
	clientDataJSON, err := json.Marshal(clientData)
	require.NoError(t, err)
	return clientDataJSON
}

func setupBaseAccount(t *testing.T, ss store.KVStoreService) Account {
	t.Helper()
	deps := makeMockDependencies(ss)
	handler := directHandler{}

	createAccFn := NewAccount(
		"base",
		signing.NewHandlerMap(handler),
		WithPubKeyWithValidationFunc(func(pt *secp256k1.PubKey) error {
			_, err := dcrd_secp256k1.ParsePubKey(pt.Key)
			return err
		}),
		WithPubKeyWithValidationFunc(func(pt *authn.AuthnPubKey) error {
			return nil
		}),
	)
	_, acc, err := createAccFn(deps)
	baseAcc := acc.(Account)
	require.NoError(t, err)

	return baseAcc
}

func TestInit(t *testing.T) {
	ctx, ss := newMockContext(t)
	baseAcc := setupBaseAccount(t, ss)
	_, err := baseAcc.Init(ctx, &v1.MsgInit{
		PubKey: toAnyPb(t, secp256k1.GenPrivKey().PubKey()),
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
				PubKey: toAnyPb(t, secp256k1.GenPrivKey().PubKey()),
			},
			false,
		},

		{
			"invalid pubkey",
			&v1.MsgInit{
				PubKey: toAnyPb(t, &secp256k1.PubKey{Key: []byte("invalid")}),
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
		PubKey: toAnyPb(t, secp256k1.GenPrivKey().PubKey()),
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
				NewPubKey: toAnyPb(t, secp256k1.GenPrivKey().PubKey()),
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
				NewPubKey: toAnyPb(t, secp256k1.GenPrivKey().PubKey()),
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
				NewPubKey: toAnyPb(t, &secp256k1.PubKey{Key: []byte("invalid")}),
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
		PubKey: toAnyPb(t, privKey.PubKey()),
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

func toAnyPb(t *testing.T, pm gogoproto.Message) *codectypes.Any {
	t.Helper()
	if gogoproto.MessageName(pm) == gogoproto.MessageName(&types.Any{}) {
		t.Fatal("no")
	}
	pb, err := codectypes.NewAnyWithValue(pm)
	require.NoError(t, err)
	return pb
}

func TestAuthenticateAuthn(t *testing.T) {
	ctx, ss := newMockContext(t)
	baseAcc := setupBaseAccount(t, ss)
	privKey, pk := GenerateAuthnKey(t)
	pkAny, err := codectypes.NewAnyWithValue(&pk)
	require.NoError(t, err)
	_, err = baseAcc.Init(ctx, &v1.MsgInit{
		PubKey: toAnyPb(t, &pk),
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

	clientDataJson := GenerateClientData(t, signBytes)
	clientDataHash := sha256.Sum256(clientDataJson)
	authenticatorData := cometcrypto.CRandBytes(37)
	payload := append(authenticatorData, clientDataHash[:]...)

	h := crypto.SHA256.New()
	h.Write(payload)
	digest := h.Sum(nil)

	sig, err := ecdsa.SignASN1(rand.Reader, privKey, digest)

	require.NoError(t, err)

	cborSig := authn.Signature{
		AuthenticatorData: hex.EncodeToString(authenticatorData),
		ClientDataJSON:    hex.EncodeToString(clientDataJson),
		Signature:         hex.EncodeToString(sig),
	}

	sigBytes, err := json.Marshal(cborSig)

	transaction.Signatures = append(transaction.Signatures, sigBytes)

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

	invalidClientDataJson := GenerateClientData(t, signBytes)
	invalidClientDataHash := sha256.Sum256(invalidClientDataJson)
	invalidAuthenticatorData := cometcrypto.CRandBytes(37)
	invalidPayload := append(invalidAuthenticatorData, invalidClientDataHash[:]...)

	ivh := crypto.SHA256.New()
	ivh.Write(invalidPayload)
	ivdigest := ivh.Sum(nil)

	invalidSig, err := ecdsa.SignASN1(rand.Reader, privKey, ivdigest)
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
