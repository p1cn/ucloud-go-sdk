package ucloud

import (
	"errors"
	"flag"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"

	"gitlab.p1staff.com/backend/tantan-backend-cloud/config"
)

var (
	u *UcloudApiClient

	confPath = flag.String("conf", "./config.json", "config file patch")
	bucket   = flag.String("bucket", "putong-test2-image-original", "bucket name, default : putong-test2-image-original")
)

var (
	letters     = []rune("abCdEfGhIjKlMnOpQrStUvWxYz")
	bucketName  = "putong-test2-image-original"
	contentType = "image/jpeg"
)

func init() {
	flag.Parse()

	if err := config.ParseServiceConfig("ufile_test", *confPath); err != nil {
		log.Fatal(err)
	}

	bucketName = *bucket

	if strings.TrimSpace(bucketName) == "" {
		log.Fatal("Empty Bucket Name")
	}

	u = NewUcloudApiClient(
		config.Conf.Cloud.UcloudStorageDriver.PublicKey,
		config.Conf.Cloud.UcloudStorageDriver.PrivateKey,
		config.Conf.Cloud.UcloudStorageDriver.ProxyURL.URL().String(),
	)
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
	err := u.PutFile(fileName, bucketName, contentType, data)
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
	err := u.PutFile(fileName, bucketName, contentType, data)
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
	err := u.PutFile(fileName, bucketName, contentType, data)
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
