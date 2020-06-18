package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVoteUnMarshalJSON(t *testing.T) {
	tests := []struct {
		option  string
		isError bool
	}{
		{"Yes", false},
		{"No", false},
		{"Abstain", false},
		{"NoWithVeto", false},
		{"", false},
		{"misc", true},
	}
	for _, tt := range tests {
		var vo VoteOption
		data, err := json.Marshal(tt.option)
		require.NoError(t, err)

		err = vo.UnmarshalJSON(data)
		if tt.isError {
			require.Error(t, err)
			require.EqualError(t, err, fmt.Sprintf("'%s' is not a valid vote option", tt.option))
		} else {
			require.NoError(t, err)
		}
	}
}
