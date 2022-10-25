package goosecoin

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
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

type RandaoStatus int64

const (
	RandaoStatusHash RandaoStatus = iota
	RandaoStatusSeed
	RandaoStatusFinished
)

type Randao struct {
	s    *Server
	seed []byte
	hash []byte

	Status     RandaoStatus
	Validators [][32]byte
	Hashs      map[[32]byte][]byte
	Seeds      map[[32]byte][]byte
	Result     []byte
}

func (s *Server) NewRandao() *Randao {
	keys := s.Validators
	validators := make([][32]byte, len(keys))
	for i, k := range keys {
		if len(k) != 32 {
			panic("invalid public key")
		}
		copy(validators[i][:], k)
	}

	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(seed)
	r := &Randao{
		s:          s,
		seed:       seed,
		hash:       hash[:],
		Status:     RandaoStatusHash,
		Validators: validators,
		Hashs:      make(map[[32]byte][]byte),
		Seeds:      make(map[[32]byte][]byte),
	}
	var cur [32]byte
	copy(cur[:], s.PublicKey)
	r.Seeds[cur] = seed
	r.Hashs[cur] = hash[:]
	return r
}

func (r *Randao) SendHash() {
	signature := ed25519.Sign(r.s.PrivateKey, r.hash)
	for _, peer := range r.s.Config.Peers {
		req := HashRequest{
			Validator: r.s.PublicKey,
			Hash:      r.hash[:],
			Signature: signature,
		}
		data, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}
		resp, err := http.Post(peer+"/randao/hash", "application/json", bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != http.StatusOK {
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			panic(string(data))
		}
	}
}

func (r *Randao) SendSeed() {
	signature := ed25519.Sign(r.s.PrivateKey, r.seed)
	for _, peer := range r.s.Config.Peers {
		req := SeedRequest{
			Validator: r.s.PublicKey,
			Seed:      r.seed,
			Signature: signature,
		}
		data, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}
		resp, err := http.Post(peer+"/randao/seed", "application/json", bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != http.StatusOK {
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			panic(string(data))
		}
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
	if len(r.Hashs) == len(r.Validators) {
		r.Status = RandaoStatusSeed
		r.SendSeed()
	}
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
	if len(r.Seeds) == len(r.Validators) {
		r.Status = RandaoStatusFinished
		r.Result = r.result()
		log.Println(r.Result)
	}
	return nil
}

func (r *Randao) result() []byte {
	h := sha256.New()
	for _, v := range r.Validators {
		_, err := h.Write(r.Seeds[v][:])
		if err != nil {
			panic(err)
		}
	}
	return h.Sum(nil)
}
