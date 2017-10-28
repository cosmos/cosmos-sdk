/*
package filestorage provides a secure on-disk storage of private keys and
metadata.  Security is enforced by file and directory permissions, much
like standard ssh key storage.
*/
package filestorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-crypto/keys"
)

const (
	BlockType = "Tendermint Light Client"

	// PrivExt is the extension for private keys.
	PrivExt = "tlc"
	// PubExt is the extensions for public keys.
	PubExt = "pub"

	keyPerm = os.FileMode(0600)
	// pubPerm = os.FileMode(0644)
	dirPerm = os.FileMode(0700)
)

type FileStore struct {
	keyDir string
}

// New creates an instance of file-based key storage with tight permissions
//
// dir should be an absolute path of a directory owner by this user. It will
// be created if it doesn't exist already.
func New(dir string) FileStore {
	err := os.MkdirAll(dir, dirPerm)
	if err != nil {
		panic(err)
	}
	return FileStore{dir}
}

// assert FileStore satisfies keys.Storage
var _ keys.Storage = FileStore{}

// Put creates two files, one with the public info as json, the other
// with the (encoded) private key as gpg ascii-armor style
func (s FileStore) Put(name string, key []byte, info keys.Info) error {
	pub, priv := s.nameToPaths(name)

	// write public info
	err := writeInfo(pub, info)
	if err != nil {
		return err
	}

	// write private info
	return write(priv, name, key)
}

// Get loads the info and (encoded) private key from the directory
// It uses `name` to generate the filename, and returns an error if the
// files don't exist or are in the incorrect format
func (s FileStore) Get(name string) ([]byte, keys.Info, error) {
	pub, priv := s.nameToPaths(name)

	info, err := readInfo(pub)
	if err != nil {
		return nil, info, err
	}

	key, _, err := read(priv)
	return key, info.Format(), err
}

// List parses the key directory for public info and returns a list of
// Info for all keys located in this directory.
func (s FileStore) List() (keys.Infos, error) {
	dir, err := os.Open(s.keyDir)
	if err != nil {
		return nil, errors.Wrap(err, "List Keys")
	}
	defer dir.Close()

	names, err := dir.Readdirnames(0)
	if err != nil {
		return nil, errors.Wrap(err, "List Keys")
	}

	// filter names for .pub ending and load them one by one
	// half the files is a good guess for pre-allocating the slice
	infos := make([]keys.Info, 0, len(names)/2)
	for _, name := range names {
		if strings.HasSuffix(name, PubExt) {
			p := path.Join(s.keyDir, name)
			info, err := readInfo(p)
			if err != nil {
				return nil, err
			}
			infos = append(infos, info.Format())
		}
	}

	return infos, nil
}

// Delete permanently removes the public and private info for the named key
// The calling function should provide some security checks first.
func (s FileStore) Delete(name string) error {
	pub, priv := s.nameToPaths(name)
	err := os.Remove(priv)
	if err != nil {
		return errors.Wrap(err, "Deleting Private Key")
	}
	err = os.Remove(pub)
	return errors.Wrap(err, "Deleting Public Key")
}

func (s FileStore) nameToPaths(name string) (pub, priv string) {
	privName := fmt.Sprintf("%s.%s", name, PrivExt)
	pubName := fmt.Sprintf("%s.%s", name, PubExt)
	return path.Join(s.keyDir, pubName), path.Join(s.keyDir, privName)
}

func writeInfo(path string, info keys.Info) error {
	return write(path, info.Name, info.PubKey.Bytes())
}

func readInfo(path string) (info keys.Info, err error) {
	var data []byte
	data, info.Name, err = read(path)
	if err != nil {
		return
	}
	pk, err := crypto.PubKeyFromBytes(data)
	info.PubKey = pk
	return
}

func read(path string) ([]byte, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", errors.Wrap(err, "Reading data")
	}
	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, "", errors.Wrap(err, "Reading data")
	}
	block, headers, key, err := crypto.DecodeArmor(string(d))
	if err != nil {
		return nil, "", errors.Wrap(err, "Invalid Armor")
	}
	if block != BlockType {
		return nil, "", errors.Errorf("Unknown key type: %s", block)
	}
	return key, headers["name"], nil
}

func write(path, name string, key []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, keyPerm)
	if err != nil {
		return errors.Wrap(err, "Writing data")
	}
	defer f.Close()
	headers := map[string]string{"name": name}
	text := crypto.EncodeArmor(BlockType, headers, key)
	_, err = f.WriteString(text)
	return errors.Wrap(err, "Writing data")
}
