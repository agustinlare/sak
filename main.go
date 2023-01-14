package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

var Info = log.New(os.Stdout, "\u001b[34mINFO: \u001B[0m", log.LstdFlags|log.Lshortfile)
var Warning = log.New(os.Stdout, "\u001b[33mWARNING: \u001B[0m", log.LstdFlags|log.Lshortfile)
var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)
var Debug = log.New(os.Stdout, "\u001b[36mDEBUG: \u001B[0m", log.LstdFlags|log.Lshortfile)

func Router(r *gin.Engine) {
	r.GET("/", index)
	r.POST("/", indexPost)
	r.GET("/health", health)
	r.GET("/unhealth", unhealth)
	r.GET("/ping", ping)
	r.GET("/1password", rickroll)
	r.POST("/awsscaling", invokeLambda)
	r.POST("/awsscaling/newuser", newUser)
	r.POST("/awsscaling/retrive", retrive)

	// Connectivity Page
	r.GET("/con-test", conTest)
	r.POST("/con-test", conTest)

	// Veritran Page
	r.GET("/vt-release", vtDropfile)
	r.POST("/vt-release", vtDropfile)
	r.POST("/vt-upload", vtExtract)

	// Amplify Vars from Terraform
	r.GET("/amplify", amplify)
	r.POST("/amplify", amplify)
}

func main() {
	initEnvs()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Static("/css", "./static/css")
	r.Static("/img", "./static/img")
	r.StaticFile("/favicon.ico", "./img/favicon.ico")
	r.LoadHTMLGlob("templates/**/*")

	Router(r)

	log.Println("Server started")
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
