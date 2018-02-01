package stake

import (
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/rational"
)

// Params defines the high level settings for staking
type Params struct {
	HoldBonded   crypto.Address `json:"hold_bonded"`   // account  where all bonded coins are held
	HoldUnbonded crypto.Address `json:"hold_unbonded"` // account where all delegated but unbonded coins are held

	InflationRateChange rational.Rational `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        rational.Rational `json:"inflation_max"`         // maximum inflation rate
	InflationMin        rational.Rational `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          rational.Rational `json:"goal_bonded"`           // Goal of percent bonded atoms

	MaxVals          uint16 `json:"max_vals"`           // maximum number of validators
	AllowedBondDenom string `json:"allowed_bond_denom"` // bondable coin denomination

	// gas costs for txs
	GasDeclareCandidacy int64 `json:"gas_declare_candidacy"`
	GasEditCandidacy    int64 `json:"gas_edit_candidacy"`
	GasDelegate         int64 `json:"gas_delegate"`
	GasUnbond           int64 `json:"gas_unbond"`
}

func defaultParams() Params {
	return Params{
		HoldBonded:          []byte("77777777777777777777777777777777"),
		HoldUnbonded:        []byte("88888888888888888888888888888888"),
		InflationRateChange: rational.New(13, 100),
		InflationMax:        rational.New(20, 100),
		InflationMin:        rational.New(7, 100),
		GoalBonded:          rational.New(67, 100),
		MaxVals:             100,
		AllowedBondDenom:    "fermion",
		GasDeclareCandidacy: 20,
		GasEditCandidacy:    20,
		GasDelegate:         20,
		GasUnbond:           20,
	}
}

//_________________________________________________________________________

// GlobalState - dynamic parameters of the current state
type GlobalState struct {
	TotalSupply       int64             `json:"total_supply"`        // total supply of all tokens
	BondedShares      rational.Rational `json:"bonded_shares"`       // sum of all shares distributed for the Bonded Pool
	UnbondedShares    rational.Rational `json:"unbonded_shares"`     // sum of all shares distributed for the Unbonded Pool
	BondedPool        int64             `json:"bonded_pool"`         // reserve of bonded tokens
	UnbondedPool      int64             `json:"unbonded_pool"`       // reserve of unbonded tokens held with candidates
	InflationLastTime int64             `json:"inflation_last_time"` // block which the last inflation was processed // TODO make time
	Inflation         rational.Rational `json:"inflation"`           // current annual inflation rate
}

// XXX define globalstate interface?

func initialGlobalState() *GlobalState {
	return &GlobalState{
		TotalSupply:       0,
		BondedShares:      rational.Zero,
		UnbondedShares:    rational.Zero,
		BondedPool:        0,
		UnbondedPool:      0,
		InflationLastTime: 0,
		Inflation:         rational.New(7, 100),
	}
}

// get the bond ratio of the global state
func (gs *GlobalState) bondedRatio() rational.Rational {
	if gs.TotalSupply > 0 {
		return rational.New(gs.BondedPool, gs.TotalSupply)
	}
	return rational.Zero
}

// get the exchange rate of bonded token per issued share
func (gs *GlobalState) bondedShareExRate() rational.Rational {
	if gs.BondedShares.IsZero() {
		return rational.One
	}
	return gs.BondedShares.Inv().Mul(rational.New(gs.BondedPool))
}

// get the exchange rate of unbonded tokens held in candidates per issued share
func (gs *GlobalState) unbondedShareExRate() rational.Rational {
	if gs.UnbondedShares.IsZero() {
		return rational.One
	}
	return gs.UnbondedShares.Inv().Mul(rational.New(gs.UnbondedPool))
}

func (gs *GlobalState) addTokensBonded(amount int64) (issuedShares rational.Rational) {
	issuedShares = gs.bondedShareExRate().Inv().Mul(rational.New(amount)) // (tokens/shares)^-1 * tokens
	gs.BondedPool += amount
	gs.BondedShares = gs.BondedShares.Add(issuedShares)
	return
}

func (gs *GlobalState) removeSharesBonded(shares rational.Rational) (removedTokens int64) {
	removedTokens = gs.bondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	gs.BondedShares = gs.BondedShares.Sub(shares)
	gs.BondedPool -= removedTokens
	return
}

