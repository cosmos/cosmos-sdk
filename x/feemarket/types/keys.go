package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the name of the feemarket module.
	ModuleName = "feemarket"
	// StoreKey is the store key string for the feemarket module.
	StoreKey = ModuleName

	// FeeCollectorName is the root string for the fee market fee collector account address.
	FeeCollectorName = "feemarket-fee-collector"
)

var (
	// KeyParams is the store key for the feemarket module's parameters.
	KeyParams = collections.NewPrefix(1)

	// KeyState is the store key for the feemarket module's data.
	KeyState = collections.NewPrefix(2)

	// KeyEnabledHeight is the store key for the feemarket module's enabled height.
	KeyEnabledHeight = collections.NewPrefix(3)
	KeyBasePrice     = collections.NewPrefix(4)
	KeyLearning      = collections.NewPrefix(5)
	KeyIndex         = collections.NewPrefix(6)

	EventTypeFeePay      = "fee_pay"
	EventTypeTipPay      = "tip_pay"
	AttributeKeyTip      = "tip"
	AttributeKeyTipPayer = "tip_payer"
	AttributeKeyTipPayee = "tip_payee"
)
