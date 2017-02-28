package keys

// Storage has many implementation, based on security and sharing requirements
// like disk-backed, mem-backed, vault, db, etc.
type Storage interface {
	Put(name string, key []byte, info KeyInfo) error
	Get(name string) ([]byte, KeyInfo, error)
	List() ([]KeyInfo, error)
	Delete(name string) error
}
