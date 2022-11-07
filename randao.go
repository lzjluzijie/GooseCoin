package goosecoin

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

type HashRequest struct {
	ID        string
	Validator ed25519.PublicKey
	Hash      []byte
	Signature []byte
}

type SeedRequest struct {
	ID        string
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

	ID       string
	OnFinish func(result []byte)

	Status     RandaoStatus
	Validators [][32]byte
	Hashs      map[[32]byte][]byte
	Seeds      map[[32]byte][]byte
	Result     []byte
}

func (s *Server) NewRandao(id string, onFinish func(result []byte)) *Randao {
	keys := s.network.Validators
	validators := make([][32]byte, len(keys))
	for i, k := range keys {
		if len(k.PublicKey) != 32 {
			panic("invalid public key")
		}
		copy(validators[i][:], k.PublicKey)
	}

	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(seed)
	r := &Randao{
		s:    s,
		seed: seed,
		hash: hash[:],

		ID:       id,
		OnFinish: onFinish,

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
	req := HashRequest{
		ID:        r.ID,
		Validator: r.s.PublicKey,
		Hash:      r.hash[:],
		Signature: signature,
	}
	data, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	for _, peer := range r.s.Config.Peers {
		resp, err := http.Post(peer+"/randao/hash", "application/json", bytes.NewReader(data))
		if err != nil {
			log.Println(base64.StdEncoding.EncodeToString(r.s.PublicKey), peer)
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
	log.Println("send hash ok", base64.StdEncoding.EncodeToString(r.s.PublicKey))
}

func (r *Randao) SendSeed() {
	signature := ed25519.Sign(r.s.PrivateKey, r.seed)
	req := SeedRequest{
		ID:        r.ID,
		Validator: r.s.PublicKey,
		Seed:      r.seed,
		Signature: signature,
	}
	data, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	for _, peer := range r.s.Config.Peers {
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
	log.Println("send seed ok", base64.StdEncoding.EncodeToString(r.s.PublicKey))
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
		go r.SendSeed()
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
		// log.Println(r.Result)
		if r.OnFinish != nil {
			go r.OnFinish(r.Result)
		}
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
