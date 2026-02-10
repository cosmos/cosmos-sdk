package internal

import (
	"encoding/binary"
	"os"
)

type OrphanWriter struct {
	*FileWriter
	curVersion    uint32
	curCheckpoint uint32
}

func NewOrphanWriter(file *os.File) *OrphanWriter {
	return &OrphanWriter{
		FileWriter: NewFileWriter(file),
	}
}

func (ow *OrphanWriter) WriteOrphan(version uint32, id NodeID) error {
	var bz [8]byte
	checkpoint := id.Checkpoint()
	if ow.Size() > 0 {
		if version != ow.curVersion {
			// write 8 bytes of zeros to indicate a version change
			if _, err := ow.Write(bz[:]); err != nil {
				return err
			}
			// write version and checkpoint for the new version
			binary.LittleEndian.PutUint32(bz[0:4], version)
			binary.LittleEndian.PutUint32(bz[4:8], checkpoint)
			if _, err := ow.Write(bz[:]); err != nil {
				return err
			}
		} else if checkpoint != ow.curCheckpoint {
			// write 4 bytes of zeros to indicate a checkpoint change
			if _, err := ow.Write(bz[0:4]); err != nil {
				return err
			}
			// write checkpoint for the new checkpoint
			binary.LittleEndian.PutUint32(bz[0:4], checkpoint)
			if _, err := ow.Write(bz[0:4]); err != nil {
				return err
			}
		}
	} else {
		// for the first entry, write version and checkpoint
		binary.LittleEndian.PutUint32(bz[0:4], version)
		binary.LittleEndian.PutUint32(bz[4:8], checkpoint)
		if _, err := ow.Write(bz[:]); err != nil {
			return err
		}
	}
	ow.curVersion = version
	ow.curCheckpoint = checkpoint
	// write flag index for the orphan node
	binary.LittleEndian.PutUint32(bz[0:4], uint32(id.flagIndex))
	if _, err := ow.Write(bz[0:4]); err != nil {
		return err
	}
	return nil
}

/*func ReadOrphanLog(file *os.File) (map[NodeID]uint32, error) {
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
		nodeCheckpoint := binary.LittleEndian.Uint32(buf[4:8])
		flagIndex := binary.LittleEndian.Uint32(buf[8:12])
		id := NodeID{
			flagIndex:  nodeFlagIndex(flagIndex),
			checkpoint: nodeCheckpoint,
		}
		if _, exists := orphanMap[id]; !exists {
			orphanMap[id] = version
		}
	}
}
*/
