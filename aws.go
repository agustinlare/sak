package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
)

type LambdaEvent struct {
	Name string `json:"name"`
}

type LambdaResponse struct {
	StatusCode int `json:"statusCode"`
	Body       struct {
		Result string `json:"result"`
		Error  string `json:"error"`
		Data   []struct {
			Item string `json:"item"`
		} `json:"data"`
	} `json:"body"`
}

func uploadToBucket(f, b, r string) error {
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

func actionHandler(a, e, u string) error {
	conf := getConfig()
	var rds_id string
	var ec2_id string
	var asg_id string
	var asg_config []int64

	if a == "start" {
		asg_config = []int64{15, 5, 11}
	} else {
		asg_config = []int64{0, 0, 0}
	}

	if e == "dev" {
		asg_id = conf.Lambda.Dev.Asc
		switch a {
		case "start":
			rds_id = conf.Lambda.Dev.RdsStart
			ec2_id = conf.Lambda.Dev.Ec2Start

		case "stop":
			rds_id = conf.Lambda.Dev.RdsStop
			ec2_id = conf.Lambda.Dev.Ec2Stop
		}
	} else if e == "qa" {
		asg_id = conf.Lambda.Qa.Asc

		switch a {
		case "start":
			rds_id = conf.Lambda.Qa.RdsStart
			ec2_id = conf.Lambda.Qa.Ec2Start
		case "stop":
			rds_id = conf.Lambda.Qa.RdsStop
			ec2_id = conf.Lambda.Qa.Ec2Stop
		}
	}

	// RDS Primero
	err := lambdaHandler(rds_id, u)
	if err != nil {
		return err
	}

	time.Sleep(180 * time.Second)

	// EC2 Segundo
	err = lambdaHandler(ec2_id, u)
	if err != nil {
		return err
	}

	// ASG Tercero
	err = asgUpdate(asg_id, asg_config)
	if err != nil {
		return err
	}

	return nil
}

func lambdaHandler(s, u string) error {
	var request LambdaEvent
	request.Name = u

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := lambda.New(sess, &aws.Config{Region: aws.String(os.Getenv("AWS_S3_REGION"))})
	payload, _ := json.Marshal(request)

	result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String(s), Payload: payload})
	if err != nil {
		return err
	}

	var resp LambdaResponse
	err = json.Unmarshal(result.Payload, &resp)
	if err != nil {
		return err
	}

	if resp.StatusCode != 0 {
		return errors.New("unexpected error")
	}

	if resp.Body.Result == "failure" {
		return errors.New("lambda failed execution")
	}

	return nil
}

func asgUpdate(a string, c []int64) error {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := autoscaling.New(sess)
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(a),
		MaxSize:              aws.Int64(c[0]),
		MinSize:              aws.Int64(c[1]),
		DesiredCapacity:      aws.Int64(c[2]),
	}

	_, err := svc.UpdateAutoScalingGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case autoscaling.ErrCodeScalingActivityInProgressFault:
				fmt.Println(autoscaling.ErrCodeScalingActivityInProgressFault, aerr.Error())
				return aerr
			case autoscaling.ErrCodeResourceContentionFault:
				fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
				return aerr
			case autoscaling.ErrCodeServiceLinkedRoleFailure:
				fmt.Println(autoscaling.ErrCodeServiceLinkedRoleFailure, aerr.Error())
				return aerr
			default:
				fmt.Println(aerr.Error())
				return aerr
			}
		} else {
			fmt.Println(err.Error())
			return aerr
		}
	}
	return nil
}
