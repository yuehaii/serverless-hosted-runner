// Package runner of micro service runner
package runner

import (
	"context"
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
	containerType   string
	runnerID        string
	runnerToken     string
	runnerRepoURL   string
	runnerOrg       string
	runnerRepoName  string
	runnerAction    string
	runnerRepoOwner string
	imageVer        string
	homePath        string
	tokenFqdn       string
	tokenPath       string
	interval        int64
	poolPrefix      string
	entokenFqdn     string
	gitID           string
	enID            string
	tk              string
	ephemeral       string
	jitEnabled      bool
	jitEncoded      string
	jitPath         string
	poolSufix       string
	removePath      string
	defaultLabels   string
	sacrifyInterval int64
	sacrityTime     int64
	runnerGroup     string
	remoteDockerd   bool
	maxIdle         int64
	sysCtl          common.ISysCtl
	retryTimes      int
	retryBackoff    int64
	pr              string
	disIP           string
	tkInvalid       int64
	cmdTimeout      int64
	gitAgent        agent.IGit
	ctxLogLevel     string
	repoRegToken    string
}

func EciRunnerCreator(containerType string, runnerID string, runnerToken string,
	runnerRepoURL string, runnerOrg string, runnerRepoName string,
	runnerAction string, runnerRepoOwner string, imageVer string, runnerLabels string,
	runnerGroup string, cloudPr string, disIP string, ctxLogLevel string, repoRegToken string) Runner {
	additionalLabels := ""
	if len(runnerLabels) > 0 && runnerLabels != "none" {
		additionalLabels = "," + runnerLabels
	}
	return &EciRunner{containerType, runnerID, runnerToken, runnerRepoURL, runnerOrg,
		runnerRepoName, runnerAction, runnerRepoOwner, imageVer,
		"/go/bin/", "https://api.github.com/", "/actions/runners/registration-token",
		int64(10), agent.NotificationQueue, "https://git.build.ingka.ikea.com/api/v3/",
		"api.github.com", "git.build.ingka.ikea.com", "", "", false, "",
		"/actions/runners/generate-jitconfig", "-" + os.Getenv("runner"),
		"/actions/runners/remove-token", "serverless-hosted-runner,eci-runner" + additionalLabels,
		10, 0, runnerGroup, false, 300, common.CreateUnixSysCtl(), 3, 20, cloudPr, disIP, 3600, 300,
		agent.CreateGitAgent(), ctxLogLevel, repoRegToken}
}

func FnRunnerCreator() Runner {
	return nil
}

func (runner *EciRunner) Init() {
	common.SetContextLogLevel(runner.ctxLogLevel)
	runner.runnerInit()
}

func (runner *EciRunner) Configure() {
	runnerConfig := func() (err error) {
		_ = runner.removeRunner(runner.homePath, runner.tk, false)
		if err := runner.refreshTk(); err != nil {
			logrus.Errorf("%s - fail to gen tk for config, %v", runner.runnerID, err)
			return err
		} else {
			if runner.runnerOrg != "none" {
				out, err := runner.configRunner(runner.runnerOrg, runner.runnerName(), runner.runnerRepoURL)
				if err != nil {
					logrus.Errorf("%s - fai to config runner, %v, err:%v", runner.runnerID, out, err)
				}
				return err
			} else if runner.runnerRepoName != "" {
				out, err := runner.configRunner(runner.runnerRepoName, runner.runnerName(), runner.runnerRepoURL)
				if err != nil {
					logrus.Errorf("%s - fai to config runner, %v, err:%v", runner.runnerID, out, err)
				}
				return err
			}
		}
		return errors.New("unknow runner type for config")
	}
	times := 0
	for times < runner.retryTimes {
		if runnerConfig() == nil {
			break
		}
		times += 1
		time.Sleep(time.Duration(runner.retryBackoff) * time.Second)
		logrus.Warnf("%s - fail to config runner url: %s, label: %s, retry %v time", runner.runnerID, runner.runnerRepoURL, runner.defaultLabels, times)
	}
}

