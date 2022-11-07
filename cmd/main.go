package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	goosecoin "github.com/lzjluzijie/GooseCoin"
)

type Config struct {
	Validators []goosecoin.Validator
	Nodes      []goosecoin.ServerConfig
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.DefaultClient.Timeout = time.Second

	configData, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		panic(err)
	}

	network := &goosecoin.Network{
		Validators: config.Validators,
	}

	servers := make([]*goosecoin.Server, len(config.Nodes))
	for i, config := range config.Nodes {
		server := network.NewServer(config)
		go server.Run()
		servers[i] = server
	}
	for _, server := range servers {
		server.Randaos[goosecoin.RandaoID(1)].SendHash()
	}
	time.Sleep(99999999 * time.Second)
}
