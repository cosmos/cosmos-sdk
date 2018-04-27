package merkle

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/tendermint/go-wire"
	"github.com/tendermint/iavl"
	"github.com/tendermint/tmlibs/merkle"
)

func computeProofFromAunts(index int, total int, inners [][]byte) (res []Node, err error) {
	if index >= total || index < 0 || total <= 0 {
		err = fmt.Errorf("Invalid SimpleProof")
		return
	}

	switch total {
	case 0:
		err = fmt.Errorf("Cannot call computeHashFromAunts() with 0 total")
		return
	case 1:
		if len(inners) != 0 {
			err = fmt.Errorf("Inner hashes length not match")
			return
		}
		return nil, nil
	default:
		if len(inners) == 0 {
			err = fmt.Errorf("Inner hashes length not match")
			return
		}
		numLeft := (total + 1) / 2
		if index < numLeft {
			prefix := new(bytes.Buffer)
			suffix := new(bytes.Buffer)

			err = encodeUvarint(prefix, 20)
			if err != nil {
				return
			}

			err = encodeByteSlice(suffix, inners[len(inners)-1])
			if err != nil {
				return
			}

			res, err = computeProofFromAunts(index, numLeft, inners[:len(inners)-1])
			if err != nil {
				return
			}
			res = append(res, Node{Prefix: prefix.Bytes(), Suffix: suffix.Bytes(), Op: RIPEMD160})
			return

		} else {
			prefix := new(bytes.Buffer) /*
				err = amino.EncodeByteSlice(prefix, inners[len(inners)-1])
				if err != nil {
					return
				}
				err = amino.EncodeUvarint(prefix, 20) // length of RIPEMD160
				if err != nil {
					return
				}*/
			err = encodeByteSlice(prefix, inners[len(inners)-1])
			if err != nil {
				return
			}
			err = encodeUvarint(prefix, 20) // length of ripemd160
			if err != nil {
				return
			}

			res, err = computeProofFromAunts(index-numLeft, total-numLeft, inners[:len(inners)-1])
			if err != nil {
				return
			}
			res = append(res, Node{Prefix: prefix.Bytes(), Op: RIPEMD160})
			return

		}
	}
}

func FromSimpleProof(p *merkle.SimpleProof, index int, total int, root []byte) (res ExistsProof, err error) {
	data := ExistsData{
		Op: RIPEMD160,
	}

	nodes, err := computeProofFromAunts(index, total, p.Aunts)
	if err != nil {
		return
	}

	return ExistsProof{
		Data:     data,
		Nodes:    nodes,
		RootHash: root,
	}, nil
}

func encodeByteSlice(w io.Writer, bz []byte) (err error) {
	err = encodeUvarint(w, uint64(len(bz)))
	if err != nil {
		return
	}
	_, err = w.Write(bz)
	return
}

func encodeUvarint(w io.Writer, i uint64) (err error) {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], i)
	_, err = w.Write(buf[0:n])
	return
}

func FromKeyProof(p iavl.KeyProof) (KeyProof, error) {
	if p == nil {
		return nil, fmt.Errorf("Proof is empty")
	}
	switch p := p.(type) {
	case *iavl.KeyExistsProof:
		return FromKeyExistsProof(p)
	case *iavl.KeyAbsentProof:
		return FromKeyAbsentProof(p)
	default:
		return nil, fmt.Errorf("Invalid proof")
	}
}

func FromKeyExistsProof(p *iavl.KeyExistsProof) (KeyProof, error) {
	prefix := new(bytes.Buffer)
	/*err := amino.EncodeInt8(prefix, 0)
	if err == nil {
		err = amino.EncodeInt64(prefix, 1)
	}
	if err == nil {
		err = amino.EncodeInt64(prefix, p.Version)
	}*/
	n, err := int(0), error(nil)

	wire.WriteInt8(0, prefix, &n, &err)
	wire.WriteInt64(1, prefix, &n, &err)
	wire.WriteInt64(p.Version, prefix, &n, &err)

	if err != nil {
		return nil, err
	}

	data := ExistsData{
		Prefix: prefix.Bytes(),
		Op:     RIPEMD160,
	}
	path := p.PathToKey.InnerNodes

	nodes := make([]Node, len(path))
	for i, inner := range path {
		prefix := new(bytes.Buffer)
		suffix := new(bytes.Buffer)
		/*
			err := amino.EncodeInt8(prefix, inner.Height)
			if err == nil {
				err = amino.EncodeInt64(prefix, inner.Size)
			}
			if err == nil {
				err = amino.EncodeInt64(prefix, inner.Version)
			}
			if len(inner.Left) == 0 {
				if err == nil {
					err = amino.EncodeByteSlice(suffix, inner.Right)
				}
			} else {
				if err == nil {
					err = amino.EncodeByteSlice(prefix, inner.Left)
				}
			}
		*/

		n, err := int(0), error(nil)
		wire.WriteInt8(inner.Height, prefix, &n, &err)
		wire.WriteInt64(inner.Size, prefix, &n, &err)
		wire.WriteInt64(inner.Version, prefix, &n, &err)
		if len(inner.Left) == 0 {
			n = 0
			wire.WriteByteSlice(inner.Right, suffix, &n, &err)
		} else {
			wire.WriteByteSlice(inner.Left, prefix, &n, &err)
		}

		if err != nil {
			return nil, err
		}

		nodes[i] = Node{
			Prefix: prefix.Bytes(),
			Suffix: suffix.Bytes(),
			Op:     RIPEMD160,
		}
	}

	return ExistsProof{
		Data:     data,
		Nodes:    nodes,
		RootHash: p.Root(),
	}, nil
}

func FromKeyAbsentProof(p *iavl.KeyAbsentProof) (KeyProof, error) {
	return nil, nil
}

func Leaf(key []byte, value []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	n, err := int(0), error(nil)

	wire.WriteByteSlice(key, buf, &n, &err)
	wire.WriteByteSlice(value, buf, &n, &err)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func SimpleLeaf(key []byte, value merkle.Hasher) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := encodeByteSlice(buf, merkle.SimpleHashFromBytes(key))
	if err != nil {
		return nil, err
	}

	err = encodeByteSlice(buf, value.Hash())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Plain(kv merkle.KVPair) []byte {
	buf := new(bytes.Buffer)

	err := encodeByteSlice(buf, kv.Key)
	if err != nil {
		panic(err)
	}

	err = encodeByteSlice(buf, kv.Value)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}
