package container

// Scope is a special type of provider argument which provider's can use to provide
// "scoped" dependency. A scoped dependency is one which is dependent on the scope.
// A scoped provider should have a first argument of type Scope and return a (possibly)
// different dependency for each scope. This can be used to configure things like
// scoped store access where the requesting scope gets a private store key that cannot
// be accessed by providers in other scopes. Scopes can also be used to configure
// security.
//
// Ex:
// func KVStoreKeyProvider(scope Scope) *sdk.KVStoreKey {
//	 return sdk.NewKVStoreKey(string(scope))
// }
type Scope string
