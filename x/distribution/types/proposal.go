package types

import (
	"fmt"
	"strings"
)

// String implements the Stringer interface.
func (csp CommunityPoolSpendProposal) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, `Community Pool Spend Proposal:
  Title:       %s
  Description: %s
  Recipient:   %s
  Amount:      %s
`, csp.Title, csp.Description, csp.Recipient, csp.Amount)
	return b.String()
}
