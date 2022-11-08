package goosecoin

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	Addr       string
	Peers      []string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

type Server struct {
	Config ServerConfig
	*Node
	Randaos      map[string]*Randao
	NextProposer Validator

	r *gin.Engine
}

func (n *Network) NewServer(config ServerConfig) *Server {
	node := NewNodeWithKey(config.PublicKey, config.PrivateKey, n)

	r := gin.Default()
	s := &Server{
		Config:  config,
		Node:    node,
		Randaos: map[string]*Randao{},
		r:       r,
	}
	s.Randaos[RandaoID(1)] = s.NewRandao(RandaoID(1), s.OnRandaoFinish)

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
		s.Mine()
		c.JSON(http.StatusOK, node.Head)
	})

	r.GET("/message", func(c *gin.Context) {
		node.AddMessage(RawMessage([]byte(c.Query("data"))))
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
		time.Sleep(time.Second / 10)
		if s.Randaos[RandaoID(block.Height)].Status != RandaoStatusFinished {
			c.JSON(http.StatusBadRequest, gin.H{"error": "randao not finished"})
			return
		}
		if !bytes.Equal(block.Proposer.PublicKey, s.NextProposer.PublicKey) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid validator"})
			return
		}
		block.Attestations = NewAttestations()
		node.AddBlock(block)
		for _, peer := range s.Config.Peers {
			go func(peer string) {
				signature := ed25519.Sign(s.PrivateKey, block.Hash)
				req := &Attestation{
					Validator: Validator{PublicKey: s.PublicKey},
					BlockHash: block.Hash,
					Signature: signature,
				}
				data, err := json.Marshal(req)
				if err != nil {
					panic(err)
				}
				_, err = http.Post(peer+"/attestation", "application/json", bytes.NewReader(data))
				if err != nil {
					panic(err)
				}
			}(peer)
		}
		c.String(http.StatusOK, "OK")
	})

	r.POST("/attestation", func(c *gin.Context) {
		var req Attestation
		c.BindJSON(&req)
		// todo: verify attestation
		s.Head.Attestations.Add(req)
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
		if s.Randaos[req.ID] == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid randao id"})
			return
		}
		err = s.Randaos[req.ID].AddHash(req)
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
		if s.Randaos[req.ID] == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid randao id"})
			return
		}
		err = s.Randaos[req.ID].AddSeed(req)
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

func (s *Server) Mine() {
	block := s.Node.Mine()
	data, err := json.Marshal(block)
	if err != nil {
		panic(err)
	}
	for _, peer := range s.Config.Peers {
		resp, err := http.Post(peer+"/newblock", "application/json", bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != http.StatusOK {
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			panic(string(data))
		}
	}
}

func (s *Server) OnRandaoFinish(result []byte) {
	i := Mod(result, int64(len(s.network.Validators)))
	log.Println(i)
	s.NextProposer = s.network.Validators[i]
	next := RandaoID(s.Head.Height + 2)
	s.Randaos[next] = s.NewRandao(next, s.OnRandaoFinish)
	log.Println(next)
	go func() {
		time.Sleep(time.Second * 12)
		s.Randaos[next].SendHash()
	}()
	if bytes.Equal(s.NextProposer.PublicKey, s.Config.PublicKey) {
		s.Mine()
	}
}
