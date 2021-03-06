package ucloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	httpUrlPrefix   = "www."
	ContentTypeJson = "application/json"
)

var defaultTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

type UcloudApiClient struct {
	publicKey  string
	privateKey string
	proxyURL   string
	baseHost   string
	conn       *http.Client
}

func NewUcloudApiClient(publicKey, privateKey, proxyURL string) *UcloudApiClient {
	host, err := bucketBaseHost(proxyURL)
	if err != nil {
		panic(err)
	}

	return &UcloudApiClient{publicKey,
		privateKey,
		proxyURL,
		host,
		&http.Client{
			Transport: defaultTransport,
			Timeout: time.Minute,
		},
	}
}

func signatureUFile(privateKey string, stringToSign string) string {
	mac := hmac.New(sha1.New, []byte(privateKey))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

type SignParam struct {
	HttpVerb                   string
	ContentMd5                 string
	ContentType                string
	Date                       string
	CanonicalizedUCloudHeaders string
	CanonicalizedResource      string
}

func (self SignParam) String() string {
	return self.HttpVerb + "\n" +
		self.ContentMd5 + "\n" +
		self.ContentType + "\n" +
		self.Date + "\n" +
		self.CanonicalizedUCloudHeaders +
		self.CanonicalizedResource
}

func (self *UcloudApiClient) genUFileAuth(param *SignParam) (authorization string) {
	return "UCloud" + " " + self.publicKey + ":" + signatureUFile(self.privateKey, fmt.Sprint(param))
}

type UcloudResponse struct {
	ContentLength int64
	ContentType   string
	ContentRange  string
	Etag          string
	StatusCode    int
	XsessionId    string
	RetCode       int
	ErrMsg        string
	Content       []byte
}

func (self *UcloudApiClient) HeadFile(fileName, bucketName string) (int64, bool, error) {
	len, exist, _, err := self.HeadFileWithEtag(fileName, bucketName)
	return len, exist, err
}

func (self *UcloudApiClient) GetFile(fileName, bucketName string) ([]byte, error) {
	resp, err := self.doHttpRequest(fileName, bucketName, "GET")
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, errors.New("content not found on ucloud")
	case http.StatusOK:
		return resp.Content, nil
	}
	return nil, fmt.Errorf("Internal Server Error, ucloud resp: %+v", *resp)
}

func (self *UcloudApiClient) PutFile(fileName, bucketName, contentType string, data []byte) error {
	resp, err := self.doHttpRequest(fileName, bucketName, "PUT", contentType, string(data))
	if err != nil || resp.StatusCode != http.StatusOK {
		time.Sleep(time.Second * 1)
		resp, err := self.doHttpRequest(fileName, bucketName, "PUT", contentType, string(data))
		if err != nil {
			return fmt.Errorf("Internal Server Error: %+v", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Internal Server Error, ucloud resp: %+v", resp)
		}
	}
	return nil
}

func parseHttpResp(httpResp *http.Response, httpVerb string) (*UcloudResponse, error) {
	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	resp := &UcloudResponse{}
	resp.ContentType = httpResp.Header.Get("Content-Type")
	resp.XsessionId = httpResp.Header.Get("X-SessionId")
	resp.Etag = httpResp.Header.Get("ETag")
	resp.StatusCode = httpResp.StatusCode
	resp.ContentLength = httpResp.ContentLength

	if resp.StatusCode == http.StatusOK {
		if httpVerb == "GET" {
			resp.Content = body
			return resp, nil
		}
		return resp, nil
	}
	if resp.StatusCode == http.StatusNotFound && httpVerb == "HEAD" {
		return resp, nil
	}

	// Only json content unmarshal the body
	if resp.ContentType == ContentTypeJson && resp.ContentLength > 0 {
		if err = json.Unmarshal(body, resp); err != nil {
			return nil, err
		}
	} else {
		resp.Content = body
	}

	return resp, nil
}

func (self *UcloudApiClient) doHttpRequest(fileName, bucketName, httpVerb string, args ...string) (*UcloudResponse, error) {
	var httpReq *http.Request
	var err error

	signParam := &SignParam{
		HttpVerb:              httpVerb,
		CanonicalizedResource: "/" + bucketName + "/" + fileName,
	}
	u := fmt.Sprintf("%s/%s", self.proxyURL, fileName)
	if httpVerb == "PUT" {
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments. Expected: %v, Got %v", 2, len(args))
		}
		contentType := args[0]
		data := []byte(args[1])
		signParam.ContentType = contentType
		httpReq, err = http.NewRequest(httpVerb, u, bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Add("Content-Type", contentType)
	} else {
		httpReq, err = http.NewRequest(httpVerb, u, nil)
		if err != nil {
			return nil, err
		}
	}
	httpReq.Host = fmt.Sprintf("%s.%s", bucketName, self.baseHost)
	httpReq.Header.Add("Authorization", self.genUFileAuth(signParam))

	httpResp, err := self.conn.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return parseHttpResp(httpResp, httpVerb)
}

func bucketBaseHost(proxyURL string) (string, error) {
	url, err := url.Parse(proxyURL)
	if err != nil {
		return "", err
	}

	host := url.Hostname()

	return strings.TrimLeft(host, httpUrlPrefix), nil
}

func (self *UcloudApiClient) HeadFileWithEtag(fileName, bucketName string) (int64, bool, string, error) {
	resp, err := self.doHttpRequest(fileName, bucketName, "HEAD")
	if err != nil {
		return 0, false, "", err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return resp.ContentLength, true, resp.Etag, nil
	case http.StatusNotFound:
		return 0, false, "", nil
	}
	return 0, false, "", fmt.Errorf("Internal Server Error, ucloud resp: %+v", *resp)
}
