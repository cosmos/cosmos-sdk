package blockstm

import "sync"

type validationStatus struct {
	TriggeredWave Wave // Wave counter when this transaction triggers wave validation
	RequiredWave  Wave // The wave number when validation is triggered specifically for this transaction
	ValidatedWave int  // Maximum wave number that validation is successfully completed, -1 if not validated yet
}

// ValidationStatus records wave numbers related to transaction validation and commit logic.
type ValidationStatus struct {
	sync.Mutex
	validationStatus
}

func (vs *ValidationStatus) Init() {
	vs.ValidatedWave = -1
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
	if int(wave) > vs.ValidatedWave {
		vs.ValidatedWave = int(wave)
	}
	vs.Unlock()
}

func (vs *ValidationStatus) Read() (s validationStatus) {
	vs.Lock()
	s = vs.validationStatus
	vs.Unlock()
	return
}
