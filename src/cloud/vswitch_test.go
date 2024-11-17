package cloud

import (
	"os"
	"testing"

	alisdk "github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	alisvc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/stretchr/testify/assert"
)

func TestLeftIPs(t *testing.T) {
	request := alisvc.CreateDescribeVSwitchAttributesRequest()
	request.VSwitchId = "vsw-uf69o7i8c0x8etrz3zjjt"
	response := alisvc.CreateDescribeVSwitchAttributesResponse()
	client, err := alisdk.NewClientWithAccessKey(os.Getenv("ALICLOUD_REGION"), os.Getenv("ALICLOUD_ACCESS_KEY"), os.Getenv("ALICLOUD_SECRET_KEY"))
	assert.Equal(t, err, nil)
	err = client.DoAction(request, response)
	assert.Equal(t, err, nil)
	assert.Greater(t, response.AvailableIpAddressCount, 0)
}
