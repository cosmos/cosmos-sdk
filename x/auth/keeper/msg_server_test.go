package keeper_test

import (
	"fmt"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (s *KeeperTestSuite) TestUpdateParams() {
	testCases := []struct {
		name      string
		req       *types.MsgUpdateParams
		expectErr bool
		expErrMsg string
	}{
		{
			name: "set invalid authority",
			req: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
			expErrMsg: "invalid authority",
		},
		{
			name: "set invalid max memo characters",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      0,
					TxSigLimit:             9,
					TxSizeCostPerByte:      5,
					SigVerifyCostED25519:   694,
					SigVerifyCostSecp256k1: 511,
					SigVerifyCostMlDsa65:   750,
					PubKeyChangeCost:       types.DefaultPubKeyChangeCost,
				},
			},
			expectErr: true,
			expErrMsg: "invalid max memo characters",
		},
		{
			name: "set invalid tx sig limit",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      140,
					TxSigLimit:             0,
					TxSizeCostPerByte:      5,
					SigVerifyCostED25519:   694,
					SigVerifyCostSecp256k1: 511,
					SigVerifyCostMlDsa65:   750,
					PubKeyChangeCost:       types.DefaultPubKeyChangeCost,
				},
			},
			expectErr: true,
			expErrMsg: "invalid tx signature limit",
		},
		{
			name: "set invalid tx size cost per bytes",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      140,
					TxSigLimit:             9,
					TxSizeCostPerByte:      0,
					SigVerifyCostED25519:   694,
					SigVerifyCostSecp256k1: 511,
					SigVerifyCostMlDsa65:   750,
					PubKeyChangeCost:       types.DefaultPubKeyChangeCost,
				},
			},
			expectErr: true,
			expErrMsg: "invalid tx size cost per byte",
		},
		{
			name: "set invalid sig verify cost ED25519",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      140,
					TxSigLimit:             9,
					TxSizeCostPerByte:      5,
					SigVerifyCostED25519:   0,
					SigVerifyCostSecp256k1: 511,
					SigVerifyCostMlDsa65:   750,
					PubKeyChangeCost:       types.DefaultPubKeyChangeCost,
				},
			},
			expectErr: true,
			expErrMsg: "invalid ED25519 signature verification cost",
		},
		{
			name: "set invalid sig verify cost Secp256k1",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      140,
					TxSigLimit:             9,
					TxSizeCostPerByte:      5,
					SigVerifyCostED25519:   694,
					SigVerifyCostSecp256k1: 0,
					SigVerifyCostMlDsa65:   750,
					PubKeyChangeCost:       types.DefaultPubKeyChangeCost,
				},
			},
			expectErr: true,
			expErrMsg: "invalid SECP256k1 signature verification cost",
		},
		{
			name: "set invalid pubkey change cost",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      140,
					TxSigLimit:             9,
					TxSizeCostPerByte:      5,
					SigVerifyCostED25519:   694,
					SigVerifyCostSecp256k1: 511,
					SigVerifyCostMlDsa65:   750,
					PubKeyChangeEnabled:    true,
					PubKeyChangeCost:       0,
				},
			},
			expectErr: true,
			expErrMsg: "invalid pubkey change cost",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.req)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateParamsAuthority() {
	keeperAuthority := s.accountKeeper.GetAuthority()
	overrideAuthority := sdk.AccAddress("override_authority___").String()

	validParams := types.Params{
		MaxMemoCharacters:      140,
		TxSigLimit:             9,
		TxSizeCostPerByte:      5,
		SigVerifyCostED25519:   694,
		SigVerifyCostSecp256k1: 511,
		SigVerifyCostMlDsa65:   750,
		PubKeyChangeCost:       types.DefaultPubKeyChangeCost,
	}

	s.Run("fallback to keeper authority", func() {
		// No consensus params authority set, keeper authority should work
		_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
			Authority: keeperAuthority,
			Params:    validParams,
		})
		s.Require().NoError(err)

		// A different address should fail
		_, err = s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
			Authority: overrideAuthority,
			Params:    validParams,
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "invalid authority")
	})

	s.Run("consensus params authority takes precedence", func() {
		ctx := s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Authority: &cmtproto.AuthorityParams{Authority: overrideAuthority},
		})

		// Override authority should now succeed
		_, err := s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: overrideAuthority,
			Params:    validParams,
		})
		s.Require().NoError(err)

		// Keeper authority should now fail
		_, err = s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: keeperAuthority,
			Params:    validParams,
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "invalid authority")
	})
}

