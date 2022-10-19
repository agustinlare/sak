package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func vtExtract(c *gin.Context) {
	if c.Request.Method == "GET" {
		reply(c, true, "GET")
		c.HTML(
			http.StatusOK,
			"views/vt-deploy.html",
			gin.H{},
		)

	} else {
		var resp string
		var headers gin.H
		file := os.Getenv("UPLOADED_VT_FILE")
		path := getFileDir(file)
		fmt.Println(path)
		err := extractZipWithPassword(file, c.PostForm("upload_password"))

		if err == nil {
			os.Remove(file)
			files, _ := ioutil.ReadDir(path)

			for _, f := range files {
				err = uploadToBucket(
					path+"/"+f.Name(),
					os.Getenv("AWS_S3_BUCKET"),
					os.Getenv("AWS_S3_REGION"),
				)
			}
		}

		if err != nil {
			reply(c, false, err.Error())
			headers = gin.H{
				"error": err.Error(),
			}
		} else {
			reply(c, true, resp)
			headers = gin.H{
				"messages": "File uploaded " + getFileName(file),
			}
		}

		reply(c, true, file)
		c.HTML(
			http.StatusOK,
			"views/vt-deploy.html",
			headers,
		)
	}
}

func vtUpload(c *gin.Context) (string, error) {
	file, err := c.FormFile("file")
	dest := fmt.Sprintf("/tmp/%s", strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename)))

	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		os.RemoveAll(dest)
	}

	os.Mkdir(dest, 0755)
	file_path := dest + "/" + file.Filename

	if err != nil {
		log.Println(err)
	}

	err = c.SaveUploadedFile(file, file_path)

	return file_path, err
}

func extractZipWithPassword(s string, p string) error {
	path := getFileDir(s)
	commandString := fmt.Sprintf(`unzip -P %s %s -d %s`, p, s, path)
	commandSlice := strings.Fields(commandString)
	c := exec.Command(commandSlice[0], commandSlice[1:]...)
	err := c.Run()

	if err != nil {
		return err
	}

	return nil
}
