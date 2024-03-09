package types

import (
	"fmt"
	"strings"
)

func (req AccountsQueryRequest) Validate() error {
	if strings.TrimSpace(req.Owner) == "" || strings.TrimSpace(req.AccountType) == "" {
		return fmt.Errorf("invalid request, request field must not be empty")
	}

	return nil
}