const rekeyChainID = "rekey-test-chain"

type changePubKeyFixture struct {
	rekeyFixture
	ms types.MsgServer
}

func newChangePubKeyFixture(t *testing.T, enabled bool) changePubKeyFixture {
	t.Helper()
	f := newRekeyFixture(t)
	f.ctx = f.ctx.WithChainID(rekeyChainID)
	params := types.DefaultParams()
	params.PubKeyChangeEnabled = enabled
	require.NoError(t, f.ak.Params.Set(f.ctx, params))
	return changePubKeyFixture{rekeyFixture: f, ms: keeper.NewMsgServerImpl(f.ak)}
}

// changePubKeyMsg builds a MsgChangePubKey for acc rotating to newPk. Each
// signer signs the proof doc; one signer gives a single-signature proof and a
// multisig newPk gives a MultiSignatureData built from all signers.
func changePubKeyMsg(t *testing.T, chainID string, acc sdk.AccountI, newPk cryptotypes.PubKey, signers ...cryptotypes.PrivKey) *types.MsgChangePubKey {
	t.Helper()
	anyPk, err := codectypes.NewAnyWithValue(newPk)
	require.NoError(t, err)
	addr := acc.GetAddress().String()
	doc := types.ChangePubKeyProofDoc{
		ChainId:       chainID,
		AccountNumber: acc.GetAccountNumber(),
		Address:       addr,
		NewPubKey:     anyPk,
	}
	signBytes, err := types.ChangePubKeyProofSignBytes(authcodec.NewBech32Codec("cosmos"), doc, newPk)
	require.NoError(t, err)

	sign := func(sk cryptotypes.PrivKey) *signing.SingleSignatureData {
		sig, err := sk.Sign(signBytes)
		require.NoError(t, err)
		return &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, Signature: sig}
	}

	var proof signing.SignatureData
	if msig, ok := newPk.(*kmultisig.LegacyAminoPubKey); ok {
		keys := msig.GetPubKeys()
		multi := multisig.NewMultisig(len(keys))
		for _, sk := range signers {
			require.NoError(t, multisig.AddSignatureV2(multi, signing.SignatureV2{PubKey: sk.PubKey(), Data: sign(sk)}, keys))
		}
		proof = multi
	} else {
		require.Len(t, signers, 1)
		proof = sign(signers[0])
	}
	proofBz, err := signing.SignatureDataToProto(proof).Marshal()
	require.NoError(t, err)

	return &types.MsgChangePubKey{Address: addr, NewPubKey: anyPk, Proof: proofBz}
}

func TestChangePubKey_Disabled(t *testing.T) {
	f := newChangePubKeyFixture(t, false)
	acc := newAccountWithKey(t, f.ctx, f.ak, secp256k1.GenPrivKey().PubKey())
	sk1, pk1 := genMlDsa65Key(t)

	_, err := f.ms.ChangePubKey(f.ctx, changePubKeyMsg(t, rekeyChainID, acc, pk1, &sk1))
	require.ErrorIs(t, err, types.ErrPubKeyChangeDisabled)
}

