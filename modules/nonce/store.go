package nonce

import (
	"fmt"

	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
)

func getSeq(store state.KVStore, key []byte) (seq uint32, err error) {
	// fmt.Printf("load:  %X\n", key)
	data := store.Get(key)
	if len(data) == 0 {
		return seq, errors.ErrNoAccount()
	}
	err = wire.ReadBinaryBytes(data, &seq)
	if err != nil {
		msg := fmt.Sprintf("Error reading sequence for %X", key)
		return seq, errors.ErrInternal(msg)
	}
	return seq, nil
}

func setSeq(store state.KVStore, key []byte, seq uint32) error {
	bin := wire.BinaryBytes(seq)
	store.Set(key, bin)
	return nil // real stores can return error...
}
