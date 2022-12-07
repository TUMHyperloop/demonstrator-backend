package main

import (
	"encoding/json"
	"os"
	"strconv"

	ads "github.com/beranek1/ads-bridge-go-lib"
	"github.com/gin-gonic/gin"
)

var addr = ":8080"
var adsBridgeAddr = "http://localhost:1234"

var adsBridge ads.ADSBridge

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
		adsStateStr, hasADS := c.GetPostForm("adsState")
		deviceStateStr, hasDevice := c.GetPostForm("deviceState")
		if hasADS {
			val1, err1 := strconv.ParseUint(adsStateStr, 10, 16)
			if err1 != nil {
				c.String(400, "{\"error\":\""+err1.Error()+"\"}")
			} else {
				adsState := uint16(val1)
				if hasDevice {
					val2, err2 := strconv.ParseUint(deviceStateStr, 10, 16)
					if err2 != nil {
						c.String(400, "{\"error\":\""+err2.Error()+"\"}")
					} else {
						deviceState := uint16(val2)
						dat, err := adsBridge.WriteControl(adsState, deviceState)
						returnADSResult(c, dat, err)
					}
				} else {
					dat, err := adsBridge.WriteControl(adsState, 0)
					returnADSResult(c, dat, err)
				}
			}
		} else if hasDevice {
			val1, err1 := strconv.ParseUint(deviceStateStr, 10, 16)
			if err1 != nil {
				c.String(400, "{\"error\":\""+err1.Error()+"\"}")
			} else {
				deviceState := uint16(val1)
				dat, err := adsBridge.WriteControl(0, deviceState)
				returnADSResult(c, dat, err)
			}
		} else {
			dat, err := adsBridge.GetState()
			returnADSResult(c, dat, err)
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
		dataStr, hasData := c.GetPostForm("data")
		if hasData {
			dat, err := adsBridge.SetSymbolValue(name, dataStr)
			returnADSResult(c, dat, err)
		} else {
			dat, err := adsBridge.GetSymbolValue(name)
			returnADSResult(c, dat, err)
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

	r := setupRouter()

	r.Run(addr)
}