func TestChangePubKey_Secp256k1ToMlDsa65(t *testing.T) {
	f := newChangePubKeyFixture(t, true)
	k0 := secp256k1.GenPrivKey().PubKey()
	acc := newAccountWithKey(t, f.ctx, f.ak, k0)
	addr := acc.GetAddress()
	sk1, pk1 := genMlDsa65Key(t)

	ctx := f.ctx.WithEventManager(sdk.NewEventManager())
	_, err := f.ms.ChangePubKey(ctx, changePubKeyMsg(t, rekeyChainID, acc, pk1, &sk1))
	require.NoError(t, err)

	got := f.ak.GetAccount(ctx, addr)
	require.True(t, got.GetPubKey().Equals(pk1))
	require.Equal(t, acc.GetAccountNumber(), got.GetAccountNumber())
	require.Len(t, historyFor(t, ctx, f.ak, addr), 1)

	var found bool
	for _, ev := range ctx.EventManager().Events() {
		if ev.Type != types.EventTypeChangePubKey {
			continue
		}
		found = true
		attrs := map[string]string{}
		for _, a := range ev.Attributes {
			attrs[a.Key] = a.Value
		}
		require.Equal(t, addr.String(), attrs[types.AttributeKeyAddress])
		require.Equal(t, sdk.AccAddress(k0.Address()).String(), attrs[types.AttributeKeyOldPubKeyAddress])
		require.Equal(t, sdk.AccAddress(pk1.Address()).String(), attrs[types.AttributeKeyNewPubKeyAddress])
	}
	require.True(t, found, "change_pubkey event not emitted")
}

func TestChangePubKey_MultisigMembershipChanges(t *testing.T) {
	f := newChangePubKeyFixture(t, true)
	acc := newAccountWithKey(t, f.ctx, f.ak, secp256k1.GenPrivKey().PubKey())
	addr := acc.GetAddress()

	skA := secp256k1.GenPrivKey()
	skB, _ := genMlDsa65Key(t)
	skC := secp256k1.GenPrivKey()
	skD, _ := genMlDsa65Key(t)

	steps := []struct {
		name    string
		newPk   cryptotypes.PubKey
		signers []cryptotypes.PrivKey
	}{
		{
			name:    "secp256k1 to 2-of-3",
			newPk:   kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{skA.PubKey(), skB.PubKey(), skC.PubKey()}),
			signers: []cryptotypes.PrivKey{skA, &skB},
		},
		{
			name:    "2-of-3 to 3-of-4 (add a member)",
			newPk:   kmultisig.NewLegacyAminoPubKey(3, []cryptotypes.PubKey{skA.PubKey(), skB.PubKey(), skC.PubKey(), skD.PubKey()}),
			signers: []cryptotypes.PrivKey{skA, &skB, &skD},
		},
		{
			name:    "3-of-4 to 2-of-2 (remove members)",
			newPk:   kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{skA.PubKey(), skD.PubKey()}),
			signers: []cryptotypes.PrivKey{skA, &skD},
		},
	}
	for i, step := range steps {
		cur := f.ak.GetAccount(f.ctx, addr)
		_, err := f.ms.ChangePubKey(f.ctx, changePubKeyMsg(t, rekeyChainID, cur, step.newPk, step.signers...))
		require.NoError(t, err, step.name)

		got := f.ak.GetAccount(f.ctx, addr)
		require.NotNil(t, got, step.name)
		require.Equal(t, addr, got.GetAddress(), step.name)
		require.True(t, got.GetPubKey().Equals(step.newPk), step.name)
		require.Len(t, historyFor(t, f.ctx, f.ak, addr), i+1, step.name)
	}
}

func TestChangePubKey_DecodedMultisigMsg(t *testing.T) {
	// A msg that went through the codec (as in a tx) must have its multisig
	// subkeys unpacked by MsgChangePubKey.UnpackInterfaces.
	f := newChangePubKeyFixture(t, true)
	acc := newAccountWithKey(t, f.ctx, f.ak, secp256k1.GenPrivKey().PubKey())
	skA := secp256k1.GenPrivKey()
	skB, _ := genMlDsa65Key(t)
	newPk := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{skA.PubKey(), skB.PubKey()})

	msg := changePubKeyMsg(t, rekeyChainID, acc, newPk, skA, &skB)
	bz, err := f.encCfg.Codec.Marshal(msg)
	require.NoError(t, err)
	var decoded types.MsgChangePubKey
	require.NoError(t, f.encCfg.Codec.Unmarshal(bz, &decoded))

	_, err = f.ms.ChangePubKey(f.ctx, &decoded)
	require.NoError(t, err)
	require.True(t, f.ak.GetAccount(f.ctx, acc.GetAddress()).GetPubKey().Equals(newPk))
}

