package main

import (
	"encoding/json"
	"os"

	ads "github.com/beranek1/ads-bridge-go-lib"
	"github.com/beranek1/goconfig"
	"github.com/beranek1/godata"
	"github.com/gin-gonic/gin"
)

var addr = ":8080"
var adsBridgeAddr = "http://localhost:1234"

var adsBridge ads.ADSBridge
var configManager goconfig.ConfigManager
var dataManager godata.DataManager

func returnADSResult(c *gin.Context, dat map[string]interface{}, err error) {
	if err != nil {
		c.String(500, "{\"error\":\""+err.Error()+"\"}")
	} else {
		byt, err := json.Marshal(dat)
		if err != nil {
			c.String(500, "{\"error\":\""+err.Error()+"\"}")
		} else {
			c.String(200, string(byt))
		}
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.GET("/ads/version", func(c *gin.Context) {
		dat, err := adsBridge.GetVersion()
		returnADSResult(c, dat, err)
	})

	r.GET("/ads/state", func(c *gin.Context) {
		dat, err := adsBridge.GetState()
		returnADSResult(c, dat, err)
	})

	r.POST("/ads/state", func(c *gin.Context) {
		rawData, err := c.GetRawData()
		if err == nil {
			var data map[string]interface{}
			err := json.Unmarshal(rawData, &data)
			if err == nil {
				adsStateData, hasADS := data["adsState"]
				deviceStateData, hasDevice := data["deviceState"]
				if hasADS {
					if val1, ok1 := adsStateData.(uint16); ok1 {
						adsState := uint16(val1)
						if hasDevice {
							if val2, ok2 := deviceStateData.(uint16); ok2 {
								deviceState := uint16(val2)
								dat, err := adsBridge.WriteControl(adsState, deviceState)
								returnADSResult(c, dat, err)
							} else {
								c.String(400, "{\"error\":\"Failed converting deviceState.\"}")
							}
						} else {
							dat, err := adsBridge.WriteControl(adsState, 0)
							returnADSResult(c, dat, err)
						}
					} else {
						c.String(400, "{\"error\":\"Failed converting adsState.\"}")
					}
				} else if hasDevice {
					if val1, ok1 := deviceStateData.(uint16); ok1 {
						deviceState := uint16(val1)
						dat, err := adsBridge.WriteControl(0, deviceState)
						returnADSResult(c, dat, err)
					} else {
						c.String(400, "{\"error\":\"Failed converting deviceState.\"}")
					}
				} else {
					dat, err := adsBridge.GetState()
					returnADSResult(c, dat, err)
				}
			} else {
				c.String(500, "{\"error\":\""+err.Error()+"\"}")
			}
		} else {
			c.String(500, "{\"error\":\""+err.Error()+"\"}")
		}
	})

	r.GET("/ads/deviceInfo", func(c *gin.Context) {
		dat, err := adsBridge.GetDeviceInfo()
		returnADSResult(c, dat, err)
	})

	r.GET("/ads/symbolInfo/:name", func(c *gin.Context) {
		name := c.Param("name")
		dat, err := adsBridge.GetSymbolInfo(name)
		returnADSResult(c, dat, err)
	})

	r.GET("/ads/symbolValue/:name", func(c *gin.Context) {
		name := c.Param("name")
		dat, err := adsBridge.GetSymbolValue(name)
		returnADSResult(c, dat, err)
	})

	r.POST("/ads/symbolValue/:name", func(c *gin.Context) {
		name := c.Param("name")
		rawData, err := c.GetRawData()
		if err == nil {
			var data map[string]interface{}
			err := json.Unmarshal(rawData, &data)
			if err == nil {
				if value, exists := data["data"]; exists {
					jsonValue, err := json.Marshal(value)
					if err == nil {
						dat, err := adsBridge.SetSymbolValue(name, string(jsonValue))
						returnADSResult(c, dat, err)
					} else {
						c.String(500, "{\"error\":\""+err.Error()+"\"}")
					}
				} else {
					dat, err := adsBridge.GetSymbolValue(name)
					returnADSResult(c, dat, err)
				}
			} else {
				c.String(500, "{\"error\":\""+err.Error()+"\"}")
			}
		} else {
			c.String(500, "{\"error\":\""+err.Error()+"\"}")
		}
	})

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

	configManager, err = goconfig.Manage("config")
	if err != nil {
		println("Error: goconfig failed managing config directory: ", err.Error())
	}

	dataManager, err = godata.Manage("data")
	if err != nil {
		println("Error: godata failed managing data directory: ", err.Error())
	}

	r := setupRouter()

	r.Run(addr)
}
