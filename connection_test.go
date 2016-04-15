package ucloud

import (
	"log"

	"backend/config"
)

var u *UcloudApiClient

func init() {
	if err := config.ParseGlobal("../../../../../config/sample.json"); err != nil {
		log.Fatal(err)
	}

	u = NewUcloudApiClient(
		config.Conf.Cloud.UcloudStorageDriver.BaseUrl,
		config.Conf.Cloud.UcloudStorageDriver.PublicKey,
		config.Conf.Cloud.UcloudStorageDriver.PrivateKey,
	)
}