func TestChangePubKey_MalformedNewKey(t *testing.T) {
	f := newChangePubKeyFixture(t, true)
	acc := newAccountWithKey(t, f.ctx, f.ak, secp256k1.GenPrivKey().PubKey())
	addr := acc.GetAddress().String()

	shortKey, err := codectypes.NewAnyWithValue(&secp256k1.PubKey{Key: []byte{1, 2, 3}})
	require.NoError(t, err)

	subBz, err := (&secp256k1.PubKey{Key: secp256k1.GenPrivKey().PubKey().Bytes()}).Marshal()
	require.NoError(t, err)
	unpackedMsig := &kmultisig.LegacyAminoPubKey{
		Threshold: 1,
		PubKeys:   []*codectypes.Any{{TypeUrl: "/cosmos.crypto.secp256k1.PubKey", Value: subBz}},
	}
	unpackedMsigAny, err := codectypes.NewAnyWithValue(unpackedMsig)
	require.NoError(t, err)

	// kmultisig.AminoCdc does not register secp256r1, so Address() on this
	// multisig panics; the handler must reject it before calling Address().
	r1, err := secp256r1.GenPrivKey()
	require.NoError(t, err)
	r1MsigAny, err := codectypes.NewAnyWithValue(kmultisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{
		r1.PubKey(), secp256k1.GenPrivKey().PubKey(),
	}))
	require.NoError(t, err)

	cases := map[string]*codectypes.Any{
		"multisig with secp256r1 subkey": r1MsigAny,
		"nil new key":                    nil,
		"wrong-length secp256k1 key":     shortKey,
		"multisig with un-unpacked key":  unpackedMsigAny,
		"any without a cached value":     {TypeUrl: "/cosmos.crypto.secp256k1.PubKey", Value: subBz},
	}
	for name, anyPk := range cases {
		t.Run(name, func(t *testing.T) {
			msg := &types.MsgChangePubKey{Address: addr, NewPubKey: anyPk, Proof: []byte("not a proof")}
			require.NotPanics(t, func() {
				_, err = f.ms.ChangePubKey(f.ctx, msg)
			})
			require.ErrorIs(t, err, types.ErrInvalidNewPubKey)
			require.True(t, f.ak.GetAccount(f.ctx, acc.GetAddress()).GetPubKey().Equals(acc.GetPubKey()))
		})
	}
}

