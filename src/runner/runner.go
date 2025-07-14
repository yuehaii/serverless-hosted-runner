package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	agent "serverless-hosted-runner/agent"
	common "serverless-hosted-runner/common"
	listener "serverless-hosted-runner/network/grpc"
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
	remote_dockerd    bool
	max_idle          int64
	sys_ctl           common.ISysCtl
	retry_times       int
	retry_backoff     int64
	pr                string
	dis_ip            string
	tk_invalid        int64
}

func EciRunnerCreator(container_type string, runner_id string, runner_token string,
	runner_repo_url string, runner_org string, runner_repo_name string,
	runner_action string, runner_repo_owner string, image_ver string, runner_labels string,
	runner_group string, cloud_pr string, dis_ip string) Runner {
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
		10, 0, runner_group, false, 300, common.CreateUnixSysCtl(), 3, 20, cloud_pr, dis_ip, 3600}
}

// TODO: Support Function runner
func FnRunnerCreator() Runner {
	return nil
}

func (runner *EciRunner) Init() {
	common.SetContextLogLevel(*ctx_log_level)
	runner.runnerInit()
}

func (runner *EciRunner) Configure() {
	runnerConfig := func() (err error) {
		_ = runner.removeRunner(runner.home_path, runner.tk, false)
		if err := runner.refreshTk(); err != nil {
			logrus.Errorf("%s - fail to gen tk for config, %v", runner.runner_id, err)
			return err
		} else {
			if runner.runner_org != "none" {
				out, err := runner.configRunner(runner.runner_org, runner.runnerName(), runner.runner_repo_url)
				if err != nil {
					logrus.Errorf("%s - fai to config runner, %v, err:%v", runner.runner_id, out, err)
				}
				return err
			} else if runner.runner_repo_name != "" {
				out, err := runner.configRunner(runner.runner_repo_name, runner.runnerName(), runner.runner_repo_url)
				if err != nil {
					logrus.Errorf("%s - fai to config runner, %v, err:%v", runner.runner_id, out, err)
				}
				return err
			}
		}
		return errors.New("unknow runner type for config")
	}
	times := 0
	for times < runner.retry_times {
		if runnerConfig() == nil {
			break
		}
		times += 1
		time.Sleep(time.Duration(runner.retry_backoff) * time.Second)
		logrus.Warnf("%s - fail to config runner url: %s, label: %s, retry %v time", runner.runner_id, runner.runner_repo_url, runner.default_labels, times)
	}
}

func (runner EciRunner) runnerInit() {
	logrus.Infof("Init runner...")
	if runner.remote_dockerd {
		cmd := exec.Command(runner.home_path + "_depsh/init.sh")
		out, err := cmd.Output()
		logrus.Infof("Init runner cmd output is: %s", out)
		if err != nil {
			logrus.Errorf("%s - Init runner error: %s", runner.runner_id, err)
		}
	} else {
		if runner.pr == "ali" {
			runner.sys_ctl.DockerStorageDriver("overlay2")
		}
		go runner.sys_ctl.StartProcess("dockerd")
		go runner.sys_ctl.SetResolvers()
	}
}

func (runner EciRunner) runnerName() string {
	if runner.container_type == "pool" {
		return runner.pool_prefix + "-" + runner.runner_org + runner.pool_sufix
	} else if runner.container_type == "org" {
		return runner.runner_org + "-" + runner.runner_id
	} else if runner.container_type == "repo" {
		return runner.runner_repo_name + "-" + runner.runner_id
	} else {
		return ""
	}
}

func (runner EciRunner) poolName() string {
	return runner.pool_prefix + "-" + runner.runner_org
}

