package files

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
	"github.com/cosmos/cosmos-sdk/lcd"
)

const (
	// MaxFullCommitSize is the maximum number of bytes we will
	// read in for a full commit to avoid excessive allocations
	// in the deserializer
	MaxFullCommitSize = 1024 * 1024
)

// SaveFullCommit exports the seed in binary / go-amino style
func SaveFullCommit(fc lcd.FullCommit, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	_, err = cdc.MarshalBinaryWriter(f, fc)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SaveFullCommitJSON exports the seed in a json format
func SaveFullCommitJSON(fc lcd.FullCommit, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	bz, err := cdc.MarshalJSON(fc)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = f.Write(bz)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// LoadFullCommit loads the full commit from the file system.
func LoadFullCommit(path string) (lcd.FullCommit, error) {
	var fc lcd.FullCommit
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fc, lcdErr.ErrCommitNotFound()
		}
		return fc, errors.WithStack(err)
	}
	defer f.Close()

	_, err = cdc.UnmarshalBinaryReader(f, &fc, 0)
	if err != nil {
		return fc, errors.WithStack(err)
	}
	return fc, nil
}

// LoadFullCommitJSON loads the commit from the file system in JSON format.
func LoadFullCommitJSON(path string) (lcd.FullCommit, error) {
	var fc lcd.FullCommit
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fc, lcdErr.ErrCommitNotFound()
		}
		return fc, errors.WithStack(err)
	}
	defer f.Close()

	bz, err := ioutil.ReadAll(f)
	if err != nil {
		return fc, errors.WithStack(err)
	}
	err = cdc.UnmarshalJSON(bz, &fc)
	if err != nil {
		return fc, errors.WithStack(err)
	}
	return fc, nil
}
