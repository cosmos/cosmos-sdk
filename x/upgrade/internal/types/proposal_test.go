package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

type ProposalWrapper struct {
	Prop gov.Content
}

func TestContentAccessors(t *testing.T) {
	cases := map[string]struct {
		p     gov.Content
		title string
		desc  string
		typ   string
		str   string
	}{
		"upgrade": {
			p: NewSoftwareUpgradeProposal("Title", "desc", Plan{
				Name: "due_time",
				Info: "https://foo.bar",
				Time: mustParseTime("2019-07-08T11:33:55Z"),
			}),
			title: "Title",
			desc:  "desc",
			typ:   "SoftwareUpgrade",
			str:   "Software Upgrade Proposal:\n  Title:       Title\n  Description: desc\n",
		},
		"cancel": {
			p:     NewCancelSoftwareUpgradeProposal("Cancel", "bad idea"),
			title: "Cancel",
			desc:  "bad idea",
			typ:   "CancelSoftwareUpgrade",
			str:   "Cancel Software Upgrade Proposal:\n  Title:       Cancel\n  Description: bad idea\n",
		},
	}

	cdc := codec.New()
	gov.RegisterCodec(cdc)
	RegisterCodec(cdc)

	for name, tc := range cases {
		tc := tc // copy to local variable for scopelint
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.title, tc.p.GetTitle())
			assert.Equal(t, tc.desc, tc.p.GetDescription())
			assert.Equal(t, tc.typ, tc.p.ProposalType())
			assert.Equal(t, "upgrade", tc.p.ProposalRoute())
			assert.Equal(t, tc.str, tc.p.String())

			// try to encode and decode type to ensure codec works
			wrap := ProposalWrapper{tc.p}
			bz, err := cdc.MarshalBinaryBare(&wrap)
			require.NoError(t, err)
			unwrap := ProposalWrapper{}
			err = cdc.UnmarshalBinaryBare(bz, &unwrap)
			require.NoError(t, err)

			// all methods should look the same
			assert.Equal(t, tc.title, unwrap.Prop.GetTitle())
			assert.Equal(t, tc.desc, unwrap.Prop.GetDescription())
			assert.Equal(t, tc.typ, unwrap.Prop.ProposalType())
			assert.Equal(t, "upgrade", unwrap.Prop.ProposalRoute())
			assert.Equal(t, tc.str, unwrap.Prop.String())

		})

	}
}
