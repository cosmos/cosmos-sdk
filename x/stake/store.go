package stake

import (
	"encoding/binary"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/types"
)

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

	//iterator := store.Iterator(subspace(CandidateKeyPrefix)) //smallest to largest
	//iterator := store.Iterator(CandidateKeyPrefix, []byte(nil)) //smallest to largest
	iterator := store.Iterator([]byte{}, []byte(nil)) //smallest to largest

	for i := 0; ; i++ {
		if !iterator.Valid() {
			//panic(fmt.Sprintf("debug i: %v\n", i))
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
