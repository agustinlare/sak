package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type pong struct {
	Ping string `json:"ping"`
}

func health(c *gin.Context) {
	reply(c, true, "OK")
}

func unhealth(c *gin.Context) {
	reply(c, false, "ERROR")
}

func index(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"views/index.html",
		gin.H{
			"healthstatus": getHealth(),
			"links":        getLinks(),
		},
	)
}

func indexPost(c *gin.Context) {
	if c.PostForm("healthswitch-btn") != "" {
		switchHealth()
	}

	c.HTML(
		http.StatusOK,
		"views/index.html",
		gin.H{
			"healthstatus": getHealth(),
		},
	)
}

func ping(c *gin.Context) {
	resp := pong{Ping: "pong"}
	c.JSON(http.StatusOK, resp)
}

func rickroll(c *gin.Context) {
	c.Redirect(http.StatusFound, "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	reply(c, true, "RickRoll")
}

// Conectivity test
func conTest(c *gin.Context) {
	method := c.Request.Method

	if method == "GET" {
		c.HTML(
			http.StatusOK,
			"views/con-test.html",
			gin.H{},
		)
	}

	if method == "POST" {
		var resp string
		var err error
		var headers gin.H

		if c.PostForm("ipcheck") != "" {
			resp, err = isReachable(c.PostForm("ipcheck"))
		} else if c.PostForm("dns") != "" {
			resp, err = dnsResolver(c.PostForm("dns"))
		} else if c.PostForm("mongodb") != "" {
			resp, err = mongodb(c.PostForm("mongodb"))
		}

		if err != nil {
			reply(c, false, err.Error())
			headers = gin.H{
				"error": err,
			}

		} else {
			reply(c, true, resp)
			headers = gin.H{
				"messages": resp,
			}
		}

		c.HTML(
			http.StatusOK,
			"views/con-test.html",
			headers,
		)
	}
}

// VT Release Upload
func vtDropfile(c *gin.Context) {
	if c.Request.Method == "GET" {
		reply(c, true, "GET")
		c.HTML(
			http.StatusOK,
			"views/vt-deploy.html",
			gin.H{},
		)
	}

	if c.Request.Method == "POST" {
		resp, err := vtUpload(c)

		if err != nil {
			log.Println(err)
		}

		os.Setenv("UPLOADED_VT_FILE", resp)
		reply(c, true, "POST")
		c.HTML(
			http.StatusOK,
			"views/vt-deploy.html",
			gin.H{},
		)
	}
}
