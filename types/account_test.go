package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccount(t *testing.T) {

	acc := Account{
		PubKey:   nil,
		Sequence: 0,
		Balance:  nil,
	}

	//test Copy
	accCopy := acc.Copy()
	accCopy.Sequence = 1
	t.Log(acc.Sequence)
	t.Log(accCopy.Sequence)
	assert.True(t, acc.Sequence != accCopy.Sequence, "Account Copy Error")

	//test sending nils for panic
	var nilAcc *Account
	nilAcc.String()
	nilAcc.Copy()
}
