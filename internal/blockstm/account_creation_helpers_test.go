package blockstm

import (
	"encoding/binary"
	"fmt"

	storetypes "cosmossdk.io/store/types"
)

var accountCreationSeqKey = []byte("global_account_number")

type accountCreationOpts struct {
	cacheWrap       bool
	panicOnConflict bool
}

// accountCreationTx models x/auth account creation with IndexedMap ordering:
// unique index check/write is done before writing account-by-address data.
func accountCreationTx(txIdx int, opts accountCreationOpts) Tx {
	return func(store MultiStore, _ Cache) error {
		authStore := store.GetKVStore(StoreKeyAuth)
		var cacheWrap storetypes.CacheWrap
		if opts.cacheWrap {
			cacheWrap = authStore.CacheWrap()
			authStore = cacheWrap.(storetypes.KVStore)
		}

		// NextAccountNumber: read-modify-write the global sequence.
		var n uint64
		if v := authStore.Get(accountCreationSeqKey); v != nil {
			n = binary.BigEndian.Uint64(v)
		}
		var seqBz [8]byte
		binary.BigEndian.PutUint64(seqBz[:], n+1)
		authStore.Set(accountCreationSeqKey, seqBz[:])

		addrKey := []byte(fmt.Sprintf("addr_%04d", txIdx))
		var numBz [8]byte
		binary.BigEndian.PutUint64(numBz[:], n)

		// IndexedMap.Set order: reference unique index first, then set primary map.
		uniqueKey := append([]byte("unique_acct_num:"), numBz[:]...)
		if authStore.Has(uniqueKey) {
			if opts.panicOnConflict {
				panic(fmt.Sprintf("account number uniqueness violation: %d (tx %d)", n, txIdx))
			}
			return fmt.Errorf("account number uniqueness violation: %d (tx %d)", n, txIdx)
		}
		authStore.Set(uniqueKey, addrKey)

		accountKey := append([]byte("account:"), addrKey...)
		authStore.Set(accountKey, numBz[:])

		if cacheWrap != nil {
			cacheWrap.Write()
		}
		return nil
	}
}

func accountCreationUniqueKey(num uint64) []byte {
	var numBz [8]byte
	binary.BigEndian.PutUint64(numBz[:], num)
	return append([]byte("unique_acct_num:"), numBz[:]...)
}

func accountCreationFinalSeq(authStore storetypes.KVStore) uint64 {
	if v := authStore.Get(accountCreationSeqKey); v != nil {
		return binary.BigEndian.Uint64(v)
	}
	return 0
}

func accountCreationMissingUniqueNumbers(authStore storetypes.KVStore, total int) []uint64 {
	var missing []uint64
	for n := uint64(0); n < uint64(total); n++ {
		if !authStore.Has(accountCreationUniqueKey(n)) {
			missing = append(missing, n)
		}
	}
	return missing
}
