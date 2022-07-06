package v046

const (
	// ModuleName is the name of the module
	ModuleName = "staking"
)

var (
	ParametersKey = []byte{0x51} // prefix for parameters for module x/staking

	KeyUnbondingTime     = []byte("UnbondingTime")
	KeyMaxValidators     = []byte("MaxValidators")
	KeyMaxEntries        = []byte("MaxEntries")
	KeyBondDenom         = []byte("BondDenom")
	KeyHistoricalEntries = []byte("HistoricalEntries")
	KeyMinCommissionRate = []byte("MinCommissionRate")
)
