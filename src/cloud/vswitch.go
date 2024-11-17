package cloud

import (
	alisdk "github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	alisvc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type IVSWitch interface {
	Info(string, string, string, string)
	LeftIPs() int64
}

type VSWitch struct {
	available_ip int64
	min_ip       int64
	vs_id        string
	key          string
	secret       string
	region       string
	client       *alisdk.Client
}

func CreateVSWitch() IVSWitch {
	return &VSWitch{0, 10, "", "", "", "", nil}
}

func (vs *VSWitch) Info(vs_id, key, secret, region string) {
	vs.vs_id = vs_id
	vs.key = key
	vs.secret = secret
	vs.region = region
	vs.initClient()
	vs.getAttr()
}

func (vs VSWitch) LeftIPs() int64 {
	logrus.Infof("VSWitch %s, %d IPs left", vs.vs_id, vs.available_ip)
	return vs.available_ip
}

func (vs VSWitch) infoRequest() *alisvc.DescribeVSwitchAttributesRequest {
	request := alisvc.CreateDescribeVSwitchAttributesRequest()
	request.VSwitchId = vs.vs_id
	return request
}

func (vs VSWitch) infoResponse() *alisvc.DescribeVSwitchAttributesResponse {
	response := alisvc.CreateDescribeVSwitchAttributesResponse()
	return response
}

func (vs *VSWitch) initClient() {
	var err error
	vs.client, err = alisdk.NewClientWithAccessKey(vs.region, vs.key, vs.secret)
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
	vs.available_ip = resp.AvailableIpAddressCount
}
