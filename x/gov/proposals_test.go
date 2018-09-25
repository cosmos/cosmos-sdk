package gov

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProposalKind_Format(t *testing.T) {
	typeText, _ := ProposalTypeFromString("Text")
	tests := []struct {
		pt                   ProposalKind
		sprintFArgs          string
		expectedStringOutput string
	}{
		{typeText, "%s", "Text"},
		{typeText, "%v", "1"},
	}
	for _, tt := range tests {
		got := fmt.Sprintf(tt.sprintFArgs, tt.pt)
		require.Equal(t, tt.expectedStringOutput, got)
	}
}

func TestProposalStatus_Format(t *testing.T) {
	statusDepositPeriod, _ := ProposalStatusFromString("DepositPeriod")
	tests := []struct {
		pt                   ProposalStatus
		sprintFArgs          string
		expectedStringOutput string
	}{
		{statusDepositPeriod, "%s", "DepositPeriod"},
		{statusDepositPeriod, "%v", "1"},
	}
	for _, tt := range tests {
		got := fmt.Sprintf(tt.sprintFArgs, tt.pt)
		require.Equal(t, tt.expectedStringOutput, got)
	}
}