func (runner EciRunner) runnerInit() {
	logrus.Infof("Init runner...")
	if runner.remoteDockerd {
		cmd := exec.Command(runner.homePath + "_depsh/init.sh")
		out, err := cmd.Output()
		logrus.Infof("Init runner cmd output is: %s", out)
		if err != nil {
			logrus.Errorf("%s - Init runner error: %s", runner.runnerID, err)
		}
	} else {
		switch runner.pr {
		case "ali":
			if err := runner.sysCtl.DockerStorageDriver("overlay2"); err != nil {
				logrus.Errorf("Init runner, fail to set storage driver: %v", err)
			}
		case "azure", "gcp", "gcp_dind":
			if err := runner.sysCtl.DockerStorageDriver("vfs"); err != nil {
				logrus.Errorf("Init runner, fail to set storage driver for %s: %v", runner.pr, err)
			}
		}
		go runner.sysCtl.StartProcess("dockerd")
		go runner.sysCtl.SetResolvers()
	}
}

func (runner EciRunner) runnerName() string {
	switch runner.containerType {
	case "pool":
		return runner.poolPrefix + "-" + runner.runnerOrg + runner.poolSufix
	case "org":
		return runner.runnerOrg + "-" + runner.runnerID
	case "repo":
		return runner.runnerRepoName + "-" + runner.runnerID
	default:
		return ""
	}
}

func (runner EciRunner) poolName() string {
	return runner.poolPrefix + "-" + runner.runnerOrg
}

func (runner EciRunner) configRunner(label string, name string, url string) (string, error) {
	logrus.Infof("Configuring runner... url %s, name %s, label %s", url, name, label+","+runner.defaultLabels)
	var cancelFn context.CancelFunc
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Duration(runner.cmdTimeout)*time.Second)
	defer cancelFn()
	cmd := exec.CommandContext(ctx, runner.homePath+"config.sh", "--url", url, "--token", runner.tk, "--name", name,
		"--runnergroup", runner.runnerGroup, "--labels", label+","+runner.defaultLabels, "--work", "_work",
		"--unattended", common.Ternary(runner.containerType == "pool", "", "--ephemeral").(string))
	out, err := cmd.Output()
	logrus.Infof("Config runner cmd %s, output is: %s", cmd.String(), string(out))
	if err != nil {
		logrus.Errorln(runner.runnerID+" - "+string(out), err)
		return string(out), err
	} else if !strings.Contains(string(out), "successful") {
		msg := "fail to register runner: " + string(out)
		logrus.Errorln(runner.runnerID + " - " + msg)
		return string(out), errors.New(msg)
	}
	return string(out), nil
}

func (runner *EciRunner) refreshTk() (err error) {
	crypto := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))
	if runner.runnerOrg != "none" {
		runner.tk, err = runner.getOrgRegToken(crypto.DecryptMsg(runner.runnerToken))
		if err != nil {
			logrus.Errorf("%s - fail to generate org access token, %v", runner.runnerID, err)
			return err
		}
	} else if runner.runnerRepoName != "" {
		runner.tk, err = runner.getRepoRegToken(crypto.DecryptMsg(runner.runnerToken))
		if err != nil {
			logrus.Errorf("%s - fail to generate repo access token, %v", runner.runnerID, err)
			return err
		}
	}
	return nil
}

func (runner EciRunner) idleDetection(workLog string) {
	if strings.Contains(strings.ToLower(workLog), "job completed") ||
		(runner.sacrityTime > runner.maxIdle &&
			!strings.Contains(strings.ToLower(workLog), "job message")) {
		logrus.Warn(common.Ternary(strings.Contains(strings.ToLower(workLog), "job completed"),
			"workflow finished. complete runner.", "sacrify cur_idle: "+strconv.Itoa(int(runner.sacrityTime))+
				", maxIdle: "+strconv.Itoa(int(runner.maxIdle))).(string))
		if runner.sacrityTime > runner.tkInvalid {
			if err := runner.refreshTk(); err != nil {
				logrus.Errorf("%s - fail to refresh token for remove, %v", runner.runnerID, err)
				return
			}
		}
		if err := runner.removeRunner(runner.homePath, runner.tk, true); err != nil {
			logrus.Warnf("idleDetection, fail to remove runner, %v", err)
		}
		os.Exit(0)
	} else if strings.Contains(strings.ToLower(workLog), "job message") {
		logrus.Infof("wf still running. idle %v", runner.sacrityTime)
	} else {
		logrus.Infof("wf dose not select the runner temporary. idle %v", runner.sacrityTime)
	}
}

