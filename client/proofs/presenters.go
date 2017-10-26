package proofs

import (
	"encoding/hex"

	"github.com/pkg/errors"

	data "github.com/tendermint/go-wire/data"
	cmn "github.com/tendermint/tmlibs/common"
)

const Raw = "raw"

// Presenter allows us to encode queries and parse results in an app-specific way
type Presenter interface {
	MakeKey(string) ([]byte, error)
	ParseData([]byte) (interface{}, error)
}

type Presenters map[string]Presenter

// NewPresenters gives you a default raw presenter
func NewPresenters() Presenters {
	return Presenters{}
}

// Lookup tries to find a registered presenter, or the raw presenter
func (p Presenters) Lookup(app string) (Presenter, error) {
	if app == Raw {
		return RawPresenter{}, nil
	}
	res, ok := p[app]
	if !ok {
		return nil, errors.Errorf("No presenter registered for %s", app)
	}
	return res, nil
}

// Register adds this app to the lookup table to parse it
func (p Presenters) Register(app string, pres Presenter) {
	p[app] = pres
}

// BruteForce will try all regitered parsers in random order,
// before calling RawPresenter.  Use if we have no idea how to
// interpret the data (eg. decoding all tx in a block)
func (p Presenters) BruteForce(raw []byte) (interface{}, error) {
	for _, pr := range p {
		res, err := pr.ParseData(raw)
		if err == nil {
			return res, err
		}
	}
	// no luck with any of them...just go raw
	return RawPresenter{}.ParseData(raw)
}

var _ Presenter = RawPresenter{}

// RawPresenter just hex-encodes/decodes text.  Useful as default,
// or to embed in other structs for the MakeKey implementation
//
// If you set a prefix, it will be prepended to all your data
// after hex-decoding them
type RawPresenter struct {
	KeyMaker
}

// ParseData on the raw-presenter, just provides a hex-encoding of the bytes
func (p RawPresenter) ParseData(raw []byte) (interface{}, error) {
	return data.Bytes(raw), nil
}

// KeyMaker can be embedded for a basic and flexible key encoder
type KeyMaker struct {
	Prefix []byte
}

func (k KeyMaker) MakeKey(str string) ([]byte, error) {
	r, err := hex.DecodeString(cmn.StripHex(str))
	if err == nil && len(k.Prefix) > 0 {
		r = append(k.Prefix, r...)
	}
	return r, errors.WithStack(err)
}

func ParseHexKey(str string) ([]byte, error) {
	return KeyMaker{}.MakeKey(str)
}
