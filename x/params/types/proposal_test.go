package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParameterChangeProposal(t *testing.T) {
	pc1 := NewParamChange("sub", "foo", "baz")
	pc2 := NewParamChangeWithSubkey("sub", "bar", "cat", "dog")
	pcp := NewParameterChangeProposal("test title", "test description", []ParamChange{pc1, pc2})

	require.Equal(t, "test title", pcp.GetTitle())
	require.Equal(t, "test description", pcp.GetDescription())
	require.Equal(t, RouterKey, pcp.ProposalRoute())
	require.Equal(t, ProposalTypeChange, pcp.ProposalType())
	require.Nil(t, pcp.ValidateBasic())

	pc3 := NewParamChangeWithSubkey("", "bar", "cat", "dog")
	pcp = NewParameterChangeProposal("test title", "test description", []ParamChange{pc3})
	require.Error(t, pcp.ValidateBasic())

	pc4 := NewParamChangeWithSubkey("sub", "", "cat", "dog")
	pcp = NewParameterChangeProposal("test title", "test description", []ParamChange{pc4})
	require.Error(t, pcp.ValidateBasic())

	pc5 := NewParamChangeWithSubkey("sub", "foo", "cat", "")
	pcp = NewParameterChangeProposal("test title", "test description", []ParamChange{pc5})
	require.Error(t, pcp.ValidateBasic())
}
