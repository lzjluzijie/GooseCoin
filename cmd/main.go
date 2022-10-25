package main

import (
	"bytes"
	"crypto/ed25519"
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
	Addr       string
	Peers      []string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
	Validators []ed25519.PublicKey
}

func runWithConfig(config Config) {
	// node := goosecoin.NewNode()
	node := goosecoin.NewNodeWithKey(config.PublicKey, config.PrivateKey)
	node.Validators = config.Validators

	// randaoSync := func() {
	// 	seed := make([]byte, 32)
	// 	_, err := rand.Read(seed)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	hash := sha256.Sum256(seed)

	// }

	r := gin.Default()
	r.GET("/node", func(c *gin.Context) {
		c.JSON(http.StatusOK, node)
	})

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
		err := c.BindJSON(&block)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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

	randao := goosecoin.NewRandao(config.Validators)
	r.POST("/randao/hash", func(c *gin.Context) {
		var req goosecoin.HashRequest
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = randao.AddHash(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})
	r.POST("/randao/seed", func(c *gin.Context) {
		var req goosecoin.SeedRequest
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = randao.AddSeed(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
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
