package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	bankKeeper quarantine.BankKeeper

	fundsHolder sdk.AccAddress
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper quarantine.BankKeeper, fundsHolder sdk.AccAddress) Keeper {
	if len(fundsHolder) == 0 {
		fundsHolder = authtypes.NewModuleAddress(quarantine.ModuleName)
	}
	rv := Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		bankKeeper:  bankKeeper,
		fundsHolder: fundsHolder,
	}
	bankKeeper.SetQuarantineKeeper(rv)
	return rv
}

// GetFundsHolder returns the account address that holds quarantined funds.
func (k Keeper) GetFundsHolder() sdk.AccAddress {
	return k.fundsHolder
}

// SetOptIn records that an address has opted into quarantine.
func (k Keeper) SetOptIn(ctx sdk.Context, toAddr sdk.AccAddress) error {
	key := quarantine.CreateOptInKey(toAddr)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, []byte{0x00})
	return ctx.EventManager().EmitTypedEvent(&quarantine.EventOptIn{ToAddress: toAddr.String()})
}

// SetOptOut removes an address' quarantine opt-in record.
func (k Keeper) SetOptOut(ctx sdk.Context, toAddr sdk.AccAddress) error {
	key := quarantine.CreateOptInKey(toAddr)
	store := ctx.KVStore(k.storeKey)
	store.Delete(key)
	return ctx.EventManager().EmitTypedEvent(&quarantine.EventOptOut{ToAddress: toAddr.String()})
}

// IsQuarantinedAddr returns true if the given address has opted into quarantine.
func (k Keeper) IsQuarantinedAddr(ctx sdk.Context, toAddr sdk.AccAddress) bool {
	key := quarantine.CreateOptInKey(toAddr)
	store := ctx.KVStore(k.storeKey)
	return store.Has(key)
}

// getQuarantinedAccountsPrefixStore returns a kv store prefixed for quarantine opt-in entries, and the prefix bytes.
func (k Keeper) getQuarantinedAccountsPrefixStore(ctx sdk.Context) (sdk.KVStore, []byte) {
	return prefix.NewStore(ctx.KVStore(k.storeKey), quarantine.OptInPrefix), quarantine.OptInPrefix
}

// IterateQuarantinedAccounts iterates over all quarantine account addresses.
// The callback function should accept the to address (that has quarantine enabled).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k Keeper) IterateQuarantinedAccounts(ctx sdk.Context, cb func(toAddr sdk.AccAddress) (stop bool)) {
	store, pre := k.getQuarantinedAccountsPrefixStore(ctx)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		addr := quarantine.ParseOptInKey(quarantine.MakeKey(pre, iter.Key()))
		if cb(addr) {
			break
		}
	}
}

// SetAutoResponse sets the auto response of sends to toAddr from fromAddr.
// If the response is AUTO_RESPONSE_UNSPECIFIED, the auto-response record is deleted,
// otherwise it is created/updated with the given setting.
func (k Keeper) SetAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) {
	key := quarantine.CreateAutoResponseKey(toAddr, fromAddr)
	val := quarantine.ToAutoB(response)
	store := ctx.KVStore(k.storeKey)
	if val == quarantine.NoAutoB {
		store.Delete(key)
	} else {
		store.Set(key, []byte{val})
	}
}

// GetAutoResponse returns the quarantine auto-response for the given to/from addresses.
func (k Keeper) GetAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) quarantine.AutoResponse {
	if toAddr.Equals(fromAddr) {
		return quarantine.AUTO_RESPONSE_ACCEPT
	}
	key := quarantine.CreateAutoResponseKey(toAddr, fromAddr)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	return quarantine.ToAutoResponse(bz)
}

// IsAutoAccept returns true if the to address has enabled auto-accept for ALL the from address.
func (k Keeper) IsAutoAccept(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) bool {
	for _, fromAddr := range fromAddrs {
		if !k.GetAutoResponse(ctx, toAddr, fromAddr).IsAccept() {
			return false
		}
	}
	return true
}

// IsAutoDecline returns true if the to address has enabled auto-decline for ANY of the from address.
func (k Keeper) IsAutoDecline(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) bool {
	for _, fromAddr := range fromAddrs {
		if k.GetAutoResponse(ctx, toAddr, fromAddr).IsDecline() {
			return true
		}
	}
	return false
}

