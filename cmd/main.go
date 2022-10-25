package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	goosecoin "github.com/lzjluzijie/GooseCoin"
)

type Config struct {
	Nodes []goosecoin.ServerConfig
}

func main() {
	configData, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		panic(err)
	}

	servers := make([]*goosecoin.Server, len(config.Nodes))
	for i, config := range config.Nodes {
		server := goosecoin.NewServer(config)
		go server.Run()
		servers[i] = server
	}
	for _, server := range servers {
		log.Println(server)
		server.Randao.SendHash()
	}
	time.Sleep(99999999 * time.Second)
}
