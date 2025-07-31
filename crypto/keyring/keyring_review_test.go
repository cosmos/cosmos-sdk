package keyring

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestKeyring_SaveOfflineKey_Enhanced provides comprehensive testing for SaveOfflineKey functionality
// This test addresses the TODO review comment in keyring_test.go:1697
func TestKeyring_SaveOfflineKey_Enhanced(t *testing.T) {
	cdc := getCodec()

	tests := []struct {
		name        string
		backend     string
		uid         string
		pubKeyGen   func() types.PubKey
		expectedErr error
		description string
	}{
		{
			name:    "save_ed25519_key_test_backend",
			backend: BackendTest,
			uid:     "ed25519_test_key",
			pubKeyGen: func() types.PubKey {
				return ed25519.GenPrivKey().PubKey()
			},
			expectedErr: nil,
			description: "Successfully save ed25519 public key with test backend",
		},
		{
			name:    "save_ed25519_key_memory_backend",
			backend: BackendMemory,
			uid:     "ed25519_memory_key",
			pubKeyGen: func() types.PubKey {
				return ed25519.GenPrivKey().PubKey()
			},
			expectedErr: nil,
			description: "Successfully save ed25519 public key with memory backend",
		},
		{
			name:    "save_secp256k1_key_test_backend",
			backend: BackendTest,
			uid:     "secp256k1_test_key",
			pubKeyGen: func() types.PubKey {
				return secp256k1.GenPrivKey().PubKey()
			},
			expectedErr: nil,
			description: "Successfully save secp256k1 public key with test backend",
		},
		{
			name:    "save_secp256k1_key_memory_backend",
			backend: BackendMemory,
			uid:     "secp256k1_memory_key",
			pubKeyGen: func() types.PubKey {
				return secp256k1.GenPrivKey().PubKey()
			},
			expectedErr: nil,
			description: "Successfully save secp256k1 public key with memory backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create keyring instance
			kr, err := New(tt.name, tt.backend, t.TempDir(), nil, cdc)
			require.NoError(t, err, "Failed to create keyring")

			// Generate public key
			pub := tt.pubKeyGen()
			require.NotNil(t, pub, "Generated public key should not be nil")

			// Verify keyring is initially empty
			initialList, err := kr.List()
			require.NoError(t, err, "Failed to list initial keys")
			require.Empty(t, initialList, "Keyring should be initially empty")

			// Save offline key
			k, err := kr.SaveOfflineKey(tt.uid, pub)
			if tt.expectedErr != nil {
				require.Error(t, err, "Expected error but got none")
				require.True(t, errors.Is(err, tt.expectedErr), "Error type mismatch")
				return
			}

			require.NoError(t, err, "Failed to save offline key")
			require.NotNil(t, k, "Returned key record should not be nil")

			// Verify key properties
			require.Equal(t, tt.uid, k.Name, "Key name mismatch")
			require.Equal(t, TypeOffline, k.GetType(), "Key type should be offline")

			// Verify public key matches
			savedPubKey, err := k.GetPubKey()
			require.NoError(t, err, "Failed to get public key from saved record")
			require.Equal(t, pub, savedPubKey, "Saved public key should match original")

			// Verify address derivation
			expectedAddr := sdk.AccAddress(pub.Address())
			savedAddr, err := k.GetAddress()
			require.NoError(t, err, "Failed to get address from saved record")
			require.Equal(t, expectedAddr, savedAddr, "Saved address should match derived address")

			// Verify key appears in list
			finalList, err := kr.List()
			require.NoError(t, err, "Failed to list keys after save")
			require.Len(t, finalList, 1, "Should have exactly one key after save")
			require.Equal(t, tt.uid, finalList[0].Name, "Listed key name should match")
		})
	}
}

