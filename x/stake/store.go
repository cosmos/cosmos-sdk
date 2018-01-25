package stake

import (
	"encoding/binary"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/types"
)

/////////////////////////////////////////////////////////// temp types

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

///////////////////////////////////////////////////////////j

// nolint
var (
	// Keys for store prefixes
	CandidatesPubKeysKey = []byte{0x01} // key for all candidates' pubkeys
	ParamKey             = []byte{0x02} // key for global parameters relating to staking
	GlobalStateKey       = []byte{0x03} // key for global parameters relating to staking

	// Key prefixes
	CandidateKeyPrefix      = []byte{0x04} // prefix for each key to a candidate
	ValidatorKeyPrefix      = []byte{0x05} // prefix for each key to a candidate
	DelegatorBondKeyPrefix  = []byte{0x06} // prefix for each key to a delegator's bond
	DelegatorBondsKeyPrefix = []byte{0x07} // prefix for each key to a delegator's bond
)

// GetCandidateKey - get the key for the candidate with pubKey
func GetCandidateKey(pubKey crypto.PubKey) []byte {
	return append(CandidateKeyPrefix, pubKey.Bytes()...)
}

// GetValidatorKey - get the key for the validator used in the power-store
func GetValidatorKey(pubKey crypto.PubKey, power int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(power))
	return append(ValidatorKeyPrefix, append(b, pubKey.Bytes()...)...) // TODO does this need prefix if its in its own store
}

// GetDelegatorBondKey - get the key for delegator bond with candidate
func GetDelegatorBondKey(delegator crypto.Address, candidate crypto.PubKey) []byte {
	return append(GetDelegatorBondKeyPrefix(delegator), candidate.Bytes()...)
}

// GetDelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
func GetDelegatorBondKeyPrefix(delegator crypto.Address) []byte {
	res, err := cdc.MarshalBinary(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondKeyPrefix, res...)
}

// GetDelegatorBondsKey - get the key for list of all the delegator's bonds
func GetDelegatorBondsKey(delegator crypto.Address) []byte {
	res, err := cdc.MarshalBinary(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondsKeyPrefix, res...)
}

//---------------------------------------------------------------------

func loadCandidate(store types.KVStore, pubKey crypto.PubKey) *Candidate {
	//if pubKey.Empty() {
	//return nil
	//}
	b := store.Get(GetCandidateKey(pubKey))
	if b == nil {
		return nil
	}
	candidate := new(Candidate)
	err := cdc.UnmarshalBinary(b, candidate)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return candidate
}

func saveCandidate(store types.KVStore, candidate *Candidate) {

	removeValidatorFromKey(store, candidate.PubKey)
	validator := &Validator{candidate.PubKey, candidate.VotingPower}
	saveValidator(store, validator)

	b, err := cdc.MarshalBinary(*candidate)
	if err != nil {
		panic(err)
	}
	store.Set(GetCandidateKey(candidate.PubKey), b)
}

func removeCandidate(store types.KVStore, pubKey crypto.PubKey) {
	removeValidatorFromKey(store, pubKey)
	store.Delete(GetCandidateKey(pubKey))
}

//---------------------------------------------------------------------

func loadValidator(store types.KVStore, pubKey crypto.PubKey, votingPower int64) *Validator {
	b := store.Get(GetValidatorKey(pubKey, votingPower))
	if b == nil {
		return nil
	}
	validator := new(Validator)
	err := cdc.UnmarshalBinary(b, validator)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return validator
}

func saveValidator(store types.KVStore, validator *Validator) {
	b, err := cdc.MarshalBinary(*validator)
	if err != nil {
		panic(err)
	}
	store.Set(GetValidatorKey(validator.PubKey, validator.VotingPower), b)
}

func removeValidator(store types.KVStore, pubKey crypto.PubKey, votingPower int64) {
	store.Delete(GetValidatorKey(pubKey, votingPower))
}

func removeValidatorFromKey(store types.KVStore, pubKey crypto.PubKey) {
	// remove validator if already there, then add new validator
	candidate := loadCandidate(store, pubKey)
	if candidate != nil {
		removeValidator(store, pubKey, candidate.VotingPower)
	}
}

// Validators - get the most recent updated validator set from the
// Candidates. These bonds are already sorted by VotingPower from
// the UpdateVotingPower function which is the only function which
// is to modify the VotingPower
func getValidators(store types.KVStore, maxVal int) Validators {

	iterator := store.Iterator(subspace(ValidatorKeyPrefix)) //smallest to largest

	validators := make(Validators, maxVal)
	for i := 0; ; i++ {
		if !iterator.Valid() || i > maxVal {
			iterator.Close()
			break
		}
		valBytes := iterator.Value()
		var val Validator
		err := cdc.UnmarshalBinary(valBytes, &val)
		if err != nil {
			panic(err)
		}
		validators[i] = val
		iterator.Next()
	}

	return validators
}

//---------------------------------------------------------------------

