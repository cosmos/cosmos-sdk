package iavl

type Options struct {
	FsyncWAL bool `json:"fsync_wal"`
	WriteWAL bool `json:"write_wal"`
}
