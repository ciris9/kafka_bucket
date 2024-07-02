package obs

import (
	huawei_obs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"kafka-bucket/config"
	"os"
)

var Client *huawei_obs.ObsClient
var err error

func init() {
	ak := os.Getenv("AccessKeyID")
	sk := os.Getenv("SecretAccessKey")
	Client, err = huawei_obs.New(ak, sk, config.ObsConfig.Endpoint, huawei_obs.WithSignature(huawei_obs.SignatureObs))
	if err != nil {
		panic(err)
	}
}
