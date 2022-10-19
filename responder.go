package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type response struct {
	Endpoint string `json:"endpoint"`
	Ip       string `json:"ip"`
	Counter  int    `json:"counter"`
	Status   int    `json:"status"`
	Message  string `json:"message"`
}

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func reply(c *gin.Context, b bool, s string) {
	code := http.StatusOK

	if !b {
		code = http.StatusInternalServerError
	}

	urlSource := c.FullPath()
	ipSource := getRealIp(c)
	hits := hitCounter()

	respMap := response{
		Endpoint: urlSource,
		Ip:       ipSource,
		Counter:  hits,
		Status:   code,
		Message:  s,
	}

	respByte, _ := json.Marshal(respMap)
	log.Println(string(respByte))
	exceptions := []string{"", "1pass", "con-test", "vt-release", "vt-upload", "amplify"}

	if !stringInSlice(strings.TrimLeft(c.FullPath(), "/"), exceptions) {
		c.JSON(code, respMap)
	}
}
