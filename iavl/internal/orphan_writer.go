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
