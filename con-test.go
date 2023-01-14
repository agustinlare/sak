package main

import (
	"context"
	"net"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func isReachable(s string) (string, error) {
	if !strings.Contains(s, ":") {
		s = s + ":443"
	}

	timeout := 5 * time.Second
	_, err := net.DialTimeout("tcp", s, timeout)

	if err != nil {
		return "Site unreachable, check logs", err
	}

	return "Connection established", nil
}

func dnsResolver(s string) (string, error) {
	ips, err := net.LookupIP(s)

	if err != nil {
		return "Cloud not retrive any IP for the DNS", err
	}

	var m []string
	var resp string

	for _, ip := range ips {
		m = append(m, ip.String()+", ")
		resp = strings.TrimLeft(strings.Join(m, ip.String()), ", ")
	}

	return resp, nil
}

func mongodb(uri string) (string, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		return "Unable to resolve uri, check the connection string and try again", err
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		return "Unable to connect to MongoDB, check logs", err
	} else {
		return "Connection established", nil
	}
}
