package stake

import (
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

var cdc = wire.NewCodec()

//nolint
type Params struct {
	HoldBonded   crypto.Address `json:"hold_bonded"`   // account  where all bonded coins are held
	HoldUnbonded crypto.Address `json:"hold_unbonded"` // account where all delegated but unbonded coins are held

	InflationRateChange int64 `json:"inflation_rate_change"` // XXX maximum annual change in inflation rate
	InflationMax        int64 `json:"inflation_max"`         // XXX maximum inflation rate
	InflationMin        int64 `json:"inflation_min"`         // XXX minimum inflation rate
	GoalBonded          int64 `json:"goal_bonded"`           // XXX Goal of percent bonded atoms

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
		InflationRateChange: 13, //rational.New(13, 100),
		InflationMax:        20, //rational.New(20, 100),
		InflationMin:        7,  //rational.New(7, 100),
		GoalBonded:          67, //rational.New(67, 100),
		MaxVals:             100,
		AllowedBondDenom:    "fermion",
		GasDeclareCandidacy: 20,
		GasEditCandidacy:    20,
		GasDelegate:         20,
		GasUnbond:           20,
	}
}

// GlobalState - dynamic parameters of the current state
type GlobalState struct {
	TotalSupply       int64 `json:"total_supply"`        // total supply of all tokens
	BondedShares      int64 `json:"bonded_shares"`       // sum of all shares distributed for the Bonded Pool
	UnbondedShares    int64 `json:"unbonded_shares"`     // sum of all shares distributed for the Unbonded Pool
	BondedPool        int64 `json:"bonded_pool"`         // reserve of bonded tokens
	UnbondedPool      int64 `json:"unbonded_pool"`       // reserve of unbonded tokens held with candidates
	InflationLastTime int64 `json:"inflation_last_time"` // block which the last inflation was processed // TODO make time
	Inflation         int64 `json:"inflation"`           // current annual inflation rate
}

// XXX define globalstate interface?

func initialGlobalState() *GlobalState {
	return &GlobalState{
		TotalSupply:       0,
		BondedShares:      0, //rational.Zero,
		UnbondedShares:    0, //rational.Zero,
		BondedPool:        0,
		UnbondedPool:      0,
		InflationLastTime: 0,
		Inflation:         0, //rational.New(7, 100),
	}
}

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
	Owner       crypto.Address  `json:"owner"`        // Sender of BondTx - UnbondTx returns here
	Assets      int64           `json:"assets"`       // total shares of a global hold pools TODO custom type PoolShares
	Liabilities int64           `json:"liabilities"`  // total shares issued to a candidate's delegators TODO custom type DelegatorShares
	VotingPower int64           `json:"voting_power"` // Voting power if considered a validator
	Description Description     `json:"description"`  // Description terms for the candidate
}

//nolint
type Candidates []*Candidate
type Validator struct {
	PubKey      crypto.PubKey `json:"pub_key"`      // Pubkey of candidate
	VotingPower int64         `json:"voting_power"` // Voting power if considered a validator
}
type Validators []Validator

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
		Assets:      0, // rational.Zero,
		Liabilities: 0, // rational.Zero,
		VotingPower: 0, //rational.Zero,
		Description: description,
	}
}

//nolint
type DelegatorBond struct {
	PubKey crypto.PubKey `json:"pub_key"`
	Shares int64         `json:"shares"`
}
