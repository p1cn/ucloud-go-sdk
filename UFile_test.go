package ucloud

import (
	"errors"
	"log"
	"math/rand"
	"testing"
	"time"

	"backend/config"
)

var u *UcloudApiClient

func init() {
	pathToConfig := "../backend/config/sample.json"
	if err := config.ParseGlobal(pathToConfig); err != nil {
		log.Fatal(err)
	}

	u = NewUcloudApiClient(
		config.Conf.Cloud.UcloudStorageDriver.PublicKey,
		config.Conf.Cloud.UcloudStorageDriver.PrivateKey,
	)
}

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

func TestEnv(t *testing.T) {
	if u.privateKey == "" || u.publicKey == "" {
		t.Error("public/private key don't exist")
	}
}

func TestPutFile(t *testing.T) {
	fileName := randSeq(18) + ".jpg"
	data := []byte(randSeq(12))
	err := u.PutFile(fileName, bucketName, contentType, data, 1)
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}
}

func TestGetNonexistFile(t *testing.T) {
	fileName := randSeq(18) + ".jpg"
	_, err := u.GetFile(fileName, bucketName)
	if err == nil {
		t.Errorf("%+v", errors.New("shoud be not exist"))
		t.FailNow()
	}
}

func TestGetExistFile(t *testing.T) {
	fileName := randSeq(18) + ".jpg"
	data := []byte(randSeq(12))
	err := u.PutFile(fileName, bucketName, contentType, data, 1)
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}

	_, err = u.GetFile(fileName, bucketName)
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}
}

func TestHeadFileSucc(t *testing.T) {
	fileName := randSeq(18) + ".jpg"
	data := []byte(randSeq(12))
	err := u.PutFile(fileName, bucketName, contentType, data, 1)
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}

	time.Sleep(time.Second * 1)

	_, exist, err := u.HeadFile(fileName, bucketName)
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}
	if !exist {
		t.FailNow()
	}
}

func TestHeadFileFail(t *testing.T) {
	fileName := randSeq(18) + ".jpg"
	_, exist, err := u.HeadFile(fileName, bucketName)
	if err != nil {
		t.Errorf("%+v", err)
		t.FailNow()
	}
	if exist {
		t.FailNow()
	}
}
