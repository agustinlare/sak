package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func uploadToBucket(f string, b string, r string) error {
	session, err := session.NewSession(&aws.Config{Region: aws.String(r)})
	if err != nil {
		log.Println(err)
	}

	upFile, err := os.Open(f)
	if err != nil {
		log.Println(err)
	}

	defer upFile.Close()

	upFileInfo, _ := upFile.Stat()
	var fileSize int64 = upFileInfo.Size()
	fileBuffer := make([]byte, fileSize)
	upFile.Read(fileBuffer)

	if strings.Contains(f, "tmp") {
		f = strings.ReplaceAll(f, "/tmp/", "")
	}
	_, err = s3.New(session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(b),
		Key:                  aws.String(f),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(fileBuffer),
		ContentLength:        aws.Int64(fileSize),
		ContentType:          aws.String(http.DetectContentType(fileBuffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	return err
}
