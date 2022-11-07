package goosecoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Hash []byte

type Block struct {
	Hash        Hash
	HexHash     string
	Height      uint64
	PrevHash    Hash
	HexPrevHash string
	TimeStamp   time.Time
	Data        []Message

	Validator Validator
	Signature []byte

	prev *Block
}

func (block *Block) ComputeHash() Hash {
	data := new(bytes.Buffer)
	data.Write(Uint64ToBytes(block.Height))
	data.Write(block.PrevHash[:])
	data.Write(Uint64ToBytes(uint64(block.TimeStamp.UnixNano())))
	for _, message := range block.Data {
		data.Write(message)
	}
	hash := sha256.Sum256(data.Bytes())
	return hash[:]
}

func NewBlock(prev *Block, data []Message) *Block {
	block := &Block{
		Height:      prev.Height + 1,
		PrevHash:    prev.Hash,
		TimeStamp:   time.Now(),
		Data:        data,
		prev:        prev,
		HexPrevHash: hex.EncodeToString(prev.Hash[:]),
	}

	hash := block.ComputeHash()
	block.Hash = hash
	block.HexHash = hex.EncodeToString(hash[:])

	return block
}

func Genesis(t time.Time) *Block {
	block := &Block{
		Height:    0,
		PrevHash:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		TimeStamp: t,
		Data:      make([]Message, 0),
	}

	hash := block.ComputeHash()
	block.Hash = hash

	return block
}

var genesis = Genesis(time.Unix(0, 0))
