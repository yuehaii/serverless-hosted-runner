package cloud

import (
	"os"
	"testing"

	alisdk "github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	alicred "github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	alisvc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLeftIPs(t *testing.T) {
	request := alisvc.CreateDescribeVSwitchAttributesRequest()
	request.VSwitchId = "vsw-uf69o7i8c0x8etrz3zjjt"
	response := alisvc.CreateDescribeVSwitchAttributesResponse()

	var err error
	credential, err := alicred.NewStaticAKCredentialsProviderBuilder().
		WithAccessKeyId(os.Getenv("ALICLOUD_ACCESS_KEY")).WithAccessKeySecret(os.Getenv("ALICLOUD_SECRET_KEY")).Build()
	if err != nil {
		logrus.Errorf("VSWitch NewStaticAKCredentialsProviderBuilder failure: %s", err)
		return
	}
	client, err := alisdk.NewClientWithOptions(os.Getenv("ALICLOUD_REGION"), alisdk.NewConfig(), credential)
	if err != nil {
		logrus.Errorf("VSWitch NewClientWithAccessKey failure: %s", err)
	}
	assert.Equal(t, err, nil)
	err = client.DoAction(request, response)
	assert.Equal(t, err, nil)
	assert.Greater(t, response.AvailableIpAddressCount, 0)
}
