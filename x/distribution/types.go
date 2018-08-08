package stake

//// GenesisState - all staking state that must be provided at genesis
//type GenesisState struct {
//Pool   Pool   `json:"pool"`
//Params Params `json:"params"`
//}

//func NewGenesisState(pool Pool, params Params, candidates []Candidate, bonds []Delegation) GenesisState {
//return GenesisState{
//Pool:   pool,
//Params: params,
//}
//}

//// get raw genesis raw message for testing
//func DefaultGenesisState() GenesisState {
//return GenesisState{
//Pool:   initialPool(),
//Params: defaultParams(),
//}
//}

//// fee information for a validator
//type Validator struct {
//Adjustments      []sdk.Rat `json:"fee_adjustments"`    // XXX Adjustment factors for lazy fee accounting, couples with Params.BondDenoms
//PrevBondedShares sdk.Rat   `json:"prev_bonded_shares"` // total shares of a global hold pools
//}

////_________________________________________________________________________

//// Params defines the high level settings for staking
//type Params struct {
//FeeDenoms      []string `json:"fee_denoms"`       // accepted fee denoms
//ReservePoolFee sdk.Rat  `json:"reserve_pool_fee"` // percent of fees which go to reserve pool
//}

//func (p Params) equal(p2 Params) bool {
//return p.BondDenom == p2.BondDenom &&
//p.ReservePoolFee.Equal(p2.ReservePoolFee)
//}

//func defaultParams() Params {
//return Params{
//FeeDenoms:      []string{"steak"},
//ReservePoolFee: sdk.NewRat(5, 100),
//}
//}

////_________________________________________________________________________

//// Pool - dynamic parameters of the current state
//type Pool struct {
//FeeReservePool   sdk.Coins `json:"fee_reserve_pool"`   // XXX reserve pool of collected fees for use by governance
//FeePool          sdk.Coins `json:"fee_pool"`           // XXX fee pool for all the fee shares which have already been distributed
//FeeSumReceived   sdk.Coins `json:"fee_sum_received"`   // XXX sum of all fees received, post reserve pool `json:"fee_sum_received"`
//FeeRecent        sdk.Coins `json:"fee_recent"`         // XXX most recent fee collected
//FeeAdjustments   []sdk.Rat `json:"fee_adjustments"`    // XXX Adjustment factors for lazy fee accounting, couples with Params.BondDenoms
//PrevBondedShares sdk.Rat   `json:"prev_bonded_shares"` // XXX last recorded bonded shares
//}

//func (p Pool) equal(p2 Pool) bool {
//return p.FeeReservePool.IsEqual(p2.FeeReservePool) &&
//p.FeePool.IsEqual(p2.FeePool) &&
//p.FeeSumReceived.IsEqual(p2.FeeSumReceived) &&
//p.FeeRecent.IsEqual(p2.FeeRecent) &&
//sdk.RatsEqual(p.FeeAdjustments, p2.FeeAdjustments) &&
//p.PrevBondedShares.Equal(p2.PrevBondedShares)
//}

//// initial pool for testing
//func initialPool() Pool {
//return Pool{
//FeeReservePool:   sdk.Coins(nil),
//FeePool:          sdk.Coins(nil),
//FeeSumReceived:   sdk.Coins(nil),
//FeeRecent:        sdk.Coins(nil),
//FeeAdjustments:   []sdk.Rat{sdk.ZeroRat()},
//PrevBondedShares: sdk.ZeroRat(),
//}
//}

////_________________________________________________________________________

//// Used in calculation of fee shares, added to a queue for each block where a power change occures
//type PowerChange struct {
//Height      int64     `json:"height"`        // block height at change
//Power       sdk.Rat   `json:"power"`         // total power at change
//PrevPower   sdk.Rat   `json:"prev_power"`    // total power at previous height-1
//FeesIn      sdk.Coins `json:"fees_in"`       // fees in at block height
//PrevFeePool sdk.Coins `json:"prev_fee_pool"` // total fees in at previous block height
//}

////_________________________________________________________________________
//// KEY MANAGEMENT

//var (
//// Keys for store prefixes
//PowerChangeKey = []byte{0x09} // prefix for power change object
//)

//// get the key for the accumulated update validators
//func GetPowerChangeKey(height int64) []byte {
//heightBytes := make([]byte, binary.MaxVarintLen64)
//binary.BigEndian.PutUint64(heightBytes, ^uint64(height)) // invert height (older validators first)
//return append(PowerChangeKey, heightBytes...)
//}
