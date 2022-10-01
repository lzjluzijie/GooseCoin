package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	goosecoin "github.com/lzjluzijie/GooseCoin"
)

func main() {
	addr := os.Args[1]
	node := goosecoin.NewNode()
	peers := make([]string, 0)

	r := gin.Default()

	r.GET("/block/:n", func(c *gin.Context) {
		n, err := strconv.Atoi(c.Param("n"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, node.Blocks[n])
	})

	r.GET("/mine", func(c *gin.Context) {
		node.Mine()
		data, err := json.Marshal(node.Blocks)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		for _, peer := range peers {
			_, err := http.Post(peer+"/sync", "application/json", bytes.NewReader(data))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"head": node.Head,
		})
	})

	r.GET("/message", func(c *gin.Context) {
		node.AddMessage([]byte(c.Query("data")))
		c.String(http.StatusOK, "OK")
	})

	r.POST("/sync", func(ctx *gin.Context) {
		var blocks []*goosecoin.Block
		ctx.BindJSON(&blocks)
		if len(blocks) <= len(node.Blocks) {
			ctx.String(http.StatusOK, "not longer")
			return
		}

		node.Blocks = blocks
		node.Head = blocks[len(blocks)-1]
		ctx.String(http.StatusOK, "OK")
	})

	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
