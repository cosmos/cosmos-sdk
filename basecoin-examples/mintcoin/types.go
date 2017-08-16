package mintcoin

import (
	"bytes"

	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
)

type MintState struct {
	Issuers Issuers
}

type Issuer []byte

type Issuers []Issuer

func (s *MintState) AddIssuer(addr []byte) {
	if !s.IsIssuer(addr) {
		s.Issuers = append(s.Issuers, addr)
	}
}

func (s *MintState) RemoveIssuer(addr []byte) {
	b := s.Issuers
	for i := range b {
		if bytes.Equal(addr, b[i]) {
			s.Issuers = append(b[:i], b[i+1:]...)
			return
		}
	}
}

func (s *MintState) IsIssuer(addr []byte) bool {
	for _, b := range s.Issuers {
		if bytes.Equal(b, addr) {
			return true
		}
	}
	return false
}

type MintTx struct {
	Credits Credits
}

type Credits []Credit

type Credit struct {
	Addr   []byte
	Amount types.Coins
}

func (tx MintTx) Serialize() []byte {
	return wire.BinaryBytes(tx)
}
