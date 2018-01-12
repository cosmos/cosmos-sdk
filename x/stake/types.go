package stake

import (
	"bytes"
	"sort"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/rational"
)

// Params defines the high level settings for staking
type Params struct {
	HoldBonded   sdk.Actor `json:"hold_bonded"`   // account  where all bonded coins are held
	HoldUnbonded sdk.Actor `json:"hold_unbonded"` // account where all delegated but unbonded coins are held

	InflationRateChange rational.Rat `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        rational.Rat `json:"inflation_max"`         // maximum inflation rate
	InflationMin        rational.Rat `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          rational.Rat `json:"goal_bonded"`           // Goal of percent bonded atoms

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
		HoldBonded:          sdk.NewActor(stakingModuleName, []byte("77777777777777777777777777777777")),
		HoldUnbonded:        sdk.NewActor(stakingModuleName, []byte("88888888888888888888888888888888")),
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
	TotalSupply       int64        `json:"total_supply"`        // total supply of all tokens
	BondedShares      rational.Rat `json:"bonded_shares"`       // sum of all shares distributed for the Bonded Pool
	UnbondedShares    rational.Rat `json:"unbonded_shares"`     // sum of all shares distributed for the Unbonded Pool
	BondedPool        int64        `json:"bonded_pool"`         // reserve of bonded tokens
	UnbondedPool      int64        `json:"unbonded_pool"`       // reserve of unbonded tokens held with candidates
	InflationLastTime int64        `json:"inflation_last_time"` // block which the last inflation was processed // TODO make time
	Inflation         rational.Rat `json:"inflation"`           // current annual inflation rate
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
func (gs *GlobalState) bondedRatio() rational.Rat {
	if gs.TotalSupply > 0 {
		return rational.New(gs.BondedPool, gs.TotalSupply)
	}
	return rational.Zero
}

// get the exchange rate of bonded token per issued share
func (gs *GlobalState) bondedShareExRate() rational.Rat {
	if gs.BondedShares.IsZero() {
		return rational.One
	}
	return gs.BondedShares.Inv().Mul(rational.New(gs.BondedPool))
}

// get the exchange rate of unbonded tokens held in candidates per issued share
func (gs *GlobalState) unbondedShareExRate() rational.Rat {
	if gs.UnbondedShares.IsZero() {
		return rational.One
	}
	return gs.UnbondedShares.Inv().Mul(rational.New(gs.UnbondedPool))
}

func (gs *GlobalState) addTokensBonded(amount int64) (issuedShares rational.Rat) {
	issuedShares = gs.bondedShareExRate().Inv().Mul(rational.New(amount)) // (tokens/shares)^-1 * tokens
	gs.BondedPool += amount
	gs.BondedShares = gs.BondedShares.Add(issuedShares)
	return
}

func (gs *GlobalState) removeSharesBonded(shares rational.Rat) (removedTokens int64) {
	removedTokens = gs.bondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	gs.BondedShares = gs.BondedShares.Sub(shares)
	gs.BondedPool -= removedTokens
	return
}

func (gs *GlobalState) addTokensUnbonded(amount int64) (issuedShares rational.Rat) {
	issuedShares = gs.unbondedShareExRate().Inv().Mul(rational.New(amount)) // (tokens/shares)^-1 * tokens
	gs.UnbondedShares = gs.UnbondedShares.Add(issuedShares)
	gs.UnbondedPool += amount
	return
}

func (gs *GlobalState) removeSharesUnbonded(shares rational.Rat) (removedTokens int64) {
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
	Status      CandidateStatus `json:"status"`       // Bonded status
	PubKey      crypto.PubKey   `json:"pub_key"`      // Pubkey of candidate
	Owner       sdk.Actor       `json:"owner"`        // Sender of BondTx - UnbondTx returns here
	Assets      rational.Rat    `json:"assets"`       // total shares of a global hold pools TODO custom type PoolShares
	Liabilities rational.Rat    `json:"liabilities"`  // total shares issued to a candidate's delegators TODO custom type DelegatorShares
	VotingPower rational.Rat    `json:"voting_power"` // Voting power if considered a validator
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
func NewCandidate(pubKey crypto.PubKey, owner sdk.Actor, description Description) *Candidate {
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
func (c *Candidate) delegatorShareExRate() rational.Rat {
	if c.Liabilities.IsZero() {
		return rational.One
	}
	return c.Assets.Quo(c.Liabilities)
}

// add tokens to a candidate
func (c *Candidate) addTokens(amount int64, gs *GlobalState) (issuedDelegatorShares rational.Rat) {

	exRate := c.delegatorShareExRate()

	var receivedGlobalShares rational.Rat
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
func (c *Candidate) removeShares(shares rational.Rat, gs *GlobalState) (removedTokens int64) {

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
	return Validator(*c)
}

// Validator is one of the top Candidates
type Validator Candidate

// ABCIValidator - Get the validator from a bond value
func (v Validator) ABCIValidator() *abci.Validator {
	return &abci.Validator{
		PubKey: wire.BinaryBytes(v.PubKey),
		Power:  v.VotingPower.Evaluate(),
	}
}

//_________________________________________________________________________

// TODO replace with sorted multistore functionality

// Candidates - list of Candidates
type Candidates []*Candidate

var _ sort.Interface = Candidates{} //enforce the sort interface at compile time

// nolint - sort interface functions
func (cs Candidates) Len() int      { return len(cs) }
func (cs Candidates) Swap(i, j int) { cs[i], cs[j] = cs[j], cs[i] }
func (cs Candidates) Less(i, j int) bool {
	vp1, vp2 := cs[i].VotingPower, cs[j].VotingPower
	pk1, pk2 := cs[i].PubKey.Bytes(), cs[j].PubKey.Bytes()

	//note that all ChainId and App must be the same for a group of candidates
	if vp1 != vp2 {
		return vp1.GT(vp2)
	}
	return bytes.Compare(pk1, pk2) == -1
}

// Sort - Sort the array of bonded values
func (cs Candidates) Sort() {
	sort.Sort(cs)
}

// update the voting power and save
func (cs Candidates) updateVotingPower(store state.SimpleDB, gs *GlobalState, params Params) Candidates {

	// update voting power
	for _, c := range cs {
		if !c.VotingPower.Equal(c.Assets) {
			c.VotingPower = c.Assets
		}
	}
	cs.Sort()
	for i, c := range cs {
		// truncate the power
		if i >= int(params.MaxVals) {
			c.VotingPower = rational.Zero
			if c.Status == Bonded {
				// XXX to replace this with handler.bondedToUnbondePool function
				// XXX waiting for logic with new SDK to update account balance here
				tokens := gs.removeSharesBonded(c.Assets)
				c.Assets = gs.addTokensUnbonded(tokens)
				c.Status = Unbonded
			}
		} else {
			c.Status = Bonded
		}
		saveCandidate(store, c)
	}
	return cs
}

// Validators - get the most recent updated validator set from the
// Candidates. These bonds are already sorted by VotingPower from
// the UpdateVotingPower function which is the only function which
// is to modify the VotingPower
func (cs Candidates) Validators() Validators {

	//test if empty
	if len(cs) == 1 {
		if cs[0].VotingPower.IsZero() {
			return nil
		}
	}

	validators := make(Validators, len(cs))
	for i, c := range cs {
		if c.VotingPower.IsZero() { //exit as soon as the first Voting power set to zero is found
			return validators[:i]
		}
		validators[i] = c.validator()
	}

	return validators
}

//_________________________________________________________________________

// Validators - list of Validators
type Validators []Validator

var _ sort.Interface = Validators{} //enforce the sort interface at compile time

// nolint - sort interface functions
func (vs Validators) Len() int      { return len(vs) }
func (vs Validators) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }
func (vs Validators) Less(i, j int) bool {
	pk1, pk2 := vs[i].PubKey.Bytes(), vs[j].PubKey.Bytes()
	return bytes.Compare(pk1, pk2) == -1
}

// Sort - Sort validators by pubkey
func (vs Validators) Sort() {
	sort.Sort(vs)
}

// determine all updated validators between two validator sets
func (vs Validators) validatorsUpdated(vs2 Validators) (updated []*abci.Validator) {

	//first sort the validator sets
	vs.Sort()
	vs2.Sort()

	max := len(vs) + len(vs2)
	updated = make([]*abci.Validator, max)
	i, j, n := 0, 0, 0 //counters for vs loop, vs2 loop, updated element

	for i < len(vs) && j < len(vs2) {

		if !vs[i].PubKey.Equals(vs2[j].PubKey) {
			// pk1 > pk2, a new validator was introduced between these pubkeys
			if bytes.Compare(vs[i].PubKey.Bytes(), vs2[j].PubKey.Bytes()) == 1 {
				updated[n] = vs2[j].ABCIValidator()
				n++
				j++
				continue
			} // else, the old validator has been removed
			updated[n] = &abci.Validator{vs[i].PubKey.Bytes(), 0}
			n++
			i++
			continue
		}

		if vs[i].VotingPower != vs2[j].VotingPower {
			updated[n] = vs2[j].ABCIValidator()
			n++
		}
		j++
		i++
	}

	// add any excess validators in set 2
	for ; j < len(vs2); j, n = j+1, n+1 {
		updated[n] = vs2[j].ABCIValidator()
	}

	// remove any excess validators left in set 1
	for ; i < len(vs); i, n = i+1, n+1 {
		updated[n] = &abci.Validator{vs[i].PubKey.Bytes(), 0}
	}

	return updated[:n]
}

// UpdateValidatorSet - Updates the voting power for the candidate set and
// returns the subset of validators which have been updated for Tendermint
func UpdateValidatorSet(store state.SimpleDB, gs *GlobalState, params Params) (change []*abci.Validator, err error) {

	// get the validators before update
	candidates := loadCandidates(store)

	v1 := candidates.Validators()
	v2 := candidates.updateVotingPower(store, gs, params).Validators()

	change = v1.validatorsUpdated(v2)
	return
}

//_________________________________________________________________________

// DelegatorBond represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
type DelegatorBond struct {
	PubKey crypto.PubKey `json:"pub_key"`
	Shares rational.Rat  `json:"shares"`
}
