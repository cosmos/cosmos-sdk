package types

import (
	"encoding/hex"
	"encoding/json"

	crypto "github.com/tendermint/go-crypto"
)

/***
This code is here for demo purposes, I think it belongs in go-common (HexData)
and go-crypto (JSONPubKey), but easier to review one repo, and get a thumbs up
or down first.
***/

type HexData []byte

func (h *HexData) UnmarshalJSON(b []byte) (err error) {
	// unmarshal into a string
	var s string
	err = json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	// and interpret that string as hex
	bin, err := hex.DecodeString(s)
	*h = bin
	return err
}

func (h HexData) MarshalJSON() ([]byte, error) {
	s := hex.EncodeToString(h)
	return json.Marshal(s)
}

type JSONPubKey struct {
	crypto.PubKey
}

func (p *JSONPubKey) UnmarshalJSON(b []byte) error {
	var data HexData
	err := data.UnmarshalJSON(b)
	if err != nil {
		return err
	}
	p.PubKey, err = crypto.PubKeyFromBytes(data)
	return err
}

func (p JSONPubKey) MarshalJSON() ([]byte, error) {
	var data []byte
	if p.PubKey != nil {
		data = p.PubKey.Bytes()
	}
	return HexData(data).MarshalJSON()
}
