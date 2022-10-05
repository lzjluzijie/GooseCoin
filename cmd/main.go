package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	goosecoin "github.com/lzjluzijie/GooseCoin"
)

type Config struct {
	Addr  string
	Peers []string
}

func runWithConfig(config Config) {
	node := goosecoin.NewNode()

	r := gin.Default()
	r.GET("/head", func(c *gin.Context) {
		c.JSON(http.StatusOK, node.Head)
	})

	r.GET("/block/:n", func(c *gin.Context) {
		n, err := strconv.Atoi(c.Param("n"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, node.Blocks[n])
	})

	r.GET("/mine", func(c *gin.Context) {
		node.Mine()
		data, err := json.Marshal(node.Head)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		for _, peer := range config.Peers {
			_, err := http.Post(peer+"/newblock", "application/json", bytes.NewReader(data))
			if err != nil {
				log.Println(err.Error())
			}
		}
		c.JSON(http.StatusOK, node.Head)
	})

	r.GET("/message", func(c *gin.Context) {
		node.AddMessage([]byte(c.Query("data")))
		c.String(http.StatusOK, "OK")
	})

	r.POST("/newblock", func(c *gin.Context) {
		var block *goosecoin.Block
		c.BindJSON(&block)

		if !node.VerifyBlock(block) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block"})
			return
		}
		if block.Height != node.Head.Height+1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block height"})
			return
		}
		node.Blocks = append(node.Blocks, block)
		node.Head = block
		c.String(http.StatusOK, "OK")
	})

	r.POST("/sync", func(c *gin.Context) {
		var blocks []*goosecoin.Block
		c.BindJSON(&blocks)
		if len(blocks) <= len(node.Blocks) {
			c.String(http.StatusOK, "not longer")
			return
		}

		node.Blocks = blocks
		node.Head = blocks[len(blocks)-1]
		c.String(http.StatusOK, "OK")
	})

	err := http.ListenAndServe(config.Addr, r)
	if err != nil {
		panic(err)
	}
}

func main() {
	for _, configPath := range os.Args[1:] {
		configData, err := os.ReadFile(configPath)
		if err != nil {
			panic(err)
		}
		var config Config
		err = json.Unmarshal(configData, &config)
		if err != nil {
			panic(err)
		}
		go runWithConfig(config)
	}
	time.Sleep(99999999 * time.Second)
}
