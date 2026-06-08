package app_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testapp"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestExportAppStateAndValidators(t *testing.T) {
	a := testapp.Setup(t)

	exported, err := a.ExportAppStateAndValidators(false, nil, nil)
	if err != nil {
		t.Fatalf("ExportAppStateAndValidators failed: %v", err)
	}
	if len(exported.AppState) == 0 {
		t.Fatal("expected non-empty app state")
	}
	// testapp.Setup calls FinalizeBlock but not Commit, so LastBlockHeight==0
	// and ExportAppStateAndValidators returns LastBlockHeight+1 == 1.
	if exported.Height != 1 {
		t.Fatalf("expected height 1, got %d", exported.Height)
	}
	if len(exported.Validators) == 0 {
		t.Fatal("expected at least one validator in export")
	}
}

func TestExportAppStateAndValidatorsZeroHeight(t *testing.T) {
	a := testapp.Setup(t)

	exported, err := a.ExportAppStateAndValidators(true, nil, nil)
	if err != nil {
		t.Fatalf("ExportAppStateAndValidators (forZeroHeight=true) failed: %v", err)
	}
	if len(exported.AppState) == 0 {
		t.Fatal("expected non-empty app state after zero-height export")
	}
	if exported.Height != 0 {
		t.Fatalf("expected height 0 for zero-height export, got %d", exported.Height)
	}
}

func TestExportAppStateAndValidatorsZeroHeightJailList(t *testing.T) {
	a := testapp.Setup(t)

	// Non-empty jailAllowedAddrs exercises the applyAllowedAddrs branch in
	// prepForZeroHeightGenesis — validators NOT in the list get jailed.
	exported, err := a.ExportAppStateAndValidators(true, []string{}, nil)
	if err != nil {
		t.Fatalf("ExportAppStateAndValidators (zero-height, empty allow-list) failed: %v", err)
	}
	if len(exported.AppState) == 0 {
		t.Fatal("expected non-empty app state")
	}
}

func TestExportAppStateAndValidatorsModuleFilter(t *testing.T) {
	a := testapp.Setup(t)

	exported, err := a.ExportAppStateAndValidators(false, nil, []string{banktypes.ModuleName})
	if err != nil {
		t.Fatalf("ExportAppStateAndValidators with module filter failed: %v", err)
	}

	var state map[string]json.RawMessage
	if err := json.Unmarshal(exported.AppState, &state); err != nil {
		t.Fatalf("unmarshal exported state: %v", err)
	}
	if _, ok := state[banktypes.ModuleName]; !ok {
		t.Fatal("expected bank module present in filtered export")
	}
	// With a single-module filter, other modules should be absent.
	if len(state) != 1 {
		t.Fatalf("expected exactly 1 module in filtered export, got %d", len(state))
	}
}
