package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"

	goosecoin "github.com/lzjluzijie/GooseCoin"
)

type Config struct {
	Validators []goosecoin.Validator
	Nodes      []goosecoin.ServerConfig
}

func main() {
	s := 32
	config := Config{
		Validators: []goosecoin.Validator{},
		Nodes:      []goosecoin.ServerConfig{},
	}

	for i := 0; i < s; i++ {
		publicKey, privateKey, err := ed25519.GenerateKey(nil)
		if err != nil {
			panic(err)
		}
		config.Validators = append(config.Validators, goosecoin.Validator{
			PublicKey: publicKey,
		})
		peers := []string{}
		for j := 0; j < s; j++ {
			if i != j {
				peers = append(peers, fmt.Sprintf("http://localhost:%d", 8000+j))
			}
		}
		config.Nodes = append(config.Nodes, goosecoin.ServerConfig{
			Addr:       fmt.Sprintf("localhost:%d", 8000+i),
			Peers:      peers,
			PublicKey:  publicKey,
			PrivateKey: privateKey,
		})
	}
	data, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("config.json", data, 0644)
	if err != nil {
		panic(err)
	}

}
