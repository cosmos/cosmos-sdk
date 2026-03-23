package internal

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/store/v2/types/maps"
)

type CommitID struct {
	Version int64
	Hash    []byte
}

type StoreInfo struct {
	Name     string
	CommitId CommitID
}

type CommitInfo struct {
	Version    int64
	Timestamp  time.Time
	StoreInfos []StoreInfo
}

func saveCommitInfo(outDir string, ci *CommitInfo) error {
	ciFilename := filepath.Join(outDir, commitInfoSubPath, fmt.Sprintf("%d", ci.Version))
	err := os.MkdirAll(filepath.Dir(ciFilename), 0700)
	if err != nil {
		return fmt.Errorf("failed to create commit info directory: %w", err)
	}
	ciFile, err := os.OpenFile(ciFilename, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to create commit info file: %w", err)
	}
	err = writeCommitInfo(ciFile, ci)
	if err != nil {
		return fmt.Errorf("failed to write commit info: %w", err)
	}
	err = ciFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync commit info file: %w", err)
	}
	err = ciFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close commit info file: %w", err)
	}

	return nil
}

func writeCommitInfo(writer io.Writer, info *CommitInfo) error {
	err := writeCommitInfoHeader(writer, info)
	if err != nil {
		return fmt.Errorf("failed to write commit info header: %w", err)
	}

	err = writeCommitInfoFooter(writer, info)
	if err != nil {
		return fmt.Errorf("failed to write commit info footer: %w", err)
	}

	return nil
}

func writeCommitInfoHeader(headerBuf io.Writer, info *CommitInfo) error {
	// write version as litte-endian uint32
	var scratchBuf [binary.MaxVarintLen64]byte
	binary.LittleEndian.PutUint32(scratchBuf[:4], uint32(info.Version))
	_, err := headerBuf.Write(scratchBuf[:4])
	if err != nil {
		return fmt.Errorf("failed to write commit info version: %w", err)
	}

	// write timestamp as unix nano int64
	binary.LittleEndian.PutUint64(scratchBuf[:8], uint64(info.Timestamp.UnixNano()))
	_, err = headerBuf.Write(scratchBuf[:8])
	if err != nil {
		return fmt.Errorf("failed to write commit info timestamp: %w", err)
	}

	// write the number of store infos as little-endian uint32
	binary.LittleEndian.PutUint32(scratchBuf[:4], uint32(len(info.StoreInfos)))
	_, err = headerBuf.Write(scratchBuf[:4])
	if err != nil {
		return fmt.Errorf("failed to write commit info store info count: %w", err)
	}

	// write each store name as a length-prefixed string
	// use index-based access to read-only Name, avoiding a full StoreInfo copy
	// that would race with concurrent CommitId writes from hash workers
	for i := range info.StoreInfos {
		// varint length prefix
		name := info.StoreInfos[i].Name
		nameLen := uint64(len(name))
		n := binary.PutUvarint(scratchBuf[:], nameLen)
		_, err := headerBuf.Write(scratchBuf[:n])
		if err != nil {
			return fmt.Errorf("failed to write commit info store info name length: %w", err)
		}
		_, err = headerBuf.Write([]byte(name))
		if err != nil {
			return fmt.Errorf("failed to write commit info store info name: %w", err)
		}
	}

	return nil
}

func writeCommitInfoFooter(writer io.Writer, info *CommitInfo) error {
	var scratchBuf [binary.MaxVarintLen64]byte

	// append each store hash to the file
	for _, storeInfo := range info.StoreInfos {
		// length-prefixed hash
		hashLen := uint64(len(storeInfo.CommitId.Hash))
		n := binary.PutUvarint(scratchBuf[:], hashLen)
		_, err := writer.Write(scratchBuf[:n])
		if err != nil {
			return fmt.Errorf("failed to write commit info store info hash length: %w", err)
		}

		_, err = writer.Write(storeInfo.CommitId.Hash)
		if err != nil {
			return fmt.Errorf("failed to write commit info store info hash: %w", err)
		}
	}

	return nil
}

const commitInfoSubPath = "commit_info"

// loadLatestCommitInfo loads the highest version number commit info file from the commit_info directory
// if any exist, returning the version and CommitInfo.
func loadLatestCommitInfo(dir string) (ci *CommitInfo, earliestVersion int64, err error) {
	commitInfoDir := filepath.Join(dir, commitInfoSubPath)
	err = os.MkdirAll(commitInfoDir, 0o700)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create commit info dir: %w", err)
	}

	entries, err := os.ReadDir(commitInfoDir)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read commit info dir: %w", err)
	}

	// find the latest version by looking for the highest numbered file
	var latestVersion int64
	for _, entry := range entries {
		// clean up incomplete pending commit info files from interrupted commits
		if strings.HasPrefix(entry.Name(), ".pending.") {
			_ = os.Remove(filepath.Join(commitInfoDir, entry.Name()))
			continue
		}

		var version int64
		_, err := fmt.Sscanf(entry.Name(), "%d", &version)
		if err != nil {
			// skip non-numeric files
			continue
		}
		if version > latestVersion {
			latestVersion = version
		}
		if version < earliestVersion || earliestVersion == 0 {
			earliestVersion = version
		}
	}

	if latestVersion == 0 {
		// no versions found, no commit info to load
		return nil, 0, nil
	}

	commitInfo, err := loadCommitInfo(dir, latestVersion)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to load commit info for version %d: %w", latestVersion, err)
	}

	if commitInfo.Version != latestVersion {
		return nil, 0, fmt.Errorf("commit info version mismatch: expected %d, got %d", latestVersion, commitInfo.Version)
	}

	return commitInfo, earliestVersion, nil
}

