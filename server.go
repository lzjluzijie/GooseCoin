package goosecoin

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	Addr       string
	Peers      []string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
	Validators []ed25519.PublicKey
}

type Server struct {
	Config ServerConfig
	*Node
	Randao *Randao

	r *gin.Engine
}

func NewServer(config ServerConfig) *Server {
	node := NewNodeWithKey(config.PublicKey, config.PrivateKey)
	node.Validators = config.Validators

	r := gin.Default()
	s := &Server{
		Config: config,
		Node:   node,
		r:      r,
	}
	s.Randao = s.NewRandao()

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
		var block *Block
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
		var blocks []*Block
		c.BindJSON(&blocks)
		if len(blocks) <= len(node.Blocks) {
			c.String(http.StatusOK, "not longer")
			return
		}

		node.Blocks = blocks
		node.Head = blocks[len(blocks)-1]
		c.String(http.StatusOK, "OK")
	})

	r.POST("/randao/hash", func(c *gin.Context) {
		var req HashRequest
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = s.Randao.AddHash(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})
	r.POST("/randao/seed", func(c *gin.Context) {
		var req SeedRequest
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = s.Randao.AddSeed(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})

	return s
}

func (s *Server) Run() {
	err := http.ListenAndServe(s.Config.Addr, s.r)
	if err != nil {
		panic(err)
	}
}