// TestChangePubKey_TooDeeplyNestedMultisig checks that an account cannot
// rotate to a multisig nested deeper than the ante handler can flatten:
// every later tx from it would fail, including one rotating back.
func TestChangePubKey_TooDeeplyNestedMultisig(t *testing.T) {
	for _, tc := range []struct {
		levels int
		ok     bool
	}{
		{levels: 2, ok: true},
		{levels: 3, ok: false},
	} {
		t.Run(fmt.Sprintf("%d levels", tc.levels), func(t *testing.T) {
			f := newChangePubKeyFixture(t, true)
			acc := newAccountWithKey(t, f.ctx, f.ak, secp256k1.GenPrivKey().PubKey())
			addr := acc.GetAddress().String()

			// newPk is 1-of-1(...1-of-1(sk)...) with tc.levels multisigs.
			sk := secp256k1.GenPrivKey()
			newPk := sk.PubKey()
			for range tc.levels {
				newPk = kmultisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{newPk})
			}
			anyPk, err := codectypes.NewAnyWithValue(newPk)
			require.NoError(t, err)
			doc := types.ChangePubKeyProofDoc{
				ChainId:       rekeyChainID,
				AccountNumber: acc.GetAccountNumber(),
				Address:       addr,
				NewPubKey:     anyPk,
			}
			signBytes, err := types.ChangePubKeyProofSignBytes(authcodec.NewBech32Codec("cosmos"), doc, newPk)
			require.NoError(t, err)
			sig, err := sk.Sign(signBytes)
			require.NoError(t, err)

			// A valid proof: the same nesting of 1-of-1 multisignatures.
			var proof signing.SignatureData = &signing.SingleSignatureData{
				SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, Signature: sig,
			}
			for range tc.levels {
				multi := multisig.NewMultisig(1)
				multi.BitArray.SetIndex(0, true)
				multi.Signatures = []signing.SignatureData{proof}
				proof = multi
			}
			require.NoError(t, types.VerifyChangePubKeyProof(newPk, signBytes, proof))
			proofBz, err := signing.SignatureDataToProto(proof).Marshal()
			require.NoError(t, err)

			_, err = f.ms.ChangePubKey(f.ctx, &types.MsgChangePubKey{Address: addr, NewPubKey: anyPk, Proof: proofBz})
			if tc.ok {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, types.ErrInvalidNewPubKey)
			require.True(t, f.ak.GetAccount(f.ctx, acc.GetAddress()).GetPubKey().Equals(acc.GetPubKey()))
		})
	}
}

func TestChangePubKey_NotRekeyable(t *testing.T) {
	f := newChangePubKeyFixture(t, true)
	sk1, pk1 := genMlDsa65Key(t)

	cred, err := types.NewModuleCredential("group", []byte{1})
	require.NoError(t, err)
	credAddr := sdk.AccAddress(cred.Address())
	credAcc, err := types.NewBaseAccountWithPubKey(cred)
	require.NoError(t, err)
	f.ak.SetAccount(f.ctx, f.ak.NewAccount(f.ctx, credAcc))

	unknown := types.NewBaseAccountWithAddress(sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))

	// An account with no stored pubkey, such as a contract or interchain
	// account dispatching msgs as itself.
	noKeyAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	f.ak.SetAccount(f.ctx, f.ak.NewAccountWithAddress(f.ctx, noKeyAddr))

	cases := map[string]sdk.AccountI{
		"account without a pubkey":  f.ak.GetAccount(f.ctx, noKeyAddr),
		"module account":            f.ak.GetModuleAccount(f.ctx, types.FeeCollectorName),
		"module credential account": f.ak.GetAccount(f.ctx, credAddr),
		"unknown account":           unknown,
	}
	for name, acc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := f.ms.ChangePubKey(f.ctx, changePubKeyMsg(t, rekeyChainID, acc, pk1, &sk1))
			require.ErrorIs(t, err, types.ErrAccountNotRekeyable)
		})
	}
	require.True(t, f.ak.GetAccount(f.ctx, credAddr).GetPubKey().Equals(cred))
	require.Nil(t, f.ak.GetAccount(f.ctx, noKeyAddr).GetPubKey())
}

// TestChangePubKey_ModuleAccountWithPubKey checks that a module account is
// rejected even when it has a stored secp256k1 pubkey, so the ModuleAccountI
// check is the only thing that blocks the rekey (the nil-pubkey and
// ModuleCredential checks do not apply).
func TestChangePubKey_ModuleAccountWithPubKey(t *testing.T) {
	f := newChangePubKeyFixture(t, true)
	sk1, pk1 := genMlDsa65Key(t)

	oldPk := secp256k1.GenPrivKey().PubKey()
	ba := types.NewBaseAccountWithAddress(types.NewModuleAddress("rekeytest"))
	require.NoError(t, ba.SetPubKey(oldPk))
	f.ak.SetAccount(f.ctx, f.ak.NewAccount(f.ctx, types.NewModuleAccount(ba, "rekeytest")))

	acc := f.ak.GetAccount(f.ctx, ba.GetAddress())
	_, isModAcc := acc.(sdk.ModuleAccountI)
	require.True(t, isModAcc)
	require.True(t, acc.GetPubKey().Equals(oldPk))

	_, err := f.ms.ChangePubKey(f.ctx, changePubKeyMsg(t, rekeyChainID, acc, pk1, &sk1))
	require.ErrorIs(t, err, types.ErrAccountNotRekeyable)
	require.ErrorContains(t, err, "is a module account")

	after := f.ak.GetAccount(f.ctx, ba.GetAddress())
	require.True(t, after.GetPubKey().Equals(oldPk))
	require.Empty(t, historyFor(t, f.ctx, f.ak, ba.GetAddress()))
}

