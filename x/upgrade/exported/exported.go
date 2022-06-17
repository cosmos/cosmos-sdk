package exported

// AppVersionManager defines the interface which allows managing the appVersion field.
type AppVersionManager interface {
	GetAppVersion() (uint64, error)
	SetAppVersion(version uint64) error
}
