package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "distribution"

	// StoreKey is the store key string for distribution
	StoreKey = ModuleName

	// RouterKey is the message route for distribution
	RouterKey = ModuleName

	// QuerierRoute is the querier route for distribution
	QuerierRoute = ModuleName
)

// Keys for distribution store
// Items are stored with the following key: values
//
// - 0x00<proposalID_Bytes>: FeePol
//
// - 0x01: sdk.ConsAddress
//
// - 0x02<valAddr_Bytes>: ValidatorOutstandingRewards
//
// - 0x03<accAddr_Bytes>: sdk.AccAddress
//
// - 0x04<valAddr_Bytes><accAddr_Bytes>: DelegatorStartingInfo
//
// - 0x05<valAddr_Bytes><period_Bytes>: ValidatorHistoricalRewards
//
// - 0x06<valAddr_Bytes>: ValidatorCurrentRewards
//
// - 0x07<valAddr_Bytes>: ValidatorCurrentRewards
//
// - 0x08<valAddr_Bytes><height>: ValidatorSlashEvent
var (
	FeePoolKey                        = []byte{0x00} // key for global distribution state
	ProposerKey                       = []byte{0x01} // key for the proposer operator address
	ValidatorOutstandingRewardsPrefix = []byte{0x02} // key for outstanding rewards

	DelegatorWithdrawAddrPrefix          = []byte{0x03} // key for delegator withdraw address
	DelegatorStartingInfoPrefix          = []byte{0x04} // key for delegator starting info
	ValidatorHistoricalRewardsPrefix     = []byte{0x05} // key for historical validators rewards / stake
	ValidatorCurrentRewardsPrefix        = []byte{0x06} // key for current validator rewards
	ValidatorAccumulatedCommissionPrefix = []byte{0x07} // key for accumulated validator commission
	ValidatorSlashEventPrefix            = []byte{0x08} // key for validator slash fraction
)
