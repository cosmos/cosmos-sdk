package mint

var (
	minterKey = 0x00 // the one key to use for the keeper

	// params store
	ParamStoreKeyInflationRateChange = []byte("inflation_rate_change")
	ParamStoreKeyInflationMax        = []byte("inflation_max")
	ParamStoreKeyInflationMin        = []byte("inflation_min")
	ParamStoreKeyGoalBonded          = []byte("goal_bonded")
)

const (
	// default paramspace for params keeper
	DefaultParamspace = "mint"
)
