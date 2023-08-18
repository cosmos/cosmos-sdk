package math

// Dec is an alias of the LegacyDec.
// N.B. Original Dec's API was broken to LegacyDec.
// However, the new Dec has not yet been introduced. As a result, to avoid breaking
// the API for clients upgrading from sdk v45, we keep this alias until the
// new Dec is implemented. This will minimize redundant API breakage.
type Dec = LegacyDec

var (
	NewDec                   = LegacyNewDec
	NewDecWithPrec           = LegacyNewDecWithPrec
	NewDecFromBigInt         = LegacyNewDecFromBigInt
	NewDecFromBigIntWithPrec = LegacyNewDecFromBigIntWithPrec
	NewDecFromInt            = LegacyNewDecFromInt
	NewDecFromStr            = LegacyNewDecFromStr

	MustNewDecFromStr = LegacyMustNewDecFromStr

	ZeroDec     = LegacyZeroDec
	OneDec      = LegacyOneDec
	SmallestDec = LegacySmallestDec

	DecsEqual   = LegacyDecsEqual
	MinDec      = LegacyMinDec
	MaxDec      = LegacyMaxDec
	DecEq       = LegacyDecEq
	DecApproxEq = LegacyDecApproxEq
)
