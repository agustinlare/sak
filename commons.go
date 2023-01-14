package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Links struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

type Config struct {
	Token        string `json:"token"`
	Secretstring string `json:"secretstring"`
	Postgres     struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Dbname   string `json:"dbname"`
		Table    string `json:"table"`
	} `json:"postgres"`
	Lambda struct {
		Dev struct {
			Ec2Stop  string `json:"ec2_stop"`
			Ec2Start string `json:"ec2_start"`
			RdsStop  string `json:"rds_stop"`
			RdsStart string `json:"rds_start"`
			Asc      string `json:"asc"`
		} `json:"dev"`
		Qa struct {
			Ec2Stop  string `json:"ec2_stop"`
			Ec2Start string `json:"ec2_start"`
			RdsStop  string `json:"rds_stop"`
			RdsStart string `json:"rds_start"`
			Asc      string `json:"asc"`
		} `json:"qa"`
	} `json:"lambda"`
	Terraform []struct {
		Name   string `json:"name"`
		Varset string `json:"varset"`
	} `json:"terraform"`
	Links []struct {
		Name string `json:"name"`
		Href string `json:"href"`
	} `json:"links"`
}

func getRealIp(c *gin.Context) string {
	resp := c.GetHeader("X-FORWARDED-FOR")

	if len(resp) == 0 {
		return c.ClientIP()
	}

	return resp
}

func hitCounter() int {
	flag := os.Getenv("COUNTER_HIT_GOLANG")
	count, err := strconv.Atoi(flag)

	if err != nil {
		panic(err)
	}

	quick_math := count + 1
	newValue := strconv.Itoa(quick_math)
	os.Setenv("COUNTER_HIT_GOLANG", newValue)

	return quick_math
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getHealth() bool {
	boolValue, err := strconv.ParseBool(os.Getenv("HEALTHCHECK_STATUS"))

	if err != nil {
		log.Println(err)
	}

	return boolValue
}

func switchHealth() {
	os.Setenv("HEALTHCHECK_STATUS", strconv.FormatBool(!getHealth()))
}

func initEnvs() {
	check := []string{"HEALTHCHECK_STATUS", "COUNTER_HIT_GOLANG", "UPLOADED_VT_FILE", "AWS_S3_BUCKET", "AWS_S3_REGION"}

	for _, v := range check {
		_, present := os.LookupEnv(v)
		if !present {
			msg := fmt.Sprintf("Environments %s not set", v)
			log.Fatal(msg)
		}
	}
}

func getFileDir(s string) string {
	return filepath.Dir(s)
}

func getFileName(s string) string {
	return filepath.Base(s)
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func getConfig() Config {
	content, err := ioutil.ReadFile(os.Getenv("SAK_CONFIG_FILE"))

	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var payload Config
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	return payload
}

func getLinks() []Links {
	var resp []Links
	for _, v := range getConfig().Links {
		resp = append(resp, v)
	}

	return resp
}