// loadCandidates - get the active list of all candidates TODO replace with  multistore
func loadCandidates(store types.KVStore) (candidates Candidates) {

	iterator := store.Iterator(subspace(CandidateKeyPrefix)) //smallest to largest

	for i := 0; ; i++ {
		if !iterator.Valid() {
			iterator.Close()
			break
		}
		candidateBytes := iterator.Value()
		var candidate Candidate
		err := cdc.UnmarshalBinary(candidateBytes, &candidate)
		if err != nil {
			panic(err)
		}
		candidates[i] = &candidate
		iterator.Next()
	}

	return candidates
}

//---------------------------------------------------------------------

// load the pubkeys of all candidates a delegator is delegated too
func loadDelegatorCandidates(store types.KVStore,
	delegator crypto.Address) (candidates []crypto.PubKey) {

	candidateBytes := store.Get(GetDelegatorBondsKey(delegator))
	if candidateBytes == nil {
		return nil
	}

	err := cdc.UnmarshalBinary(candidateBytes, &candidates)
	if err != nil {
		panic(err)
	}
	return
}

//---------------------------------------------------------------------

func loadDelegatorBond(store types.KVStore,
	delegator crypto.Address, candidate crypto.PubKey) *DelegatorBond {

	delegatorBytes := store.Get(GetDelegatorBondKey(delegator, candidate))
	if delegatorBytes == nil {
		return nil
	}

	bond := new(DelegatorBond)
	err := cdc.UnmarshalBinary(delegatorBytes, bond)
	if err != nil {
		panic(err)
	}
	return bond
}

func saveDelegatorBond(store types.KVStore, delegator crypto.Address, bond *DelegatorBond) {

	// if a new bond add to the list of bonds
	if loadDelegatorBond(store, delegator, bond.PubKey) == nil {
		pks := loadDelegatorCandidates(store, delegator)
		pks = append(pks, (*bond).PubKey)
		b, err := cdc.MarshalBinary(pks)
		if err != nil {
			panic(err)
		}
		store.Set(GetDelegatorBondsKey(delegator), b)
	}

	// now actually save the bond
	b, err := cdc.MarshalBinary(*bond)
	if err != nil {
		panic(err)
	}
	store.Set(GetDelegatorBondKey(delegator, bond.PubKey), b)
	//updateDelegatorBonds(store, delegator)
}

func removeDelegatorBond(store types.KVStore, delegator crypto.Address, candidate crypto.PubKey) {
	// TODO use list queries on multistore to remove iterations here!
	// first remove from the list of bonds
	pks := loadDelegatorCandidates(store, delegator)
	for i, pk := range pks {
		if candidate.Equals(pk) {
			pks = append(pks[:i], pks[i+1:]...)
		}
	}
	b, err := cdc.MarshalBinary(pks)
	if err != nil {
		panic(err)
	}
	store.Set(GetDelegatorBondsKey(delegator), b)

	// now remove the actual bond
	store.Delete(GetDelegatorBondKey(delegator, candidate))
	//updateDelegatorBonds(store, delegator)
}

//func updateDelegatorBonds(store types.KVStore,
//delegator crypto.Address) {

//var bonds []*DelegatorBond

//prefix := GetDelegatorBondKeyPrefix(delegator)
//l := len(prefix)
//delegatorsBytes := store.List(prefix,
//append(prefix[:l-1], (prefix[l-1]+1)), loadParams(store).MaxVals)

//for _, delegatorBytesModel := range delegatorsBytes {
//delegatorBytes := delegatorBytesModel.Value
//if delegatorBytes == nil {
//continue
//}

//bond := new(DelegatorBond)
//err := wire.UnmarshalBinary(delegatorBytes, bond)
//if err != nil {
//panic(err)
//}
//bonds = append(bonds, bond)
//}

//if len(bonds) == 0 {
//store.Remove(GetDelegatorBondsKey(delegator))
//return
//}

//b := wire.MarshalBinary(bonds)
//store.Set(GetDelegatorBondsKey(delegator), b)
//}

//_______________________________________________________________________

// load/save the global staking params
func loadParams(store types.KVStore) (params Params) {
	b := store.Get(ParamKey)
	if b == nil {
		return defaultParams()
	}

	err := cdc.UnmarshalBinary(b, &params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}

	return
}
func saveParams(store types.KVStore, params Params) {
	b, err := cdc.MarshalBinary(params)
	if err != nil {
		panic(err)
	}
	store.Set(ParamKey, b)
}

//_______________________________________________________________________

// load/save the global staking state
func loadGlobalState(store types.KVStore) (gs *GlobalState) {
	b := store.Get(GlobalStateKey)
	if b == nil {
		return initialGlobalState()
	}
	gs = new(GlobalState)
	err := cdc.UnmarshalBinary(b, gs)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}
func saveGlobalState(store types.KVStore, gs *GlobalState) {
	b, err := cdc.MarshalBinary(*gs)
	if err != nil {
		panic(err)
	}
	store.Set(GlobalStateKey, b)
}
