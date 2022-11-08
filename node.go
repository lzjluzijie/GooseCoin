package goosecoin

import (
	"crypto/ed25519"
	"reflect"
)

type Node struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey

	Head     *Block
	Blocks   []*Block
	Messages []Message

	network *Network
}

func NewNode(n *Network) *Node {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	return NewNodeWithKey(publicKey, privateKey, n)
}

func NewNodeWithKey(publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey, n *Network) *Node {
	blocks := make([]*Block, 0)
	blocks = append(blocks, genesis)
	return &Node{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Head:       genesis,
		Blocks:     blocks,
		Messages:   make([]Message, 0),
		network:    n,
	}
}

func (n *Node) Validator() Validator {
	return Validator{
		PublicKey: n.PublicKey,
	}
}

func (n *Node) VerifyBlock(block *Block) bool {
	if !reflect.DeepEqual(block.ComputeHash(), block.Hash) {
		return false
	}
	for _, m := range block.Data {
		if !m.Verify() {
			return false
		}
	}

	return ed25519.Verify(block.Proposer.PublicKey, block.Hash, block.Signature)
}

func (n *Node) AddMessage(m Message) {
	n.Messages = append(n.Messages, m)
}

func (n *Node) AddBlock(block *Block) {
	if !n.VerifyBlock(block) {
		panic("invalid block")
	}

	n.Head = block
	n.Blocks = append(n.Blocks, block)
	n.Messages = make([]Message, 0)
}

func (n *Node) Mine() *Block {
	block := NewBlock(n.Head, n.Messages)
	block.Proposer = n.Validator()
	block.Signature = ed25519.Sign(n.PrivateKey, block.Hash[:])
	n.AddBlock(block)
	return block
}
