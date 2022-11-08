package goosecoin

import "crypto/ed25519"

type Validator struct {
	PublicKey ed25519.PublicKey
	// Stake     uint64
}

func (v Validator) Key() [32]byte {
	var key [32]byte
	copy(key[:], v.PublicKey)
	return key
}

func (v Validator) Verify(data []byte, signature []byte) bool {
	return ed25519.Verify(v.PublicKey, data, signature)
}

type Network struct {
	// TotalStake uint64
	Validators []Validator
}
