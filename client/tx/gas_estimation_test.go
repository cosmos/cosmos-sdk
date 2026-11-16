package tx

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type gasSimulationRecorder struct {
	mockContext
	t      *testing.T
	config client.TxConfig
	calls  []sdk.FeeTx
	result func(sdk.FeeTx, int) (*txtypes.SimulateResponse, error)
}

func (m *gasSimulationRecorder) Invoke(_ context.Context, method string, req, reply any, _ ...grpc.CallOption) error {
	require.Equal(m.t, "/cosmos.tx.v1beta1.Service/Simulate", method)
	decoded, err := m.config.TxDecoder()(req.(*txtypes.SimulateRequest).TxBytes)
	require.NoError(m.t, err)
	feeTx := decoded.(sdk.FeeTx)
	m.calls = append(m.calls, feeTx)
	res, err := m.result(feeTx, len(m.calls))
	if res != nil {
		*reply.(*txtypes.SimulateResponse) = *res
	}
	return err
}

func gasResponse(used uint64) *txtypes.SimulateResponse {
	return &txtypes.SimulateResponse{GasInfo: &sdk.GasInfo{GasUsed: used}}
}

func TestCalculateGasFeeBearingCandidates(t *testing.T) {
	for _, adjustment := range []float64{0.5, 1, 1.05, 1.5} {
		t.Run(fmt.Sprint(adjustment), func(t *testing.T) {
			config, _ := newTestTxConfig()
			f := Factory{}.WithTxConfig(config).WithChainID("test").WithGasPrices("0.025stake").WithGasAdjustment(adjustment)
			recorder := &gasSimulationRecorder{t: t, config: config, result: func(tx sdk.FeeTx, _ int) (*txtypes.SimulateResponse, error) {
				used := uint64(36722)
				if !tx.GetFee().IsZero() {
					used = 53751
				}
				return gasResponse(used), nil
			}}
			res, limit, err := CalculateGas(recorder, f)
			require.NoError(t, err)
			require.EqualValues(t, 53751, res.GasInfo.GasUsed)
			want, err := adjustedGas(53751, adjustment)
			require.NoError(t, err)
			require.Equal(t, want, limit)
			require.Len(t, recorder.calls, 3)
			require.Zero(t, recorder.calls[0].GetGas())
			require.True(t, recorder.calls[0].GetFee().IsZero())
			defaultMode, err := authsigning.APISignModeToInternal(config.SignModeHandler().DefaultMode())
			require.NoError(t, err)
			for _, call := range recorder.calls[1:] {
				require.Positive(t, call.GetGas())
				require.False(t, call.GetFee().IsZero())
				sigs, err := call.(authsigning.SigVerifiableTx).GetSignaturesV2()
				require.NoError(t, err)
				require.Equal(t, defaultMode, sigs[0].Data.(*signing.SingleSignatureData).SignMode)
				require.Empty(t, sigs[0].Data.(*signing.SingleSignatureData).Signature)
			}
			built, err := f.WithGas(limit).BuildUnsignedTx()
			require.NoError(t, err)
			require.Equal(t, built.GetTx().GetFee(), recorder.calls[2].GetFee())
			require.Equal(t, limit, recorder.calls[2].GetGas())
			require.Zero(t, f.Gas(), "estimation must not mutate the factory")
		})
	}
}

func TestCalculateGasFeeOptions(t *testing.T) {
	config, _ := newTestTxConfig()
	payer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	granter := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	for _, tc := range []struct {
		name string
		f    Factory
		fees sdk.Coins
	}{
		{"explicit fees", Factory{}.WithFees("1378stake"), sdk.NewCoins(sdk.NewInt64Coin("stake", 1378))},
		{"zero prices", Factory{}.WithGasPrices("0stake"), nil},
		{"no fees", Factory{}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.f.WithTxConfig(config).WithChainID("test").WithGas(200000).WithGasAdjustment(1).
				WithFeePayer(payer).WithFeeGranter(granter).WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
			recorder := &gasSimulationRecorder{t: t, config: config, result: func(tx sdk.FeeTx, _ int) (*txtypes.SimulateResponse, error) {
				require.True(t, tc.fees.Equal(tx.GetFee()))
				require.Equal(t, []byte(payer), tx.FeePayer())
				require.Equal(t, []byte(granter), tx.FeeGranter())
				sigs, err := tx.(authsigning.SigVerifiableTx).GetSignaturesV2()
				require.NoError(t, err)
				require.Equal(t, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, sigs[0].Data.(*signing.SingleSignatureData).SignMode)
				return gasResponse(100), nil
			}}
			_, limit, err := CalculateGas(recorder, f)
			require.NoError(t, err)
			require.EqualValues(t, 100, limit)
			require.Zero(t, recorder.calls[0].GetGas(), "dry-run must not bootstrap from manual gas")
		})
	}
}

