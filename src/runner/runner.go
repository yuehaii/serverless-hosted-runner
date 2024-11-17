package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	agent "serverless-hosted-runner/agent"
	common "serverless-hosted-runner/common"
	"strconv"
	"strings"
	"time"

	ali_mns "github.com/aliyun/aliyun-mns-go-sdk"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type Runner interface {
	Monitor(obj interface{}, para interface{}) bool
	Init()
	Configure()
	Start() error
	Info() interface{}
}

type EciRunner struct {
	container_type    string
	runner_id         string
	runner_token      string
	runner_repo_url   string
	runner_org        string
	runner_repo_name  string
	runner_action     string
	runner_repo_owner string
	image_ver         string
	home_path         string
	token_fqdn        string
	token_path        string
	interval          int64
	pool_prefix       string
	entoken_fqdn      string
	git_id            string
	en_id             string
	tk                string
	ephemeral         string
	jit_enabled       bool
	jit_encoded       string
	jit_path          string
	pool_sufix        string
	remove_path       string
	default_labels    string
	sacrify_interval  int64
	sacrity_time      int64
	runner_group      string
}

func EciRunnerCreator(container_type string, runner_id string, runner_token string,
	runner_repo_url string, runner_org string, runner_repo_name string,
	runner_action string, runner_repo_owner string, image_ver string, runner_labels string,
	runner_group string) Runner {
	additional_labels := ""
	if len(runner_labels) > 0 && runner_labels != "none" {
		additional_labels = "," + runner_labels
	}
	return &EciRunner{container_type, runner_id, runner_token, runner_repo_url, runner_org,
		runner_repo_name, runner_action, runner_repo_owner, image_ver,
		"/go/bin/", "https://api.github.com/", "/actions/runners/registration-token",
		int64(10), agent.NOTIFICATION_Q, "https://git.build.ingka.ikea.com/api/v3/",
		"api.github.com", "git.build.ingka.ikea.com", "", "", false, "",
		"/actions/runners/generate-jitconfig", "-" + os.Getenv("runner"),
		"/actions/runners/remove-token", "serverless-hosted-runner,eci-runner" + additional_labels,
		10, 0, runner_group}
}

// TODO: Support Function runner
func FnRunnerCreator() Runner {
	return nil
}

func (runer *EciRunner) Init() {
	common.SetContextLogLevel(*ctx_log_level)
	runer.runner_init()
}

func (runer *EciRunner) Configure() {
	crypto := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))
	_ = runer.remove_runner(runer.home_path, runer.tk)
	if runer.runner_org != "none" {
		runer.tk, _ = runer.get_org_registration_token(crypto.DecryptMsg(runer.runner_token))
		_, _ = runer.config_runner(runer.runner_org, runer.runner_name(), runer.runner_repo_url)
	} else if runer.runner_repo_name != "" {
		runer.tk, _ = runer.get_repo_registration_token(crypto.DecryptMsg(runer.runner_token))
		_, _ = runer.config_runner(runer.runner_repo_name, runer.runner_name(), runer.runner_repo_url)
	}
}

func (runer EciRunner) runner_init() {
	logrus.Infof("Init runner...")
	cmd := exec.Command(runer.home_path + "init.sh")
	out, err := cmd.Output()
	logrus.Infof("Init runner cmd output is: %s", out)
	if err != nil {
		logrus.Errorf("Init runner error: %s", err)
	}
}

func (runer EciRunner) runner_name() string {
	if runer.container_type == "pool" {
		return runer.pool_prefix + "-" + runer.runner_org + runer.pool_sufix
	} else if runer.container_type == "org" {
		return runer.runner_org + "-" + runer.runner_id
	} else if runer.container_type == "repo" {
		return runer.runner_repo_name + "-" + runer.runner_id
	} else {
		return ""
	}
}

func (runer EciRunner) pool_name() string {
	return runer.pool_prefix + "-" + runer.runner_org
}

func (runer EciRunner) config_runner(label string, name string, url string) (string, error) {
	logrus.Infof("Configuring runner... url %s, name %s, label %s", url, name, label+","+runer.default_labels)
	cmd := exec.Command(runer.home_path+"config.sh", "--url", url, "--token", runer.tk, "--name", name,
		"--runnergroup", runer.runner_group,
		"--labels", label+","+runer.default_labels, "--work", "_work", "--replace",
		common.Ternary(runer.container_type == "pool", "", "--ephemeral").(string))
	logrus.Infof("Config runner cmd is: %s", cmd.String())
	out, err := cmd.Output()
	logrus.Infof("Config runner cmd output is: %s", string(out))
	if err != nil {
		logrus.Errorln(string(out), err)
		return fmt.Sprint(out), err
	}
	return fmt.Sprint(out), nil
}

func (runer *EciRunner) need_sacrifice() {
	for runer.container_type != "pool" {
		runcmd := exec.Command(runer.home_path+"sacrify.sh", strconv.FormatInt(runer.sacrity_time, 10))
		out, err := runcmd.Output()
		logrus.Infof("need_sacrifice cmd output is: %s, sacrity_time is: %d",
			string(out), runer.sacrity_time)
		if err != nil {
			logrus.Errorln(string(out), err)
		}
		runer.sacrity_time += runer.sacrify_interval
		time.Sleep(time.Duration(runer.sacrify_interval) * time.Second)
	}
}

func (runer EciRunner) Start() error {
	runcmd := exec.Command(runer.home_path + "run.sh")
	err := runcmd.Start()
	if err != nil {
		logrus.Errorf("Unable to start runner: %s", err)
		return err
	}
	go runer.need_sacrifice()
	return nil
}

func (runer EciRunner) Info() interface{} {
	logrus.Infof("Info tk %s", runer.tk)
	return runer
}

