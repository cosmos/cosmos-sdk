package stake

import (
	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/rational"

	"github.com/cosmos/cosmos-sdk/types"
)

// nolint
var (

	// internal wire codec
	cdc *wire.Codec

	// Keys for store prefixes
	CandidatesPubKeysKey = []byte{0x01} // key for all candidates' pubkeys
	ParamKey             = []byte{0x02} // key for global parameters relating to staking
	GlobalStateKey       = []byte{0x03} // key for global parameters relating to staking

	// Key prefixes
	CandidateKeyPrefix        = []byte{0x04} // prefix for each key to a candidate
	ValidatorKeyPrefix        = []byte{0x05} // prefix for each key to a candidate
	ValidatorUpdatesKeyPrefix = []byte{0x06} // prefix for each key to a candidate
	DelegatorBondKeyPrefix    = []byte{0x07} // prefix for each key to a delegator's bond
	DelegatorBondsKeyPrefix   = []byte{0x08} // prefix for each key to a delegator's bond
)

func init() {
	cdc = wire.NewCodec()
	cdc.RegisterInterface((*rational.Rational)(nil), nil)
	cdc.RegisterConcrete(rational.Rat{}, "rat", nil)
}

// GetCandidateKey - get the key for the candidate with pubKey
func GetCandidateKey(pubKey crypto.PubKey) []byte {
	return append(CandidateKeyPrefix, pubKey.Bytes()...)
}

// GetValidatorKey - get the key for the validator used in the power-store
func GetValidatorKey(pubKey crypto.PubKey, power rational.Rational) []byte {
	b, _ := cdc.MarshalJSON(power)                                     // TODO need to handle error here?
	return append(ValidatorKeyPrefix, append(b, pubKey.Bytes()...)...) // TODO does this need prefix if its in its own store
}

// GetValidatorUpdatesKey - get the key for the validator used in the power-store
func GetValidatorUpdatesKey(pubKey crypto.PubKey) []byte {
	return append(ValidatorUpdatesKeyPrefix, pubKey.Bytes()...) // TODO does this need prefix if its in its own store
}

// GetDelegatorBondKey - get the key for delegator bond with candidate
func GetDelegatorBondKey(delegator crypto.Address, candidate crypto.PubKey) []byte {
	return append(GetDelegatorBondKeyPrefix(delegator), candidate.Bytes()...)
}

// GetDelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
func GetDelegatorBondKeyPrefix(delegator crypto.Address) []byte {
	res, err := cdc.MarshalJSON(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondKeyPrefix, res...)
}

// GetDelegatorBondsKey - get the key for list of all the delegator's bonds
func GetDelegatorBondsKey(delegator crypto.Address) []byte {
	res, err := cdc.MarshalJSON(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondsKeyPrefix, res...)
}

//---------------------------------------------------------------------

