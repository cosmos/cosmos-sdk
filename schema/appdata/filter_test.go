package appdata

import "testing"

func TestModuleFilter(t *testing.T) {
	excludedMods := map[string]bool{"b": true, "d": true}
	filter := func(modName string) bool {
		return !excludedMods[modName]
	}

	var initialized []string
	var onKvPair []string
	var onObjectUpdate []string

	listener := Listener{
		InitializeModuleData: func(data ModuleInitializationData) error {
			if excludedMods[data.ModuleName] {
				t.Errorf("module %s should be excluded", data.ModuleName)
			}
			initialized = append(initialized, data.ModuleName)
			return nil
		},
		OnKVPair: func(data KVPairData) error {
			for _, update := range data.Updates {
				if excludedMods[update.ModuleName] {
					t.Errorf("module %s should be excluded", update.ModuleName)
				}
				onKvPair = append(onKvPair, update.ModuleName)
			}
			return nil
		},
		OnObjectUpdate: func(data ObjectUpdateData) error {
			if excludedMods[data.ModuleName] {
				t.Errorf("module %s should be excluded", data.ModuleName)
			}
			onObjectUpdate = append(onObjectUpdate, data.ModuleName)
			return nil
		},
	}

	listener = ModuleFilter(listener, filter)

	err := listener.OnKVPair(KVPairData{
		Updates: []ModuleKVPairUpdate{
			{ModuleName: "a"},
			{ModuleName: "b"},
			{ModuleName: "c"},
			{ModuleName: "d"},
		},
	})
	requireNoError(t, err)

	expectedMods := []string{"a", "c"}
	expectStrings(t, onKvPair, expectedMods)

	requireNoError(t, listener.InitializeModuleData(ModuleInitializationData{ModuleName: "a"}))
	requireNoError(t, listener.InitializeModuleData(ModuleInitializationData{ModuleName: "b"}))
	requireNoError(t, listener.InitializeModuleData(ModuleInitializationData{ModuleName: "c"}))
	requireNoError(t, listener.InitializeModuleData(ModuleInitializationData{ModuleName: "d"}))

	expectStrings(t, initialized, expectedMods)

	requireNoError(t, listener.OnObjectUpdate(ObjectUpdateData{ModuleName: "a"}))
	requireNoError(t, listener.OnObjectUpdate(ObjectUpdateData{ModuleName: "b"}))
	requireNoError(t, listener.OnObjectUpdate(ObjectUpdateData{ModuleName: "c"}))
	requireNoError(t, listener.OnObjectUpdate(ObjectUpdateData{ModuleName: "d"}))

	expectStrings(t, onObjectUpdate, expectedMods)
}

// requireNoError is a helper to avoid adding any test dependencies to this go.mod
func requireNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func expectStrings(t *testing.T, actual, expected []string) {
	if len(actual) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(actual))
	}

	for i, s := range actual {
		if s != expected[i] {
			t.Fatalf("expected %s at index %d, got %s", expected[i], i, s)
		}
	}
}
