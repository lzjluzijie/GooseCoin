package goosecoin

type Message interface {
	Data() []byte
	Verify() bool
}

type RawMessage []byte

func (m RawMessage) Data() []byte {
	return m
}

func (m RawMessage) Verify() bool {
	return true
}
