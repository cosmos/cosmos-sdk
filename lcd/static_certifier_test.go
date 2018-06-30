package lcd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/lcd"
	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
)

func TestStaticCert(t *testing.T) {
	// assert, require := assert.New(t), require.New(t)
	assert := assert.New(t)
	// require := require.New(t)

	keys := lcd.GenValKeys(4)
	// 20, 30, 40, 50 - the first 3 don't have 2/3, the last 3 do!
	vals := keys.ToValidators(20, 10)
	// and a certifier based on our known set
	chainID := "test-static"
	cert := lcd.NewStaticCertifier(chainID, vals)

	cases := []struct {
		keys        lcd.ValKeys
		vals        *types.ValidatorSet
		height      int64
		first, last int  // who actually signs
		proper      bool // true -> expect no error
		changed     bool // true -> expect validator change error
	}{
		// perfect, signed by everyone
		{keys, vals, 1, 0, len(keys), true, false},
		// skip little guy is okay
		{keys, vals, 2, 1, len(keys), true, false},
		// but not the big guy
		{keys, vals, 3, 0, len(keys) - 1, false, false},
		// even changing the power a little bit breaks the static validator
		// the sigs are enough, but the validator hash is unknown
		{keys, keys.ToValidators(20, 11), 4, 0, len(keys), false, true},
	}

	for _, tc := range cases {
		check := tc.keys.GenCommit(chainID, tc.height, nil, tc.vals,
			[]byte("foo"), []byte("params"), []byte("results"), tc.first, tc.last)
		err := cert.Certify(check)
		if tc.proper {
			assert.Nil(err, "%+v", err)
		} else {
			assert.NotNil(err)
			if tc.changed {
				assert.True(lcdErr.IsValidatorsChangedErr(err), "%+v", err)
			}
		}
	}

}
