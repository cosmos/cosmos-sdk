package types

// CreateKeyRequest is sent to create a new key
type CreateKeyRequest struct {
	Name       string `json:"name" validate:"required,min=4,printascii"`
	Passphrase string `json:"passphrase" validate:"required,min=10"`
	Algo       string `json:"algo"`
}

// DeleteKeyRequest to destroy a key permanently (careful!)
type DeleteKeyRequest struct {
	Name       string `json:"name" validate:"required,min=4,printascii"`
	Passphrase string `json:"passphrase" validate:"required,min=10"`
}

// UpdateKeyRequest is sent to update the passphrase for an existing key
type UpdateKeyRequest struct {
	Name    string `json:"name" validate:"required,min=4,printascii"`
	OldPass string `json:"passphrase"  validate:"required,min=10"`
	NewPass string `json:"new_passphrase" validate:"required,min=10"`
}

// ErrorResponse is returned for 4xx and 5xx errors
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"` // error message if Success is false
	Code    int    `json:"code"`  // error code if Success is false
}