func loadCommitInfo(dir string, version int64) (*CommitInfo, error) {
	commitInfoDir := filepath.Join(dir, commitInfoSubPath)
	err := os.MkdirAll(commitInfoDir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit info dir: %w", err)
	}

	commitInfoPath := filepath.Join(commitInfoDir, fmt.Sprintf("%d", version))
	bz, err := os.ReadFile(commitInfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit info file for version %d: %w", version, err)
	}

	rdr := bytes.NewReader(bz)

	// read version
	var storedVersion uint32
	err = binary.Read(rdr, binary.LittleEndian, &storedVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit info version for version %d: %w", version, err)
	}
	if int64(storedVersion) != version {
		return nil, fmt.Errorf("commit info version mismatch: expected %d, got %d", version, storedVersion)
	}

	// read timestamp
	var timestampNano uint64
	err = binary.Read(rdr, binary.LittleEndian, &timestampNano)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit info timestamp for version %d: %w", version, err)
	}

	// read store count
	var storeCount uint32
	err = binary.Read(rdr, binary.LittleEndian, &storeCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit info store count for version %d: %w", version, err)
	}

	commitInfo := &CommitInfo{
		StoreInfos: make([]StoreInfo, storeCount),
		Timestamp:  time.Unix(0, int64(timestampNano)),
		Version:    version,
	}

	// read each store info
	for i := uint32(0); i < storeCount; i++ {
		// read name length
		nameLen, err := binary.ReadUvarint(rdr)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit info store info name length for version %d: %w", version, err)
		}
		nameBytes := make([]byte, nameLen)
		_, err = io.ReadFull(rdr, nameBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit info store info name for version %d: %w", version, err)
		}
		commitInfo.StoreInfos[i].Name = string(nameBytes)
	}

	// Hashes are appended after the header without an additional fsync, so they may be
	// missing or incomplete if the process crashed or the write hasn't completed yet.
	// This is expected — the header is the durable part, hashes are best-effort.
	// If hashes are missing, we return the commit info with empty commit IDs.
	for i := uint32(0); i < storeCount; i++ {
		hashLen, err := binary.ReadUvarint(rdr)
		if err != nil {
			// no more hash data available — return commit info without hashes
			return commitInfo, nil
		}

		hashBytes := make([]byte, hashLen)
		_, err = io.ReadFull(rdr, hashBytes)
		if err != nil {
			// partial hash data — return what we have so far
			return commitInfo, nil
		}

		commitInfo.StoreInfos[i].CommitId = CommitID{
			Version: version,
			Hash:    hashBytes,
		}
	}

	return commitInfo, nil
}

// GetHash returns the GetHash from the CommitID.
// This is used in CommitInfo.Hash()
//
// When we commit to this in a merkle proof, we create a map of storeInfo.Name -> storeInfo.GetHash()
// and build a merkle proof from that.
// This is then chained with the substore proof, so we prove the root hash from the substore before this
// and need to pass that (unmodified) as the leaf value of the multistore proof.
func (si StoreInfo) GetHash() []byte {
	return si.CommitId.Hash
}

func (ci CommitInfo) toMap() map[string][]byte {
	m := make(map[string][]byte, len(ci.StoreInfos))
	for _, storeInfo := range ci.StoreInfos {
		m[storeInfo.Name] = storeInfo.GetHash()
	}

	return m
}

// Hash returns the simple merkle root hash of the stores sorted by name.
func (ci CommitInfo) Hash() []byte {
	// we need a special case for empty set, as SimpleProofsFromMap requires at least one entry
	if len(ci.StoreInfos) == 0 {
		emptyHash := sha256.Sum256([]byte{})
		return emptyHash[:]
	}

	rootHash, _, _ := maps.ProofsFromMap(ci.toMap())

	if len(rootHash) == 0 {
		emptyHash := sha256.Sum256([]byte{})
		return emptyHash[:]
	}

	return rootHash
}

func (ci CommitInfo) ProofOp(storeName string) cmtprotocrypto.ProofOp {
	ret, err := ProofOpFromMap(ci.toMap(), storeName)
	if err != nil {
		panic(err)
	}
	return ret
}

func (ci CommitInfo) CommitID() CommitID {
	return CommitID{
		Version: ci.Version,
		Hash:    ci.Hash(),
	}
}
