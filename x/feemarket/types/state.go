package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

// NewState instantiates a new fee market state instance. This is utilized
// to implement both the base EIP-1559 fee market implementation and the
// AIMD EIP-1559 fee market implementation. Note that on init, you initialize
// both the minimum and current base gas price to the same value.
func NewState(
	windowSize uint64,
	baseGasPrice math.LegacyDec,
	learningRate math.LegacyDec,
) State {
	return State{
		Window:       make([]uint64, windowSize),
		BaseGasPrice: baseGasPrice,
		Index:        0,
		LearningRate: learningRate,
	}
}

// Update updates the block utilization for the current height with the given
// transaction utilization i.e. gas limit.
func (s *State) Update(gas uint64, params Params) error {
	update := s.Window[s.Index] + gas
	if update > params.MaxBlockUtilization {
		return fmt.Errorf("block utilization of %d cannot exceed max block utilization of %d", update, params.MaxBlockUtilization)
	}

	s.Window[s.Index] = update
	return nil
}

// IncrementHeight increments the current height of the state.
func (s *State) IncrementHeight() {
	s.Index = (s.Index + 1) % uint64(len(s.Window))
	s.Window[s.Index] = 0
}

// UpdateBaseGasPrice updates the learning rate and base gas price based on the AIMD
// learning rate adjustment algorithm. The learning rate is updated
// based on the average utilization of the block window. The base gas price is
// update using the new learning rate and the delta adjustment. Please
// see the EIP-1559 specification for more details.
func (s *State) UpdateBaseGasPrice(params Params) (gasPrice math.LegacyDec) {
	// Panic catch in case there is an overflow
	defer func() {
		if rec := recover(); rec != nil {
			s.BaseGasPrice = params.MinBaseGasPrice
			gasPrice = s.BaseGasPrice
		}
	}()

	// Calculate the new base gasPrice with the learning rate adjustment.
	currentBlockSize := math.LegacyNewDecFromInt(math.NewIntFromUint64(s.Window[s.Index]))
	targetBlockSize := math.LegacyNewDecFromInt(math.NewIntFromUint64(params.TargetBlockUtilization()))
	utilization := (currentBlockSize.Sub(targetBlockSize)).Quo(targetBlockSize)

	// Truncate the learning rate adjustment to an integer.
	//
	// This is equivalent to
	// 1 + (learningRate * (currentBlockSize - targetBlockSize) / targetBlockSize)
	learningRateAdjustment := math.LegacyOneDec().Add(s.LearningRate.Mul(utilization))

	// Calculate the delta adjustment.
	net := math.LegacyNewDecFromInt(s.GetNetUtilization(params)).Mul(params.Delta)

	// Update the base gasPrice.
	gasPrice = s.BaseGasPrice.Mul(learningRateAdjustment).Add(net)

	// Ensure the base gasPrice is greater than the minimum base gasPrice.
	if gasPrice.LT(params.MinBaseGasPrice) {
		gasPrice = params.MinBaseGasPrice
	}

	s.BaseGasPrice = gasPrice
	return s.BaseGasPrice
}

// UpdateLearningRate updates the learning rate based on the AIMD
// learning rate adjustment algorithm. The learning rate is updated
// based on the average utilization of the block window. There are
// two cases that can occur:
//
//  1. The average utilization is above the target threshold. In this
//     case, the learning rate is increased by the alpha parameter. This occurs
//     when blocks are nearly full or empty.
//  2. The average utilization is below the target threshold. In this
//     case, the learning rate is decreased by the beta parameter. This occurs
//     when blocks are relatively close to the target block utilization.
//
// For more details, please see the EIP-1559 specification.
func (s *State) UpdateLearningRate(params Params) (lr math.LegacyDec) {
	// Panic catch in case there is an overflow
	defer func() {
		if rec := recover(); rec != nil {
			s.LearningRate = params.MinLearningRate
			lr = s.LearningRate
		}
	}()

	// Calculate the average utilization of the block window.
	avg := s.GetAverageUtilization(params)

	// Determine if the average utilization is above or below the target
	// threshold and adjust the learning rate accordingly.
	if avg.LTE(params.Gamma) || avg.GTE(math.LegacyOneDec().Sub(params.Gamma)) {
		lr = params.Alpha.Add(s.LearningRate)
		if lr.GT(params.MaxLearningRate) {
			lr = params.MaxLearningRate
		}
	} else {
		lr = s.LearningRate.Mul(params.Beta)
		if lr.LT(params.MinLearningRate) {
			lr = params.MinLearningRate
		}
	}

	// Update the current learning rate.
	s.LearningRate = lr
	return s.LearningRate
}

// GetNetUtilization returns the net utilization of the block window.
func (s *State) GetNetUtilization(params Params) math.Int {
	net := math.NewInt(0)

	targetUtilization := math.NewIntFromUint64(params.TargetBlockUtilization())
	for _, utilization := range s.Window {
		diff := math.NewIntFromUint64(utilization).Sub(targetUtilization)
		net = net.Add(diff)
	}

	return net
}

// GetAverageUtilization returns the average utilization of the block
// window.
func (s *State) GetAverageUtilization(params Params) math.LegacyDec {
	var total uint64
	for _, utilization := range s.Window {
		total += utilization
	}

	sum := math.LegacyNewDecFromInt(math.NewIntFromUint64(total))

	multiple := math.LegacyNewDecFromInt(math.NewIntFromUint64(uint64(len(s.Window))))
	divisor := math.LegacyNewDecFromInt(math.NewIntFromUint64(params.MaxBlockUtilization)).Mul(multiple)

	return sum.Quo(divisor)
}

// ValidateBasic performs basic validation on the state.
func (s *State) ValidateBasic() error {
	if s.Window == nil {
		return fmt.Errorf("block utilization window cannot be nil or empty")
	}

	if s.BaseGasPrice.IsNil() || s.BaseGasPrice.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("base gas price must be positive")
	}

	if s.LearningRate.IsNil() || s.LearningRate.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("learning rate must be positive")
	}

	return nil
}
