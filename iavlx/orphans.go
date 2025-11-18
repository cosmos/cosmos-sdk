package iavlx

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type OrphanWriter struct {
	*FileWriter
}

func NewOrphanWriter(file *os.File) *OrphanWriter {
	return &OrphanWriter{
		FileWriter: NewFileWriter(file),
	}
}

func (ow *OrphanWriter) WriteOrphan(version uint32, id NodeID) error {
	var bz [12]byte
	binary.LittleEndian.PutUint32(bz[0:4], version)
	binary.LittleEndian.PutUint64(bz[4:12], uint64(id))
	_, err := ow.Write(bz[:])
	return err
}

func (ow *OrphanWriter) WriteOrphanMap(orphanMap map[NodeID]uint32) error {
	for id, version := range orphanMap {
		if err := ow.WriteOrphan(version, id); err != nil {
			return err
		}
	}

	return ow.Flush()
}

func ReadOrphanMap(file *os.File) (map[NodeID]uint32, error) {
	file2, err := os.Open(file.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open orphan file for reading: %w", err)
	}
	orphanMap := make(map[NodeID]uint32)
	rdr := bufio.NewReader(file2)
	var buf [12]byte
	for {
		_, err := rdr.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				return orphanMap, nil
			}
			return nil, err
		}
		version := binary.LittleEndian.Uint32(buf[0:4])
		id := NodeID(binary.LittleEndian.Uint64(buf[4:12]))
		if _, exists := orphanMap[id]; !exists {
			orphanMap[id] = version
		}
	}
}