// getAutoResponsesPrefixStore returns a kv store prefixed for quarantine auto-responses and the prefix used.
// If a toAddr is provided, the store is prefixed for just the given address.
// If toAddr is empty, it will be prefixed for all quarantine auto-responses.
func (k Keeper) getAutoResponsesPrefixStore(ctx sdk.Context, toAddr sdk.AccAddress) (sdk.KVStore, []byte) {
	pre := quarantine.AutoResponsePrefix
	if len(toAddr) > 0 {
		pre = quarantine.CreateAutoResponseToAddrPrefix(toAddr)
	}
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre), pre
}

// IterateAutoResponses iterates over the auto-responses for a given recipient address,
// or if no address is provided, iterates over all auto-response entries.
// The callback function should accept a to address, from address, and auto-response setting (in that order).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k Keeper) IterateAutoResponses(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) (stop bool)) {
	store, pre := k.getAutoResponsesPrefixStore(ctx, toAddr)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kToAddr, kFromAddr := quarantine.ParseAutoResponseKey(quarantine.MakeKey(pre, iter.Key()))
		val := quarantine.ToAutoResponse(iter.Value())
		if cb(kToAddr, kFromAddr, val) {
			break
		}
	}
}

// SetQuarantineRecord sets a quarantine record.
// Panics if the record is nil.
// If the record is fully accepted, it is deleted.
// Otherwise, it is saved.
func (k Keeper) SetQuarantineRecord(ctx sdk.Context, toAddr sdk.AccAddress, record *quarantine.QuarantineRecord) {
	if record == nil {
		panic("record cannot be nil")
	}
	fromAddrs := record.GetAllFromAddrs()
	key := quarantine.CreateRecordKey(toAddr, fromAddrs...)
	store := ctx.KVStore(k.storeKey)

	if record.IsFullyAccepted() {
		store.Delete(key)
		if len(fromAddrs) > 1 {
			_, suffix := quarantine.ParseRecordIndexKey(key)
			k.deleteQuarantineRecordSuffixIndexes(store, toAddr, fromAddrs, suffix)
		}
	} else {
		val := k.cdc.MustMarshal(record)
		store.Set(key, val)
		if len(fromAddrs) > 1 {
			_, suffix := quarantine.ParseRecordIndexKey(key)
			k.addQuarantineRecordSuffixIndexes(store, toAddr, fromAddrs, suffix)
		}
	}
}

// bzToQuarantineRecord converts the given byte slice into a QuarantineRecord or returns an error.
// If the byte slice is nil or empty, a default QuarantineRecord is returned with zero coins.
func (k Keeper) bzToQuarantineRecord(bz []byte) (*quarantine.QuarantineRecord, error) {
	qf := quarantine.QuarantineRecord{
		Coins: sdk.Coins{},
	}
	if len(bz) > 0 {
		err := k.cdc.Unmarshal(bz, &qf)
		if err != nil {
			return &qf, err
		}
	}
	return &qf, nil
}

// mustBzToQuarantineRecord returns bzToQuarantineRecord but panics on error.
func (k Keeper) mustBzToQuarantineRecord(bz []byte) *quarantine.QuarantineRecord {
	qf, err := k.bzToQuarantineRecord(bz)
	if err != nil {
		panic(err)
	}
	return qf
}

// GetQuarantineRecord gets the single quarantine record to toAddr from all the fromAddrs.
// If the record doesn't exist, nil is returned.
//
// If you want all records from any of the fromAddrs, use GetQuarantineRecords.
func (k Keeper) GetQuarantineRecord(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) *quarantine.QuarantineRecord {
	store := ctx.KVStore(k.storeKey)
	key := quarantine.CreateRecordKey(toAddr, fromAddrs...)
	if store.Has(key) {
		bz := store.Get(key)
		qr := k.mustBzToQuarantineRecord(bz)
		return qr
	}
	return nil
}

