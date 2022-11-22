package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

var addr = ":8080"

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	return r
}

func main() {

	if len(os.Args) > 1 {
		addr = os.Args[2]
	}

	r := setupRouter()

	r.Run(addr)
}