func loadCandidate(store types.KVStore, pubKey crypto.PubKey) *Candidate {
	b := store.Get(GetCandidateKey(pubKey))
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

func saveCandidate(store types.KVStore, candidate *Candidate) {

	// XXX should only remove validator if we know candidate is a validator
	removeValidator(store, candidate.PubKey)
	validator := &Validator{candidate.PubKey, candidate.VotingPower}
	updateValidator(store, validator)

	b, err := cdc.MarshalJSON(*candidate)
	if err != nil {
		panic(err)
	}
	store.Set(GetCandidateKey(candidate.PubKey), b)
}

func removeCandidate(store types.KVStore, pubKey crypto.PubKey) {

	// XXX should only remove validator if we know candidate is a validator
	removeValidator(store, pubKey)
	store.Delete(GetCandidateKey(pubKey))
}

//---------------------------------------------------------------------

//func loadValidator(store types.KVStore, pubKey crypto.PubKey, votingPower rational.Rational) *Validator {
//b := store.Get(GetValidatorKey(pubKey, votingPower))
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
func updateValidator(store types.KVStore, validator *Validator) {

	b, err := cdc.MarshalJSON(*validator)
	if err != nil {
		panic(err)
	}

	// add to the validators to update list if necessary
	store.Set(GetValidatorUpdatesKey(validator.PubKey), b)

	// update the list ordered by voting power
	store.Set(GetValidatorKey(validator.PubKey, validator.VotingPower), b)
}

func removeValidator(store types.KVStore, pubKey crypto.PubKey) {

	//add validator with zero power to the validator updates
	b, err := cdc.MarshalJSON(Validator{pubKey, rational.Zero})
	if err != nil {
		panic(err)
	}
	store.Set(GetValidatorUpdatesKey(pubKey), b)

	// now actually delete from the validator set
	candidate := loadCandidate(store, pubKey)
	if candidate != nil {
		store.Delete(GetValidatorKey(pubKey, candidate.VotingPower))
	}
}

// get the most recent updated validator set from the Candidates. These bonds
// are already sorted by VotingPower from the UpdateVotingPower function which
// is the only function which is to modify the VotingPower
func getValidators(store types.KVStore, maxVal int) (validators []Validator) {

	iterator := store.Iterator(subspace(ValidatorKeyPrefix)) //smallest to largest

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

//---------------------------------------------------------------------

// get the most updated validators
func getValidatorUpdates(store types.KVStore) (updates []Validator) {

	iterator := store.Iterator(subspace(ValidatorUpdatesKeyPrefix)) //smallest to largest

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
func clearValidatorUpdates(store types.KVStore, maxVal int) {
	iterator := store.Iterator(subspace(ValidatorUpdatesKeyPrefix))
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key()) // XXX write test for this, may need to be in a second loop
	}
	iterator.Close()
}

//---------------------------------------------------------------------

// loadCandidates - get the active list of all candidates TODO replace with  multistore
func loadCandidates(store types.KVStore) (candidates Candidates) {

	iterator := store.Iterator(subspace(CandidateKeyPrefix))
	//iterator := store.Iterator(CandidateKeyPrefix, []byte(nil))
	//iterator := store.Iterator([]byte{}, []byte(nil))

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

//---------------------------------------------------------------------

// load the pubkeys of all candidates a delegator is delegated too
func loadDelegatorCandidates(store types.KVStore,
	delegator crypto.Address) (candidates []crypto.PubKey) {

	candidateBytes := store.Get(GetDelegatorBondsKey(delegator))
	if candidateBytes == nil {
		return nil
	}

	err := cdc.UnmarshalJSON(candidateBytes, &candidates)
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
	err := cdc.UnmarshalJSON(delegatorBytes, bond)
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
		b, err := cdc.MarshalJSON(pks)
		if err != nil {
			panic(err)
		}
		store.Set(GetDelegatorBondsKey(delegator), b)
	}

	// now actually save the bond
	b, err := cdc.MarshalJSON(*bond)
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
	b, err := cdc.MarshalJSON(pks)
	if err != nil {
		panic(err)
	}
	store.Set(GetDelegatorBondsKey(delegator), b)

	// now remove the actual bond
	store.Delete(GetDelegatorBondKey(delegator, candidate))
	//updateDelegatorBonds(store, delegator)
}

//_______________________________________________________________________

// load/save the global staking params
func loadParams(store types.KVStore) (params Params) {
	b := store.Get(ParamKey)
	if b == nil {
		return defaultParams()
	}

	err := cdc.UnmarshalJSON(b, &params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}

	return
}
func saveParams(store types.KVStore, params Params) {
	b, err := cdc.MarshalJSON(params)
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
	err := cdc.UnmarshalJSON(b, gs)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}
func saveGlobalState(store types.KVStore, gs *GlobalState) {
	b, err := cdc.MarshalJSON(*gs)
	if err != nil {
		panic(err)
	}
	store.Set(GlobalStateKey, b)
}
