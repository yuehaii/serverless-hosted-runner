// Package agent for external services
package agent

import (
	"os"
	"time"

	common "serverless-hosted-runner/common"

	ali_mns "github.com/aliyun/aliyun-mns-go-sdk"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

var (
	DefaultPoolQueue  = "registration"
	NotificationQueue = "notification"
)

type Agent interface {
	InitAgent()
	MonitorOnAgent()
	NotifyAgent(msg string)
}

func CreateAliMNSAgent(url string, key string, secret string, q string, fn common.MnsProcess, fnArg interface{}) Agent {
	return MnsQueue{url, key, secret, q, fn, fnArg, nil, false}
}

func CreateAzureServiceBusAgent(url string, connectionString string, q string) Agent {
	return nil
}

type MnsQueue struct {
	MnsURL, AccessKey, AccessSecret, Queue string
	MsgProcess                             common.MnsProcess
	ParaProcess                            interface{}
	queue                                  ali_mns.AliMNSQueue
	activeMsg                              bool
}

func (q MnsQueue) InitAgent() {
	err := q.initQueue()
	if err != nil {
		logrus.Errorln(err)
	}
}
func (q MnsQueue) MonitorOnAgent() {
	q.queue = q.initQueue()
	err := q.listenMessage()
	if err != nil {
		logrus.Errorln(err)
	}
}
func (q MnsQueue) NotifyAgent(msg string) {
	err := q.sendMsg(msg)
	if err != nil {
		logrus.Errorln(err)
	}
}

func (q MnsQueue) sendMsg(msg string) error {
	client := ali_mns.NewClient(q.MnsURL)
	queueManager := ali_mns.NewMNSQueueManager(client)

	err := queueManager.CreateQueue(q.Queue, 0, 65536, 604800, 30, 3, 3)
	if err != nil && !ali_mns.ERR_MNS_QUEUE_ALREADY_EXIST_AND_HAVE_SAME_ATTR.IsEqual(err) {
		logrus.Errorln(err)
	}
	if err == nil {
		time.Sleep(time.Duration(2) * time.Second)
	}

	queue := ali_mns.NewMNSQueue(q.Queue, client)
	mnsMsg := ali_mns.MessageSendRequest{
		MessageBody:  msg,
		DelaySeconds: 0,
		Priority:     8}
	_, err = queue.SendMessage(mnsMsg)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	return err
}

func (q MnsQueue) initQueue() ali_mns.AliMNSQueue {
	if err := os.Setenv(ali_mns.AliyunAkEnvKey, q.AccessKey); err != nil {
		logrus.Errorf("fail to set env %s, %v", ali_mns.AliyunAkEnvKey, err)
	}
	if err := os.Setenv(ali_mns.AliyunSkEnvKey, q.AccessSecret); err != nil {
		logrus.Errorf("fail to set env %s, %v", ali_mns.AliyunSkEnvKey, err)
	}
	client := ali_mns.NewClient(q.MnsURL)
	queueManager := ali_mns.NewMNSQueueManager(client)
	err := queueManager.CreateQueue(q.Queue, 0, 65536, 604800, 30, 3, 3)
	if err != nil && !ali_mns.ERR_MNS_QUEUE_ALREADY_EXIST_AND_HAVE_SAME_ATTR.IsEqual(err) {
		logrus.Println(err, q.MnsURL, q.Queue)
	}
	if err == nil {
		time.Sleep(time.Duration(2) * time.Second)
	}
	newq := ali_mns.NewMNSQueue(q.Queue, client)
	return newq
}

func (q MnsQueue) listenMessage() error {
	if len(q.MnsURL) == 0 {
		return nil
	}
	if q.activeMsg {
		q.activeMessage()
	}
	for !q.MsgProcess(q.queue, q.ParaProcess) {
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (q MnsQueue) activeMessage() int64 {
	client := ali_mns.NewClient(q.MnsURL)
	queueManager := ali_mns.NewMNSQueueManager(client)
	attr, err := queueManager.GetQueueAttributes(q.Queue)
	if err != nil {
		logrus.Printf("error %s", err)
		return 0
	}
	return attr.ActiveMessages
}
