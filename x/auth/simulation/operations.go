package simulation

import (
	"context"
	"math/rand"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgChangePubKey = "op_weight_msg_change_pub_key"

	DefaultWeightMsgChangePubKey = 10
)

// AccountKeeper defines the account keeper methods used by the auth
// simulation operations.
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetParams(ctx context.Context) types.Params
	AddressCodec() address.Codec
}

// WeightedOperations returns all the operations from the auth module with
// their respective weights.
//
// MsgChangePubKey is simulated as a legacy operation rather than a simsx
// message factory: after a rotation the simulation must sign the account's
// later txs with the new key, and only a legacy operation has access to the
// simulation's account list.
func WeightedOperations(appParams simtypes.AppParams, txGen client.TxConfig, ak AccountKeeper) simulation.WeightedOperations {
	var weightMsgChangePubKey int
	appParams.GetOrGenerate(OpWeightMsgChangePubKey, &weightMsgChangePubKey, nil, func(_ *rand.Rand) {
		weightMsgChangePubKey = DefaultWeightMsgChangePubKey
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(weightMsgChangePubKey, SimulateMsgChangePubKey(txGen, ak)),
	}
}

// SimulateMsgChangePubKey rotates a random account to a newly generated
// secp256k1 or ML-DSA-65 key. On success it replaces the account's keys in
// the simulation's account list, so later operations sign with the new key.
func SimulateMsgChangePubKey(txGen client.TxConfig, ak AccountKeeper) simtypes.Operation {
	return simulateMsgChangePubKey(txGen, ak, genRekeyPrivKey)
}

// genRekeyPrivKey deterministically generates a secp256k1 or ML-DSA-65 private
// key from r.
func genRekeyPrivKey(r *rand.Rand) (cryptotypes.PrivKey, error) {
	if r.Intn(2) == 0 {
		// don't need that much entropy for simulation
		seed := make([]byte, 15)
		if _, err := r.Read(seed); err != nil {
			return nil, err
		}
		return secp256k1.GenPrivKeyFromSecret(seed), nil
	}
	seed := make([]byte, 32)
	if _, err := r.Read(seed); err != nil {
		return nil, err
	}
	priv, err := mldsa65.GenPrivKeyFromSeed(seed)
	if err != nil {
		return nil, err
	}
	return &priv, nil
}

func simulateMsgChangePubKey(
	txGen client.TxConfig,
	ak AccountKeeper,
	genKey func(*rand.Rand) (cryptotypes.PrivKey, error),
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgChangePubKey{})

		params := ak.GetParams(ctx)
		if !params.PubKeyChangeEnabled {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "pubkey change disabled"), nil, nil
		}

		simAccount, idx := simtypes.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAccount.Address)
		if acc == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "account not found"), nil, nil
		}
		if _, ok := acc.(sdk.ModuleAccountI); ok {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "module account"), nil, nil
		}
		// An account that has not signed a tx yet has no stored pubkey; the
		// ante handler sets it to the simulation account's key before the msg
		// runs.
		current := acc.GetPubKey()
		if current == nil {
			current = simAccount.PubKey
		}
		if !current.Equals(simAccount.PubKey) {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "simulation key does not match the account pubkey"), nil, nil
		}

		newPriv, err := genKey(r)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate new key"), nil, err
		}
		newPk := newPriv.PubKey()
		if err := types.ValidateRekeyPubKey(current, newPk, params); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "invalid new pubkey"), nil, nil
		}

		newPkAny, err := codectypes.NewAnyWithValue(newPk)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to pack new pubkey"), nil, err
		}
		addr, err := ak.AddressCodec().BytesToString(simAccount.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "invalid address"), nil, err
		}

		doc := types.ChangePubKeyProofDoc{
			ChainId:       ctx.ChainID(),
			AccountNumber: acc.GetAccountNumber(),
			Address:       addr,
			NewPubKey:     newPkAny,
		}
		signBytes, err := types.ChangePubKeyProofSignBytes(ak.AddressCodec(), doc, newPk)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to build proof sign bytes"), nil, err
		}
		sig, err := newPriv.Sign(signBytes)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to sign proof"), nil, err
		}
		proof, err := signing.SignatureDataToProto(&signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
			Signature: sig,
		}).Marshal()
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to encode proof"), nil, err
		}

		msg := &types.MsgChangePubKey{
			Address:   addr,
			NewPubKey: newPkAny,
			Proof:     proof,
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txGen,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			ModuleName:    types.ModuleName,
		}
		// No fees: simapp sims set no minimum gas price. Later ops from other
		// modules cover fee-paying txs signed with the rotated key.
		opMsg, futureOps, err := simulation.GenAndDeliverTx(txCtx, nil)
		if err != nil {
			return opMsg, futureOps, err
		}

		// Keep the simulation's keys in sync with the chain so later
		// operations sign for this account with the new key.
		accs[idx].PrivKey = newPriv
		accs[idx].PubKey = newPk

		return opMsg, futureOps, nil
	}
}
