package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeWeightedVoteOptions(t *testing.T) {
	cases := map[string]struct {
		options    string
		normalized string
	}{
		"simple Yes": {
			options:    "Yes",
			normalized: "VOTE_OPTION_YES=1",
		},
		"simple yes": {
			options:    "yes",
			normalized: "VOTE_OPTION_YES=1",
		},
		"formal yes": {
			options:    "yes=1",
			normalized: "VOTE_OPTION_YES=1",
		},
		"half yes half no": {
			options:    "yes=0.5,no=0.5",
			normalized: "VOTE_OPTION_YES=0.5,VOTE_OPTION_NO=0.5",
		},
		"3 options": {
			options:    "Yes=0.5,No=0.4,NoWithVeto=0.1",
			normalized: "VOTE_OPTION_YES=0.5,VOTE_OPTION_NO=0.4,VOTE_OPTION_NO_WITH_VETO=0.1",
		},
		"zero weight option": {
			options:    "Yes=0.5,No=0.5,NoWithVeto=0",
			normalized: "VOTE_OPTION_YES=0.5,VOTE_OPTION_NO=0.5,VOTE_OPTION_NO_WITH_VETO=0",
		},
		"minus weight option": {
			options:    "Yes=0.5,No=0.6,NoWithVeto=-0.1",
			normalized: "VOTE_OPTION_YES=0.5,VOTE_OPTION_NO=0.6,VOTE_OPTION_NO_WITH_VETO=-0.1",
		},
		"empty options": {
			options:    "",
			normalized: "=1",
		},
		"not available option": {
			options:    "Yessss=1",
			normalized: "Yessss=1",
		},
	}

	for _, tc := range cases {
		normalized := NormalizeWeightedVoteOptions(tc.options)
		require.Equal(t, normalized, tc.normalized)
	}
}
