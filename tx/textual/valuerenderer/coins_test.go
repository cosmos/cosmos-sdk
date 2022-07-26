package valuerenderer_test

import (
	"encoding/json"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
)

type coinsTest struct {
	coins       []*basev1beta1.Coin
	metadataMap map[string]coinTestMetadata
	expRes      string
}

func (t *coinsTest) UnmarshalJSON(b []byte) error {
	a := []interface{}{&t.coins, &t.metadataMap, &t.expRes}
	return json.Unmarshal(b, &a)
}