func (runner *EciRunner) sacrifyCtl() {
	logrus.Infof("sacrify ctl start")
	entries, err := os.ReadDir(runner.homePath + "_diag/")
	if err == nil {
		workerExist := false
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "Worker_") {
				fWorker, err := os.ReadFile(runner.homePath + "_diag/" + entry.Name())
				if err != nil {
					logrus.Warnf("%s - fail to read the work log %s, %s", runner.runnerID, entry.Name(), err)
					continue
				}
				workerExist = true
				runner.idleDetection(string(fWorker))
			}
		}
		if !workerExist {
			logrus.Warnf("%s - Worker_ dose not exist temporary", runner.runnerID)
			runner.idleDetection("")
		}
	} else {
		logrus.Errorf("%s - fail to list diag dir, %v", runner.runnerID, err)
		runner.idleDetection("")
	}
}

func (runner *EciRunner) sacrifyCmd() {
	runcmd := exec.Command(runner.homePath+"_depsh/sacrify.sh", strconv.FormatInt(runner.sacrityTime, 10))
	out, err := runcmd.Output()
	if err != nil {
		logrus.Errorln(runner.runnerID+" - "+string(out), err)
	}
}

func (runner *EciRunner) needSacrifice() {
	for runner.containerType != "pool" {
		if runner.remoteDockerd {
			runner.sacrifyCmd()
		} else {
			runner.sacrifyCtl()
		}
		runner.sacrityTime += runner.sacrifyInterval
		time.Sleep(time.Duration(runner.sacrifyInterval) * time.Second)
	}
}

func (runner EciRunner) Start() error {
	runcmd := exec.Command(runner.homePath + "run.sh")
	err := runcmd.Start()
	if err != nil {
		logrus.Errorf("%s - Unable to start runner: %s", runner.runnerID, err)
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
					logrus.Infof("Found msg... url %s", runner.runnerRepoURL)
					if ret, e := q.ChangeMessageVisibility(resp.ReceiptHandle, 5); e != nil {
						fmt.Println(e)
					} else {
						if e := q.DeleteMessage(ret.ReceiptHandle); e != nil {
							fmt.Println(e)
						}

						crypto := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))

						pref := common.Ternary(strings.Contains(runner.runnerRepoURL, runner.enID),
							runner.entokenFqdn, runner.tokenFqdn).(string)
						url := common.Ternary(runner.runnerOrg != "none", pref+"orgs/"+runner.runnerOrg+runner.removePath,
							pref+"repos/"+runner.runnerRepoOwner+"/"+runner.runnerRepoName+runner.removePath).(string)
						rTk, _ := runner.getTk(crypto.DecryptMsg(runner.runnerToken), url, nil)
						if err := runner.removeRunner(runner.homePath, rTk, true); err != nil {
							logrus.Errorf("fail to remove runner in Monitor, %v", err)
						}

						endChan <- true
					}
				} else {
					logrus.Infof("pool name %s", runner.poolName())
				}
				endChan <- false
			}
		case err := <-errChan:
			{
				if err != nil && !ali_mns.ERR_MNS_MESSAGE_NOT_EXIST.IsEqual(err) {
					logrus.Errorln(runner.runnerID+" - ", err)
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
		logrus.Errorf("%s - unable to remove runner, err - %s, out - %s", runner.runnerID, err, out)
		time.Sleep(time.Duration(10) * time.Second)
		if err := runner.refreshTk(); err != nil {
			logrus.Errorf("%s - retry and fail to refresh token for remove, %v", runner.runnerID, err)
		}
		if out, err = rmRunner(); err != nil {
			logrus.Errorf("%s - retry and fail to remove runner, err - %s, out - %s", runner.runnerID, err, out)
		}
	}
	if validIP := net.ParseIP(runner.disIP); validIP != nil && notify {
		state := "Finished"
		stateMsg := "Runner job finished"
		runnerName := runner.runnerName()
		labels := runner.runnerRepoName + "," + runner.defaultLabels
		notifier := listener.CreateNotifier(runner.disIP)
		notifier.Notify(&listener.RunnerState{
			RunnerID:  &runner.runnerID,
			State:     &state,
			StateMsg:  &stateMsg,
			Act:       &runner.runnerAction,
			RunerName: &runnerName,
			RepoName:  &runner.runnerRepoName,
			OrgName:   &runner.runnerOrg,
			RunWf:     &runner.runnerID,
			Labels:    &labels,
			URL:       &runner.runnerRepoURL,
			Owner:     &runner.runnerRepoOwner,
		})
	}
	logrus.Infof("Finish notify removing runner %s, out - %s", runner.disIP, out)
	return nil
}