// GetQuarantineRecords gets all the quarantine records to toAddr that involved any of the fromAddrs.
//
// If you want a single record from all the fromAddrs, use GetQuarantineRecord.
func (k Keeper) GetQuarantineRecords(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) []*quarantine.QuarantineRecord {
	store := ctx.KVStore(k.storeKey)
	allSuffixes := k.getQuarantineRecordSuffixes(store, toAddr, fromAddrs)
	var rv []*quarantine.QuarantineRecord
	for _, suffix := range allSuffixes {
		key := quarantine.CreateRecordKey(toAddr, suffix)
		if store.Has(key) {
			bz := store.Get(key)
			rv = append(rv, k.mustBzToQuarantineRecord(bz))
		}
	}
	return rv
}

// AddQuarantinedCoins records that some new funds have been quarantined.
func (k Keeper) AddQuarantinedCoins(ctx sdk.Context, coins sdk.Coins, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) error {
	qr := k.GetQuarantineRecord(ctx, toAddr, fromAddrs...)
	if qr != nil {
		qr.AddCoins(coins...)
	} else {
		qr = &quarantine.QuarantineRecord{
			Coins: coins,
		}
		for _, fromAddr := range fromAddrs {
			if k.IsAutoAccept(ctx, toAddr, fromAddr) {
				qr.AcceptedFromAddresses = append(qr.AcceptedFromAddresses, fromAddr)
			} else {
				qr.UnacceptedFromAddresses = append(qr.UnacceptedFromAddresses, fromAddr)
			}
		}
	}
	// Regardless of if its new or existing, set declined based on current auto-decline info.
	qr.Declined = k.IsAutoDecline(ctx, toAddr, fromAddrs...)
	k.SetQuarantineRecord(ctx, toAddr, qr)
	return ctx.EventManager().EmitTypedEvent(&quarantine.EventFundsQuarantined{
		ToAddress: toAddr.String(),
		Coins:     coins,
	})
}

// AcceptQuarantinedFunds looks up all quarantined funds to toAddr from any of the fromAddrs.
// It marks and saves each as accepted and, if fully accepted, releases (sends) the funds to toAddr.
// Returns total funds released and possibly an error.
func (k Keeper) AcceptQuarantinedFunds(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) (sdk.Coins, error) {
	fundsReleased := sdk.Coins{}
	for _, record := range k.GetQuarantineRecords(ctx, toAddr, fromAddrs...) {
		if record.AcceptFrom(fromAddrs) {
			if record.IsFullyAccepted() {
				err := k.bankKeeper.SendCoinsBypassQuarantine(ctx, k.fundsHolder, toAddr, record.Coins)
				if err != nil {
					return nil, err
				}
				fundsReleased = fundsReleased.Add(record.Coins...)

				err = ctx.EventManager().EmitTypedEvent(&quarantine.EventFundsReleased{
					ToAddress: toAddr.String(),
					Coins:     record.Coins,
				})
				if err != nil {
					return nil, err
				}
			} else {
				// update declined to false unless one of the unaccepted from addresses is set to auto-decline.
				record.Declined = k.IsAutoDecline(ctx, toAddr, record.UnacceptedFromAddresses...)
			}
			k.SetQuarantineRecord(ctx, toAddr, record)
		}
	}

	return fundsReleased, nil
}

// DeclineQuarantinedFunds marks as declined, all quarantined funds to toAddr where any fromAddr is a sender.
func (k Keeper) DeclineQuarantinedFunds(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) {
	for _, record := range k.GetQuarantineRecords(ctx, toAddr, fromAddrs...) {
		if record.DeclineFrom(fromAddrs) {
			k.SetQuarantineRecord(ctx, toAddr, record)
		}
	}
}

// getQuarantineRecordPrefixStore returns a kv store prefixed for quarantine records and the prefix used.
// If a toAddr is provided, the store is prefixed for just the given address.
// If toAddr is empty, it will be prefixed for all quarantine records.
func (k Keeper) getQuarantineRecordPrefixStore(ctx sdk.Context, toAddr sdk.AccAddress) (sdk.KVStore, []byte) {
	pre := quarantine.RecordPrefix
	if len(toAddr) > 0 {
		pre = quarantine.CreateRecordToAddrPrefix(toAddr)
	}
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre), pre
}

