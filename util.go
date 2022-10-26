package goosecoin

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

func Uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

func Mod(data []byte, mod int64) int64 {
	b := new(big.Int).SetBytes(data)
	var ret big.Int
	ret.Mod(b, big.NewInt(mod))
	return ret.Int64()
}

func RandaoID(height uint64) string {
	return fmt.Sprintf("block-%d", height)
}
