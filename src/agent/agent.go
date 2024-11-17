package agent

import (
	"fmt"
	"time"

	common "serverless-hosted-runner/common"

	ali_mns "github.com/aliyun/aliyun-mns-go-sdk"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

var (
	DEFAULT_POOL_Q = "registration"
	NOTIFICATION_Q = "notification"
)

type Agent interface {
	InitAgent()
	MonitorOnAgent()
	NotifyAgent(msg string)
}

func CreateAliMNSAgent(url string, key string, secret string, q string, fn common.Mns_Process, fn_arg interface{}) Agent {
	return MnsQueue{url, key, secret, q, fn, fn_arg}
}

func createAzureServiceBusAgent(url string, connectionString string, q string) Agent {
	return nil
}

type MnsQueue struct {
	MnsUrl, AccessKey, AccessSecret, Queue string
	MsgProcess                             common.Mns_Process
	ParaProcess                            interface{}
}

func (q MnsQueue) InitAgent() {
	err := q.initQueue()
	if err != nil {
		logrus.Errorln(err)
	}
}
func (q MnsQueue) MonitorOnAgent() {
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
	client := ali_mns.NewAliMNSClient(q.MnsUrl, q.AccessKey, q.AccessSecret)
	queueManager := ali_mns.NewMNSQueueManager(client)

	err := queueManager.CreateQueue(q.Queue, 0, 65536, 604800, 30, 3, 3)
	if err != nil && !ali_mns.ERR_MNS_QUEUE_ALREADY_EXIST_AND_HAVE_SAME_ATTR.IsEqual(err) {
		logrus.Errorln(err)
	}
	if err == nil {
		time.Sleep(time.Duration(2) * time.Second)
	}

	queue := ali_mns.NewMNSQueue(q.Queue, client)
	mns_msg := ali_mns.MessageSendRequest{
		MessageBody:  msg,
		DelaySeconds: 0,
		Priority:     8}
	_, err = queue.SendMessage(mns_msg)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	return err
}

func (q MnsQueue) initQueue() ali_mns.AliMNSQueue {
	client := ali_mns.NewAliMNSClient(q.MnsUrl, q.AccessKey, q.AccessSecret)
	queueManager := ali_mns.NewMNSQueueManager(client)
	err := queueManager.CreateQueue(q.Queue, 0, 65536, 604800, 30, 3, 3)
	if err != nil && !ali_mns.ERR_MNS_QUEUE_ALREADY_EXIST_AND_HAVE_SAME_ATTR.IsEqual(err) {
		fmt.Println(err, q.MnsUrl, q.Queue)
	}
	if err == nil {
		time.Sleep(time.Duration(2) * time.Second)
	}
	newq := ali_mns.NewMNSQueue(q.Queue, client)
	return newq
}

func (q MnsQueue) listenMessage() error {
	if len(q.MnsUrl) == 0 {
		return nil
	}
	queue := q.initQueue()
	for {
		if q.MsgProcess(queue, q.ParaProcess) {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (q MnsQueue) activeMessage() int64 {
	client := ali_mns.NewAliMNSClient(q.MnsUrl, q.AccessKey, q.AccessSecret)
	queueManager := ali_mns.NewMNSQueueManager(client)
	attr, err := queueManager.GetQueueAttributes(q.Queue)
	if err != nil {
		fmt.Printf("error %s", err)
		return 0
	}
	return attr.ActiveMessages
}