func (runner EciRunner) configRunner(label string, name string, url string) (string, error) {
	logrus.Infof("Configuring runner... url %s, name %s, label %s", url, name, label+","+runner.default_labels)
	cmd := exec.Command(runner.home_path+"config.sh", "--url", url, "--token", runner.tk, "--name", name,
		"--runnergroup", runner.runner_group,
		"--labels", label+","+runner.default_labels, "--work", "_work",
		common.Ternary(runner.container_type == "pool", "", "--ephemeral").(string))
	out, err := cmd.Output()
	logrus.Infof("Config runner cmd %s, output is: %s", cmd.String(), string(out))
	if err != nil {
		logrus.Errorln(runner.runner_id+" - "+string(out), err)
		return fmt.Sprint(out), err
	} else if !strings.Contains(string(out), "successful") {
		msg := "fail to register runner: " + string(out)
		logrus.Errorln(runner.runner_id + " - " + msg)
		return fmt.Sprint(out), errors.New(msg)
	}
	return fmt.Sprint(out), nil
}

func (runner *EciRunner) refreshTk() (err error) {
	crypto := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))
	if runner.runner_org != "none" {
		runner.tk, err = runner.getOrgRegToken(crypto.DecryptMsg(runner.runner_token))
		if err != nil {
			logrus.Errorf("%s - fail to generate org access token, %v", runner.runner_id, err)
			return err
		}
	} else if runner.runner_repo_name != "" {
		runner.tk, err = runner.getRepoRegToken(crypto.DecryptMsg(runner.runner_token))
		if err != nil {
			logrus.Errorf("%s - fail to generate repo access token, %v", runner.runner_id, err)
			return err
		}
	}
	return nil
}

func (runner EciRunner) idleDetection(work_log string) {
	if strings.Contains(strings.ToLower(work_log), "job completed") ||
		(runner.sacrity_time > runner.max_idle &&
			!strings.Contains(strings.ToLower(work_log), "job message")) {
		logrus.Warnf(common.Ternary(strings.Contains(strings.ToLower(work_log), "job completed"),
			"workflow finished. complete runner.", "sacrify cur_idle: "+strconv.Itoa(int(runner.sacrity_time))+
				", max_idle: "+strconv.Itoa(int(runner.max_idle))).(string))
		if runner.sacrity_time > runner.tk_invalid {
			if err := runner.refreshTk(); err != nil {
				logrus.Errorf("%s - fail to refresh token for remove, %v", runner.runner_id, err)
				return
			}
		}
		runner.removeRunner(runner.home_path, runner.tk, true)
		os.Exit(0)
	} else if strings.Contains(strings.ToLower(work_log), "job message") {
		logrus.Infof("wf still running. idle %v", runner.sacrity_time)
	} else {
		logrus.Infof("wf dose not select the runner temporary. idle %v", runner.sacrity_time)
	}
}

func (runner *EciRunner) sacrifyCtl() {
	logrus.Infof("sacrify ctl start")
	entries, err := os.ReadDir(runner.home_path + "_diag/")
	if err == nil {
		worker_exist := false
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "Worker_") {
				f_b, err := os.ReadFile(runner.home_path + "_diag/" + entry.Name())
				if err != nil {
					logrus.Warnf("%s - fail to read the work log %s, %s", runner.runner_id, entry.Name(), err)
					continue
				}
				worker_exist = true
				runner.idleDetection(string(f_b))
			}
		}
		if !worker_exist {
			logrus.Warnf("%s - Worker_ dose not exist temporary", runner.runner_id)
			runner.idleDetection("")
		}
	} else {
		logrus.Errorf("%s - fail to list diag dir, %v", runner.runner_id, err)
		runner.idleDetection("")
	}
}

func (runner *EciRunner) sacrifyCmd() {
	runcmd := exec.Command(runner.home_path+"_depsh/sacrify.sh", strconv.FormatInt(runner.sacrity_time, 10))
	out, err := runcmd.Output()
	if err != nil {
		logrus.Errorln(runner.runner_id+" - "+string(out), err)
	}
}

