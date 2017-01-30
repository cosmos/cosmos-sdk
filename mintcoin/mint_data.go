package mintnet

import (
	"bytes"

	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
)

type MintState struct {
	Bankers [][]byte
}

func (s *MintState) AddBanker(addr []byte) {
	if !s.IsBanker(addr) {
		s.Bankers = append(s.Bankers, addr)
	}
}

func (s *MintState) RemoveBanker(addr []byte) {
	b := s.Bankers
	for i := range b {
		if bytes.Equal(addr, b[i]) {
			s.Bankers = append(b[:i], b[i+1:]...)
			return
		}
	}
}

func (s *MintState) IsBanker(addr []byte) bool {
	for _, b := range s.Bankers {
		if bytes.Equal(b, addr) {
			return true
		}
	}
	return false
}

type MintTx struct {
	Winners []Winner
}

type Winner struct {
	Addr   []byte
	Amount types.Coins
}

func (tx MintTx) Serialize() []byte {
	return wire.BinaryBytes(tx)
}
