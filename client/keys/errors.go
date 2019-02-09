package keys

import "fmt"

func errKeyNameConflict(name string) error {
	return fmt.Errorf("account with name %s already exists", name)
}

func errMissingName() error {
	return fmt.Errorf("you have to specify a name for the locally stored account")
}

func errMissingPassword() error {
	return fmt.Errorf("you have to specify a password for the locally stored account")
}

func errMissingMnemonic() error {
	return fmt.Errorf("you have to specify a mnemonic for key recovery")
}

func errInvalidMnemonic() error {
	return fmt.Errorf("the mnemonic is invalid")
}

func errInvalidAccountNumber() error {
	return fmt.Errorf("the account number is invalid")
}

func errInvalidIndexNumber() error {
	return fmt.Errorf("the index number is invalid")
}
