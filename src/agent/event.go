package agent

import (
	"encoding/json"
	"fmt"
	common "serverless-hosted-runner/common"
	"strconv"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type IEventAgent interface {
	Agent
}

type IKafkaEventAgent interface {
	IEventAgent
}

type KafkaEvent struct {
	KafkaInstance
	matchEvents []common.PoolMsg
}

type KafkaInstance struct {
	create                 common.CreateRunner
	destroy                common.DestroyRunner
	config                 *kafka.ConfigMap
	producer               *kafka.Producer
	consumer               *kafka.Consumer
	topic                  string
	groupID                string
	bootstrapServer        string
	securityProtocol       string
	saslMechanism          string
	saslUserName           string
	saslPassword           string
	apiVerRequest          string
	autoOffsetReset        string
	heartbeatIntervalMs    string
	sessionTimeOutMS       string
	maxPollIntervalMS      string
	fetchMaxBytes          string
	maxPartitionFetchBytes string
	iv                     int
}

func CreateKafkaEventAgent(crt common.CreateRunner, des common.DestroyRunner,
	userName, pwd, endpoint, topic, consumer string, events []common.PoolMsg) IKafkaEventAgent {
	return &KafkaEvent{KafkaInstance{crt, des, &kafka.ConfigMap{}, nil, nil, topic, consumer, endpoint,
		"SASL_SSL", "PLAIN", userName, pwd, "true", "latest", "3000", "10000", "300000", "5120000", "1024000", 2}, events}
}

func (agent *KafkaEvent) InitAgent() {
	logrus.Warnf("InitAgent loading...")
	var err error
	if agent.producer, err = agent.initProducer(); err != nil {
		logrus.Errorf("fail to init kafka producer: %v", err)
	}
	if agent.consumer, err = agent.initConsumer(); err != nil {
		logrus.Errorf("fail to init kafka consumer: %v", err)
	}
}

func (agent *KafkaEvent) MonitorOnAgent() {
	logrus.Warnf("MonitorOnAgent loading...")
	if agent.InitAgent(); agent.consumer != nil {
		if err := agent.consumer.SubscribeTopics([]string{agent.topic}, nil); err != nil {
			logrus.Errorf("kafka event SubscribeTopics error: %v", err)
		}
		consumerClose := func() {
			if err := agent.consumer.Close(); err != nil {
				logrus.Errorf("fail to close consumer %v", err)
			}
		}
		defer consumerClose()
		for {
			event, err := agent.consumer.ReadMessage(-1)
			if err == nil {
				logrus.Warnf("kafka event consumer message on %s: %s\n", event.TopicPartition, string(event.Value))
				if err = agent.processEvent(event.Value); err != nil {
					logrus.Errorf("kafka event process error: %v (%v)\n", err, event)
				}
			} else {
				logrus.Errorf("kafka event consumer error: %v (%v)\n", err, event)
			}
			time.Sleep(time.Duration(agent.iv) * time.Second)
		}
	} else {
		logrus.Errorf("fail to create consumer, stop monitoring kafka topic %s", agent.topic)
	}
}

func (agent KafkaEvent) NotifyAgent(msg string) {
	// grpc is a better choice for bi-direction communication
}

func (agent KafkaEvent) initProducer() (*kafka.Producer, error) {
	return nil, nil
}

func (agent *KafkaEvent) addConsumerConfig(key string, val kafka.ConfigValue) {
	if err := agent.config.SetKey(key, val); err != nil {
		logrus.Errorf("fail to addConsumerConfig key %s, %v", key, err)
	}
}

func (agent *KafkaEvent) initConsumer() (*kafka.Consumer, error) {
	agent.addConsumerConfig("api.version.request", agent.apiVerRequest)
	agent.addConsumerConfig("auto.offset.reset", agent.autoOffsetReset)
	agent.addConsumerConfig("heartbeat.interval.ms", agent.heartbeatIntervalMs)
	agent.addConsumerConfig("session.timeout.ms", agent.sessionTimeOutMS)
	agent.addConsumerConfig("max.poll.interval.ms", agent.maxPollIntervalMS)
	agent.addConsumerConfig("fetch.max.bytes", agent.fetchMaxBytes)
	agent.addConsumerConfig("max.partition.fetch.bytes", agent.maxPartitionFetchBytes)
	agent.addConsumerConfig("bootstrap.servers", agent.bootstrapServer)
	agent.addConsumerConfig("group.id", agent.groupID)

	switch agent.securityProtocol {
	case "PLAINTEXT":
		agent.addConsumerConfig("security.protocol", "plaintext")
	case "SASL_SSL":
		agent.addConsumerConfig("security.protocol", "sasl_ssl")
		agent.addConsumerConfig("ssl.ca.location", "/go/bin/certs/kafka.ins.ca.cert")
		agent.addConsumerConfig("sasl.username", agent.saslUserName)
		agent.addConsumerConfig("sasl.password", agent.saslPassword)
		agent.addConsumerConfig("sasl.mechanism", agent.saslMechanism)
		agent.addConsumerConfig("ssl.endpoint.identification.algorithm", "None")
		agent.addConsumerConfig("enable.ssl.certificate.verification", "false")
	case "SASL_PLAINTEXT":
		agent.addConsumerConfig("security.protocol", "sasl_plaintext")
		agent.addConsumerConfig("sasl.username", agent.saslUserName)
		agent.addConsumerConfig("sasl.password", agent.saslPassword)
		agent.addConsumerConfig("sasl.mechanism", agent.saslMechanism)
	default:
		return nil, fmt.Errorf("kafka consumer init. unknown protocol")
	}

	consumer, err := kafka.NewConsumer(agent.config)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

func (agent *KafkaEvent) processEvent(rawEvent []byte) (err error) {
	jsonEvent := common.KafkaEvent{}
	if err = json.Unmarshal(rawEvent, &jsonEvent); err != nil {
		return err
	}
	eventType, match := agent.isMatchEvent(jsonEvent)
	if match && eventType == WfStatusQueued {
		runID := strconv.FormatFloat(jsonEvent.Data.Body.WorkflowJob.RunID, 'f', -1, 64)
		jobID := strconv.FormatFloat(jsonEvent.Data.Body.WorkflowJob.ID, 'f', -1, 64)
		logrus.Warnf("kafka event, creating runner %s, paras: %s, %s, %s, %s, %s",
			runID+"-"+jobID, jsonEvent.Data.Body.Repository.Name,
			jsonEvent.Data.Body.Repository.HTMLURL, jsonEvent.Data.Body.Repository.Owner.Login,
			jsonEvent.Data.Body.Repository.Owner.Login, jsonEvent.Data.Body.WorkflowJob.Labels)
		go agent.create(WfStatusQueued, runID+"-"+jobID, jsonEvent.Data.Body.Repository.Name,
			jsonEvent.Data.Body.Repository.HTMLURL, jsonEvent.Data.Body.Repository.Owner.Login,
			jsonEvent.Data.Body.Repository.Owner.Login, jsonEvent.Data.Body.WorkflowJob.Labels)
	} else if match && eventType == WfStatusCompleted {
		runID := strconv.FormatFloat(jsonEvent.Data.Body.WorkflowJob.RunID, 'f', -1, 64)
		jobID := strconv.FormatFloat(jsonEvent.Data.Body.WorkflowJob.ID, 'f', -1, 64)
		logrus.Warnf("kafka event, destroy runner %s, repo %s, org %s, label %s, url %s, runner id %s",
			jsonEvent.Data.Body.WorkflowJob.RunnerName, jsonEvent.Data.Body.Repository.Name,
			jsonEvent.Data.Body.Repository.Name, jsonEvent.Data.Body.WorkflowJob.Labels,
			jsonEvent.Data.Body.Repository.HTMLURL, runID+"-"+jobID)
		go agent.destroy(WfStatusCompleted, jsonEvent.Data.Body.WorkflowJob.RunnerName,
			jsonEvent.Data.Body.Repository.Name, jsonEvent.Data.Body.Repository.Owner.Login,
			runID+"-"+jobID, jsonEvent.Data.Body.WorkflowJob.Labels, jsonEvent.Data.Body.Repository.HTMLURL,
			jsonEvent.Data.Body.Repository.Owner.Login)
	}

	return nil
}

func (agent *KafkaEvent) isMatchEvent(event common.KafkaEvent) (string, bool) {
	if (event.Data.Body.Action == WfStatusQueued || event.Data.Body.Action == WfStatusCompleted) &&
		len(agent.matchEvents) > 0 {
		for _, item := range agent.matchEvents {
			logrus.Infof("matchEvents, repos: %s", item.Repos)
			repos := strings.Split(item.Repos, ",")
			url := item.URL
			for _, repo := range repos {
				if len(repo) > 0 {
					matchURL := url + "/" + repo
					logrus.Infof("matchEvents, matchURL: %s", matchURL)
					if matchURL == event.Data.Body.Repository.HTMLURL {
						return event.Data.Body.Action, true
					}
				}
			}
		}
	}
	return "", false
}
