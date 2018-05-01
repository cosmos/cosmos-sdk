package stake

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/wire"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Pool       Pool            `json:"pool"`
	Params     Params          `json:"params"`
	Candidates []Candidate     `json:"candidates"`
	Bonds      []DelegatorBond `json:"bonds"`
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Pool:   initialPool(),
		Params: defaultParams(),
	}
}

//_________________________________________________________________________

// Params defines the high level settings for staking
type Params struct {
	InflationRateChange sdk.Rat `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Rat `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Rat `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Rat `json:"goal_bonded"`           // Goal of percent bonded atoms

	MaxValidators uint16 `json:"max_validators"` // maximum number of validators
	BondDenom     string `json:"bond_denom"`     // bondable coin denomination

	ReservePoolFee sdk.Rat `json:"reserve_pool_fee"` // percent of fees which go to reserve pool
}

func (p Params) equal(p2 Params) bool {
	return p.InflationRateChange.Equal(p2.InflationRateChange) &&
		p.InflationMax.Equal(p2.InflationMax) &&
		p.InflationMin.Equal(p2.InflationMin) &&
		p.GoalBonded.Equal(p2.GoalBonded) &&
		p.MaxValidators == p2.MaxValidators &&
		p.BondDenom == p2.BondDenom &&
		p.ReservePoolFee.Equal(p2.ReservePoolFee)
}

func defaultParams() Params {
	return Params{
		InflationRateChange: sdk.NewRat(13, 100),
		InflationMax:        sdk.NewRat(20, 100),
		InflationMin:        sdk.NewRat(7, 100),
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "steak",
		ReservePoolFee:      sdk.NewRat(5, 100),
	}
}

//_________________________________________________________________________

// Pool - dynamic parameters of the current state
type Pool struct {
	TotalSupply       int64   `json:"total_supply"`        // total supply of all tokens
	BondedShares      sdk.Rat `json:"bonded_shares"`       // sum of all shares distributed for the Bonded Pool
	UnbondedShares    sdk.Rat `json:"unbonded_shares"`     // sum of all shares distributed for the Unbonded Pool
	BondedPool        int64   `json:"bonded_pool"`         // reserve of bonded tokens
	UnbondedPool      int64   `json:"unbonded_pool"`       // reserve of unbonded tokens held with candidates
	InflationLastTime int64   `json:"inflation_last_time"` // block which the last inflation was processed // TODO make time
	Inflation         sdk.Rat `json:"inflation"`           // current annual inflation rate

	DateLastCommissionReset int64 `json:"date_last_commission_reset"` // unix timestamp for last commission accounting reset (daily)

	// XXX need to use special sdk.Rat amounts in coins here because added at small increments
	ReservePool     sdk.Coins `json:"reserve_pool"`      // XXX reserve pool of collected fees for use by governance
	FeePool         sdk.Coins `json:"fee_pool"`          // XXX fee pool for all the fee shares which have already been distributed
	SumFeesReceived sdk.Coins `json:"sum_fees_received"` // XXX sum of all fees received
	RecentFee       sdk.Coins `json:"recent_fee"`        // XXX most recent fee collected
	Adjustment      sdk.Rat   `json:"adjustment"`        // XXX Adjustment factor for calculating global fee accum
}

func (p Pool) equal(p2 Pool) bool {
	return p.TotalSupply == p2.TotalSupply &&
		p.BondedShares.Equal(p2.BondedShares) &&
		p.UnbondedShares.Equal(p2.UnbondedShares) &&
		p.BondedPool == p2.BondedPool &&
		p.UnbondedPool == p2.UnbondedPool &&
		p.InflationLastTime == p2.InflationLastTime &&
		p.Inflation.Equal(p2.Inflation) &&
		p.DateLastCommissionReset == p2.DateLastCommissionReset &&
		p.ReservePool.IsEqual(p2.ReservePool) &&
		p.FeePool.IsEqual(p2.FeePool) &&
		p.SumFeesReceived.IsEqual(p2.SumFeesReceived) &&
		p.RecentFee.IsEqual(p2.RecentFee) &&
		p.Adjustment.Equal(p2.Adjustment)
}

// initial pool for testing
func initialPool() Pool {
	return Pool{
		TotalSupply:             0,
		BondedShares:            sdk.ZeroRat(),
		UnbondedShares:          sdk.ZeroRat(),
		BondedPool:              0,
		UnbondedPool:            0,
		InflationLastTime:       0,
		Inflation:               sdk.NewRat(7, 100),
		DateLastCommissionReset: 0,
		ReservePool:             sdk.Coins{},
		FeePool:                 sdk.Coins{},
		SumFeesReceived:         sdk.Coins{},
		RecentFee:               sdk.Coins{},
		Adjustment:              sdk.ZeroRat(),
	}
}

//_________________________________________________________________________

// CandidateStatus - status of a validator-candidate
type CandidateStatus byte

const (
	// nolint
	Bonded   CandidateStatus = 0x00
	Unbonded CandidateStatus = 0x01
	Revoked  CandidateStatus = 0x02
)

// Candidate defines the total amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// candidate, the candidate is credited with a DelegatorBond whose number of
// bond shares is based on the amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
type Candidate struct {
	Status                CandidateStatus `json:"status"`                  // Bonded status
	Address               sdk.Address     `json:"owner"`                   // Sender of BondTx - UnbondTx returns here
	PubKey                crypto.PubKey   `json:"pub_key"`                 // Pubkey of candidate
	Assets                sdk.Rat         `json:"assets"`                  // total shares of a global hold pools
	Liabilities           sdk.Rat         `json:"liabilities"`             // total shares issued to a candidate's delegators
	Description           Description     `json:"description"`             // Description terms for the candidate
	ValidatorBondHeight   int64           `json:"validator_bond_height"`   // Earliest height as a bonded validator
	ValidatorBondCounter  int16           `json:"validator_bond_counter"`  // Block-local tx index of validator change
	ProposerRewardPool    sdk.Coins       `json:"proposer_reward_pool"`    // XXX reward pool collected from being the proposer
	Commission            sdk.Rat         `json:"commission"`              // XXX the commission rate of fees charged to any delegators
	CommissionMax         sdk.Rat         `json:"commission_max"`          // XXX maximum commission rate which this candidate can ever charge
	CommissionChangeRate  sdk.Rat         `json:"commission_change_rate"`  // XXX maximum daily increase of the candidate commission
	CommissionChangeToday sdk.Rat         `json:"commission_change_today"` // XXX commission rate change today, reset each day (UTC time)
}

// Candidates - list of Candidates
type Candidates []Candidate

// NewCandidate - initialize a new candidate
func NewCandidate(address sdk.Address, pubKey crypto.PubKey, description Description) Candidate {
	return Candidate{
		Status:                Unbonded,
		Address:               address,
		PubKey:                pubKey,
		Assets:                sdk.ZeroRat(),
		Liabilities:           sdk.ZeroRat(),
		Description:           description,
		ValidatorBondHeight:   int64(0),
		ValidatorBondCounter:  int16(0),
		ProposerRewardPool:    sdk.Coins{},
		Commission:            sdk.ZeroRat(),
		CommissionMax:         sdk.ZeroRat(),
		CommissionChangeRate:  sdk.ZeroRat(),
		CommissionChangeToday: sdk.ZeroRat(),
	}
}

func (c Candidate) equal(c2 Candidate) bool {
	return c.Status == c2.Status &&
		c.PubKey.Equals(c2.PubKey) &&
		bytes.Equal(c.Address, c2.Address) &&
		c.Assets.Equal(c2.Assets) &&
		c.Liabilities.Equal(c2.Liabilities) &&
		c.Description == c2.Description &&
		c.ValidatorBondHeight == c2.ValidatorBondHeight &&
		c.ValidatorBondCounter == c2.ValidatorBondCounter &&
		c.ProposerRewardPool.IsEqual(c2.ProposerRewardPool) &&
		c.Commission.Equal(c2.Commission) &&
		c.CommissionMax.Equal(c2.CommissionMax) &&
		c.CommissionChangeRate.Equal(c2.CommissionChangeRate) &&
		c.CommissionChangeToday.Equal(c2.CommissionChangeToday)
}

// Description - description fields for a candidate
type Description struct {
	Moniker  string `json:"moniker"`
	Identity string `json:"identity"`
	Website  string `json:"website"`
	Details  string `json:"details"`
}

func NewDescription(moniker, identity, website, details string) Description {
	return Description{
		Moniker:  moniker,
		Identity: identity,
		Website:  website,
		Details:  details,
	}
}

// get the exchange rate of global pool shares over delegator shares
func (c Candidate) delegatorShareExRate() sdk.Rat {
	if c.Liabilities.IsZero() {
		return sdk.OneRat()
	}
	return c.Assets.Quo(c.Liabilities)
}

// Validator returns a copy of the Candidate as a Validator.
// Should only be called when the Candidate qualifies as a validator.
func (c Candidate) validator() Validator {
	return Validator{
		Address: c.Address,
		PubKey:  c.PubKey,
		Power:   c.Assets,
		Height:  c.ValidatorBondHeight,
		Counter: c.ValidatorBondCounter,
	}
}

//XXX updateDescription function
//XXX enforce limit to number of description characters

//______________________________________________________________________

// Validator is one of the top Candidates
type Validator struct {
	Address sdk.Address   `json:"address"`
	PubKey  crypto.PubKey `json:"pub_key"`
	Power   sdk.Rat       `json:"voting_power"`
	Height  int64         `json:"height"`  // Earliest height as a validator
	Counter int16         `json:"counter"` // Block-local tx index for resolving equal voting power & height
}

func (v Validator) equal(v2 Validator) bool {
	return bytes.Equal(v.Address, v2.Address) &&
		v.PubKey.Equals(v2.PubKey) &&
		v.Power.Equal(v2.Power) &&
		v.Height == v2.Height &&
		v.Counter == v2.Counter
}

// abci validator from stake validator type
func (v Validator) abciValidator(cdc *wire.Codec) sdk.Validator {
	return sdk.Validator{
		PubKey: v.PubKey.Bytes(),
		Power:  v.Power.Evaluate(),
	}
}

// abci validator from stake validator type
// with zero power used for validator updates
func (v Validator) abciValidatorZero(cdc *wire.Codec) sdk.Validator {
	return sdk.Validator{
		PubKey: v.PubKey.Bytes(),
		Power:  0,
	}
}

//_________________________________________________________________________

// DelegatorBond represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
// TODO better way of managing space
type DelegatorBond struct {
	DelegatorAddr sdk.Address `json:"delegator_addr"`
	CandidateAddr sdk.Address `json:"candidate_addr"`
	Shares        sdk.Rat     `json:"shares"`
	Height        int64       `json:"height"` // Last height bond updated
}

func (b DelegatorBond) equal(b2 DelegatorBond) bool {
	return bytes.Equal(b.DelegatorAddr, b2.DelegatorAddr) &&
		bytes.Equal(b.CandidateAddr, b2.CandidateAddr) &&
		b.Height == b2.Height &&
		b.Shares.Equal(b2.Shares)
}
