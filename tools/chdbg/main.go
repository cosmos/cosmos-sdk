package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
)

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "chdbg: exactly two directories must be specified\n")
		os.Exit(2)
	}
	if err := diff(flag.Arg(0), flag.Arg(1)); err != nil {
		fmt.Fprintf(os.Stderr, "chdbg: %v\n", err)
		os.Exit(2)
	}
}

func diff(dbdir1, dbdir2 string) error {
	trees := make([]*iavl.MutableTree, 2)
	for i, dir := range []string{dbdir1, dbdir2} {
		db, err := openDB(dir)
		if err != nil {
			return err
		}
		defer db.Close()
		// Match iavlviewer.
		const cacheSize = 10000
		tree, err := iavl.NewMutableTree(db, cacheSize, false)
		if err != nil {
			return fmt.Errorf("%s: %w", dir, err)
		}
		if _, err := tree.LoadVersion(0); err != nil {
			return fmt.Errorf("%s: %w", dir, err)
		}
		trees[i] = tree
	}
	t1, t2 := trees[0], trees[1]
	versions1, versions2 := t1.AvailableVersions(), t2.AvailableVersions()
	if len(versions1) == 0 {
		return fmt.Errorf("%s is empty", dbdir1)
	}
	if len(versions2) == 0 {
		return fmt.Errorf("%s is empty", dbdir2)
	}
	min := versions1[0]
	if min2 := versions2[0]; min2 > min {
		min = min2
	}
	max := versions1[len(versions1)-1]
	if max2 := versions2[len(versions2)-1]; max2 < max {
		max = max2
	}
	for ver := min; ver <= max; ver++ {
		imt1, err := t1.GetImmutable(int64(ver))
		if err != nil {
			return err
		}
		it1, err := imt1.Iterator(nil, nil, true)
		if err != nil {
			return err
		}
		imt2, err := t2.GetImmutable(int64(ver))
		if err != nil {
			return err
		}
		it2, err := imt2.Iterator(nil, nil, true)
		if err != nil {
			return err
		}
		h1, err := imt1.Hash()
		if err != nil {
			return err
		}
		h2, err := imt2.Hash()
		if err != nil {
			return err
		}
		diff := !bytes.Equal(h1, h2)
		if diff {
			fmt.Printf("chdbg: hash mismatch: %X != %X\n", h1, h2)
		}
		reports := 0
		reportf := func(format string, args ...any) {
			reports++
			const max = 10
			if reports > max {
				if reports == max+1 {
					fmt.Fprintf(os.Stderr, "chdgb: ... (additional diffs omitted)\n")
				}
				return
			}
			fmt.Fprintf(os.Stderr, "chdbg: "+format, args...)
		}
	loop:
		for {
			switch {
			case it1.Valid() && it2.Valid():
				k1, v1 := it1.Key(), it1.Value()
				k2, v2 := it2.Key(), it2.Value()
				cmp := bytes.Compare(k1, k2)
				switch cmp {
				case -1:
					it1.Next()
					reportf("%s: missing key %s\n", dbdir2, parseWeaveKey(k1))
				case +1:
					it2.Next()
					reportf("%s: missing key %s\n", dbdir1, parseWeaveKey(k2))
				default:
					it1.Next()
					it2.Next()
					eq := bytes.Equal(v1, v2)
					if !eq {
						reportf("key %s: value mismatch\n", parseWeaveKey(k1))
					} else {
						proof1, err := imt1.GetProof(k1)
						if err != nil {
							return err
						}
						ok, err := imt2.VerifyProof(proof1, k1)
						if err != nil {
							return err
						}
						if !ok {
							reportf("key %s: key proofs differ\n", parseWeaveKey(k1))
						}
					}
				}
			case it1.Valid():
				it1.Next()
				k1 := it1.Key()
				reportf("%s: missing key %s\n", dbdir2, parseWeaveKey(k1))
			case it2.Valid():
				it2.Next()
				k2 := it2.Key()
				reportf("%s: missing key %s\n", dbdir1, parseWeaveKey(k2))
			default:
				break loop
			}
		}
		it1.Close()
		it2.Close()
		if diff {
			return fmt.Errorf("database mismatch at version %d with %d differences", ver, reports)
		}
	}
	return nil
}

// parseWeaveKey assumes a separating : where all in front should be ascii,
// and all afterwards may be ascii or binary
func parseWeaveKey(key []byte) string {
	cut := bytes.IndexRune(key, ':')
	if cut == -1 {
		return encodeID(key)
	}
	prefix := key[:cut]
	id := key[cut+1:]
	return fmt.Sprintf("%s:%s", encodeID(prefix), encodeID(id))
}

// casts to a string if it is printable ascii, hex-encodes otherwise
func encodeID(id []byte) string {
	for _, b := range id {
		if b < 0x20 || b >= 0x80 {
			return strings.ToUpper(hex.EncodeToString(id))
		}
	}
	return string(id)
}

func openDB(dir string) (dbm.DB, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(dir, ".db") {
		return nil, fmt.Errorf("database directory %q must end with .db", dir)
	}
	dir = dir[:len(dir)-len(".db")]

	name := filepath.Base(dir)
	parent := filepath.Dir(dir)
	return dbm.NewGoLevelDB(name, parent, nil)
}
