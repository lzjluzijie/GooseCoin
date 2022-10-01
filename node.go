package goosecoin

import (
	"crypto/ed25519"
	"reflect"
)

type Node struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey

	Head     *Block
	Blocks   []*Block
	Messages []Message
}

func NewNode() *Node {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	blocks := make([]*Block, 0)
	blocks = append(blocks, genesis)
	return &Node{
		publicKey:  publicKey,
		privateKey: privateKey,
		Head:       genesis,
		Blocks:     blocks,
		Messages:   make([]Message, 0),
	}
}

func (n *Node) VerifyBlock(block *Block) bool {
	if !reflect.DeepEqual(block.ComputeHash(), block.Hash) {
		return false
	}

	if !ed25519.Verify(block.Validator, block.Hash, block.Signature) {
		return false
	}
	return true
}

func (n *Node) AddMessage(m Message) {
	n.Messages = append(n.Messages, m)
}

func (n *Node) Mine() {
	block := NewBlock(n.Head, n.Messages)
	block.Validator = n.publicKey
	block.Signature = ed25519.Sign(n.privateKey, block.Hash[:])

	n.Head = block
	n.Blocks = append(n.Blocks, block)
	n.Messages = make([]Message, 0)
}
