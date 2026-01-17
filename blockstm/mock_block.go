package blockstm

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	storetypes "cosmossdk.io/store/types"
)

const SigVerificationResultCacheKey = "ante:SigVerificationResult"

var (
	StoreKeyAuth = storetypes.NewKVStoreKey("acc")
	StoreKeyBank = storetypes.NewKVStoreKey("bank")
)

type Cache map[string]interface{}
type Tx func(MultiStore, Cache) error

type MockBlock struct {
	Txs     []Tx
	Results []error
}

func NewMockBlock(txs []Tx) *MockBlock {
	return &MockBlock{
		Txs:     txs,
		Results: make([]error, len(txs)),
	}
}

func (b *MockBlock) Size() int {
	return len(b.Txs)
}

func (b *MockBlock) ExecuteTx(txn TxnIndex, store MultiStore, cache Cache) {
	b.Results[txn] = b.Txs[txn](store, cache)
}

// Simulated transaction logic for tests and benchmarks

// NoopTx verifies a signature and increases the nonce of the sender
func NoopTx(i int, sender string) Tx {
	verifySig := genRandomSignature()
	return func(store MultiStore, cache Cache) error {
		verifySig(cache)
		return increaseNonce(i, sender, store.GetKVStore(StoreKeyAuth))
	}
}

func BankTransferTx(i int, sender, receiver string, amount uint64) Tx {
	base := NoopTx(i, sender)
	return func(store MultiStore, cache Cache) error {
		if err := base(store, cache); err != nil {
			return err
		}

		return bankTransfer(i, sender, receiver, amount, store.GetKVStore(StoreKeyBank))
	}
}

func IterateTx(i int, sender, receiver string, amount uint64) Tx {
	base := BankTransferTx(i, sender, receiver, amount)
	return func(store MultiStore, cache Cache) error {
		if err := base(store, cache); err != nil {
			return err
		}

		// find a nearby account, do a bank transfer
		accStore := store.GetKVStore(StoreKeyAuth)

		{
			it := accStore.Iterator([]byte("nonce"+sender), nil)
			defer it.Close()

			var j int
			for ; it.Valid(); it.Next() {
				j++
				if j > 5 {
					recipient := strings.TrimPrefix(string(it.Key()), "nonce")
					return bankTransfer(i, sender, recipient, amount, store.GetKVStore(StoreKeyBank))
				}
			}
		}

		{
			it := accStore.ReverseIterator([]byte("nonce"), []byte("nonce"+sender))
			defer it.Close()

			var j int
			for ; it.Valid(); it.Next() {
				j++
				if j > 5 {
					recipient := strings.TrimPrefix(string(it.Key()), "nonce")
					return bankTransfer(i, sender, recipient, amount, store.GetKVStore(StoreKeyBank))
				}
			}
		}

		return nil
	}
}

func genRandomSignature() func(Cache) {
	privKey := secp256k1.GenPrivKey()
	signBytes := make([]byte, 1024)
	if _, err := cryptorand.Read(signBytes); err != nil {
		panic(err)
	}
	sig, _ := privKey.Sign(signBytes)
	pubKey := privKey.PubKey()

	return func(cache Cache) {
		if cache != nil {
			if _, ok := cache[SigVerificationResultCacheKey]; ok {
				return
			}
		}
		pubKey.VerifySignature(signBytes, sig)
		if cache != nil {
			cache[SigVerificationResultCacheKey] = struct{}{}
		}
	}
}

func increaseNonce(i int, sender string, store storetypes.KVStore) error {
	nonceKey := []byte("nonce" + sender)
	var nonce uint64
	v := store.Get(nonceKey)
	if v != nil {
		nonce = binary.BigEndian.Uint64(v)
	}

	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], nonce+1)
	store.Set(nonceKey, bz[:])

	v = store.Get(nonceKey)
	if binary.BigEndian.Uint64(v) != nonce+1 {
		return fmt.Errorf("nonce not incremented: %d", binary.BigEndian.Uint64(v))
	}

	return nil
}

func bankTransfer(i int, sender, receiver string, amount uint64, store storetypes.KVStore) error {
	senderKey := []byte("balance" + sender)
	receiverKey := []byte("balance" + receiver)

	var senderBalance, receiverBalance uint64
	v := store.Get(senderKey)
	if v != nil {
		senderBalance = binary.BigEndian.Uint64(v)
	}

	v = store.Get(receiverKey)
	if v != nil {
		receiverBalance = binary.BigEndian.Uint64(v)
	}

	if senderBalance >= amount {
		// avoid the failure
		senderBalance -= amount
	}

	receiverBalance += amount

	var bz1, bz2 [8]byte
	binary.BigEndian.PutUint64(bz1[:], senderBalance)
	store.Set(senderKey, bz1[:])

	binary.BigEndian.PutUint64(bz2[:], receiverBalance)
	store.Set(receiverKey, bz2[:])

	return nil
}
