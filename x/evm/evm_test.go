package evm_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ethereumproject/sputnikvm-ffi/go/sputnikvm"
)

func TestEVM(t *testing.T) {
	assert.True(t, true)

	account := sputnikvm.AccountChangeStorageItem{
		Key:   big.NewInt(100),
		Value: big.NewInt(19),
	}

	_ = account

	assert.True(t, true)
}
