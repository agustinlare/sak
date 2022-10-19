package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	// home
	r.GET("/", index)
	r.POST("/", indexPost)
	r.GET("/health", health)
	r.GET("/unhealth", unhealth)
	r.GET("/ping", ping)
	r.GET("/1password", rickroll)
	// con-test
	r.GET("/con-test", conTest)
	r.POST("/con-test", conTest)
	// vt-deploy
	r.GET("/vt-release", vtDropfile)
	r.POST("/vt-release", vtDropfile)
	r.POST("/vt-upload", vtExtract)
	// Amplify
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
