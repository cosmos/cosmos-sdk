package stake

import (
	"encoding/json"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"

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
	DelegatorBondKeyPrefix  = []byte{0x05} // prefix for each key to a delegator's bond
	DelegatorBondsKeyPrefix = []byte{0x06} // prefix for each key to a delegator's bond
)

// GetCandidateKey - get the key for the candidate with pubKey
func GetCandidateKey(pubKey crypto.PubKey) []byte {
	return append(CandidateKeyPrefix, pubKey.Bytes()...)
}

// GetDelegatorBondKey - get the key for delegator bond with candidate
func GetDelegatorBondKey(delegator crypto.Address, candidate crypto.PubKey) []byte {
	return append(GetDelegatorBondKeyPrefix(delegator), candidate.Bytes()...)
}

// GetDelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
func GetDelegatorBondKeyPrefix(delegator crypto.Address) []byte {
	res, err := wire.MarshalBinary(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondKeyPrefix, res...)
}

// GetDelegatorBondsKey - get the key for list of all the delegator's bonds
func GetDelegatorBondsKey(delegator crypto.Address) []byte {
	res, err := wire.MarshalBinary(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondsKeyPrefix, res...)
}

//---------------------------------------------------------------------

// Get the active list of all the candidate pubKeys and owners
func loadCandidatesPubKeys(store types.KVStore) (pubKeys []crypto.PubKey) {
	bytes := store.Get(CandidatesPubKeysKey)
	if bytes == nil {
		return
	}
	err := wire.UnmarshalBinary(bytes, &pubKeys)
	if err != nil {
		panic(err)
	}
	return
}
func saveCandidatesPubKeys(store types.KVStore, pubKeys []crypto.PubKey) {
	b, err := wire.MarshalBinary(pubKeys)
	if err != nil {
		panic(err)
	}
	store.Set(CandidatesPubKeysKey, b)
}

// loadCandidates - get the active list of all candidates TODO replace with  multistore
func loadCandidates(store types.KVStore) (candidates Candidates) {
	pks := loadCandidatesPubKeys(store)
	for _, pk := range pks {
		candidates = append(candidates, loadCandidate(store, pk))
	}
	return
}

//---------------------------------------------------------------------

// loadCandidate - loads the candidate object for the provided pubkey
func loadCandidate(store types.KVStore, pubKey crypto.PubKey) *Candidate {
	//if pubKey.Empty() {
	//return nil
	//}
	b := store.Get(GetCandidateKey(pubKey))
	if b == nil {
		return nil
	}
	candidate := new(Candidate)
	err := json.Unmarshal(b, candidate)
	if err != nil {
		panic(err) // This error should never occure big problem if does
	}
	return candidate
}

func saveCandidate(store types.KVStore, candidate *Candidate) {

	if !store.Has(GetCandidateKey(candidate.PubKey)) {
		// TODO to be replaced with iteration in the multistore?
		pks := loadCandidatesPubKeys(store)
		saveCandidatesPubKeys(store, append(pks, candidate.PubKey))
	}

	b, err := json.Marshal(*candidate)
	if err != nil {
		panic(err)
	}
	store.Set(GetCandidateKey(candidate.PubKey), b)
}

func removeCandidate(store types.KVStore, pubKey crypto.PubKey) {
	store.Delete(GetCandidateKey(pubKey))

	// TODO to be replaced with iteration in the multistore?
	pks := loadCandidatesPubKeys(store)
	for i := range pks {
		if pks[i].Equals(pubKey) {
			saveCandidatesPubKeys(store,
				append(pks[:i], pks[i+1:]...))
			break
		}
	}
}

//---------------------------------------------------------------------

// load the pubkeys of all candidates a delegator is delegated too
func loadDelegatorCandidates(store types.KVStore,
	delegator crypto.Address) (candidates []crypto.PubKey) {

	candidateBytes := store.Get(GetDelegatorBondsKey(delegator))
	if candidateBytes == nil {
		return nil
	}

	err := wire.UnmarshalBinary(candidateBytes, &candidates)
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
	err := json.Unmarshal(delegatorBytes, bond)
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
		b, err := wire.MarshalBinary(pks)
		if err != nil {
			panic(err)
		}
		store.Set(GetDelegatorBondsKey(delegator), b)
	}

	// now actually save the bond
	b, err := json.Marshal(*bond)
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
	b, err := wire.MarshalBinary(pks)
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

	err := json.Unmarshal(b, &params)
	if err != nil {
		panic(err) // This error should never occure big problem if does
	}

	return
}
func saveParams(store types.KVStore, params Params) {
	b, err := json.Marshal(params)
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
	err := json.Unmarshal(b, gs)
	if err != nil {
		panic(err) // This error should never occure big problem if does
	}
	return
}
func saveGlobalState(store types.KVStore, gs *GlobalState) {
	b, err := json.Marshal(*gs)
	if err != nil {
		panic(err)
	}
	store.Set(GlobalStateKey, b)
}
