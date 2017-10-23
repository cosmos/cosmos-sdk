/*
package filestorage provides a secure on-disk storage of private keys and
metadata.  Security is enforced by file and directory permissions, much
like standard ssh key storage.
*/
package filestorage

import (
	"encoding/hex"
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
	// BlockType is the type of block.
	BlockType = "Tendermint Light Client"

	// PrivExt is the extension for private keys.
	PrivExt = "tlc"
	// PubExt is the extensions for public keys.
	PubExt = "pub"

	keyPerm = os.FileMode(0600)
	// pubPerm = os.FileMode(0644)
	dirPerm = os.FileMode(0700)
)

// FileStore is a file-based key storage with tight permissions.
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
func (s FileStore) Put(name string, salt, key []byte, info keys.Info) error {
	pub, priv := s.nameToPaths(name)

	// write public info
	err := writeInfo(pub, info)
	if err != nil {
		return err
	}

	// write private info
	return write(priv, name, salt, key)
}

// Get loads the info and (encoded) private key from the directory
// It uses `name` to generate the filename, and returns an error if the
// files don't exist or are in the incorrect format
func (s FileStore) Get(name string) (salt []byte, key []byte, info keys.Info, err error) {
	pub, priv := s.nameToPaths(name)

	info, err = readInfo(pub)
	if err != nil {
		return nil, nil, info, err
	}

	salt, key, _, err = read(priv)
	return salt, key, info.Format(), err
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

func readInfo(path string) (info keys.Info, err error) {
	f, err := os.Open(path)
	if err != nil {
		return info, errors.Wrap(err, "Reading data")
	}
	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		return info, errors.Wrap(err, "Reading data")
	}

	block, headers, key, err := crypto.DecodeArmor(string(d))
	if err != nil {
		return info, errors.Wrap(err, "Invalid Armor")
	}

	if block != BlockType {
		return info, errors.Errorf("Unknown key type: %s", block)
	}

	pk, _ := crypto.PubKeyFromBytes(key)
	info.Name = headers["name"]
	info.PubKey = pk

	return info, nil
}

func read(path string) (salt, key []byte, name string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "Reading data")
	}
	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "Reading data")
	}

	block, headers, key, err := crypto.DecodeArmor(string(d))
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "Invalid Armor")
	}

	if block != BlockType {
		return nil, nil, "", errors.Errorf("Unknown key type: %s", block)
	}

	if headers["kdf"] != "bcrypt" {
		return nil, nil, "", errors.Errorf("Unrecognized KDF type: %v", headers["kdf"])
	}

	if headers["salt"] == "" {
		return nil, nil, "", errors.Errorf("Missing salt bytes")
	}

	salt, err = hex.DecodeString(headers["salt"])
	if err != nil {
		return nil, nil, "", errors.Errorf("Error decoding salt: %v", err.Error())
	}

	return salt, key, headers["name"], nil
}

func writeInfo(path string, info keys.Info) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, keyPerm)
	if err != nil {
		return errors.Wrap(err, "Writing data")
	}
	defer f.Close()

	headers := map[string]string{"name": info.Name}
	text := crypto.EncodeArmor(BlockType, headers, info.PubKey.Bytes())
	_, err = f.WriteString(text)

	return errors.Wrap(err, "Writing data")
}

func write(path, name string, salt, key []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, keyPerm)
	if err != nil {
		return errors.Wrap(err, "Writing data")
	}
	defer f.Close()

	headers := map[string]string{
		"name": name,
		"kdf":  "bcrypt",
		"salt": fmt.Sprintf("%X", salt),
	}

	text := crypto.EncodeArmor(BlockType, headers, key)
	_, err = f.WriteString(text)

	return errors.Wrap(err, "Writing data")
}
