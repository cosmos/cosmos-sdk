package types

import "strings"

func (m MsgRegister) Validate() error {
	if strings.TrimSpace(m.Account) == "" || strings.TrimSpace(m.Owner) == "" {
		return ErrInvalidAddress
	}

	return nil
}
