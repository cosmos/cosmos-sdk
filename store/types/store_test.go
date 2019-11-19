package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreUpgrades(t *testing.T) {
	type toDelete struct {
		key    string
		delete bool
	}
	type toRename struct {
		newkey string
		result string
	}

	cases := map[string]struct {
		upgrades     *StoreUpgrades
		expectDelete []toDelete
		expectRename []toRename
	}{
		"empty upgrade": {
			expectDelete: []toDelete{{"foo", false}},
			expectRename: []toRename{{"foo", ""}},
		},
		"simple matches": {
			upgrades: &StoreUpgrades{
				Deleted: []string{"foo"},
				Renamed: []StoreRename{{"bar", "baz"}},
			},
			expectDelete: []toDelete{{"foo", true}, {"bar", false}, {"baz", false}},
			expectRename: []toRename{{"foo", ""}, {"bar", ""}, {"baz", "bar"}},
		},
		"many data points": {
			upgrades: &StoreUpgrades{
				Deleted: []string{"one", "two", "three", "four", "five"},
				Renamed: []StoreRename{{"old", "new"}, {"white", "blue"}, {"black", "orange"}, {"fun", "boring"}},
			},
			expectDelete: []toDelete{{"four", true}, {"six", false}, {"baz", false}},
			expectRename: []toRename{{"white", ""}, {"blue", "white"}, {"boring", "fun"}, {"missing", ""}},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			for _, d := range tc.expectDelete {
				assert.Equal(t, tc.upgrades.IsDeleted(d.key), d.delete)
			}
			for _, r := range tc.expectRename {
				assert.Equal(t, tc.upgrades.RenamedFrom(r.newkey), r.result)
			}
		})
	}
}
