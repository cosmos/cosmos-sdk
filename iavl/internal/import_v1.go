package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"cosmossdk.io/log/v2"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/cosmos/iavl"
	iavldb "github.com/cosmos/iavl/db"

	storetypes "cosmossdk.io/store/types"
)

func ImportIAVLV1MultiStore(dataDir, outDir string, logger log.Logger) error {
	logger.Info("Starting import of IAVL v1 multi-store", "sourceDir", dataDir, "destDir", outDir)
	v1Db, err := dbm.NewGoLevelDB("application", dataDir, nil)
	if err != nil {
		return fmt.Errorf("failed to open source database: %w", err)
	}

	const (
		latestVersionKey = "s/latest"
		commitInfoKeyFmt = "s/%d" // s/<version>
	)
	bz, err := v1Db.Get([]byte(latestVersionKey))
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}
	var latestVersion int64
	if err := gogotypes.StdInt64Unmarshal(&latestVersion, bz); err != nil {
		panic(err)
	}

	bz, err = v1Db.Get([]byte(fmt.Sprintf(commitInfoKeyFmt, latestVersion)))
	if err != nil {
		return fmt.Errorf("failed to get commit info for latest version: %w", err)
	}

	var ci storetypes.CommitInfo
	if err := proto.Unmarshal(bz, &ci); err != nil {
		return fmt.Errorf("failed to unmarshal commit info for latest version: %w", err)
	}

	logger.Info("Successfully read latest commit info", "version", ci.Version, "timestamp", ci.Timestamp, "storeCount", len(ci.StoreInfos))

	for _, store := range ci.StoreInfos {
		importErr := importIAVLV1Store(v1Db, store.Name, outDir, logger)
		if importErr != nil {
			return fmt.Errorf("failed to import IAVL tree for store %s: %w", store.Name, importErr)
		}
	}

	// TODO save commit info

	return nil
}

func importIAVLV1Store(v1Db dbm.DB, store, multiStoreDir string, log log.Logger) error {
	treeDir := filepath.Join(multiStoreDir, "stores", fmt.Sprintf("%s.iavl", store))
	err := os.MkdirAll(treeDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to mkdir %s: %w", treeDir, err)
	}

	v1Prefix := "s/k:" + store + "/"
	v1Db = dbm.NewPrefixDB(v1Db, []byte(v1Prefix))
	tree := iavl.NewMutableTree(iavldb.NewWrapper(v1Db), 0, false, log)
	_, err = tree.Load()
	if err != nil {
		return fmt.Errorf("failed to load IAVL tree: %w", err)
	}

	version, err := tree.GetLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest version of IAVL tree: %w", err)
	}

	imTree, err := tree.GetImmutable(version)
	if err != nil {
		return fmt.Errorf("failed to get immutable tree for version %d: %w", version, err)
	}

	exporter, err := imTree.Export()
	if err != nil {
		return fmt.Errorf("failed to create exporter for version %d: %w", version, err)
	}

	importer, err := NewImporter(uint32(version), treeDir, log)
	if err != nil {
		return fmt.Errorf("failed to create importer for version %d: %w", version, err)
	}

	err = importer.importExporter(exporter)
	if err != nil {
		return fmt.Errorf("failed to import exported nodes for version %d: %w", version, err)
	}

	return importer.Finalize()
}
