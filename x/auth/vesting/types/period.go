package types

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Periods stores all vesting periods passed as part of a PeriodicVestingAccount
type Periods []Period

// String Period implements stringer interface
func (p Period) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// String Periods implements stringer interface
func (vp Periods) String() string {
	periodsListString := make([]string, len(vp))
	for _, period := range vp {
		periodsListString = append(periodsListString, period.String())
	}

	return strings.TrimSpace(fmt.Sprintf(`Vesting Periods:
		%s`, strings.Join(periodsListString, ", ")))
}
