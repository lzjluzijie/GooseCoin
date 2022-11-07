package goosecoin

import "crypto/ed25519"

type Validator struct {
	PublicKey ed25519.PublicKey
	// Stake     uint64
}

type Network struct {
	// TotalStake uint64
	Validators []Validator
}
