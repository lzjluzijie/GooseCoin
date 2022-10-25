package goosecoin

import (
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
)

type HashRequest struct {
	Validator ed25519.PublicKey
	Hash      []byte
	Signature []byte
}

type SeedRequest struct {
	Validator ed25519.PublicKey
	Seed      []byte
	Signature []byte
}

type Randao struct {
	Validators [][32]byte
	Hashs      map[[32]byte][]byte
	Seeds      map[[32]byte][]byte
}

func NewRandao(keys []ed25519.PublicKey) *Randao {
	validators := make([][32]byte, len(keys))
	for i, k := range keys {
		if len(k) != 32 {
			panic("invalid public key")
		}
		copy(validators[i][:], k)
	}
	return &Randao{
		Validators: validators,
		Hashs:      make(map[[32]byte][]byte),
		Seeds:      make(map[[32]byte][]byte),
	}
}

func (r *Randao) AddHash(req HashRequest) error {
	ok := ed25519.Verify(req.Validator, req.Hash[:], req.Signature)
	if !ok {
		return errors.New("invalid signature")
	}
	v := (*[32]byte)(req.Validator)
	if _, ok := r.Hashs[*v]; ok {
		return errors.New("hash already exists")
	}
	r.Hashs[*v] = req.Hash
	return nil
}

func (r *Randao) AddSeed(req SeedRequest) error {
	ok := ed25519.Verify(req.Validator, req.Seed[:], req.Signature)
	if !ok {
		return errors.New("invalid signature")
	}
	v := (*[32]byte)(req.Validator)
	if _, ok := r.Seeds[*v]; ok {
		return errors.New("hash already exists")
	}
	r.Seeds[*v] = req.Seed
	return nil
}

func (r *Randao) Result() []byte {
	h := sha256.New()
	for _, v := range r.Validators {
		_, err := h.Write(r.Seeds[v][:])
		if err != nil {
			panic(err)
		}
	}
	return h.Sum(nil)
}
