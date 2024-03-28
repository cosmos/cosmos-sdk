package encoding

import (
	"bytes"
	"fmt"

	corestore "cosmossdk.io/core/store"
)

// encodedSize returns the size of the encoded Changeset.
func encodedSize(cs *corestore.Changeset) int {
	size := EncodeUvarintSize(uint64(len(cs.Changes)))
	for _, changes := range cs.Changes {
		size += EncodeBytesSize(changes.Actor)
		size += EncodeUvarintSize(uint64(len(changes.StateChanges)))
		for _, pair := range changes.StateChanges {
			size += EncodeBytesSize(pair.Key)
			size += EncodeUvarintSize(1) // pair.Remove
			if !pair.Remove {
				size += EncodeBytesSize(pair.Value)
			}
		}
	}
	return size
}

// MarshalChangeset returns the encoded byte representation of Changeset.
// NOTE: The Changeset is encoded as follows:
// - number of store keys (uvarint)
// - for each store key:
// -- store key (bytes)
// -- number of pairs (uvarint)
// -- for each pair:
// --- key (bytes)
// --- remove (1 byte)
// --- value (bytes)
func MarshalChangeset(cs *corestore.Changeset) ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(encodedSize(cs))

	if err := EncodeUvarint(&buf, uint64(len(cs.Changes))); err != nil {
		return nil, err
	}
	for _, changes := range cs.Changes {
		if err := EncodeBytes(&buf, changes.Actor); err != nil {
			return nil, err
		}
		if err := EncodeUvarint(&buf, uint64(len(changes.StateChanges))); err != nil {
			return nil, err
		}
		for _, pair := range changes.StateChanges {
			if err := EncodeBytes(&buf, pair.Key); err != nil {
				return nil, err
			}
			if pair.Remove {
				if err := EncodeUvarint(&buf, 1); err != nil {
					return nil, err
				}
			} else {
				if err := EncodeUvarint(&buf, 0); err != nil {
					return nil, err
				}
				if err := EncodeBytes(&buf, pair.Value); err != nil {
					return nil, err
				}
			}
		}
	}

	return buf.Bytes(), nil
}

// UnmarshalChangeset decodes the Changeset from the given byte slice.
func UnmarshalChangeset(cs *corestore.Changeset, buf []byte) error {
	storeCount, n, err := DecodeUvarint(buf)
	if err != nil {
		return err
	}
	buf = buf[n:]
	changes := make([]corestore.StateChanges, storeCount)
	for i := uint64(0); i < storeCount; i++ {
		storeKey, n, err := DecodeBytes(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]

		pairCount, n, err := DecodeUvarint(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]

		pairs := make([]corestore.KVPair, pairCount)
		for j := uint64(0); j < pairCount; j++ {
			pairs[j].Key, n, err = DecodeBytes(buf)
			if err != nil {
				return err
			}
			buf = buf[n:]

			remove, n, err := DecodeUvarint(buf)
			if err != nil {
				return err
			}
			buf = buf[n:]
			if remove == 0 {
				pairs[j].Remove = false
				pairs[j].Value, n, err = DecodeBytes(buf)
				if err != nil {
					return err
				}
				buf = buf[n:]
			} else if remove == 1 {
				pairs[j].Remove = true
			} else {
				return fmt.Errorf("invalid remove flag: %d", remove)
			}
		}
		changes[i] = corestore.StateChanges{Actor: storeKey, StateChanges: pairs}
	}
	cs.Changes = changes

	return nil
}
