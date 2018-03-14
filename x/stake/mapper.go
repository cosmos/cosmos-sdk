package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

//nolint
var (
	// Keys for store prefixes
	CandidatesAddrKey = []byte{0x01} // key for all candidates' addresses
	ParamKey          = []byte{0x02} // key for global parameters relating to staking
	GlobalStateKey    = []byte{0x03} // key for global parameters relating to staking

	// Key prefixes
	CandidateKeyPrefix        = []byte{0x04} // prefix for each key to a candidate
	ValidatorKeyPrefix        = []byte{0x05} // prefix for each key to a candidate
	ValidatorUpdatesKeyPrefix = []byte{0x06} // prefix for each key to a candidate
	DelegatorBondKeyPrefix    = []byte{0x07} // prefix for each key to a delegator's bond
	DelegatorBondsKeyPrefix   = []byte{0x08} // prefix for each key to a delegator's bond
)

// GetCandidateKey - get the key for the candidate with address
func GetCandidateKey(address sdk.Address) []byte {
	return append(CandidateKeyPrefix, address.Bytes()...)
}

// GetValidatorKey - get the key for the validator used in the power-store
func GetValidatorKey(address sdk.Address, power sdk.Rational) []byte {
	b, _ := cdc.MarshalJSON(power)                                      // TODO need to handle error here?
	return append(ValidatorKeyPrefix, append(b, address.Bytes()...)...) // TODO does this need prefix if its in its own store
}

// GetValidatorUpdatesKey - get the key for the validator used in the power-store
func GetValidatorUpdatesKey(address sdk.Address) []byte {
	return append(ValidatorUpdatesKeyPrefix, address.Bytes()...) // TODO does this need prefix if its in its own store
}

// GetDelegatorBondKey - get the key for delegator bond with candidate
func GetDelegatorBondKey(delegator, candidate sdk.Address) []byte {
	return append(GetDelegatorBondKeyPrefix(delegator), candidate.Bytes()...)
}

// GetDelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
func GetDelegatorBondKeyPrefix(delegator sdk.Address) []byte {
	res, err := cdc.MarshalJSON(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondKeyPrefix, res...)
}

// GetDelegatorBondsKey - get the key for list of all the delegator's bonds
func GetDelegatorBondsKey(delegator sdk.Address) []byte {
	res, err := cdc.MarshalJSON(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondsKeyPrefix, res...)
}

//___________________________________________________________________________

// mapper of the staking store
type Mapper struct {
	store sdk.KVStore
	cdc   *wire.Codec
}

func NewMapper(ctx sdk.Context, cdc *wire.Codec, key sdk.StoreKey) Mapper {
	return StakeMapper{
		store: ctx.KVStore(m.key),
		cdc:   cdc,
	}
}

func (m Mapper) loadCandidate(address sdk.Address) *Candidate {
	b := m.store.Get(GetCandidateKey(address))
	if b == nil {
		return nil
	}
	candidate := new(Candidate)
	err := cdc.UnmarshalJSON(b, candidate)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return candidate
}

func (m Mapper) saveCandidate(candidate *Candidate) {

	// XXX should only remove validator if we know candidate is a validator
	removeValidator(m.store, candidate.Address)
	validator := &Validator{candidate.Address, candidate.VotingPower}
	updateValidator(m.store, validator)

	b, err := cdc.MarshalJSON(*candidate)
	if err != nil {
		panic(err)
	}
	m.store.Set(GetCandidateKey(candidate.Address), b)
}

func (m Mapper) removeCandidate(address sdk.Address) {

	// XXX should only remove validator if we know candidate is a validator
	removeValidator(m.store, address)
	m.store.Delete(GetCandidateKey(address))
}

//___________________________________________________________________________

//func loadValidator(m.store sdk.KVStore, address sdk.Address, votingPower sdk.Rational) *Validator {
//b := m.store.Get(GetValidatorKey(address, votingPower))
//if b == nil {
//return nil
//}
//validator := new(Validator)
//err := cdc.UnmarshalJSON(b, validator)
//if err != nil {
//panic(err) // This error should never occur big problem if does
//}
//return validator
//}

// updateValidator - update a validator and create accumulate any changes
// in the changed validator substore
func (m Mapper) updateValidator(validator *Validator) {

	b, err := cdc.MarshalJSON(*validator)
	if err != nil {
		panic(err)
	}

	// add to the validators to update list if necessary
	m.store.Set(GetValidatorUpdatesKey(validator.Address), b)

	// update the list ordered by voting power
	m.store.Set(GetValidatorKey(validator.Address, validator.VotingPower), b)
}

func (m Mapper) removeValidator(address sdk.Address) {

	//add validator with zero power to the validator updates
	b, err := cdc.MarshalJSON(Validator{address, sdk.ZeroRat})
	if err != nil {
		panic(err)
	}
	m.store.Set(GetValidatorUpdatesKey(address), b)

	// now actually delete from the validator set
	candidate := loadCandidate(m.store, address)
	if candidate != nil {
		m.store.Delete(GetValidatorKey(address, candidate.VotingPower))
	}
}

