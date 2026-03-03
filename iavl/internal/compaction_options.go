package internal

type CompactionOptions struct {
	RetainVersion          uint32
	CompactionRolloverSize int64
}
