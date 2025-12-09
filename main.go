package main

import (
	"crypto/rand"
	"encoding/hex"

	sdkmath "cosmossdk.io/math"
)

// Mock Bitcoin transaction
type MockBTCTx struct {
	TxID   string
	Amount sdkmath.Int
}

// Mock Osmosis pool
type MockOsmosisPool struct {
	BTCBalance sdkmath.Int
}

// Generate a mock Bitcoin transaction
func generateMockBTCTx(i int) MockBTCTx {
	txID := make([]byte, 32)
	_, err := rand.Read(txID)
	if err != nil {
		panic(err)
	}

	return MockBTCTx{
		TxID:   hex.EncodeToString(txID),
		Amount: sdkmath.NewInt(int64(i) * 1000),
	}
}

// Simulate Bitcoin interoperability
func simulateBitcoinInteroperability(pool *MockOsmosisPool, tx MockBTCTx) {
	pool.BTCBalance = pool.BTCBalance.Add(tx.Amount)
}