func (runner *EciRunner) needSacrifice() {
	for runner.container_type != "pool" {
		if runner.remote_dockerd {
			runner.sacrifyCmd()
		} else {
			runner.sacrifyCtl()
		}
		runner.sacrity_time += runner.sacrify_interval
		time.Sleep(time.Duration(runner.sacrify_interval) * time.Second)
	}
}

func (runner EciRunner) Start() error {
	runcmd := exec.Command(runner.home_path + "run.sh")
	err := runcmd.Start()
	if err != nil {
		logrus.Errorf("%s - Unable to start runner: %s", runner.runner_id, err)
		return err
	}
	go runner.needSacrifice()
	return nil
}

func (runner EciRunner) Info() interface{} {
	return runner
}

func (runner *EciRunner) Monitor(obj interface{}, para interface{}) bool {
	q, _ := obj.(ali_mns.AliMNSQueue)
	// rChan := make(chan EciRunner, 1)
	endChan := make(chan bool)
	respChan := make(chan ali_mns.MessageReceiveResponse)
	errChan := make(chan error)
	go func() {
		select {
		case resp := <-respChan:
			{
				logrus.Infof("runner received a msg: %s, name %s", resp.MessageBody, runner.runnerName())
				if strings.Compare(resp.MessageBody, runner.runnerName()) == 0 {
					logrus.Infof("Found msg... url %s", runner.runner_repo_url)
					if ret, e := q.ChangeMessageVisibility(resp.ReceiptHandle, 5); e != nil {
						fmt.Println(e)
					} else {
						if e := q.DeleteMessage(ret.ReceiptHandle); e != nil {
							fmt.Println(e)
						}

						crypto := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))

						pref := common.Ternary(strings.Contains(runner.runner_repo_url, runner.en_id),
							runner.entoken_fqdn, runner.token_fqdn).(string)
						url := common.Ternary(runner.runner_org != "none", pref+"orgs/"+runner.runner_org+runner.remove_path,
							pref+"repos/"+runner.runner_repo_owner+"/"+runner.runner_repo_name+runner.remove_path).(string)
						r_tk, _ := runner.getTk(crypto.DecryptMsg(runner.runner_token), url, nil)
						runner.removeRunner(runner.home_path, r_tk, true)

						endChan <- true
					}
				}
				endChan <- false
			}
		case err := <-errChan:
			{
				if err != nil && !ali_mns.ERR_MNS_MESSAGE_NOT_EXIST.IsEqual(err) {
					logrus.Errorln(runner.runner_id+" - ", err)
				}
				endChan <- false
			}
		}
	}()
	q.ReceiveMessage(respChan, errChan, runner.interval)
	return <-endChan
}

func (runner EciRunner) removeRunner(path string, tk string, notify bool) error {
	rmRunner := func() ([]byte, error) {
		cmd := exec.Command(path+"config.sh", "remove", "--token", tk)
		return cmd.Output()
	}
	out, err := rmRunner()
	if err != nil {
		// network issue: https://github.com/ingka-group-digital/serverless-hosted-runner/issues/39
		logrus.Errorf("%s - unable to remove runner, err - %s, out - %s", runner.runner_id, err, out)
		time.Sleep(time.Duration(10) * time.Second)
		if err := runner.refreshTk(); err != nil {
			logrus.Errorf("%s - retry and fail to refresh token for remove, %v", runner.runner_id, err)
		}
		if out, err = rmRunner(); err != nil {
			logrus.Errorf("%s - retry and fail to remove runner, err - %s, out - %s", runner.runner_id, err, out)
		}
	}
	if valid_ip := net.ParseIP(runner.dis_ip); valid_ip != nil && notify {
		state := "Finished"
		state_msg := "Runner job finished"
		runner_name := runner.runnerName()
		labels := runner.runner_repo_name + "," + runner.default_labels
		notifier := listener.CreateNotifier(runner.dis_ip)
		notifier.Notify(listener.RunnerState{
			RunnerId:  &runner.runner_id,
			State:     &state,
			StateMsg:  &state_msg,
			Act:       &runner.runner_action,
			RunerName: &runner_name,
			RepoName:  &runner.runner_repo_name,
			OrgName:   &runner.runner_org,
			RunWf:     &runner.runner_id,
			Labels:    &labels,
			Url:       &runner.runner_repo_url,
			Owner:     &runner.runner_repo_owner,
		})
	}
	logrus.Infof("Finish notify removing runner %s, out - %s", runner.dis_ip, out)
	return nil
}

