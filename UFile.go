package ucloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (u *UcloudApiClient) SignatureUFile(privateKey string, stringToSign string) string {
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

func (param *SignParam) String() string {
	return param.HttpVerb + "\n" +
		param.ContentMd5 + "\n" +
		param.ContentType + "\n" +
		param.Date + "\n" +
		param.CanonicalizedUCloudHeaders +
		param.CanonicalizedResource
}

func (u *UcloudApiClient) GenUFileAuth(param *SignParam) (authorization string) {
	return "UCloud" + " " + u.publicKey + ":" + u.SignatureUFile(u.privateKey, fmt.Sprint(param))
}

func (u *UcloudApiClient) DownloadURL(bucketName, fileName string) (url string) {
	return u.HeadURL(bucketName, fileName)
}

func (u *UcloudApiClient) UploadURL(bucketName, fileName string) (url string) {
	return "http://" + bucketName + ".ufile.ucloud.cn" + "/" + fileName
}

func (u *UcloudApiClient) HeadURL(bucketName, fileName string) (url string) {
	return "http://" + bucketName + ".ufile.ucloud.com.cn" + "/" + fileName
}

// Head File
type headFileResponse struct {
	contentLength int64
	contentType   string
	contentRange  string
	etag          string
	statusCode    int

	xsessionId string //err header

	RetCode int    // err
	ErrMsg  string //err
}

type headFile struct {
	authorization string
	resp          *headFileResponse
}

func NewHeadFileWithAuth(authorization string) *headFile {
	return &headFile{
		authorization: authorization,
		resp:          &headFileResponse{},
	}
}

func NewHeadFile() *headFile {
	return &headFile{
		resp: &headFileResponse{},
	}
}

// change the name later
func (self *headFile) R() *headFileResponse {
	return self.resp
}

func (self *headFile) StatusCode() int {
	return self.resp.statusCode
}

func (self *headFile) ContentLength() int64 {
	return self.resp.contentLength
}

func (self *headFile) DoHttpRequest(apiClient *UcloudApiClient, uri string) error {
	httpReq, err := http.NewRequest("HEAD", uri, nil)
	if err != nil {
		return err
	}

	if self.authorization != "" {
		httpReq.Header.Add("Authorization", self.authorization)
	}
	httpResp, err := apiClient.conn.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	self.resp.statusCode = httpResp.StatusCode
	self.resp.contentLength = httpResp.ContentLength
	self.resp.contentType = httpResp.Header.Get("Content-Type")

	if self.resp.statusCode == http.StatusOK || self.resp.statusCode == http.StatusNotFound {
		self.resp.etag = httpResp.Header.Get("ETag")
		return nil
	}

	self.resp.xsessionId = httpResp.Header.Get("X-SessionId")
	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, self.resp)
}

// Download file
type getFileResponse struct {
	contentLength int64
	contentType   string
	contentRange  string
	etag          string
	statusCode    int
	//range_        string `ucloud:"header,optional" rename:"Range"` // feipian

	xsessionId string // err header
	ErrMsg     string //err
	RetCode    int    // err

	content []byte
}

type getFile struct {
	authorization string
	//ifModifySince string `ucloud:"header,optional" rename:"If-Modify-Since"`
	//range_        string `ucloud:"header,optional" rename:"Range"`
	resp *getFileResponse
}

func NewGetFile() *getFile {
	return &getFile{
		resp: &getFileResponse{},
	}
}
func NewGetFileWithAuth(authorization string) *getFile {
	return &getFile{
		authorization: authorization,
		resp:          &getFileResponse{},
	}
}
func (self *getFile) R() *getFileResponse {
	return self.resp
}

func (self *getFile) StatusCode() int {
	return self.resp.statusCode
}

func (self *getFile) Data() []byte {
	return self.resp.content
}

func (self *getFile) DoHttpRequest(apiClient *UcloudApiClient, uri string) error {
	httpReq, err := http.NewRequest("Get", uri, nil)
	if err != nil {
		return err
	}
	if self.authorization != "" {
		httpReq.Header.Add("Authorization", self.authorization)
	}
	httpResp, err := apiClient.conn.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	self.resp.statusCode = httpResp.StatusCode
	self.resp.contentLength = httpResp.ContentLength
	self.resp.contentType = httpResp.Header.Get("Content-Type")
	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if self.resp.statusCode == http.StatusOK {
		self.resp.etag = httpResp.Header.Get("ETag")

		self.resp.content = body
		return nil
	}
	self.resp.xsessionId = httpResp.Header.Get("X-SessionId")
	return json.Unmarshal(body, self.resp)
}

// Upload file
type putFileResponse struct {
	contentLength int64
	contentType   string
	etag          string
	statusCode    int

	xsessionId string // err header
	ErrMsg     string // err
	RetCode    int    // err

}

type putFile struct {
	authorization string
	contentLength int64
	contentType   string
	contentMD5    string

	data []byte

	resp *putFileResponse
}

func (self *putFile) StatusCode() int {
	return self.resp.statusCode
}

func (self *putFile) R() *putFileResponse {
	return self.resp
}

func NewPutFile(authorization string, contentLength int64, contentType string, data []byte) *putFile {
	return &putFile{
		authorization: authorization,
		contentLength: contentLength,
		contentType:   contentType,
		data:          data,

		resp: &putFileResponse{},
	}
}

func (self *putFile) DoHttpRequest(apiClient *UcloudApiClient, uri string) error {
	httpReq, err := http.NewRequest("PUT", uri, bytes.NewBuffer(self.data))
	if err != nil {
		return err
	}
	httpReq.Header.Add("Authorization", self.authorization)
	httpReq.Header.Add("Content-Type", self.contentType)
	httpResp, err := apiClient.conn.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	self.resp.statusCode = httpResp.StatusCode
	self.resp.contentLength = httpResp.ContentLength
	self.resp.contentType = httpResp.Header.Get("Content-Type")

	if self.resp.statusCode == http.StatusOK {
		self.resp.etag = httpResp.Header.Get("ETag")
		return nil
	}
	self.resp.xsessionId = httpResp.Header.Get("X-SessionId")
	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, self.resp)
}
