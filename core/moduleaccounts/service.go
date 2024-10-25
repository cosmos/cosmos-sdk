package moduleaccounts

type Service interface {
	Register(name string, perms []string) error
	Address(name string) []byte         // TODO: should we return an empty byte slice if it wasn't registered or should we just register it?
	IsModuleAccount(addr []byte) string // Needed in burn coins and ante handler
}

type ServiceWithPerms interface {
	Service

	AllAccounts() map[string][]byte
	HasPermission(name string, perm string) bool
}
