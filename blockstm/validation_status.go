package blockstm

import "sync"

// ValidationStatus records wave numbers related to transaction validation and commit logic.
type ValidationStatus struct {
	sync.Mutex

	TriggeredWave Wave // Wave counter when this transaction triggers wave validation
	RequiredWave  Wave // The wave number when validation is triggered specifically for this transaction
	ValidatedWave Wave // Maximum wave number that validation is successfully completed
}

func (vs *ValidationStatus) SetTriggeredWave(wave Wave) {
	vs.Lock()
	if wave > vs.TriggeredWave {
		vs.TriggeredWave = wave
	}
	vs.Unlock()
}

func (vs *ValidationStatus) SetValidatedWave(wave Wave) {
	vs.Lock()
	if wave > vs.ValidatedWave {
		vs.ValidatedWave = wave
	}
	vs.Unlock()
}

func (vs *ValidationStatus) SetRequiredWave(wave Wave) {
	vs.Lock()
	if wave > vs.RequiredWave {
		vs.RequiredWave = wave
	}
	vs.Unlock()
}
