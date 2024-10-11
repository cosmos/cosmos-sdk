package runtime

import (
	"testing"

	corestore "cosmossdk.io/core/store"
)

func TestCheckStoreUpgrade(t *testing.T) {
	tests := []struct {
		name          string
		storeUpgrades *corestore.StoreUpgrades
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "Nil StoreUpgrades",
			storeUpgrades: nil,
			wantErr:       true,
			errMsg:        "store upgrades cannot be nil",
		},
		{
			name: "Valid StoreUpgrades",
			storeUpgrades: &corestore.StoreUpgrades{
				Added:   []string{"store1", "store2"},
				Deleted: []string{"store3", "store4"},
			},
			wantErr: false,
		},
		{
			name: "Duplicate key in Added",
			storeUpgrades: &corestore.StoreUpgrades{
				Added:   []string{"store1", "store2", "store1"},
				Deleted: []string{"store3"},
			},
			wantErr: true,
			errMsg:  "store upgrade has duplicate key store1 in added",
		},
		{
			name: "Duplicate key in Deleted",
			storeUpgrades: &corestore.StoreUpgrades{
				Added:   []string{"store1"},
				Deleted: []string{"store2", "store3", "store2"},
			},
			wantErr: true,
			errMsg:  "store upgrade has duplicate key store2 in deleted",
		},
		{
			name: "Key in both Added and Deleted",
			storeUpgrades: &corestore.StoreUpgrades{
				Added:   []string{"store1", "store2"},
				Deleted: []string{"store2", "store3"},
			},
			wantErr: true,
			errMsg:  "store upgrade has key store2 in both added and deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkStoreUpgrade(tt.storeUpgrades)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkStoreUpgrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("checkStoreUpgrade() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}
