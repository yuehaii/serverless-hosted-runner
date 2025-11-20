package main

import (
	common "serverless-hosted-runner/common"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

func main() {
	crypt := common.RSACryptography("")
	if err, _, _ := crypt.GenCertificate(nil, nil); err == nil {
		// if err, _, _ = crypt.GenCertificate(ca, caPri); err != nil {
		// 	   logrus.Errorf("Fail to gen leaf certs: %v", err)
		// } else {
        //     logrus.Infof("Finish generation of certificates")
        // }
		logrus.Infof("Finish generation ca certificates")
	} else {
		logrus.Errorf("Fail to gen ca certs: %v", err)
	}
}
