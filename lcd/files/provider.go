/*
Package files defines a Provider that stores all data in the filesystem

We assume the same validator hash may be reused by many different
headers/Commits, and thus store it separately. This leaves us
with three issues:

  1. Given a validator hash, retrieve the validator set if previously stored
  2. Given a block height, find the Commit with the highest height <= h
  3. Given a FullCommit, store it quickly to satisfy 1 and 2

Note that we do not worry about caching, as that can be achieved by
pairing this with a MemStoreProvider and CacheProvider from certifiers
*/
package files

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"

	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
	"github.com/cosmos/cosmos-sdk/lcd"
)

// nolint
const (
	Ext      = ".tsd"
	ValDir   = "validators"
	CheckDir = "checkpoints"
	dirPerm  = os.FileMode(0755)
	//filePerm = os.FileMode(0644)
)

type provider struct {
	valDir   string
	checkDir string
}

// NewProvider creates the parent dir and subdirs
// for validators and checkpoints as needed
func NewProvider(dir string) lcd.Provider {
	valDir := filepath.Join(dir, ValDir)
	checkDir := filepath.Join(dir, CheckDir)
	for _, d := range []string{valDir, checkDir} {
		err := os.MkdirAll(d, dirPerm)
		if err != nil {
			panic(err)
		}
	}
	return &provider{valDir: valDir, checkDir: checkDir}
}

func (p *provider) encodeHash(hash []byte) string {
	return hex.EncodeToString(hash) + Ext
}

func (p *provider) encodeHeight(h int64) string {
	// pad up to 10^12 for height...
	return fmt.Sprintf("%012d%s", h, Ext)
}

// StoreCommit saves a full commit after it has been verified.
func (p *provider) StoreCommit(fc lcd.FullCommit) error {
	// make sure the fc is self-consistent before saving
	err := fc.ValidateBasic(fc.Commit.Header.ChainID)
	if err != nil {
		return err
	}

	paths := []string{
		filepath.Join(p.checkDir, p.encodeHeight(fc.Height())),
		filepath.Join(p.valDir, p.encodeHash(fc.Header.ValidatorsHash)),
	}
	for _, path := range paths {
		err := SaveFullCommit(fc, path)
		// unknown error in creating or writing immediately breaks
		if err != nil {
			return err
		}
	}
	return nil
}

// GetByHeight returns the closest commit with height <= h.
func (p *provider) GetByHeight(h int64) (lcd.FullCommit, error) {
	// first we look for exact match, then search...
	path := filepath.Join(p.checkDir, p.encodeHeight(h))
	fc, err := LoadFullCommit(path)
	if lcdErr.IsCommitNotFoundErr(err) {
		path, err = p.searchForHeight(h)
		if err == nil {
			fc, err = LoadFullCommit(path)
		}
	}
	return fc, err
}

// LatestCommit returns the newest commit stored.
func (p *provider) LatestCommit() (fc lcd.FullCommit, err error) {
	// Note to future: please update by 2077 to avoid rollover
	return p.GetByHeight(math.MaxInt32 - 1)
}

// search for height, looks for a file with highest height < h
// return certifiers.ErrCommitNotFound() if not there...
func (p *provider) searchForHeight(h int64) (string, error) {
	d, err := os.Open(p.checkDir)
	if err != nil {
		return "", errors.WithStack(err)
	}
	files, err := d.Readdirnames(0)

	d.Close()
	if err != nil {
		return "", errors.WithStack(err)
	}

	desired := p.encodeHeight(h)
	sort.Strings(files)
	i := sort.SearchStrings(files, desired)
	if i == 0 {
		return "", lcdErr.ErrCommitNotFound()
	}
	found := files[i-1]
	path := filepath.Join(p.checkDir, found)
	return path, errors.WithStack(err)
}

// GetByHash returns a commit exactly matching this validator hash.
func (p *provider) GetByHash(hash []byte) (lcd.FullCommit, error) {
	path := filepath.Join(p.valDir, p.encodeHash(hash))
	return LoadFullCommit(path)
}