// get the most recent updated validator set from the Candidates. These bonds
// are already sorted by VotingPower from the UpdateVotingPower function which
// is the only function which is to modify the VotingPower
func (m Mapper) getValidators(maxVal int) (validators []Validator) {

	iterator := m.store.Iterator(subspace(ValidatorKeyPrefix)) //smallest to largest

	validators = make([]Validator, maxVal)
	for i := 0; ; i++ {
		if !iterator.Valid() || i > maxVal {
			iterator.Close()
			break
		}
		valBytes := iterator.Value()
		var val Validator
		err := cdc.UnmarshalJSON(valBytes, &val)
		if err != nil {
			panic(err)
		}
		validators[i] = val
		iterator.Next()
	}

	return
}

//_________________________________________________________________________

// get the most updated validators
func (m Mapper) getValidatorUpdates() (updates []Validator) {

	iterator := m.store.Iterator(subspace(ValidatorUpdatesKeyPrefix)) //smallest to largest

	for ; iterator.Valid(); iterator.Next() {
		valBytes := iterator.Value()
		var val Validator
		err := cdc.UnmarshalJSON(valBytes, &val)
		if err != nil {
			panic(err)
		}
		updates = append(updates, val)
	}

	iterator.Close()
	return
}

// remove all validator update entries
func (m Mapper) clearValidatorUpdates(maxVal int) {
	iterator := m.store.Iterator(subspace(ValidatorUpdatesKeyPrefix))
	for ; iterator.Valid(); iterator.Next() {
		m.store.Delete(iterator.Key()) // XXX write test for this, may need to be in a second loop
	}
	iterator.Close()
}

//---------------------------------------------------------------------

// loadCandidates - get the active list of all candidates TODO replace with  multistore
func (m Mapper) loadCandidates() (candidates Candidates) {

	iterator := m.store.Iterator(subspace(CandidateKeyPrefix))
	//iterator := m.store.Iterator(CandidateKeyPrefix, []byte(nil))
	//iterator := m.store.Iterator([]byte{}, []byte(nil))

	for ; iterator.Valid(); iterator.Next() {
		candidateBytes := iterator.Value()
		var candidate Candidate
		err := cdc.UnmarshalJSON(candidateBytes, &candidate)
		if err != nil {
			panic(err)
		}
		candidates = append(candidates, &candidate)
	}
	iterator.Close()
	return candidates
}

//_____________________________________________________________________

// load the pubkeys of all candidates a delegator is delegated too
func (m Mapper) loadDelegatorCandidates(delegator sdk.Address) (candidateAddrs []sdk.Address) {

	candidateBytes := m.store.Get(GetDelegatorBondsKey(delegator))
	if candidateBytes == nil {
		return nil
	}

	err := cdc.UnmarshalJSON(candidateBytes, &candidateAddrs)
	if err != nil {
		panic(err)
	}
	return
}

//_____________________________________________________________________

func (m Mapper) loadDelegatorBond(delegator, candidate sdk.Address) *DelegatorBond {

	delegatorBytes := m.store.Get(GetDelegatorBondKey(delegator, candidate))
	if delegatorBytes == nil {
		return nil
	}

	bond := new(DelegatorBond)
	err := cdc.UnmarshalJSON(delegatorBytes, bond)
	if err != nil {
		panic(err)
	}
	return bond
}

func (m Mapper) saveDelegatorBond(delegator sdk.Address,
	bond *DelegatorBond) {

	// if a new bond add to the list of bonds
	if loadDelegatorBond(m.store, delegator, bond.Address) == nil {
		pks := loadDelegatorCandidates(m.store, delegator)
		pks = append(pks, (*bond).Address)
		b, err := cdc.MarshalJSON(pks)
		if err != nil {
			panic(err)
		}
		m.store.Set(GetDelegatorBondsKey(delegator), b)
	}

	// now actually save the bond
	b, err := cdc.MarshalJSON(*bond)
	if err != nil {
		panic(err)
	}
	m.store.Set(GetDelegatorBondKey(delegator, bond.Address), b)
	//updateDelegatorBonds(store, delegator)
}

func (m Mapper) removeDelegatorBond(delegator sdk.Address, candidate sdk.Address) {
	// TODO use list queries on multistore to remove iterations here!
	// first remove from the list of bonds
	pks := loadDelegatorCandidates(m.store, delegator)
	for i, pk := range pks {
		if candidate.Equals(pk) {
			pks = append(pks[:i], pks[i+1:]...)
		}
	}
	b, err := cdc.MarshalJSON(pks)
	if err != nil {
		panic(err)
	}
	m.store.Set(GetDelegatorBondsKey(delegator), b)

	// now remove the actual bond
	m.store.Delete(GetDelegatorBondKey(delegator, candidate))
	//updateDelegatorBonds(store, delegator)
}

//_______________________________________________________________________

// load/save the global staking params
func (m Mapper) loadParams() (params Params) {
	b := m.store.Get(ParamKey)
	if b == nil {
		return defaultParams()
	}

	err := cdc.UnmarshalJSON(b, &params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}
func (m Mapper) saveParams(params Params) {
	b, err := cdc.MarshalJSON(params)
	if err != nil {
		panic(err)
	}
	m.store.Set(ParamKey, b)
}

//_______________________________________________________________________

// load/save the global staking state
func (m Mapper) loadGlobalState() (gs *GlobalState) {
	b := m.store.Get(GlobalStateKey)
	if b == nil {
		return initialGlobalState()
	}
	gs = new(GlobalState)
	err := cdc.UnmarshalJSON(b, gs)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}

func (m Mapper) saveGlobalState(gs *GlobalState) {
	b, err := cdc.MarshalJSON(*gs)
	if err != nil {
		panic(err)
	}
	m.store.Set(GlobalStateKey, b)
}
