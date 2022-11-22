package main

import (
	"fmt"
	"os"
	"time"

	tmdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
)

// stores is the list of stores in the CosmosHub database
// FIXME would be nice to autodetect this
var stores = []string{
	"acc",
	"distribution",
	"evidence",
	"god",
	"main",
	"mint",
	"params",
	"slashing",
	"staking",
	"supply",
	"upgrade",
}

// Stats track import/export statistics
type Stats struct {
	nodes     uint64
	leafNodes uint64
	size      uint64
	duration  time.Duration
}

func (s *Stats) Add(o Stats) {
	s.nodes += o.nodes
	s.leafNodes += o.leafNodes
	s.size += o.size
	s.duration += o.duration
}

func (s *Stats) AddDurationSince(started time.Time) {
	s.duration += time.Since(started)
}

func (s *Stats) AddNode(node *iavl.ExportNode) {
	s.nodes++
	if node.Height == 0 {
		s.leafNodes++
	}
	s.size += uint64(len(node.Key) + len(node.Value) + 8 + 1)
}

func (s *Stats) String() string {
	return fmt.Sprintf("%v nodes (%v leaves) in %v with size %v MB",
		s.nodes, s.leafNodes, s.duration.Round(time.Millisecond), s.size/1024/1024)
}

// main runs the main program
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %v <dbpath>\n", os.Args[0])
		os.Exit(1)
	}
	err := run(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
		os.Exit(1)
	}
}

// run runs the command with normal error handling
func run(dbPath string) error {
	version, exports, err := runExport(dbPath)
	if err != nil {
		return err
	}

	err = runImport(version, exports)
	if err != nil {
		return err
	}
	return nil
}

// runExport runs an export benchmark and returns a map of store names/export nodes
func runExport(dbPath string) (int64, map[string][]*iavl.ExportNode, error) {
	ldb, err := tmdb.NewDB("application", tmdb.GoLevelDBBackend, dbPath)
	if err != nil {
		return 0, nil, err
	}
	tree, err := iavl.NewMutableTree(tmdb.NewPrefixDB(ldb, []byte("s/k:main/")), 0, false)
	if err != nil {
		return 0, nil, err
	}
	version, err := tree.LoadVersion(0)
	if err != nil {
		return 0, nil, err
	}
	fmt.Printf("Exporting cosmoshub database at version %v\n\n", version)

	exports := make(map[string][]*iavl.ExportNode, len(stores))

	totalStats := Stats{}
	for _, name := range stores {
		db := tmdb.NewPrefixDB(ldb, []byte("s/k:"+name+"/"))
		tree, err := iavl.NewMutableTree(db, 0, false)
		if err != nil {
			return 0, nil, err
		}

		stats := Stats{}
		export := make([]*iavl.ExportNode, 0, 100000)

		storeVersion, err := tree.LoadVersion(0)
		if err != nil {
			return 0, nil, err
		}
		if storeVersion == 0 {
			fmt.Printf("%-13v: %v\n", name, stats.String())
			continue
		}

		itree, err := tree.GetImmutable(version)
		if err != nil {
			return 0, nil, err
		}
		start := time.Now().UTC()
		exporter := itree.Export()
		defer exporter.Close()
		for {
			node, err := exporter.Next()
			if err == iavl.ErrorExportDone {
				break
			} else if err != nil {
				return 0, nil, err
			}
			export = append(export, node)
			stats.AddNode(node)
		}
		stats.AddDurationSince(start)
		fmt.Printf("%-13v: %v\n", name, stats.String())
		totalStats.Add(stats)
		exports[name] = export
	}

	fmt.Printf("\nExported %v stores with %v\n\n", len(stores), totalStats.String())

	return version, exports, nil
}

// runImport runs an import benchmark with nodes exported from runExport()
func runImport(version int64, exports map[string][]*iavl.ExportNode) error {
	fmt.Print("Importing into new LevelDB stores\n\n")

	totalStats := Stats{}

	for _, name := range stores {
		tempdir, err := os.MkdirTemp("", name)
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempdir)

		start := time.Now()
		stats := Stats{}

		newDB, err := tmdb.NewDB(name, tmdb.GoLevelDBBackend, tempdir)
		if err != nil {
			return err
		}
		newTree, err := iavl.NewMutableTree(newDB, 0, false)
		if err != nil {
			return err
		}
		importer, err := newTree.Import(version)
		if err != nil {
			return err
		}
		defer importer.Close()
		for _, node := range exports[name] {
			err = importer.Add(node)
			if err != nil {
				return err
			}
			stats.AddNode(node)
		}
		err = importer.Commit()
		if err != nil {
			return err
		}
		stats.AddDurationSince(start)
		fmt.Printf("%-12v: %v\n", name, stats.String())
		totalStats.Add(stats)
	}

	fmt.Printf("\nImported %v stores with %v\n", len(stores), totalStats.String())

	return nil
}
