// Package cloud virtual switch component
package cloud

import (
	alisdk "github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	alicred "github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	alisvc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type IVSWitch interface {
	Info(string, string, string, string)
	LeftIPs() int64
}

type VSWitch struct {
	availableIP int64
	minIP       int64
	vsID        string
	key         string
	secret      string
	region      string
	client      *alisdk.Client
}

func CreateVSWitch() IVSWitch {
	return &VSWitch{0, 10, "", "", "", "", nil}
}

func (vs *VSWitch) Info(vsID, key, secret, region string) {
	vs.vsID = vsID
	vs.key = key
	vs.secret = secret
	vs.region = region
	vs.initClient()
	vs.getAttr()
}

func (vs VSWitch) LeftIPs() int64 {
	logrus.Infof("VSWitch %s, %d IPs left", vs.vsID, vs.availableIP)
	return vs.availableIP
}

func (vs VSWitch) infoRequest() *alisvc.DescribeVSwitchAttributesRequest {
	request := alisvc.CreateDescribeVSwitchAttributesRequest()
	request.VSwitchId = vs.vsID
	return request
}

func (vs VSWitch) infoResponse() *alisvc.DescribeVSwitchAttributesResponse {
	response := alisvc.CreateDescribeVSwitchAttributesResponse()
	return response
}

func (vs *VSWitch) initClient() {
	var err error
	credential, err := alicred.NewStaticAKCredentialsProviderBuilder().
		WithAccessKeyId(vs.key).WithAccessKeySecret(vs.secret).Build()
	if err != nil {
		logrus.Errorf("VSWitch NewStaticAKCredentialsProviderBuilder failure: %s", err)
		return
	}
	vs.client, err = alisdk.NewClientWithOptions(vs.region, alisdk.NewConfig(), credential)
	if err != nil {
		logrus.Errorf("VSWitch NewClientWithAccessKey failure: %s", err)
	}
}

func (vs VSWitch) getAttr() {
	resp := vs.infoResponse()
	err := vs.client.DoAction(vs.infoRequest(), resp)
	if err != nil {
		logrus.Errorf("VSWitch DoAction failure: %s", err)
	}
	vs.availableIP = resp.AvailableIpAddressCount
}