func TestChangePubKey_BadProof(t *testing.T) {
	f := newChangePubKeyFixture(t, true)
	k0 := secp256k1.GenPrivKey().PubKey()
	acc := newAccountWithKey(t, f.ctx, f.ak, k0)
	addr := acc.GetAddress()
	sk1, pk1 := genMlDsa65Key(t)
	otherSk, _ := genMlDsa65Key(t)

	cases := map[string]*types.MsgChangePubKey{
		"signed by another key":   changePubKeyMsg(t, rekeyChainID, acc, pk1, &otherSk),
		"proof for another chain": changePubKeyMsg(t, "other-chain", acc, pk1, &sk1),
		"garbage proof": func() *types.MsgChangePubKey {
			m := changePubKeyMsg(t, rekeyChainID, acc, pk1, &otherSk)
			m.Proof = []byte{0xff, 0xff, 0xff}
			return m
		}(),
	}
	for name, msg := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := f.ms.ChangePubKey(f.ctx, msg)
			require.ErrorIs(t, err, types.ErrInvalidPubKeyProof)
			got := f.ak.GetAccount(f.ctx, addr)
			require.True(t, got.GetPubKey().Equals(k0))
			require.Empty(t, historyFor(t, f.ctx, f.ak, addr))
			require.Empty(t, indexEntries(t, f.ctx, f.ak))
		})
	}
}

func TestChangePubKey_ConsumesGas(t *testing.T) {
	// Rotate to a secp256k1 key so store gas stays well below the cost, and
	// compare two runs that differ only in PubKeyChangeCost.
	run := func(cost uint64) uint64 {
		f := newChangePubKeyFixture(t, true)
		params := f.ak.GetParams(f.ctx)
		params.PubKeyChangeCost = cost
		require.NoError(t, f.ak.Params.Set(f.ctx, params))

		acc := newAccountWithKey(t, f.ctx, f.ak, secp256k1.GenPrivKey().PubKey())
		sk1 := secp256k1.GenPrivKey()
		ctx := f.ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
		_, err := f.ms.ChangePubKey(ctx, changePubKeyMsg(t, rekeyChainID, acc, sk1.PubKey(), sk1))
		require.NoError(t, err)
		return ctx.GasMeter().GasConsumed()
	}

	base := run(types.DefaultPubKeyChangeCost)
	require.GreaterOrEqual(t, base, types.DefaultPubKeyChangeCost)
	require.Equal(t, base+100_000, run(types.DefaultPubKeyChangeCost+100_000))
}

func TestChangePubKey_BadProofConsumesGas(t *testing.T) {
	// The cost is charged before the proof is decoded and verified, so a tx
	// with a failing proof still pays for the verification work.
	f := newChangePubKeyFixture(t, true)
	acc := newAccountWithKey(t, f.ctx, f.ak, secp256k1.GenPrivKey().PubKey())
	_, pk1 := genMlDsa65Key(t)
	otherSk, _ := genMlDsa65Key(t)

	ctx := f.ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
	_, err := f.ms.ChangePubKey(ctx, changePubKeyMsg(t, rekeyChainID, acc, pk1, &otherSk))
	require.ErrorIs(t, err, types.ErrInvalidPubKeyProof)
	require.GreaterOrEqual(t, ctx.GasMeter().GasConsumed(), types.DefaultPubKeyChangeCost)
}