func (runner EciRunner) parseResponse(body io.Reader) (string, error) {
	data, _ := io.ReadAll(body)
	if runner.jit_enabled && runner.container_type != "pool" {
		jitToken := JitToken{}
		json.Unmarshal(data, &jitToken)
		return jitToken.EncodedJitConfig, nil
	} else {
		regToken := RunnerToken{}
		json.Unmarshal(data, &regToken)
		return regToken.Token, nil
	}
}

func (runner EciRunner) getTk(tk string, url string, body io.Reader) (string, error) {
	client := http.Client{
		Timeout: time.Duration(60 * time.Second),
	}
	request, _ := http.NewRequest("POST", url, body)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+tk)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "serverless-hosted-runner")
	if runner.container_type == "pool" {
		// prevent to generate same token
		t, err := strconv.Atoi(os.Getenv("runner"))
		time.Sleep(time.Duration(t) * time.Second)
		if err != nil {
			logrus.Errorf("%s - strconv error: %s", runner.runner_id, err)
		}
	}
	resp, err := client.Do(request)
	if err != nil || (resp != nil && resp.StatusCode != 201) {
		logrus.Errorf("%s - Unable to get runner registrtation url %s, token %s, error %v, resp %v", runner.runner_id, url, tk, err, resp)
		return "Unable to get runner registration token", err
	}
	defer resp.Body.Close()
	return runner.parseResponse(resp.Body)
}

func (runner EciRunner) getOrgRegToken(tk string) (string, error) {
	pref := common.Ternary(strings.Contains(runner.runner_repo_url, runner.en_id), runner.entoken_fqdn, runner.token_fqdn).(string)
	if runner.jit_enabled && runner.container_type != "pool" {
		url := pref + "orgs/" + runner.runner_org + runner.jit_path
		body := "{\"name\":\"" + runner.runnerName() + "\", \"labels\":[\"" + runner.runner_org +
			"\"],\"runner_group_id\":1,\"work_folder\":\"_work\"}"
		logrus.Infof("Org jit url: %s, body: %s", url, body)
		return runner.getTk(tk, url, strings.NewReader(body))
	} else {
		url := pref + "orgs/" + runner.runner_org + runner.token_path
		logrus.Infof("Org url: %s", url)
		return runner.getTk(tk, url, nil)
	}
}

func (runner EciRunner) getRepoRegToken(tk string) (string, error) {
	pref := common.Ternary(strings.Contains(runner.runner_repo_url, runner.en_id), runner.entoken_fqdn, runner.token_fqdn).(string)
	if runner.jit_enabled && runner.container_type != "pool" {
		url := pref + "repos/" + runner.runner_repo_owner + "/" + runner.runner_repo_name + runner.jit_path
		body := "{\"name\":\"" + runner.runnerName() + "\", \"labels\":[\"" + runner.runner_repo_name +
			"\"],\"runner_group_id\":1,\"work_folder\":\"_work\"}"
		logrus.Infof("Repo jit url: %s", url)
		return runner.getTk(tk, url, strings.NewReader(body))
	} else {
		url := pref + "repos/" + runner.runner_repo_owner + "/" + runner.runner_repo_name + runner.token_path
		logrus.Infof("Repo url: %s", url)
		return runner.getTk(tk, url, nil)
	}
}
