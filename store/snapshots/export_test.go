package snapshots

// SetSyncDirFn replaces syncDirFn for testing and returns the original so the
// caller can restore it with defer.
func SetSyncDirFn(fn func(string) error) func(string) error {
	orig := syncDirFn
	syncDirFn = fn
	return orig
}
