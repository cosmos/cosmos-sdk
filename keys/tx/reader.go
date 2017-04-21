package tx

import (
	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-crypto/keys"
	data "github.com/tendermint/go-wire/data"
)

const (
	typeOneSig   = byte(0x01)
	typeMultiSig = byte(0x02)
	nameOneSig   = "sig"
	nameMultiSig = "multi"
)

var _ keys.Signable = Sig{}
var TxMapper data.Mapper

func init() {
	TxMapper = data.NewMapper(Sig{}).
		RegisterImplementation(&OneSig{}, nameOneSig, typeOneSig).
		RegisterImplementation(&MultiSig{}, nameMultiSig, typeMultiSig)
}

/*
DO NOT USE this interface.

It is public by necessity but should never be used directly
outside of this package.

Only use Sig, never SigInner
*/
type SigInner interface {
	SignBytes() []byte
	Sign(pubkey crypto.PubKey, sig crypto.Signature) error
	Signers() ([]crypto.PubKey, error)
}

// Sig is what is exported, and handles serialization
type Sig struct {
	SigInner
}

// TxBytes
func (s Sig) TxBytes() ([]byte, error) {
	return data.ToWire(s)
}

// WrapSig goes from concrete implementation to "interface" struct
func WrapSig(pk SigInner) Sig {
	if wrap, ok := pk.(Sig); ok {
		pk = wrap.Unwrap()
	}
	return Sig{pk}
}

// Unwrap recovers the concrete interface safely (regardless of levels of embeds)
func (p Sig) Unwrap() SigInner {
	pk := p.SigInner
	for wrap, ok := pk.(Sig); ok; wrap, ok = pk.(Sig) {
		pk = wrap.SigInner
	}
	return pk
}

func (p Sig) MarshalJSON() ([]byte, error) {
	return TxMapper.ToJSON(p.Unwrap())
}

func (p *Sig) UnmarshalJSON(data []byte) (err error) {
	parsed, err := TxMapper.FromJSON(data)
	if err == nil && parsed != nil {
		p.SigInner = parsed.(SigInner)
	}
	return
}
