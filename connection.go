// Package ucloud provides ...
package ucloud

import (
	"net/http"
	"time"
)

type UcloudApiClient struct {
	baseUrl    string
	publicKey  string
	privateKey string
	/*
		regionId   string
		zoneId     string
	*/
	conn *http.Client
}

func NewUcloudApiClient(baseUrl, publicKey, privateKey string) *UcloudApiClient {
	conn := &http.Client{Timeout: time.Minute}
	return &UcloudApiClient{baseUrl, publicKey, privateKey, conn}
}
