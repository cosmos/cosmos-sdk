package stake

//// keeper of the staking store
//type Keeper struct {
//storeKey   sdk.StoreKey
//cdc        *codec.Codec
//bankKeeper bank.Keeper

//// codespace
//codespace sdk.CodespaceType
//}

//func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ck bank.Keeper, codespace sdk.CodespaceType) Keeper {
//keeper := Keeper{
//storeKey:   key,
//cdc:        cdc,
//bankKeeper: ck,
//codespace:  codespace,
//}
//return keeper
//}

////_________________________________________________________________________

//// cummulative power of the non-absent prevotes
//func (k Keeper) GetTotalPrecommitVotingPower(ctx sdk.Context) sdk.Dec {
//store := ctx.KVStore(k.storeKey)

//// get absent prevote indexes
//absents := ctx.AbsentValidators()

//TotalPower := sdk.ZeroDec()
//i := int32(0)
//iterator := store.SubspaceIterator(ValidatorsBondedKey)
//for ; iterator.Valid(); iterator.Next() {

//skip := false
//for j, absentIndex := range absents {
//if absentIndex > i {
//break
//}

//// if non-voting validator found, skip adding its power
//if absentIndex == i {
//absents = append(absents[:j], absents[j+1:]...) // won't need again
//skip = true
//break
//}
//}
//if skip {
//continue
//}

//bz := iterator.Value()
//var validator Validator
//k.cdc.MustUnmarshalBinary(bz, &validator)
//TotalPower = TotalPower.Add(validator.Power)
//i++
//}
//iterator.Close()
//return TotalPower
//}

////_______________________________________________________________________

//// XXX TODO trim functionality

//// retrieve all the power changes which occur after a height
//func (k Keeper) GetPowerChangesAfterHeight(ctx sdk.Context, earliestHeight int64) (pcs []PowerChange) {
//store := ctx.KVStore(k.storeKey)

//iterator := store.SubspaceIterator(PowerChangeKey) //smallest to largest
//for ; iterator.Valid(); iterator.Next() {
//pcBytes := iterator.Value()
//var pc PowerChange
//k.cdc.MustUnmarshalBinary(pcBytes, &pc)
//if pc.Height < earliestHeight {
//break
//}
//pcs = append(pcs, pc)
//}
//iterator.Close()
//return
//}

//// set a power change
//func (k Keeper) setPowerChange(ctx sdk.Context, pc PowerChange) {
//store := ctx.KVStore(k.storeKey)
//b := k.cdc.MustMarshalBinary(pc)
//store.Set(GetPowerChangeKey(pc.Height), b)
//}