func (gs *GlobalState) addTokensUnbonded(amount int64) (issuedShares rational.Rational) {
	issuedShares = gs.unbondedShareExRate().Inv().Mul(rational.New(amount)) // (tokens/shares)^-1 * tokens
	gs.UnbondedShares = gs.UnbondedShares.Add(issuedShares)
	gs.UnbondedPool += amount
	return
}

func (gs *GlobalState) removeSharesUnbonded(shares rational.Rational) (removedTokens int64) {
	removedTokens = gs.unbondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	gs.UnbondedShares = gs.UnbondedShares.Sub(shares)
	gs.UnbondedPool -= removedTokens
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
type Candidate struct {
	Status      CandidateStatus   `json:"status"`       // Bonded status
	PubKey      crypto.PubKey     `json:"pub_key"`      // Pubkey of candidate
	Owner       crypto.Address    `json:"owner"`        // Sender of BondTx - UnbondTx returns here
	Assets      rational.Rational `json:"assets"`       // total shares of a global hold pools TODO custom type PoolShares
	Liabilities rational.Rational `json:"liabilities"`  // total shares issued to a candidate's delegators TODO custom type DelegatorShares
	VotingPower rational.Rational `json:"voting_power"` // Voting power if considered a validator
	Description Description       `json:"description"`  // Description terms for the candidate
}

// Description - description fields for a candidate
type Description struct {
	Moniker  string `json:"moniker"`
	Identity string `json:"identity"`
	Website  string `json:"website"`
	Details  string `json:"details"`
}

// NewCandidate - initialize a new candidate
func NewCandidate(pubKey crypto.PubKey, owner crypto.Address, description Description) *Candidate {
	return &Candidate{
		Status:      Unbonded,
		PubKey:      pubKey,
		Owner:       owner,
		Assets:      rational.Zero,
		Liabilities: rational.Zero,
		VotingPower: rational.Zero,
		Description: description,
	}
}

// XXX define candidate interface?

// get the exchange rate of global pool shares over delegator shares
func (c *Candidate) delegatorShareExRate() rational.Rational {
	if c.Liabilities.IsZero() {
		return rational.One
	}
	return c.Assets.Quo(c.Liabilities)
}

// add tokens to a candidate
func (c *Candidate) addTokens(amount int64, gs *GlobalState) (issuedDelegatorShares rational.Rational) {

	exRate := c.delegatorShareExRate()

	var receivedGlobalShares rational.Rational
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
func (c *Candidate) removeShares(shares rational.Rational, gs *GlobalState) (removedTokens int64) {

	globalPoolSharesToRemove := c.delegatorShareExRate().Mul(shares)

	if c.Status == Bonded {
		removedTokens = gs.removeSharesBonded(globalPoolSharesToRemove)
	} else {
		removedTokens = gs.removeSharesUnbonded(globalPoolSharesToRemove)
	}
	c.Assets = c.Assets.Sub(globalPoolSharesToRemove)

	c.Liabilities = c.Liabilities.Sub(shares)
	return
}

// Validator returns a copy of the Candidate as a Validator.
// Should only be called when the Candidate qualifies as a validator.
func (c *Candidate) validator() Validator {
	return Validator{
		PubKey:      c.PubKey,
		VotingPower: c.VotingPower,
	}
}

// Validator is one of the top Candidates
type Validator struct {
	PubKey      crypto.PubKey     `json:"pub_key"`      // Pubkey of candidate
	VotingPower rational.Rational `json:"voting_power"` // Voting power if considered a validator
}

// ABCIValidator - Get the validator from a bond value
func (v Validator) ABCIValidator() (*abci.Validator, error) {
	pkBytes, err := cdc.MarshalBinary(v.PubKey)
	if err != nil {
		return nil, err
	}
	return &abci.Validator{
		PubKey: pkBytes,
		Power:  v.VotingPower.Evaluate(),
	}, nil
}

//_________________________________________________________________________

// Candidates - list of Candidates
type Candidates []*Candidate

//_________________________________________________________________________

// DelegatorBond represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
type DelegatorBond struct {
	PubKey crypto.PubKey     `json:"pub_key"`
	Shares rational.Rational `json:"shares"`
}