func TestCalculateGasFailuresAndConvergence(t *testing.T) {
	rpcErr := errors.New("simulation unavailable")
	for _, tc := range []struct {
		name    string
		result  func(sdk.FeeTx, int) (*txtypes.SimulateResponse, error)
		calls   int
		limit   uint64
		wantErr string
	}{
		{"bootstrap failure", func(sdk.FeeTx, int) (*txtypes.SimulateResponse, error) { return nil, rpcErr }, 1, 0, "simulation unavailable"},
		{"fee payment failure", func(_ sdk.FeeTx, n int) (*txtypes.SimulateResponse, error) {
			if n == 2 {
				return nil, rpcErr
			}
			return gasResponse(100), nil
		}, 2, 0, "simulation unavailable"},
		{"missing gas information", func(sdk.FeeTx, int) (*txtypes.SimulateResponse, error) { return &txtypes.SimulateResponse{}, nil }, 1, 0, "no gas information"},
		{"nonconvergence", func(tx sdk.FeeTx, _ int) (*txtypes.SimulateResponse, error) { return gasResponse(tx.GetGas() + 1), nil }, 4, 0, "did not stabilize"},
		{"fourth call verifies final fields", func(_ sdk.FeeTx, n int) (*txtypes.SimulateResponse, error) {
			if n >= 3 {
				return gasResponse(101), nil
			}
			return gasResponse(100), nil
		}, 4, 101, ""},
		{"candidate never decreases", func(_ sdk.FeeTx, n int) (*txtypes.SimulateResponse, error) {
			if n > 1 {
				return gasResponse(90), nil
			}
			return gasResponse(100), nil
		}, 3, 100, ""},
		{"zero measurement", func(sdk.FeeTx, int) (*txtypes.SimulateResponse, error) { return gasResponse(0), nil }, 3, 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config, _ := newTestTxConfig()
			recorder := &gasSimulationRecorder{t: t, config: config, result: tc.result}
			res, limit, err := CalculateGas(recorder, Factory{}.WithTxConfig(config).WithChainID("test").WithGasAdjustment(1))
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, res)
				if tc.wantErr == "simulation unavailable" {
					require.ErrorIs(t, err, rpcErr)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.limit, limit)
				require.Equal(t, limit, recorder.calls[len(recorder.calls)-1].GetGas())
			}
			require.Len(t, recorder.calls, tc.calls)
		})
	}
}

func TestCalculateGasRejectsInvalidAdjustmentBeforeSimulation(t *testing.T) {
	for _, adjustment := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		t.Run(fmt.Sprint(adjustment), func(t *testing.T) {
			config, _ := newTestTxConfig()
			recorder := &gasSimulationRecorder{t: t, config: config}
			_, _, err := CalculateGas(recorder, Factory{}.WithTxConfig(config).WithChainID("test").WithGasAdjustment(adjustment))
			require.Error(t, err)
			require.Empty(t, recorder.calls)
		})
	}
}

func TestAdjustedGas(t *testing.T) {
	for _, tc := range []struct {
		name       string
		used       uint64
		adjustment float64
		want       uint64
		wantErr    bool
	}{
		{"round upward", 101, 1.05, 107, false},
		{"decimal product", 100, 1.1, 110, false},
		{"uint64 precision", 1<<53 + 1, 1, 1<<53 + 1, false},
		{"maximum uint64", math.MaxUint64, 1, math.MaxUint64, false},
		{"overflow", math.MaxUint64, 1.05, 0, true},
		{"zero adjustment", 101, 0, 0, false},
		{"subunit adjustment", 101, 0.5, 51, false},
		{"negative", 101, -1, 0, true},
		{"NaN", 101, math.NaN(), 0, true},
		{"infinite", 101, math.Inf(1), 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := adjustedGas(tc.used, tc.adjustment)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestCalculateGasMultisig(t *testing.T) {
	config, cdc := newTestTxConfig()
	kb := keyring.NewInMemory(cdc)
	pubkeys := []cryptotypes.PubKey{secp256k1.GenPrivKey().PubKey(), secp256k1.GenPrivKey().PubKey()}
	_, err := kb.SaveMultisig("multi", multisig.NewLegacyAminoPubKey(2, pubkeys))
	require.NoError(t, err)
	f := Factory{}.WithTxConfig(config).WithChainID("test").WithKeybase(kb).WithFromName("multi").
		WithSimulateAndExecute(true).WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON).
		WithGasPrices("0.025stake").WithGasAdjustment(1)
	recorder := &gasSimulationRecorder{t: t, config: config, result: func(tx sdk.FeeTx, _ int) (*txtypes.SimulateResponse, error) {
		sigs, err := tx.(authsigning.SigVerifiableTx).GetSignaturesV2()
		require.NoError(t, err)
		data := sigs[0].Data.(*signing.MultiSignatureData)
		require.Equal(t, 2, data.BitArray.Count())
		require.True(t, data.BitArray.GetIndex(0))
		require.True(t, data.BitArray.GetIndex(1))
		require.Len(t, data.Signatures, 2)
		return gasResponse(100), nil
	}}
	_, _, err = CalculateGas(recorder, f)
	require.NoError(t, err)
	require.Len(t, recorder.calls, 3)
}
