package valuerenderer_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
	"github.com/stretchr/testify/require"
)

func TestIntegers(t *testing.T) {
	vr := valuerenderer.DefaultValueRenderer{}

	type integerTest []string
	var testcases []integerTest
	raw, err := ioutil.ReadFile("./testdata/integers.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		output, err := vr.Format(context.Background(), tc[0])
		require.NoError(t, err)

		require.Equal(t, tc[1], output[0])
	}

	require.False(t, true)
}
