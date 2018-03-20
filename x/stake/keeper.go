package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
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

// XXX remove beggining word get from all these keys
// GetCandidateKey - get the key for the candidate with address
func GetCandidateKey(addr sdk.Address) []byte {
	return append(CandidateKeyPrefix, addr.Bytes()...)
}

// GetValidatorKey - get the key for the validator used in the power-store
func GetValidatorKey(addr sdk.Address, power sdk.Rational, cdc *wire.Codec) []byte {
	b, _ := cdc.MarshalJSON(power)                                   // TODO need to handle error here?
	return append(ValidatorKeyPrefix, append(b, addr.Bytes()...)...) // TODO does this need prefix if its in its own store
}

// GetValidatorUpdatesKey - get the key for the validator used in the power-store
func GetValidatorUpdatesKey(addr sdk.Address) []byte {
	return append(ValidatorUpdatesKeyPrefix, addr.Bytes()...) // TODO does this need prefix if its in its own store
}

// GetDelegatorBondKey - get the key for delegator bond with candidate
func GetDelegatorBondKey(delegatorAddr, candidateAddr sdk.Address, cdc *wire.Codec) []byte {
	return append(GetDelegatorBondKeyPrefix(delegatorAddr, cdc), candidateAddr.Bytes()...)
}

// GetDelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
func GetDelegatorBondKeyPrefix(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res, err := cdc.MarshalJSON(&delegatorAddr)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondKeyPrefix, res...)
}

// GetDelegatorBondsKey - get the key for list of all the delegator's bonds
func GetDelegatorBondsKey(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res, err := cdc.MarshalJSON(&delegatorAddr)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondsKeyPrefix, res...)
}

//___________________________________________________________________________

// keeper of the staking store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *wire.Codec
	coinKeeper bank.CoinKeeper

	//just caches
	gs     GlobalState
	params Params
}

func NewKeeper(ctx sdk.Context, cdc *wire.Codec, key sdk.StoreKey, ck bank.CoinKeeper) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		coinKeeper: ck,
	}
	return keeper
}

//XXX load/save -> get/set
func (m Keeper) getCandidate(ctx sdk.Context, addr sdk.Address) (candidate Candidate) {
	store := ctx.KVStore(storeKey)
	b := store.Get(GetCandidateKey(addr))
	if b == nil {
		return nil
	}
	err := m.cdc.UnmarshalJSON(b, &candidate)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}

func (m Keeper) setCandidate(ctx sdk.Context, candidate Candidate) {
	store := ctx.KVStore(storeKey)

	// XXX should only remove validator if we know candidate is a validator
	m.removeValidator(candidate.Address)
	validator := &Validator{candidate.Address, candidate.VotingPower}
	m.updateValidator(validator)

	b, err := m.cdc.MarshalJSON(candidate)
	if err != nil {
		panic(err)
	}
	store.Set(GetCandidateKey(candidate.Address), b)
}

func (m Keeper) removeCandidate(ctx sdk.Context, candidateAddr sdk.Address) {
	store := ctx.KVStore(storeKey)

	// XXX should only remove validator if we know candidate is a validator
	m.removeValidator(candidateAddr)
	store.Delete(GetCandidateKey(candidateAddr))
}

//___________________________________________________________________________

//func loadValidator(store sdk.KVStore, address sdk.Address, votingPower sdk.Rational) *Validator {
//b := store.Get(GetValidatorKey(address, votingPower))
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
func (m Keeper) updateValidator(ctx sdk.Context, validator Validator) {
	store := ctx.KVStore(storeKey)

	b, err := m.cdc.MarshalJSON(validator)
	if err != nil {
		panic(err)
	}

	// add to the validators to update list if necessary
	store.Set(GetValidatorUpdatesKey(validator.Address), b)

	// update the list ordered by voting power
	store.Set(GetValidatorKey(validator.Address, validator.VotingPower, m.cdc), b)
}

func (m Keeper) removeValidator(ctx sdk.Context, address sdk.Address) {
	store := ctx.KVStore(storeKey)

	//add validator with zero power to the validator updates
	b, err := m.cdc.MarshalJSON(Validator{address, sdk.ZeroRat})
	if err != nil {
		panic(err)
	}
	store.Set(GetValidatorUpdatesKey(address), b)

	// now actually delete from the validator set
	candidate := m.getCandidate(address)
	if candidate != nil {
		store.Delete(GetValidatorKey(address, candidate.VotingPower, m.cdc))
	}
}

// get the most recent updated validator set from the Candidates. These bonds
// are already sorted by VotingPower from the UpdateVotingPower function which
// is the only function which is to modify the VotingPower
func (m Keeper) getValidators(ctx sdk.Context, maxVal uint16) (validators []Validator) {
	store := ctx.KVStore(storeKey)

	iterator := store.Iterator(subspace(ValidatorKeyPrefix)) //smallest to largest

	validators = make([]Validator, maxVal)
	for i := 0; ; i++ {
		if !iterator.Valid() || i > int(maxVal) {
			iterator.Close()
			break
		}
		valBytes := iterator.Value()
		var val Validator
		err := m.cdc.UnmarshalJSON(valBytes, &val)
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
func (m Keeper) getValidatorUpdates(ctx sdk.Context) (updates []Validator) {
	store := ctx.KVStore(storeKey)

	iterator := store.Iterator(subspace(ValidatorUpdatesKeyPrefix)) //smallest to largest

	for ; iterator.Valid(); iterator.Next() {
		valBytes := iterator.Value()
		var val Validator
		err := m.cdc.UnmarshalJSON(valBytes, &val)
		if err != nil {
			panic(err)
		}
		updates = append(updates, val)
	}

	iterator.Close()
	return
}

// remove all validator update entries
func (m Keeper) clearValidatorUpdates(ctx sdk.Context, maxVal int) {
	store := ctx.KVStore(storeKey)
	iterator := store.Iterator(subspace(ValidatorUpdatesKeyPrefix))
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key()) // XXX write test for this, may need to be in a second loop
	}
	iterator.Close()
}

