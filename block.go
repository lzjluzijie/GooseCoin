package goosecoin

import (
	"bytes"
	"crypto/sha256"
	"time"
)

type Hash [32]byte

type Block struct {
	Hash      Hash
	Height    uint64
	PrevHash  Hash
	TimeStamp time.Time
	Data      []Message

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
	return sha256.Sum256(data.Bytes())
}

func NewBlock(prev *Block, Data []Message) *Block {
	block := &Block{
		Height:    prev.Height + 1,
		PrevHash:  prev.Hash,
		TimeStamp: time.Now(),
		prev:      prev,
	}

	hash := block.ComputeHash()
	block.Hash = hash

	return block
}

func Genesis() *Block {
	block := &Block{
		Height:    0,
		PrevHash:  [...]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		TimeStamp: time.Now(),
		Data:      make([]Message, 0),
	}

	hash := block.ComputeHash()
	block.Hash = hash

	return block
}
