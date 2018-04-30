package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Pool       Pool            `json:"pool"`
	Params     Params          `json:"params"`
	Candidates []Candidate     `json:"candidates"`
	Bonds      []DelegatorBond `json:"bonds"`
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
}

func (p Params) equal(p2 Params) bool {
	return p.InflationRateChange.Equal(p2.InflationRateChange) &&
		p.InflationMax.Equal(p2.InflationMax) &&
		p.InflationMin.Equal(p2.InflationMin) &&
		p.GoalBonded.Equal(p2.GoalBonded) &&
		p.MaxValidators == p2.MaxValidators &&
		p.BondDenom == p2.BondDenom
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
}

func (p Pool) equal(p2 Pool) bool {
	return p.BondedShares.Equal(p2.BondedShares) &&
		p.UnbondedShares.Equal(p2.UnbondedShares) &&
		p.Inflation.Equal(p2.Inflation) &&
		p.TotalSupply == p2.TotalSupply &&
		p.BondedPool == p2.BondedPool &&
		p.UnbondedPool == p2.UnbondedPool &&
		p.InflationLastTime == p2.InflationLastTime
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
	Status               CandidateStatus `json:"status"`                 // Bonded status
	Address              sdk.Address     `json:"owner"`                  // Sender of BondTx - UnbondTx returns here
	PubKey               crypto.PubKey   `json:"pub_key"`                // Pubkey of candidate
	Assets               sdk.Rat         `json:"assets"`                 // total shares of a global hold pools
	Liabilities          sdk.Rat         `json:"liabilities"`            // total shares issued to a candidate's delegators
	Description          Description     `json:"description"`            // Description terms for the candidate
	ValidatorBondHeight  int64           `json:"validator_bond_height"`  // Earliest height as a bonded validator
	ValidatorBondCounter int16           `json:"validator_bond_counter"` // Block-local tx index of validator change
}

// Candidates - list of Candidates
type Candidates []Candidate

// NewCandidate - initialize a new candidate
func NewCandidate(address sdk.Address, pubKey crypto.PubKey, description Description) Candidate {
	return Candidate{
		Status:               Unbonded,
		Address:              address,
		PubKey:               pubKey,
		Assets:               sdk.ZeroRat(),
		Liabilities:          sdk.ZeroRat(),
		Description:          description,
		ValidatorBondHeight:  int64(0),
		ValidatorBondCounter: int16(0),
	}
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

// abci validator from stake validator type
func (v Validator) abciValidator(cdc *wire.Codec) abci.Validator {
	return abci.Validator{
		PubKey: v.PubKey.Bytes(),
		Power:  v.Power.Evaluate(),
	}
}

// abci validator from stake validator type
// with zero power used for validator updates
func (v Validator) abciValidatorZero(cdc *wire.Codec) abci.Validator {
	return abci.Validator{
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
