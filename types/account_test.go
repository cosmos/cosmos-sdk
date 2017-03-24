package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilAccount(t *testing.T) {

	var acc Account

	//test Copy
	accCopy := acc.Copy()
	//note that the assert.True is used instead of assert.Equal because looking at pointers
	assert.True(t, &acc != accCopy, fmt.Sprintf("Account Copy Error, acc1: %v, acc2: %v", &acc, accCopy))
	assert.Equal(t, acc.Sequence, accCopy.Sequence)

	//test sending nils for panic
	var nilAcc *Account
	nilAcc.String()
	nilAcc.Copy()
}
