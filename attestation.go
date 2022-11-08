package goosecoin

import (
	"log"
	"sync"
)

type Attestation struct {
	Validator Validator
	BlockHash Hash
	Signature []byte
}

type Attestations struct {
	Attestations map[[32]byte]Attestation
	lock         sync.Mutex
}

func NewAttestations() Attestations {
	return Attestations{
		Attestations: make(map[[32]byte]Attestation),
	}
}

func (a *Attestations) Add(attestation Attestation) {
	if a.Attestations == nil {
		a.Attestations = make(map[[32]byte]Attestation)
	}
	a.lock.Lock()
	a.Attestations[attestation.Validator.Key()] = attestation
	a.lock.Unlock()
	if a.Size() >= 22 {
		log.Println("attestation done")
	}
}

func (a *Attestations) Size() int {
	return len(a.Attestations)
}
