package internal

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"sync"
)

func (node *MemNode) ComputeHash(scheduler HashScheduler) ([]byte, error) {
	if node.hash != nil {
		return node.hash, nil
	}

	if node.IsLeaf() {
		nodeHash, err := computeHash(node, nil, nil)
		if err != nil {
			return nil, err
		}
		node.hash = nodeHash
		return nodeHash, nil
	}

	var leftHash, rightHash []byte
	var err error
	leftNode, leftPin, err := node.Left().Resolve()
	defer leftPin.Unpin()
	if err != nil {
		return nil, fmt.Errorf("resolving left child: %w", err)
	}

	rightNode, rightPin, err := node.Right().Resolve()
	defer rightPin.Unpin()
	if err != nil {
		return nil, fmt.Errorf("resolving right child: %w", err)
	}

	leftMem, leftIsMem := leftNode.(*MemNode)
	rightMem, rightIsMem := rightNode.(*MemNode)

	if leftIsMem && rightIsMem {
		leftHash, rightHash, err = scheduler.ComputeHashes(leftMem, rightMem)
		if err != nil {
			return nil, fmt.Errorf("computing child hashes via scheduler: %w", err)
		}
	} else {
		if leftIsMem {
			leftHash, err = leftMem.ComputeHash(scheduler)
			if err != nil {
				return nil, fmt.Errorf("computing left child hash: %w", err)
			}
		} else {
			leftHash = leftNode.Hash().UnsafeBytes()
		}

		if rightIsMem {
			rightHash, err = rightMem.ComputeHash(scheduler)
			if err != nil {
				return nil, fmt.Errorf("computing right child hash: %w", err)
			}
		} else {
			rightHash = rightNode.Hash().UnsafeBytes()
		}
	}

	nodeHash, err := computeHash(node, leftHash, rightHash)
	if err != nil {
		return nil, err
	}
	node.hash = nodeHash
	return nodeHash, nil
}

var hasherPool = sync.Pool{
	New: func() any {
		return sha256.New()
	},
}

func putBackHasher(h hash.Hash) {
	h.Reset()
	hasherPool.Put(h)
}

func computeHash(node Node, leftHash, rightHash []byte) ([]byte, error) {
	hasher := hasherPool.Get().(hash.Hash)
	defer putBackHasher(hasher)
	if err := writeHashBytes(node, leftHash, rightHash, hasher); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func shaSum256(bz []byte) []byte {
	hasher := hasherPool.Get().(hash.Hash)
	defer putBackHasher(hasher)
	hasher.Write(bz)
	var sum [sha256.Size]byte
	hasher.Sum(sum[:0])
	return sum[:]
}

// Writes the node's hash to the given `io.Writer`. This function recursively calls
// children to update hashes.
func writeHashBytes(node Node, leftHash, rightHash []byte, w io.Writer) error {
	var (
		n   int
		buf [binary.MaxVarintLen64]byte
	)

	n = binary.PutVarint(buf[:], int64(node.Height()))
	if _, err := w.Write(buf[0:n]); err != nil {
		return fmt.Errorf("writing height, %w", err)
	}
	n = binary.PutVarint(buf[:], node.Size())
	if _, err := w.Write(buf[0:n]); err != nil {
		return fmt.Errorf("writing size, %w", err)
	}
	n = binary.PutVarint(buf[:], int64(node.Version()))
	if _, err := w.Write(buf[0:n]); err != nil {
		return fmt.Errorf("writing layer, %w", err)
	}

	// Key is not written for inner nodes, unlike writeBytes.

	if node.IsLeaf() {
		key, err := node.Key()
		if err != nil {
			return fmt.Errorf("getting key, %w", err)
		}

		if err := encodeVarintPrefixedBytes(w, key.UnsafeBytes()); err != nil {
			return fmt.Errorf("writing key, %w", err)
		}

		value, err := node.Value()
		if err != nil {
			return fmt.Errorf("getting value, %w", err)
		}

		// Indirection needed to provide proofs without values.
		// (e.g. ProofLeafNode.ValueHash)
		if err := encodeVarintPrefixedBytes(w, shaSum256(value.UnsafeBytes())); err != nil {
			return fmt.Errorf("writing value, %w", err)
		}
	} else {
		if err := encodeVarintPrefixedBytes(w, leftHash); err != nil {
			return fmt.Errorf("writing left hash, %w", err)
		}
		if err := encodeVarintPrefixedBytes(w, rightHash); err != nil {
			return fmt.Errorf("writing right hash, %w", err)
		}
	}

	return nil
}

// encodeVarintPrefixedBytes writes a varint length-prefixed byte slice to the writer,
// it's used for hash computation, must be compactible with the official IAVL implementation.
func encodeVarintPrefixedBytes(w io.Writer, bz []byte) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], uint64(len(bz)))
	if _, err := w.Write(buf[0:n]); err != nil {
		return err
	}
	_, err := w.Write(bz)
	return err
}
