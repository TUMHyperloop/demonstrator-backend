package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var addr = ":8080"
var adsBridgeAddr = "http://localhost:1234"

func adsProcessResponse(r io.Reader) (map[string]interface{}, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		return nil, err
	}
	return dat, nil
}

func adsBridgeGetRequest(path string) (map[string]interface{}, error) {
	var url = adsBridgeAddr + path
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return adsProcessResponse(resp.Body)
}

func adsBridgePostRequest(path string, jsonStr string) (map[string]interface{}, error) {
	var url = adsBridgeAddr + path
	resp, err := http.Post(url, "text/json", bytes.NewBufferString(jsonStr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return adsProcessResponse(resp.Body)
}

func adsGetVersion() (map[string]interface{}, error) {
	return adsBridgeGetRequest("/version")
}

func adsGetState() (map[string]interface{}, error) {
	return adsBridgeGetRequest("/state")
}

func adsGetDeviceInfo() (map[string]interface{}, error) {
	return adsBridgeGetRequest("/deviceInfo")
}

func adsGetSymbolInfo(name string) (map[string]interface{}, error) {
	return adsBridgeGetRequest("/getSymbolInfo/" + name)
}

func adsGetSymbolValue(name string) (map[string]interface{}, error) {
	return adsBridgeGetRequest("/getSymbolValue/" + name)
}

func adsSetSymbolValue(name string, value string) (map[string]interface{}, error) {
	return adsBridgePostRequest("/setSymbolValue/"+name, "{\"data\":"+value+"}")
}

func adsWriteControl(adsState uint16, deviceState uint16) (map[string]interface{}, error) {
	if adsState != 0 {
		if deviceState != 0 {
			return adsBridgePostRequest("/writeControl", "{\"adsState\":"+string(adsState)+","+"\"deviceState\":"+string(deviceState)+"}")
		} else {
			return adsBridgePostRequest("/writeControl", "{\"adsState\":"+string(adsState)+"}")
		}
	} else if deviceState != 0 {
		return adsBridgePostRequest("/writeControl", "{\"deviceState\":"+string(deviceState)+"}")
	}
	return adsBridgePostRequest("/writeControl", "{}")
}

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
		dat, err := adsGetVersion()
		returnADSResult(c, dat, err)
	})

	r.GET("/ads/state", func(c *gin.Context) {
		dat, err := adsGetState()
		returnADSResult(c, dat, err)
	})

	r.GET("/ads/deviceInfo", func(c *gin.Context) {
		dat, err := adsGetDeviceInfo()
		returnADSResult(c, dat, err)
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
