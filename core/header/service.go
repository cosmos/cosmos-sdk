package header

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"
)

// Service defines the interface in which you can get header information
type Service interface {
	HeaderInfo(context.Context) Info
}

// Info defines a struct that contains information about the header
type Info struct {
	Height  int64     // Height returns the height of the block
	Hash    []byte    // Hash returns the hash of the block header
	Time    time.Time // Time returns the time of the block
	ChainID string    // ChainId returns the chain ID of the block
	AppHash []byte    // AppHash used in the current block header
}

// Bytes encodes the Info struct into a byte slice using little-endian encoding
func (i *Info) Bytes() ([]byte, error) {
	buf := make([]byte, 0)

	// Encode Height
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(i.Height))
	buf = append(buf, heightBytes...)

	// Encode Hash
	buf = append(buf, i.Hash...)

	// Encode Time
	timeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeBytes, uint64(i.Time.Unix()))
	buf = append(buf, timeBytes...)

	// Encode ChainID
	buf = append(buf, []byte(i.ChainID)...)

	// Encode AppHash
	buf = append(buf, i.AppHash...)

	return buf, nil
}

// FromBytes decodes the byte slice into an Info struct using little-endian encoding
func (i *Info) FromBytes(bytes []byte) error {
	if len(bytes) < 40 {
		return fmt.Errorf("invalid byte slice length")
	}

	// Decode Height
	i.Height = int64(binary.LittleEndian.Uint64(bytes[:8]))
	bytes = bytes[8:]

	// Decode Hash
	i.Hash = make([]byte, 32)
	copy(i.Hash, bytes[:32])
	bytes = bytes[32:]

	// Decode Time
	unixTime := int64(binary.LittleEndian.Uint64(bytes[:8]))
	i.Time = time.Unix(unixTime, 0)
	bytes = bytes[8:]

	// Decode ChainID
	chainIDLen := int(bytes[0])
	bytes = bytes[1:]
	if len(bytes) < chainIDLen {
		return fmt.Errorf("invalid byte slice length")
	}
	i.ChainID = string(bytes[:chainIDLen])
	bytes = bytes[chainIDLen:]

	// Decode AppHash
	i.AppHash = make([]byte, len(bytes))
	copy(i.AppHash, bytes)

	return nil
}
