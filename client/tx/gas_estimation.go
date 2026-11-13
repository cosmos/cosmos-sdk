package tx

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strconv"

	gogogrpc "github.com/cosmos/gogoproto/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

// adjustedGas rounds upward using the caller's decimal adjustment without
// converting gas to float64, which loses precision for large uint64 values.
func adjustedGas(used uint64, adjustment float64) (uint64, error) {
	if math.IsNaN(adjustment) || math.IsInf(adjustment, 0) || adjustment < 0 {
		return 0, fmt.Errorf("gas adjustment must be finite and nonnegative; got %g", adjustment)
	}
	factor, ok := new(big.Rat).SetString(strconv.FormatFloat(adjustment, 'f', -1, 64))
	if !ok {
		return 0, fmt.Errorf("invalid gas adjustment %g", adjustment)
	}
	product := factor.Mul(factor, new(big.Rat).SetInt(new(big.Int).SetUint64(used)))
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(product.Num(), product.Denom(), remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if !quotient.IsUint64() {
		return 0, fmt.Errorf("adjusted gas estimate overflows uint64")
	}
	return quotient.Uint64(), nil
}

func calculateGas(conn gogogrpc.ClientConn, f Factory, msgs ...sdk.Msg) (*txtypes.SimulateResponse, uint64, error) {
	if _, err := adjustedGas(0, f.GasAdjustment()); err != nil {
		return nil, 0, err
	}
	service := txtypes.NewServiceClient(conn)
	// Bootstrap without a gas limit, then measure fee payment at a provisional
	// limit. Verify the resulting candidate with its own gas limit and fees.
	var candidate uint64
	for call := 0; call < 4; call++ {
		bz, err := f.WithGas(candidate).BuildSimTx(msgs...)
		if err != nil {
			return nil, 0, err
		}
		res, err := service.Simulate(context.Background(), &txtypes.SimulateRequest{TxBytes: bz})
		if err != nil {
			return nil, 0, fmt.Errorf("gas simulation with limit %d failed: %w", candidate, err)
		}
		if res == nil || res.GasInfo == nil {
			return nil, 0, fmt.Errorf("gas simulation returned no gas information")
		}
		required, err := adjustedGas(res.GasInfo.GasUsed, f.GasAdjustment())
		if err != nil {
			return nil, 0, err
		}
		if call >= 2 && required <= candidate {
			return res, candidate, nil
		}
		if required > candidate {
			candidate = required
		}
	}
	return nil, 0, fmt.Errorf("automatic gas estimate did not stabilize after 4 simulations; retry with an explicit --gas limit or a larger --gas-adjustment")
}