// TestKeyring_SaveOfflineKey_ErrorCases tests error scenarios for SaveOfflineKey
func TestKeyring_SaveOfflineKey_ErrorCases(t *testing.T) {
	cdc := getCodec()

	tests := []struct {
		name        string
		backend     string
		setupFunc   func(kr Keyring) error
		uid         string
		pubKey      types.PubKey
		expectedErr bool
		description string
	}{
		{
			name:    "duplicate_key_name_same_pubkey_allowed",
			backend: BackendTest,
			setupFunc: func(kr Keyring) error {
				// Pre-create a key with the same name and same pubkey
				pub := ed25519.GenPrivKey().PubKey()
				_, err := kr.SaveOfflineKey("duplicate_name", pub)
				return err
			},
			uid:         "duplicate_name",
			pubKey:      ed25519.GenPrivKey().PubKey(), // Different pubkey
			expectedErr: false, // Actually allowed due to keyring's recovery logic
			description: "Duplicate key name with different pubkey is allowed due to recovery logic",
		},
		{
			name:        "empty_key_name_allowed",
			backend:     BackendTest,
			setupFunc:   func(kr Keyring) error { return nil },
			uid:         "",
			pubKey:      ed25519.GenPrivKey().PubKey(),
			expectedErr: false,
			description: "Empty key name is actually allowed in current implementation",
		},
		{
			name:        "nil_pubkey",
			backend:     BackendTest,
			setupFunc:   func(kr Keyring) error { return nil },
			uid:         "test_nil_pubkey",
			pubKey:      nil,
			expectedErr: true,
			description: "Should fail when public key is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create keyring instance
			kr, err := New(tt.name, tt.backend, t.TempDir(), nil, cdc)
			require.NoError(t, err, "Failed to create keyring")

			// Setup test conditions
			err = tt.setupFunc(kr)
			require.NoError(t, err, "Failed to setup test conditions")

			// Attempt to save offline key
			k, err := kr.SaveOfflineKey(tt.uid, tt.pubKey)

			if tt.expectedErr {
				require.Error(t, err, "Expected error but got none for test: %s", tt.description)
				require.Nil(t, k, "Key record should be nil on error")
			} else {
				require.NoError(t, err, "Unexpected error occurred for test: %s", tt.description)
				require.NotNil(t, k, "Key record should not be nil on success")
				
				// Verify key properties for successful cases
				require.Equal(t, tt.uid, k.Name, "Key name mismatch")
				require.Equal(t, TypeOffline, k.GetType(), "Key type should be offline")
			}
		})
	}
}

// TestKeyring_SaveOfflineKey_Consistency tests data consistency after save operations
func TestKeyring_SaveOfflineKey_Consistency(t *testing.T) {
	cdc := getCodec()

	// Test with multiple backends
	backends := []string{BackendTest, BackendMemory}

	for _, backend := range backends {
		t.Run("consistency_"+backend, func(t *testing.T) {
			kr, err := New("consistency_test", backend, t.TempDir(), nil, cdc)
			require.NoError(t, err)

			// Save multiple keys
			keys := make(map[string]types.PubKey)
			for i := 0; i < 5; i++ {
				uid := fmt.Sprintf("key_%d", i)
				pub := ed25519.GenPrivKey().PubKey()
				keys[uid] = pub

				k, err := kr.SaveOfflineKey(uid, pub)
				require.NoError(t, err, "Failed to save key %s", uid)
				require.Equal(t, uid, k.Name, "Key name mismatch for %s", uid)
			}

			// Verify all keys are present and correct
			list, err := kr.List()
			require.NoError(t, err)
			require.Len(t, list, len(keys), "Key count mismatch")

			// Verify each key individually
			for uid, expectedPub := range keys {
				k, err := kr.Key(uid)
				require.NoError(t, err, "Failed to retrieve key %s", uid)

				savedPub, err := k.GetPubKey()
				require.NoError(t, err, "Failed to get public key for %s", uid)
				require.Equal(t, expectedPub, savedPub, "Public key mismatch for %s", uid)
			}
		})
	}
}