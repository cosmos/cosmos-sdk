package keys

import "fmt"

func errKeyNameConflict(name string) error {
	return fmt.Errorf("acount with name %s already exists", name)
}

func errMissingName() error {
	return fmt.Errorf("you have to specify a name for the locally stored account")
}

func errMissingPassword() error {
	return fmt.Errorf("you have to specify a password for the locally stored account")
}

func errMissingSeed() error {
	return fmt.Errorf("you have to specify seed for key recover")
}
