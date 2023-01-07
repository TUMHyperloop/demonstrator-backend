package main

import (
	"encoding/json"
	"flag"

	ads "github.com/beranek1/ads-bridge-go-lib"
	"github.com/beranek1/ginads"
	"github.com/beranek1/gindata"
	"github.com/beranek1/goconfig"
	"github.com/beranek1/godata"
	"github.com/beranek1/godatainterface"
	"github.com/gin-gonic/gin"
)

var configManager goconfig.ConfigManager
var dataStore godatainterface.DataStoreVersionedRangeFromInterval

var adsBackend *ginads.Backend
var dataStoreBackend *gindata.DataStoreBackend

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	adsBackend.AttachToRouter("/ads", r)
	dataStoreBackend.AttachToRouter("/data", r)

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

	return r
}

func main() {
	// Programm arguments
	var addr string
	var adsBridgeAddr string
	var configPath string
	var dataPath string

	// Set arguments
	flag.StringVar(&addr, "addr", ":8080", "target address of backend")
	flag.StringVar(&adsBridgeAddr, "bridge", "http://localhost:1234", "complete address of ADSBridge with protocol")
	flag.StringVar(&configPath, "config", "config", "path of config directory")
	flag.StringVar(&dataPath, "data", "data", "path of data directory")
	flag.Parse()

	adsBridge, err := ads.Connect(adsBridgeAddr)
	if err != nil {
		println("Error: Specified ADSBridge unavailable due to error: ", err.Error())
	}
	adsBackend = ginads.Create(adsBridge)

	configManager, err = goconfig.Manage("config")
	if err != nil {
		println("Error: goconfig failed managing config directory: ", err.Error())
	}

	dataStore, err = godata.Create("data")
	if err != nil {
		println("Error: godata failed managing data directory: ", err.Error())
	}
	dataStoreBackend = gindata.CreateDataStoreBackend(dataStore)

	r := setupRouter()

	r.Run(addr)
}
