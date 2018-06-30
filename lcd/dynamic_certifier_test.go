package lcd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/lcd"
	"github.com/cosmos/cosmos-sdk/lcd/errors"
)

// TestDynamicCert just makes sure it still works like StaticCert
func TestDynamicCert(t *testing.T) {
	// assert, require := assert.New(t), require.New(t)
	assert := assert.New(t)
	// require := require.New(t)

	keys := lcd.GenValKeys(4)
	// 20, 30, 40, 50 - the first 3 don't have 2/3, the last 3 do!
	vals := keys.ToValidators(20, 10)
	// and a certifier based on our known set
	chainID := "test-dyno"
	cert := lcd.NewDynamicCertifier(chainID, vals, 0)

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
			[]byte("bar"), []byte("params"), []byte("results"), tc.first, tc.last)
		err := cert.Certify(check)
		if tc.proper {
			assert.Nil(err, "%+v", err)
			assert.Equal(cert.LastHeight(), tc.height)
		} else {
			assert.NotNil(err)
			if tc.changed {
				assert.True(errors.IsValidatorsChangedErr(err), "%+v", err)
			}
		}
	}
}

// TestDynamicUpdate makes sure we update safely and sanely
func TestDynamicUpdate(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	chainID := "test-dyno-up"
	keys := lcd.GenValKeys(5)
	vals := keys.ToValidators(20, 0)
	cert := lcd.NewDynamicCertifier(chainID, vals, 40)

	// one valid block to give us a sense of time
	h := int64(100)
	good := keys.GenCommit(chainID, h, nil, vals, []byte("foo"), []byte("params"), []byte("results"), 0, len(keys))
	err := cert.Certify(good)
	require.Nil(err, "%+v", err)

	// some new sets to try later
	keys2 := keys.Extend(2)
	keys3 := keys2.Extend(4)

	// we try to update with some blocks
	cases := []struct {
		keys        lcd.ValKeys
		vals        *types.ValidatorSet
		height      int64
		first, last int  // who actually signs
		proper      bool // true -> expect no error
		changed     bool // true -> expect too much change error
	}{
		// same validator set, well signed, of course it is okay
		{keys, vals, h + 10, 0, len(keys), true, false},
		// same validator set, poorly signed, fails
		{keys, vals, h + 20, 2, len(keys), false, false},

		// shift the power a little, works if properly signed
		{keys, keys.ToValidators(10, 0), h + 30, 1, len(keys), true, false},
		// but not on a poor signature
		{keys, keys.ToValidators(10, 0), h + 40, 2, len(keys), false, false},
		// and not if it was in the past
		{keys, keys.ToValidators(10, 0), h + 25, 0, len(keys), false, false},

		// let's try to adjust to a whole new validator set (we have 5/7 of the votes)
		{keys2, keys2.ToValidators(10, 0), h + 33, 0, len(keys2), true, false},

		// properly signed but too much change, not allowed (only 7/11 validators known)
		{keys3, keys3.ToValidators(10, 0), h + 50, 0, len(keys3), false, true},
	}

	for _, tc := range cases {
		fc := tc.keys.GenFullCommit(chainID, tc.height, nil, tc.vals,
			[]byte("bar"), []byte("params"), []byte("results"), tc.first, tc.last)
		err := cert.Update(fc)
		if tc.proper {
			assert.Nil(err, "%d: %+v", tc.height, err)
			// we update last seen height
			assert.Equal(cert.LastHeight(), tc.height)
			// and we update the proper validators
			assert.EqualValues(fc.Header.ValidatorsHash, cert.Hash())
		} else {
			assert.NotNil(err, "%d", tc.height)
			// we don't update the height
			assert.NotEqual(cert.LastHeight(), tc.height)
			if tc.changed {
				assert.True(errors.IsTooMuchChangeErr(err),
					"%d: %+v", tc.height, err)
			}
		}
	}
}