// IterateQuarantineRecords iterates over the quarantine records for a given recipient address,
// or if no address is provided, iterates over all quarantine records.
// The callback function should accept a to address, record suffix, and QuarantineRecord (in that order).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k Keeper) IterateQuarantineRecords(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, recordSuffix sdk.AccAddress, record *quarantine.QuarantineRecord) (stop bool)) {
	store, pre := k.getQuarantineRecordPrefixStore(ctx, toAddr)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kToAddr, kRecordSuffix := quarantine.ParseRecordKey(quarantine.MakeKey(pre, iter.Key()))
		qf := k.mustBzToQuarantineRecord(iter.Value())

		if cb(kToAddr, kRecordSuffix, qf) {
			break
		}
	}
}

// setQuarantineRecordSuffixIndex writes the provided suffix index.
// If it is nil or there are no record suffixes, the entry is instead deleted.
func (k Keeper) setQuarantineRecordSuffixIndex(store sdk.KVStore, key []byte, value *quarantine.QuarantineRecordSuffixIndex) {
	if value == nil || len(value.RecordSuffixes) == 0 {
		store.Delete(key)
	} else {
		val := k.cdc.MustMarshal(value)
		store.Set(key, val)
	}
}

// bzToQuarantineRecordSuffixIndex converts the given byte slice into a QuarantineRecordSuffixIndex or returns an error.
// If the byte slice is nil or empty, a default QuarantineRecordSuffixIndex is returned with no suffixes.
func (k Keeper) bzToQuarantineRecordSuffixIndex(bz []byte) (*quarantine.QuarantineRecordSuffixIndex, error) {
	var si quarantine.QuarantineRecordSuffixIndex
	if len(bz) > 0 {
		err := k.cdc.Unmarshal(bz, &si)
		if err != nil {
			return &si, err
		}
	}
	return &si, nil
}

// mustBzToQuarantineRecordSuffixIndex returns bzToQuarantineRecordSuffixIndex but panics on error.
func (k Keeper) mustBzToQuarantineRecordSuffixIndex(bz []byte) *quarantine.QuarantineRecordSuffixIndex {
	si, err := k.bzToQuarantineRecordSuffixIndex(bz)
	if err != nil {
		panic(err)
	}
	return si
}

// getQuarantineRecordSuffixIndex gets a quarantine record suffix entry and it's key.
func (k Keeper) getQuarantineRecordSuffixIndex(store sdk.KVStore, toAddr, fromAddr sdk.AccAddress) (*quarantine.QuarantineRecordSuffixIndex, []byte) {
	key := quarantine.CreateRecordIndexKey(toAddr, fromAddr)
	bz := store.Get(key)
	rv := k.mustBzToQuarantineRecordSuffixIndex(bz)
	return rv, key
}

// getQuarantineRecordSuffixes gets a sorted list of known record suffixes of quarantine records to toAddr
// from any of the fromAddrs. The list will not contain duplicates, but may contain suffixes that don't point to records.
func (k Keeper) getQuarantineRecordSuffixes(store sdk.KVStore, toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress) [][]byte {
	rv := &quarantine.QuarantineRecordSuffixIndex{}
	for _, fromAddr := range fromAddrs {
		suffixes, _ := k.getQuarantineRecordSuffixIndex(store, toAddr, fromAddr)
		rv.AddSuffixes(suffixes.RecordSuffixes...)
		rv.AddSuffixes(fromAddr)
	}
	rv.Simplify()
	return rv.RecordSuffixes
}

// addQuarantineRecordSuffixIndexes adds the provided suffix to all to/from suffix index entries.
func (k Keeper) addQuarantineRecordSuffixIndexes(store sdk.KVStore, toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress, suffix []byte) {
	for _, fromAddr := range fromAddrs {
		ind, key := k.getQuarantineRecordSuffixIndex(store, toAddr, fromAddr)
		ind.AddSuffixes(suffix)
		ind.Simplify(fromAddr)
		k.setQuarantineRecordSuffixIndex(store, key, ind)
	}
}

// deleteQuarantineRecordSuffixIndexes removes the provided suffix from all to/from suffix index entries and either saves
// the updated list or deletes it if it's now empty.
func (k Keeper) deleteQuarantineRecordSuffixIndexes(store sdk.KVStore, toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress, suffix []byte) {
	for _, fromAddr := range fromAddrs {
		ind, key := k.getQuarantineRecordSuffixIndex(store, toAddr, fromAddr)
		ind.Simplify(fromAddr, suffix)
		k.setQuarantineRecordSuffixIndex(store, key, ind)
	}
}
