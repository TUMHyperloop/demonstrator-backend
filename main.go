package main

import (
	"encoding/json"
	"os"

	ads "github.com/beranek1/ads-bridge-go-lib"
	"github.com/beranek1/ginads"
	"github.com/beranek1/goconfig"
	"github.com/beranek1/godata"
	"github.com/gin-gonic/gin"
)

var addr = ":8080"
var adsBridgeAddr = "http://localhost:1234"

var adsBridge ads.ADSBridge
var configManager goconfig.ConfigManager
var dataManager godata.DataManager

var adsBackend *ginads.Backend

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	adsBackend.AttachToRouter("/ads", r)

	r.GET("/config/:name", func(c *gin.Context) {
		name := c.Param("name")
		var config map[string]interface{}
		err := configManager.Read(name, &config)
		if err == nil {
			json, err := json.Marshal(config)
			if err == nil {
				c.String(200, string(json))
			} else {
				c.String(500, "{\"error\":\""+err.Error()+"\"}")
			}
		} else {
			c.String(500, "{\"error\":\""+err.Error()+"\"}")
		}
	})

	r.POST("/config/:name", func(c *gin.Context) {
		name := c.Param("name")
		rawData, err := c.GetRawData()
		if err == nil {
			var config map[string]interface{}
			err := json.Unmarshal(rawData, &config)
			if err == nil {
				err := configManager.Write(name, config)
				if err == nil {
					c.String(200, string(rawData))
				} else {
					c.String(500, "{\"error\":\""+err.Error()+"\"}")
				}
			} else {
				c.String(500, "{\"error\":\""+err.Error()+"\"}")
			}
		} else {
			c.String(500, "{\"error\":\""+err.Error()+"\"}")
		}
	})

	r.GET("/data/:name", func(c *gin.Context) {
		name := c.Param("name")
		var data = dataManager.GetData(name)
		if data != nil {
			json, err := json.Marshal(data)
			if err == nil {
				c.String(200, string(json))
			} else {
				c.String(500, "{\"error\":\""+err.Error()+"\"}")
			}
		} else {
			c.String(404, "{\"error\":\"No data found for given key.\"}")
		}
	})

	return r
}

func main() {

	if len(os.Args) > 1 {
		addr = os.Args[2]
	}
	var err error
	adsBridge, err = ads.Connect(adsBridgeAddr)
	if err != nil {
		println("Error: Specified ADSBridge unavailable due to error: ", err.Error())
	}
	adsBackend = ginads.Create(adsBridge)

	configManager, err = goconfig.Manage("config")
	if err != nil {
		println("Error: goconfig failed managing config directory: ", err.Error())
		os.Exit(1)
	}

	dataManager, err = godata.Manage("data")
	if err != nil {
		println("Error: godata failed managing data directory: ", err.Error())
		os.Exit(1)
	}

	r := setupRouter()

	r.Run(addr)
}