func (runer *EciRunner) Monitor(obj interface{}, para interface{}) bool {
	q, _ := obj.(ali_mns.AliMNSQueue)
	// rChan := make(chan EciRunner, 1)
	endChan := make(chan bool)
	respChan := make(chan ali_mns.MessageReceiveResponse)
	errChan := make(chan error)
	go func() {
		select {
		case resp := <-respChan:
			{
				logrus.Infof("runner received a msg: %s, name %s", resp.MessageBody, runer.runner_name())
				if strings.Compare(resp.MessageBody, runer.runner_name()) == 0 {
					logrus.Infof("Found msg... url %s", runer.runner_repo_url)
					if ret, e := q.ChangeMessageVisibility(resp.ReceiptHandle, 5); e != nil {
						fmt.Println(e)
					} else {
						if e := q.DeleteMessage(ret.ReceiptHandle); e != nil {
							fmt.Println(e)
						}

						crypto := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))

						pref := common.Ternary(strings.Contains(runer.runner_repo_url, runer.en_id),
							runer.entoken_fqdn, runer.token_fqdn).(string)
						url := common.Ternary(runer.runner_org != "none", pref+"orgs/"+runer.runner_org+runer.remove_path,
							pref+"repos/"+runer.runner_repo_owner+"/"+runer.runner_repo_name+runer.remove_path).(string)
						r_tk, _ := runer.get_tk(crypto.DecryptMsg(runer.runner_token), url, nil)
						runer.remove_runner(runer.home_path, r_tk)

						endChan <- true
					}
				}
				endChan <- false
			}
		case err := <-errChan:
			{
				if err != nil && !ali_mns.ERR_MNS_MESSAGE_NOT_EXIST.IsEqual(err) {
					logrus.Errorln(err)
				}
				endChan <- false
			}
		}
	}()
	q.ReceiveMessage(respChan, errChan, runer.interval)
	return <-endChan
}

func (runer EciRunner) remove_runner(path string, tk string) error {
	cmd := exec.Command(path+"config.sh", "remove", "--token", tk)
	out, err := cmd.Output()
	if err != nil {
		logrus.Errorf("Unable to remove runner with token - %s, err - %s, out - %s", tk, err, out)
		return err
	}
	logrus.Infof("Finish removing runner %s, out - %s", tk, out)
	return nil
}

func (runer EciRunner) parse_response(body io.Reader) (string, error) {
	data, _ := io.ReadAll(body)
	if runer.jit_enabled && runer.container_type != "pool" {
		jitToken := JitToken{}
		json.Unmarshal(data, &jitToken)
		logrus.Infof("parse_response JIT output is: %s", jitToken.EncodedJitConfig)
		return jitToken.EncodedJitConfig, nil
	} else {
		regToken := RunnerToken{}
		json.Unmarshal(data, &regToken)
		logrus.Infof("parse_response output is: %s", regToken.Token)
		return regToken.Token, nil
	}
}

func (runer EciRunner) get_tk(tk string, url string, body io.Reader) (string, error) {
	client := http.Client{
		Timeout: time.Duration(60 * time.Second),
	}
	request, _ := http.NewRequest("POST", url, body)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+tk)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "serverless-hosted-runner")
	if runer.container_type == "pool" {
		// prevent to generate same token
		t, err := strconv.Atoi(os.Getenv("runner"))
		time.Sleep(time.Duration(t) * time.Second)
		if err != nil {
			logrus.Errorf("strconv error: %s", err)
		}
	}
	resp, err := client.Do(request)
	if err != nil || resp.StatusCode != 201 {
		logrus.Errorf("Unable to get runner registrtation token %s, %d: %s", tk, resp.StatusCode, err)
		logrus.Errorln(resp)
		return "Unable to get runner registration token", err
	}
	defer resp.Body.Close()
	return runer.parse_response(resp.Body)
}

func (runer EciRunner) get_org_registration_token(tk string) (string, error) {
	pref := common.Ternary(strings.Contains(runer.runner_repo_url, runer.en_id), runer.entoken_fqdn, runer.token_fqdn).(string)
	if runer.jit_enabled && runer.container_type != "pool" {
		url := pref + "orgs/" + runer.runner_org + runer.jit_path
		body := "{\"name\":\"" + runer.runner_name() + "\", \"labels\":[\"" + runer.runner_org +
			"\"],\"runner_group_id\":1,\"work_folder\":\"_work\"}"
		logrus.Infof("Org jit url: %s, body: %s", url, body)
		return runer.get_tk(tk, url, strings.NewReader(body))
	} else {
		url := pref + "orgs/" + runer.runner_org + runer.token_path
		logrus.Infof("Org url: %s", url)
		return runer.get_tk(tk, url, nil)
	}
}

func (runer EciRunner) get_repo_registration_token(tk string) (string, error) {
	pref := common.Ternary(strings.Contains(runer.runner_repo_url, runer.en_id), runer.entoken_fqdn, runer.token_fqdn).(string)
	if runer.jit_enabled && runer.container_type != "pool" {
		url := pref + "repos/" + runer.runner_repo_owner + "/" + runer.runner_repo_name + runer.jit_path
		body := "{\"name\":\"" + runer.runner_name() + "\", \"labels\":[\"" + runer.runner_repo_name +
			"\"],\"runner_group_id\":1,\"work_folder\":\"_work\"}"
		logrus.Infof("Repo jit url: %s", url)
		return runer.get_tk(tk, url, strings.NewReader(body))
	} else {
		url := pref + "repos/" + runer.runner_repo_owner + "/" + runer.runner_repo_name + runer.token_path
		logrus.Infof("Repo url: %s", url)
		return runer.get_tk(tk, url, nil)
	}
}
