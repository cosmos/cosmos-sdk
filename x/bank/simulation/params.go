package simulation

import "math/rand"

// Simulation parameter constants
const (
	SendEnabled = "send_enabled"
)

// GenParams generates random bank parameters
func GenParams(paramSims map[string]func(r *rand.Rand) interface{}) {
	paramSims[SendEnabled] = func(r *rand.Rand) interface{} {
		return r.Int63n(2) == 0
	}
}
