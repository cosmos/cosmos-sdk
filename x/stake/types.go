package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// Params defines the high level settings for staking
type StakeParams struct {
	InflationRateChange sdk.Rat `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Rat `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Rat `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Rat `json:"goal_bonded"`           // Goal of percent bonded atoms

	MaxValidators uint16 `json:"max_vals"`   // maximum number of validators
	BondDenom     string `json:"bond_denom"` // bondable coin denomination

	// gas costs for txs

}

func defaultStakeParams() StakeParams {
	return StakeParams{
		InflationRateChange: sdk.NewRat(13, 100),
		InflationMax:        sdk.NewRat(20, 100),
		InflationMin:        sdk.NewRat(7, 100),
		GoalBonded:          sdk.NewRat(67, 100),
		MaxValidators:       100,
		BondDenom:           "fermion",
	}
}

//_________________________________________________________________________

// StateState - dynamic parameters of the current state
type StakeState struct {
	TotalSupply       int64   `json:"total_supply"`        // total supply of all tokens
	BondedShares      sdk.Rat `json:"bonded_shares"`       // sum of all shares distributed for the Bonded Pool
	UnbondedShares    sdk.Rat `json:"unbonded_shares"`     // sum of all shares distributed for the Unbonded Pool
	BondedPool        int64   `json:"bonded_pool"`         // reserve of bonded tokens
	UnbondedPool      int64   `json:"unbonded_pool"`       // reserve of unbonded tokens held with candidates
	InflationLastTime int64   `json:"inflation_last_time"` // block which the last inflation was processed // TODO make time
	Inflation         sdk.Rat `json:"inflation"`           // current annual inflation rate
}

// XXX define globalstate interface?

func initialStakeState() *GlobalState {
	return &StakeState{
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
func (ss *StakeState) bondedRatio() sdk.Rat {
	if ss.TotalSupply > 0 {
		return sdk.NewRat(gs.BondedPool, ss.TotalSupply)
	}
	return sdk.ZeroRat
}

// get the exchange rate of bonded token per issued share
func (ss *StakeState) bondedShareExRate() sdk.Rat {
	if ss.BondedShares.IsZero() {
		return sdk.OneRat
	}
	return gs.BondedShares.Inv().Mul(sdk.NewRat(ss.BondedPool))
}

// get the exchange rate of unbonded tokens held in candidates per issued share
func (ss *StakeState) unbondedShareExRate() sdk.Rat {
	if ss.UnbondedShares.IsZero() {
		return sdk.OneRat
	}
	return gs.UnbondedShares.Inv().Mul(sdk.NewRat(ss.UnbondedPool))
}

// XXX XXX XXX
// expand to include the function of actually transfering the tokens

//XXX CONFIRM that use of the exRate is correct with Zarko Spec!
func (ss *StakeState) addTokensBonded(amount int64) (issuedShares sdk.Rat) {
	issuedShares = gs.bondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	ss.BondedPool += amount
	ss.BondedShares = ss.BondedShares.Add(issuedShares)
	return
}

//XXX CONFIRM that use of the exRate is correct with Zarko Spec!
func (ss *StakeState) removeSharesBonded(shares sdk.Rat) (removedTokens int64) {
	removedTokens = ss.bondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	ss.BondedShares = ss.BondedShares.Sub(shares)
	ss.BondedPool -= removedTokens
	return
}

//XXX CONFIRM that use of the exRate is correct with Zarko Spec!
func (ss *StakeState) addTokensUnbonded(amount int64) (issuedShares sdk.Rat) {
	issuedShares = ss.unbondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	ss.UnbondedShares = ss.UnbondedShares.Add(issuedShares)
	ss.UnbondedPool += amount
	return
}

//XXX CONFIRM that use of the exRate is correct with Zarko Spec!
func (ss *StakeState) removeSharesUnbonded(shares sdk.Rat) (removedTokens int64) {
	removedTokens = ss.unbondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	ss.UnbondedShares = ss.UnbondedShares.Sub(shares)
	ss.UnbondedPool -= removedTokens
	return
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

// XXX update to use Address as the main key NOT the pubkey
type Candidate struct {
	Status      CandidateStatus `json:"status"`       // Bonded status
	PubKey      crypto.PubKey   `json:"pub_key"`      // Pubkey of candidate
	Address     sdk.Address     `json:"owner"`        // Sender of BondTx - UnbondTx returns here
	Assets      sdk.Rat         `json:"assets"`       // total shares of a global hold pools TODO custom type PoolShares
	Liabilities sdk.Rat         `json:"liabilities"`  // total shares issued to a candidate's delegators TODO custom type DelegatorShares
	VotingPower sdk.Rat         `json:"voting_power"` // Voting power if considered a validator
	Description Description     `json:"description"`  // Description terms for the candidate
}

// Description - description fields for a candidate
type Description struct {
	Moniker  string `json:"moniker"`
	Identity string `json:"identity"`
	Website  string `json:"website"`
	Details  string `json:"details"`
}

// NewCandidate - initialize a new candidate
func NewCandidate(address sdk.Address, pubKey crypto.PubKey, description Description) *Candidate {
	return &Candidate{
		Status:      Unbonded,
		PubKey:      pubKey,
		Address:     address,
		Assets:      sdk.ZeroRat,
		Liabilities: sdk.ZeroRat,
		VotingPower: sdk.ZeroRat,
		Description: description,
	}
}

func (c *Candidate) updateDescription(description Description) {
	if description.Moniker != "" {
		candidate.Description.Moniker = description.Moniker
	}
	if description.Identity != "" {
		candidate.Description.Identity = description.Identity
	}
	if description.Website != "" {
		candidate.Description.Website = description.Website
	}
	if description.Details != "" {
		candidate.Description.Details = description.Details
	}
}

// get the exchange rate of global pool shares over delegator shares
func (c *Candidate) delegatorShareExRate() sdk.Rat {
	if c.Liabilities.IsZero() {
		return sdk.OneRat
	}
	return c.Assets.Quo(c.Liabilities)
}

// add tokens to a candidate
func (c *Candidate) addTokens(amount int64, gs *GlobalState) (issuedDelegatorShares sdk.Rat) {

	exRate := c.delegatorShareExRate()

	var receivedGlobalShares sdk.Rat
	if c.Status == Bonded {
		receivedGlobalShares = gs.addTokensBonded(amount)
	} else {
		receivedGlobalShares = gs.addTokensUnbonded(amount)
	}
	c.Assets = c.Assets.Add(receivedGlobalShares)

	issuedDelegatorShares = exRate.Mul(receivedGlobalShares)
	c.Liabilities = c.Liabilities.Add(issuedDelegatorShares)
	return
}

// remove shares from a candidate
func (c *Candidate) removeShares(shares sdk.Rat, gs *GlobalState) (createdCoins int64) {

	globalPoolSharesToRemove := c.delegatorShareExRate().Mul(shares)

	if c.Status == Bonded {
		createdCoins = gs.removeSharesBonded(globalPoolSharesToRemove)
	} else {
		createdCoins = gs.removeSharesUnbonded(globalPoolSharesToRemove)
	}
	c.Assets = c.Assets.Sub(globalPoolSharesToRemove)

	c.Liabilities = c.Liabilities.Sub(shares)
	return
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
type Candidates []*Candidate

//_________________________________________________________________________

// DelegatorBond represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
type DelegatorBond struct {
	Delegator sdk.Address `json:"delegator"`
	Delegatee sdk.Address `json:"delegatee"`
	Shares    sdk.Rat     `json:"shares"`
}
