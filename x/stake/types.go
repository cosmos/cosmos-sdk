package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// Params defines the high level settings for staking
type Params struct {
	InflationRateChange sdk.Rat `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Rat `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Rat `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Rat `json:"goal_bonded"`           // Goal of percent bonded atoms

	MaxValidators uint16 `json:"max_validators"` // maximum number of validators
	BondDenom     string `json:"bond_denom"`     // bondable coin denomination
}

// XXX do we want to allow for default params even or do we want to enforce that you
// need to be explicit about defining all params in genesis?
func defaultParams() Params {
	return Params{
		InflationRateChange: sdk.NewRat(13, 100),
		InflationMax:        sdk.NewRat(20, 100),
		InflationMin:        sdk.NewRat(7, 100),
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "fermion",
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
}

// XXX define globalstate interface?

func initialPool() Pool {
	return Pool{
		TotalSupply:       0,
		BondedShares:      sdk.ZeroRat,
		UnbondedShares:    sdk.ZeroRat,
		BondedPool:        0,
		UnbondedPool:      0,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
}

// get the bond ratio of the global state
func (p Pool) bondedRatio() sdk.Rat {
	if p.TotalSupply > 0 {
		return sdk.NewRat(p.BondedPool, p.TotalSupply)
	}
	return sdk.ZeroRat
}

// get the exchange rate of bonded token per issued share
func (p Pool) bondedShareExRate() sdk.Rat {
	if p.BondedShares.IsZero() {
		return sdk.OneRat
	}
	return sdk.NewRat(p.BondedPool).Quo(p.BondedShares)
}

// get the exchange rate of unbonded tokens held in candidates per issued share
func (p Pool) unbondedShareExRate() sdk.Rat {
	if p.UnbondedShares.IsZero() {
		return sdk.OneRat
	}
	return sdk.NewRat(p.UnbondedPool).Quo(p.UnbondedShares)
}

//_______________________________________________________________________________________________________

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
	Status      CandidateStatus `json:"status"`      // Bonded status
	Address     sdk.Address     `json:"owner"`       // Sender of BondTx - UnbondTx returns here
	PubKey      crypto.PubKey   `json:"pub_key"`     // Pubkey of candidate
	Assets      sdk.Rat         `json:"assets"`      // total shares of a global hold pools
	Liabilities sdk.Rat         `json:"liabilities"` // total shares issued to a candidate's delegators
	Description Description     `json:"description"` // Description terms for the candidate
}

// NewCandidate - initialize a new candidate
func NewCandidate(address sdk.Address, pubKey crypto.PubKey, description Description) Candidate {
	return Candidate{
		Status:      Unbonded,
		Address:     address,
		PubKey:      pubKey,
		Assets:      sdk.ZeroRat,
		Liabilities: sdk.ZeroRat,
		Description: description,
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
		return sdk.OneRat
	}
	return c.Assets.Quo(c.Liabilities)
}

// Validator returns a copy of the Candidate as a Validator.
// Should only be called when the Candidate qualifies as a validator.
func (c Candidate) validator() Validator {
	return Validator{
		Address:     c.Address, // XXX !!!
		VotingPower: c.Assets,
	}
}

//XXX updateDescription function
//XXX enforce limit to number of description characters

//______________________________________________________________________

// Validator is one of the top Candidates
type Validator struct {
	Address     sdk.Address `json:"address"`      // Address of validator
	VotingPower sdk.Rat     `json:"voting_power"` // Voting power if considered a validator
}

// ABCIValidator - Get the validator from a bond value
/* TODO
func (v Validator) ABCIValidator() (*abci.Validator, error) {
	pkBytes, err := wire.MarshalBinary(v.PubKey)
	if err != nil {
		return nil, err
	}
	return &abci.Validator{
		PubKey: pkBytes,
		Power:  v.VotingPower.Evaluate(),
	}, nil
}
*/

//_________________________________________________________________________

// Candidates - list of Candidates
type Candidates []Candidate

//_________________________________________________________________________

// DelegatorBond represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
// TODO better way of managing space
type DelegatorBond struct {
	DelegatorAddr sdk.Address `json:"delegatoraddr"`
	CandidateAddr sdk.Address `json:"candidate_addr"`
	Shares        sdk.Rat     `json:"shares"`
}
