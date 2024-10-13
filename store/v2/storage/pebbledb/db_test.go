package pebbledb

import (
	"testing"

	"github.com/stretchr/testify/suite"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/storage"
)

func TestStorageTestSuite(t *testing.T) {
	s := &storage.StorageTestSuite{
		NewDB: func(dir string) (*storage.StorageStore, error) {
			db, err := New(dir)
			if err == nil && db != nil {
				// We set sync=false just to speed up CI tests. Operators should take
				// careful consideration when setting this value in production environments.
				db.SetSync(false)
			}

			return storage.NewStorageStore(db, coretesting.NewNopLogger()), err
		},
		EmptyBatchSize: 12,
	}

	suite.Run(t, s)
}

// TestVersionExists tests the VersionExists method of the Database struct.
func TestVersionExists(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		setup          func(t *testing.T, db *Database)
		version        uint64
		expectedExists bool
		expectError    bool
	}{
		{
			name: "Fresh database: version 0 exists",
			setup: func(t *testing.T, db *Database) {
				t.Helper()
				// No setup needed for fresh database
			},
			version:        0,
			expectedExists: true,
			expectError:    false,
		},
		{
			name: "Fresh database: version 1 exists",
			setup: func(t *testing.T, db *Database) {
				t.Helper()
				// No setup needed for fresh database
			},
			version:        1,
			expectedExists: true,
			expectError:    false,
		},
		{
			name: "After setting latest version to 10, version 5 exists",
			setup: func(t *testing.T, db *Database) {
				t.Helper()
				err := db.SetLatestVersion(10)
				if err != nil {
					t.Fatalf("Setting latest version should not error: %v", err)
				}
			},
			version:        5,
			expectedExists: true, // Since pruning hasn't occurred, earliestVersion is still 0
			expectError:    false,
		},
		{
			name: "After setting latest version to 10 and pruning to 5, version 4 does not exist",
			setup: func(t *testing.T, db *Database) {
				t.Helper()
				err := db.SetLatestVersion(10)
				if err != nil {
					t.Fatalf("Setting latest version should not error: %v", err)
				}

				err = db.Prune(5)
				if err != nil {
					t.Fatalf("Pruning to version 5 should not error: %v", err)
				}
			},
			version:        4,
			expectedExists: false,
			expectError:    false,
		},
		{
			name: "After setting latest version to 10 and pruning to 5, version 5 does not exist",
			setup: func(t *testing.T, db *Database) {
				t.Helper()
				err := db.SetLatestVersion(10)
				if err != nil {
					t.Fatalf("Setting latest version should not error: %v", err)
				}

				err = db.Prune(5)
				if err != nil {
					t.Fatalf("Pruning to version 5 should not error: %v", err)
				}
			},
			version:        5,
			expectedExists: false,
			expectError:    false,
		},
		{
			name: "After setting latest version to 10 and pruning to 5, version 6 exists",
			setup: func(t *testing.T, db *Database) {
				t.Helper()
				err := db.SetLatestVersion(10)
				if err != nil {
					t.Fatalf("Setting latest version should not error: %v", err)
				}

				err = db.Prune(5)
				if err != nil {
					t.Fatalf("Pruning to version 5 should not error: %v", err)
				}
			},
			version:        6,
			expectedExists: true,
			expectError:    false,
		},
		{
			name: "After pruning to 0, all versions >=1 exist",
			setup: func(t *testing.T, db *Database) {
				t.Helper()
				// Prune to version 0
				err := db.Prune(0)
				if err != nil {
					t.Fatalf("Pruning to version 0 should not error: %v", err)
				}
			},
			version:        0,
			expectedExists: false,
			expectError:    false,
		},
		{
			name: "After pruning to 0, version 1 exists",
			setup: func(t *testing.T, db *Database) {
				t.Helper()
				// Prune to version 0
				err := db.Prune(0)
				if err != nil {
					t.Fatalf("Pruning to version 0 should not error: %v", err)
				}
			},
			version:        1,
			expectedExists: true,
			expectError:    false,
		},
	}

	// Iterate over each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary directory for the test database
			tempDir := t.TempDir()

			// Initialize the database
			db, err := New(tempDir)
			if err != nil {
				t.Fatalf("Initializing the database should not error: %v", err)
			}
			defer func() {
				err := db.Close()
				if err != nil {
					t.Fatalf("Closing the database should not error: %v", err)
				}
			}()

			// Setup the database state as per the test case
			tc.setup(t, db)

			// Call VersionExists with the specified version
			exists, err := db.VersionExists(tc.version)

			// Assert based on expectation
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got one: %v", err)
				}
				if exists != tc.expectedExists {
					t.Errorf("Version existence mismatch: expected %v, got %v", tc.expectedExists, exists)
				}
			}
		})
	}
}
