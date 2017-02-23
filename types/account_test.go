package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilAccount(t *testing.T) {

	acc := Account{}

	//test Copy
	accCopy := acc.Copy()
	assert.True(t, &acc != accCopy, "Account Copy Error")
	assert.True(t, acc.Sequence == accCopy.Sequence)

	//test sending nils for panic
	var nilAcc *Account
	nilAcc.String()
	nilAcc.Copy()
}