func (runner EciRunner) parseResponse(body io.Reader) (string, error) {
	data, _ := io.ReadAll(body)
	if runner.jitEnabled && runner.containerType != "pool" {
		jitToken := common.JitToken{}
		if err := json.Unmarshal(data, &jitToken); err != nil {
			logrus.Errorf("fail to Unmarshal jit tk during runner parseResponse, %v", err)
		}
		logrus.Infof("parseResponse JIT output is: %s", jitToken.EncodedJitConfig)
		return jitToken.EncodedJitConfig, nil
	} else {
		regToken := common.RunnerToken{}
		if err := json.Unmarshal(data, &regToken); err != nil {
			logrus.Errorf("fail to Unmarshal reg tk during runner parseResponse, %v", err)
		}
		logrus.Infof("parseResponse output is: %s", regToken.Token)
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
	if runner.containerType == "pool" {
		// prevent to generate same token
		t, err := strconv.Atoi(os.Getenv("runner"))
		time.Sleep(time.Duration(t) * time.Second)
		if err != nil {
			logrus.Errorf("%s - strconv error: %s", runner.runnerID, err)
		}
	}
	resp, err := client.Do(request)
	if err != nil || (resp != nil && resp.StatusCode != 201) {
		logrus.Errorf("%s - Unable to get runner registrtation url %s, token %s, error %v, resp %v", runner.runnerID, url, tk, err, resp)
		return "Unable to get runner registration token", err
	}
	bodyClose := func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("getTk, fail to close body: %v", err)
		}
	}
	defer bodyClose()
	return runner.parseResponse(resp.Body)
}

func (runner EciRunner) getOrgRegToken(tk string) (string, error) {
	pref := common.Ternary(strings.Contains(runner.runnerRepoURL, runner.enID), runner.entokenFqdn, runner.tokenFqdn).(string)
	if runner.jitEnabled && runner.containerType != "pool" {
		url := pref + "orgs/" + runner.runnerOrg + runner.jitPath
		body := "{\"name\":\"" + runner.runnerName() + "\", \"labels\":[\"" + runner.runnerOrg +
			"\"],\"runner_group_id\":1,\"work_folder\":\"_work\"}"
		logrus.Infof("Org jit url: %s, body: %s", url, body)
		return runner.getTk(tk, url, strings.NewReader(body))
	} else {
		url := pref + "orgs/" + runner.runnerOrg + runner.tokenPath
		logrus.Infof("Org url: %s", url)
		return runner.getTk(tk, url, nil)
	}
}

func (runner EciRunner) getRepoRegToken(tk string) (string, error) {
	pref := common.Ternary(strings.Contains(runner.runnerRepoURL, runner.enID), runner.entokenFqdn, runner.tokenFqdn).(string)
	if runner.jitEnabled && runner.containerType != "pool" {
		url := pref + "repos/" + runner.runnerRepoOwner + "/" + runner.runnerRepoName + runner.jitPath
		body := "{\"name\":\"" + runner.runnerName() + "\", \"labels\":[\"" + runner.runnerRepoName +
			"\"],\"runner_group_id\":1,\"work_folder\":\"_work\"}"
		logrus.Infof("Repo jit url: %s", url)
		return runner.getTk(tk, url, strings.NewReader(body))
	} else {
		if len(runner.repoRegToken) > 0 && runner.repoRegToken != "none" {
			regToken := agent.GitRegToken{}
			if err := json.Unmarshal([]byte(runner.repoRegToken), &regToken); err != nil {
				logrus.Errorf("fail to unmarshal runner regTokenStr: %s, %v", runner.repoRegToken, err)
			} else {
				if runner.gitAgent.IsTokenValid(regToken) {
					logrus.Infof("reuse the registration toke from dis")
					return regToken.Token, nil
				}
			}
		}
		url := pref + "repos/" + runner.runnerRepoOwner + "/" + runner.runnerRepoName + runner.tokenPath
		logrus.Infof("Repo url: %s", url)
		return runner.getTk(tk, url, nil)
	}
}
