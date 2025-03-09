package iavl

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/migration"
	dbm "cosmossdk.io/store/v2/db"
)

// mockReader implements store.Reader for testing purposes
type mockReader struct {
	mock.Mock
}

// Get implements store.Reader
func (m *mockReader) Get(version uint64, key []byte) ([]byte, error) {
	args := m.Called(version, key)
	return args.Get(0).([]byte), args.Error(1)
}

// Iterator implements store.Reader
func (m *mockReader) Iterator(version uint64, start, end []byte, ascending bool) (store.Iterator, error) {
	args := m.Called(version, start, end, ascending)
	return args.Get(0).(store.Iterator), args.Error(1)
}

func TestHistoricalQueries(t *testing.T) {
	testCases := []struct {
		name                  string
		enableHistorical      bool
		setupMock             func(*mockReader)
		expectedValue         []byte
		shouldCallMock        bool
		version              uint64
		migrationHeight      uint64
		expectError          bool
	}{
		{
			name:             "historical queries enabled - before migration",
			enableHistorical: true,
			version:         999,
			migrationHeight: 1000,
			expectedValue:   []byte("test-value"),
			shouldCallMock:  true,
			setupMock: func(m *mockReader) {
				m.On("Get", uint64(999), []byte("test-key")).Return([]byte("test-value"), nil)
			},
		},
		{
			name:             "historical queries disabled",
			enableHistorical: false,
			version:         999,
			migrationHeight: 1000,
			expectedValue:   []byte("new-value"),
			shouldCallMock:  false,
		},
		{
			name:             "after migration height",
			enableHistorical: true,
			version:         1001,
			migrationHeight: 1000,
			expectedValue:   []byte("new-value"),
			shouldCallMock:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := dbm.NewMemDB()
			cfg := DefaultConfig()
			cfg.EnableHistoricalQueries = tc.enableHistorical
			tree, err := NewIavlTree(db, nil, cfg)
			require.NoError(t, err)

			mockOldReader := new(mockReader)
			if tc.setupMock != nil {
				tc.setupMock(mockOldReader)
			}

			querier, err := NewHistoricalQuerier(tree, mockOldReader, db, cfg)
			require.NoError(t, err)

			err = migration.StoreMigrationHeight(db, tc.migrationHeight)
			require.NoError(t, err)

			key := []byte("test-key")
			err = tree.Set(key, []byte("new-value"))
			require.NoError(t, err)
			_, _, err = tree.Commit()
			require.NoError(t, err)

			value, err := querier.Get(tc.version, key)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, value)

			if tc.shouldCallMock {
				mockOldReader.AssertExpectations(t)
			} else {
				mockOldReader.AssertNotCalled(t, "Get")
			}
		})
	}
}

func TestHistoricalQueriesIterator(t *testing.T) {
	testCases := []struct {
		name             string
		enableHistorical bool
		version         uint64
		migrationHeight uint64
		shouldCallMock  bool
	}{
		{
			name:             "historical queries enabled - before migration",
			enableHistorical: true,
			version:         999,
			migrationHeight: 1000,
			shouldCallMock:  true,
		},
		{
			name:             "historical queries disabled",
			enableHistorical: false,
			version:         999,
			migrationHeight: 1000,
			shouldCallMock:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := dbm.NewMemDB()
			cfg := DefaultConfig()
			cfg.EnableHistoricalQueries = tc.enableHistorical
			tree, err := NewIavlTree(db, nil, cfg)
			require.NoError(t, err)

			mockOldReader := new(mockReader)
			if tc.shouldCallMock {
				mockIter := &mockIterator{}
				mockOldReader.On("Iterator", tc.version, []byte("start"), []byte("end"), true).Return(mockIter, nil)
			}

			querier, err := NewHistoricalQuerier(tree, mockOldReader, db, cfg)
			require.NoError(t, err)

			err = migration.StoreMigrationHeight(db, tc.migrationHeight)
			require.NoError(t, err)

			iter, err := querier.Iterator(tc.version, []byte("start"), []byte("end"), true)
			require.NoError(t, err)
			require.NotNil(t, iter)

			if tc.shouldCallMock {
				mockOldReader.AssertExpectations(t)
			} else {
				mockOldReader.AssertNotCalled(t, "Iterator")
			}
		})
	}
}

// mockIterator implements store.Iterator for testing
type mockIterator struct {
	store.Iterator
}

func TestHistoricalQueriesErrors(t *testing.T) {
	db := dbm.NewMemDB()
	cfg := DefaultConfig()
	cfg.EnableHistoricalQueries = true
	tree := NewIavlTree(db, nil, cfg)

	// Test without setting migration height
	_, err := tree.Get(1, []byte("key"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get migration height")

	_, err = tree.Iterator(1, []byte("start"), []byte("end"), true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get migration height")
}

func TestNewHistoricalQuerierValidation(t *testing.T) {
	testCases := []struct {
		name        string
		tree        *IavlTree
		oldReader   store.Reader
		db          store.KVStoreWithBatch
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil tree",
			tree:        nil,
			db:          dbm.NewMemDB(),
			expectError: true,
			errorMsg:    "tree cannot be nil",
		},
		{
			name:        "nil db",
			db:          nil,
			expectError: true,
			errorMsg:    "db cannot be nil",
		},
		{
			name:        "nil config",
			db:          dbm.NewMemDB(),
			config:      nil,
			expectError: false,
		},
		{
			name:        "valid parameters",
			db:          dbm.NewMemDB(),
			config:      DefaultConfig(),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var tree *IavlTree
			var err error
			if tc.tree == nil && !tc.expectError {
				tree, err = NewIavlTree(tc.db, nil, tc.config)
				require.NoError(t, err)
			} else {
				tree = tc.tree
			}

			querier, err := NewHistoricalQuerier(tree, tc.oldReader, tc.db, tc.config)
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
				require.Nil(t, querier)
			} else {
				require.NoError(t, err)
				require.NotNil(t, querier)
			}
		})
	}
} 
