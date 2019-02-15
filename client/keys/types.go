package keys

// used for outputting keys.Info over REST
type KeyOutput struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Address  string `json:"address"`
	PubKey   string `json:"pub_key"`
	Mnemonic string `json:"mnemonic,omitempty"`
}

// AddNewKey request a new key
type AddNewKey struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Mnemonic string `json:"mnemonic"`
	Account  int    `json:"account,string,omitempty"`
	Index    int    `json:"index,string,omitempty"`
}

// RecoverKeyBody recovers a key
type RecoverKey struct {
	Password string `json:"password"`
	Mnemonic string `json:"mnemonic"`
	Account  int    `json:"account,string,omitempty"`
	Index    int    `json:"index,string,omitempty"`
}

// UpdateKeyReq requests updating a key
type UpdateKeyReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// DeleteKeyReq requests deleting a key
type DeleteKeyReq struct {
	Password string `json:"password"`
}
