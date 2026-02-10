package internal

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type OrphanLogEntry struct {
	OrphanedVersion uint32
	NodeID          NodeID
}

func ReadOrphanLog(file *os.File) (*OrphanLogReader, error) {
	// re-open the file to get a separate file handle for reading, since the original file may be open for writing
	file2, err := os.OpenFile(file.Name(), os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open orphan file for reading: %w", err)
	}

	return &OrphanLogReader{
		file: file2,
		rdr:  bufio.NewReader(file2),
	}, nil
}

type OrphanLogReader struct {
	file          *os.File
	rdr           *bufio.Reader
	curVersion    uint32
	curCheckpoint uint32
}

func (olr *OrphanLogReader) Next() (OrphanLogEntry, error) {
	if olr.curVersion == 0 {
		// read initial version & checkpoint
		x, err := olr.readU32()
		if err != nil {
			return OrphanLogEntry{}, fmt.Errorf("failed to read initial version from orphan log: %w", err)
		}
		if x == 0 {
			return OrphanLogEntry{}, fmt.Errorf("orphan log cannot start with a zero version")
		}
		olr.curVersion = x
		x, err = olr.readU32()
		if err != nil {
			return OrphanLogEntry{}, fmt.Errorf("failed to read initial checkpoint from orphan log: %w", err)
		}
		if x == 0 {
			return OrphanLogEntry{}, fmt.Errorf("orphan log cannot start with a zero checkpoint")
		}
		olr.curCheckpoint = x
	}

	cur, err := olr.readU32()
	if err != nil {
		return OrphanLogEntry{}, err
	}
	if cur == 0 {
		// version or checkpoint change
		cur, err = olr.readU32()
		if err != nil {
			return OrphanLogEntry{}, err
		}
		if cur == 0 {
			// version change
			cur, err = olr.readU32()
			if err != nil {
				return OrphanLogEntry{}, err
			}
			if cur == 0 {
				return OrphanLogEntry{}, fmt.Errorf("zero is not a valid version in orphan log")
			}
			olr.curVersion = cur

			cur, err = olr.readU32()
			if err != nil {
				return OrphanLogEntry{}, err
			}
		}

		if cur == 0 {
			return OrphanLogEntry{}, fmt.Errorf("zero is not a valid checkpoint in orphan log")
		}
		olr.curCheckpoint = cur

		cur, err = olr.readU32()
		if err != nil {
			return OrphanLogEntry{}, err
		}
	}
	if cur == 0 {
		return OrphanLogEntry{}, fmt.Errorf("zero is not a valid flag index in orphan log")
	}
	flagIndex := cur
	id := NodeID{
		checkpoint: olr.curCheckpoint,
		flagIndex:  nodeFlagIndex(flagIndex),
	}
	return OrphanLogEntry{
		OrphanedVersion: olr.curVersion,
		NodeID:          id,
	}, nil
}

func (olr *OrphanLogReader) readU32() (uint32, error) {
	var bz [4]byte
	_, err := io.ReadFull(olr.rdr, bz[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(bz[:]), nil
}

func (olr *OrphanLogReader) resume() {
	olr.rdr = bufio.NewReader(olr.file)
}

func (olr *OrphanLogReader) Close() error {
	return olr.file.Close()
}
