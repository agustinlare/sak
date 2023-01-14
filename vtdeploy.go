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
		c.HTML(
			http.StatusOK,
			"views/vt-deploy.html",
			gin.H{},
		)

	} else {
		var headers gin.H
		file := os.Getenv("UPLOADED_VT_FILE")
		path := getFileDir(file)
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
			Error.Println(err.Error())
			headers = gin.H{
				"error": err.Error(),
			}
		} else {
			headers = gin.H{
				"messages": "File uploaded " + getFileName(file),
			}
		}

		Info.Printf("File uploaded %s", file)
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
