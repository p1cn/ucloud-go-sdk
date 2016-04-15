package ucloud

import (
	"math/rand"
	"net/http"
	"testing"
	"time"
)

var letters []rune

var bucketName, contentType string

func init() {
	letters = []rune("abCdEfGhIjKlMnOpQrStUvWxYz")

	bucketName = "putong-test2-image-original"

	contentType = "image/jpeg"
}

func randSeq(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func givenPutFile(bucketName, fileName, contentType string) *putFile {
	s := &SignParam{
		HttpVerb:              "PUT",
		ContentType:           contentType,
		CanonicalizedResource: "/" + bucketName + "/" + fileName,
	}
	authorization := u.GenUFileAuth(s)
	data := []byte(randSeq(12))

	return NewPutFile(authorization, int64(len(data)), contentType, data)
}

func givenHeadFile(bucketName, fileName string) *headFile {
	s := &SignParam{
		HttpVerb:              "HEAD",
		CanonicalizedResource: "/" + bucketName + "/" + fileName,
	}
	authorization := u.GenUFileAuth(s)
	return NewHeadFileWithAuth(authorization)
}

func givenGetFile(bucketName, fileName string) *getFile {
	s := &SignParam{
		HttpVerb:              "GET",
		CanonicalizedResource: "/" + bucketName + "/" + fileName,
	}

	authorization := u.GenUFileAuth(s)
	return NewGetFileWithAuth(authorization)
}

func TestEnv(t *testing.T) {
	if u.privateKey == "" || u.publicKey == "" {
		t.Error("public/private key don't exist")
	}
}

func TestPutFile(t *testing.T) {
	fileName := randSeq(18) + ".jpg"
	r := givenPutFile(bucketName, fileName, contentType)
	err := r.DoHttpRequest(u, u.UploadURL(bucketName, fileName))
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}

	if r.StatusCode() != http.StatusOK {
		t.Errorf("%+v", r.R())
		t.FailNow()
	}
}

func TestGetNonexistFile(t *testing.T) {
	fileName := randSeq(18) + ".jpg"
	r := givenGetFile(bucketName, fileName)
	err := r.DoHttpRequest(u, u.DownloadURL(bucketName, fileName))
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}

	if r.StatusCode() != http.StatusNotFound {
		t.Errorf("%+v", r.R())
		t.FailNow()
	}
}

func TestGetExistFile(t *testing.T) {
	fileName := randSeq(18) + ".jpg"
	p := givenPutFile(bucketName, fileName, contentType)
	err := p.DoHttpRequest(u, u.UploadURL(bucketName, fileName))
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}

	r := givenGetFile(bucketName, fileName)
	err = r.DoHttpRequest(u, u.DownloadURL(bucketName, fileName))
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}

	if r.StatusCode() != http.StatusOK {
		t.Errorf("%+v", r.R())
		t.FailNow()
	}
}
func TestHeadFileSucc(t *testing.T) {
	fileName := randSeq(18) + ".jpg"

	r := givenPutFile(bucketName, fileName, contentType)
	_ = r.DoHttpRequest(u, u.DownloadURL(bucketName, fileName))

	time.Sleep(time.Second * 1)

	s := givenHeadFile(bucketName, fileName)
	err := s.DoHttpRequest(u, u.HeadURL(bucketName, fileName))
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}
}
func TestHeadFileFail(t *testing.T) {
	fileName := randSeq(18) + ".jpg"
	r := givenHeadFile(bucketName, fileName)
	err := r.DoHttpRequest(u, u.HeadURL(bucketName, fileName))
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}

	if r.StatusCode() != http.StatusNotFound {
		t.Errorf("%+v", r.R())
		t.FailNow()
	}
}