//---------------------------------------------------------------------

// getCandidates - get the active list of all candidates TODO replace with  multistore
func (m Keeper) getCandidates(ctx sdk.Context) (candidates Candidates) {
	store := ctx.KVStore(storeKey)

	iterator := store.Iterator(subspace(CandidateKeyPrefix))
	//iterator := store.Iterator(CandidateKeyPrefix, []byte(nil))
	//iterator := store.Iterator([]byte{}, []byte(nil))

	for ; iterator.Valid(); iterator.Next() {
		candidateBytes := iterator.Value()
		var candidate Candidate
		err := m.cdc.UnmarshalJSON(candidateBytes, &candidate)
		if err != nil {
			panic(err)
		}
		candidates = append(candidates, &candidate)
	}
	iterator.Close()
	return candidates
}

//_____________________________________________________________________

// XXX use a store iterator to get
//// load the pubkeys of all candidates a delegator is delegated too
//func (m Keeper) getDelegatorCandidates(ctx sdk.Context, delegator sdk.Address) (candidateAddrs []sdk.Address) {
//store := ctx.KVStore(storeKey)

//candidateBytes := store.Get(GetDelegatorBondsKey(delegator, m.cdc))
//if candidateBytes == nil {
//return nil
//}

//err := m.cdc.UnmarshalJSON(candidateBytes, &candidateAddrs)
//if err != nil {
//panic(err)
//}
//return
//}

//_____________________________________________________________________

func (m Keeper) getDelegatorBond(ctx sdk.Context,
	delegator, candidate sdk.Address) (bond DelegatorBond) {

	store := ctx.KVStore(storeKey)
	delegatorBytes := store.Get(GetDelegatorBondKey(delegator, candidate, m.cdc))
	if delegatorBytes == nil {
		return nil
	}

	err := m.cdc.UnmarshalJSON(delegatorBytes, &bond)
	if err != nil {
		panic(err)
	}
	return bond
}

func (m Keeper) setDelegatorBond(ctx sdk.Context, bond DelegatorBond) {
	store := ctx.KVStore(storeKey)

	// XXX use store iterator
	// if a new bond add to the list of bonds
	//if m.getDelegatorBond(delegator, bond.Address) == nil {
	//pks := m.getDelegatorCandidates(delegator)
	//pks = append(pks, bond.Address)
	//b, err := m.cdc.MarshalJSON(pks)
	//if err != nil {
	//panic(err)
	//}
	//store.Set(GetDelegatorBondsKey(delegator, m.cdc), b)
	//}

	// now actually save the bond
	b, err := m.cdc.MarshalJSON(bond)
	if err != nil {
		panic(err)
	}
	store.Set(GetDelegatorBondKey(delegator, bond.Address, m.cdc), b)
}

func (m Keeper) removeDelegatorBond(ctx sdk.Context, bond DelegatorBond) {
	store := ctx.KVStore(storeKey)

	// XXX use store iterator
	// TODO use list queries on multistore to remove iterations here!
	// first remove from the list of bonds
	//addrs := m.getDelegatorCandidates(delegator)
	//for i, addr := range addrs {
	//if bytes.Equal(candidateAddr, addr) {
	//addrs = append(addrs[:i], addrs[i+1:]...)
	//}
	//}
	//b, err := m.cdc.MarshalJSON(addrs)
	//if err != nil {
	//panic(err)
	//}
	//store.Set(GetDelegatorBondsKey(delegator, m.cdc), b)

	// now remove the actual bond
	store.Delete(GetDelegatorBondKey(bond.delegatorAddr, bond.candidateAddr, m.cdc))
	//updateDelegatorBonds(store, delegator) //XXX remove?
}

//_______________________________________________________________________

// load/save the global staking params
func (m Keeper) getParams(ctx sdk.Context) (params Params) {
	// check if cached before anything
	if m.params != (Params{}) {
		return m.params
	}
	store := ctx.KVStore(storeKey)
	b := store.Get(ParamKey)
	if b == nil {
		return defaultParams()
	}

	err := m.cdc.UnmarshalJSON(b, &params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}
func (m Keeper) setParams(ctx sdk.Context, params Params) {
	store := ctx.KVStore(storeKey)
	b, err := m.cdc.MarshalJSON(params)
	if err != nil {
		panic(err)
	}
	store.Set(ParamKey, b)
	m.params = Params{} // clear the cache
}

//_______________________________________________________________________

// XXX nothing is this Keeper should return a pointer...!!!!!!
// load/save the global staking state
func (m Keeper) getGlobalState(ctx sdk.Context) (gs GlobalState) {
	// check if cached before anything
	if m.gs != nil {
		return m.gs
	}
	store := ctx.KVStore(storeKey)
	b := store.Get(GlobalStateKey)
	if b == nil {
		return initialGlobalState()
	}
	gs = new(GlobalState)
	err := m.cdc.UnmarshalJSON(b, &gs)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}

func (m Keeper) setGlobalState(ctx sdk.Context, gs GlobalState) {
	store := ctx.KVStore(storeKey)
	b, err := m.cdc.MarshalJSON(gs)
	if err != nil {
		panic(err)
	}
	store.Set(GlobalStateKey, b)
	m.gs = GlobalState{} // clear the cache
}
