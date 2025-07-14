package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	common "serverless-hosted-runner/common"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

var (
	Proto             = "https://"
	WfGitDM           = "github.com"
	EntDM             = "git.build.ingka.ikea.com"
	ApiPath           = "/api/v3"
	GitDM             = "api.github.com"
	WfList            = "/actions/runs"
	RepoType          = "Repo"
	OrgType           = "Org"
	ListRepoOps       = "ListRepoOps"
	ListRunsOps       = "ListRunsOps"
	ListJobsOps       = "ListJobsOps"
	GetRepoDetailOps  = "GetRepoDetail"
	WfStatusQueued    = "queued"
	WfStatusCompleted = "completed"
	RunnerLabel       = [2]string{"serverless-hosted-runner", "eci-runner"}
)

type WorkFlow struct {
	t           string
	name        string
	url         string
	org         string
	repo        string
	repos       []string
	ent         bool
	runid       string
	tk          string
	runner      string
	jobid       string
	create      common.CreateRunner
	destroy     common.DestroyRunner
	release     common.ReleaseRunner
	check       common.CheckRunner
	iv          int
	labels      string
	run_last_h  int
	complete_iv int
}

func CreateWorkflowAgent(t string, name string, url string, crt common.CreateRunner, des common.DestroyRunner,
	rel common.ReleaseRunner, ck common.CheckRunner, repo string, org string, iv int, labels string) Agent {
	return &WorkFlow{t, name, url, org, repo, nil, false, "", "", "", "", crt, des, rel, ck, iv, labels, -2, 1200}
}

func (wf *WorkFlow) InitAgent() {
	if len(wf.repo) == 0 && wf.t == RepoType {
		wf.repo = wf.name
	} else if len(wf.org) == 0 && wf.t == OrgType {
		wf.org = wf.name
	}
	if strings.Contains(wf.url, EntDM) {
		wf.ent = true
		wf.tk = os.Getenv("SLS_GITENT_TK")
	} else {
		wf.ent = false
		wf.tk = os.Getenv("SLS_GITHUB_TK")
	}
}

func (wf WorkFlow) MonitorOnAgent() {
	go wf.monitorOnQueued()
	go wf.monitorOnComplete()
}

func (wf WorkFlow) NotifyAgent(msg string) {
	wf.release(msg)
}

func (wf WorkFlow) checkWorkflows(wftype string) {
	if wf.t == RepoType {
		wf.checkRepoWorkflows(wftype)
	} else if wf.t == OrgType {
		wf.checkOrgWorkflows(wftype)
	}
}

func (wf WorkFlow) monitorOnQueued() {
	for {
		wf.checkWorkflows(WfStatusQueued)
		logrus.Infof("monitorOnQueued, wait " + strconv.Itoa(wf.iv) +
			" seconds and scan the org:" + wf.org + " repo:" + wf.repo + " workflows again..")
		time.Sleep(time.Duration(wf.iv) * time.Second)
	}
}

func (wf WorkFlow) monitorOnComplete() {
	for {
		wf.checkWorkflows(WfStatusCompleted)
		logrus.Infof("monitorOnComplete, wait " + strconv.Itoa(wf.complete_iv) +
			" seconds and scan the org:" + wf.org + " repo:" + wf.repo + " workflows again..")
		time.Sleep(time.Duration(wf.complete_iv) * time.Second)
	}
}

func (wf *WorkFlow) checkRepoWorkflows(wftype string) {
	logrus.Infof("begin checkRepoWorkflows...")
	wf.getOrg()
	wf.checkRepoRuns(wf.getRepo(), wftype)
}

func (wf *WorkFlow) checkOrgWorkflows(wftype string) {
	logrus.Infof("begin checkOrgWorkflows...")
	reps := wf.getRepos()

	// TODO: should we use goroutine? it may cause possible git API limit reaching caused deny issue
	// ref: https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api
	//      < 100 concurrent requests
	//      < 5000 git / 15000 Enteprise per hour
	// senario: security team has about 300 repos under https://git.build.ingka.ikea.com/china-digital-hub to scan
	//          300x5 = 1500 > 100
	//          300x5x4x60 = 360000 > 15000; 300x5x3600/45 = 120000 < 15000
	for _, rep := range reps {
		// go wf.checkRepoRuns(rep)
		wf.checkRepoRuns(rep, wftype)
	}
}

func (wf *WorkFlow) checkQueuedCompleteRun(runs common.WorkflowRuns, rep common.Repository) {
	for _, run := range runs.WorkflowRuns {
		logrus.Infof("checking wf run %s, status %s, conclusion %s", run.Name, run.Status, run.Conclusion)
		if run.Status == WfStatusQueued {
			wf.runid = strconv.FormatInt(run.ID, 10)
			jobs := wf.getJobs()
			num := 0
			for _, job := range jobs.Jobs {
				logrus.Infof("checking wf job %s, status %s, conclusion %s, label %s", job.Name, job.Status,
					job.Conclusion, strings.Join(job.Labels, ","))
				if job.Status == WfStatusQueued &&
					// wf.permitedLabel(job.Labels) {
					wf.check(job.Labels, wf.repo, wf.org, wf.t, wf.labels) {
					wf.jobid = strconv.FormatInt(job.ID, 10)
					wf.runner = wf.runid + "-" + wf.jobid
					num++
					logrus.Infof("start creating runner#%d, paras: %s, %s, %s, %s, %s, %s.", num, wf.runner,
						wf.repo, wf.url, wf.org, rep.Owner.Login, job.Labels)
					go wf.create(WfStatusQueued, wf.runner, wf.repo, wf.url, wf.org, rep.Owner.Login, job.Labels)
				}
			}
		} else if run.Status == WfStatusCompleted {
			wf.runid = strconv.FormatInt(run.ID, 10)
			jobs := wf.getJobs()
			for _, job := range jobs.Jobs {
				logrus.Infof("in complete, %s, status %s, conclusion %s, job status %s, job label %s",
					run.Name, run.Status, run.Conclusion, job.Status, job.Labels)
				if job.Status == WfStatusCompleted &&
					// wf.permitedLabel(job.Labels) {
					wf.check(job.Labels, wf.repo, wf.org, wf.t, wf.labels) {
					wf.jobid = strconv.FormatInt(job.ID, 10)
					wf.runner = job.RunnerName // repo - event_data.WorkflowJob.RunID - WorkflowJob.ID
					logrus.Infof("destroy runner, %s, jobid %s, repo %s, org %s, label %s, url %s, runner id %s",
						job.RunnerName, wf.jobid, wf.repo, wf.org, job.Labels, wf.url, wf.runner)
					//wf.NotifyAgent(job.RunnerName)
					go wf.destroy(WfStatusCompleted, job.RunnerName, wf.repo, wf.org,
						strconv.FormatInt(job.RunID, 10)+"-"+strconv.FormatInt(job.ID, 10),
						job.Labels, wf.url, rep.Owner.Login)
				}
			}
		}
	}
}

func (wf *WorkFlow) checkRepoRuns(rep common.Repository, wftype string) {
	logrus.Infof("checkRepoRuns, checking repo %s...", rep.Name)
	wf.repo = rep.Name

	if wftype == WfStatusQueued {
		queued_runs := wf.getWfQueuedRuns()
		logrus.Infof("checkRepoRuns, queued run number: %d", queued_runs.TotalCount)
		wf.checkQueuedCompleteRun(queued_runs, rep)
	} else if wftype == WfStatusCompleted {
		complete_runs := wf.getWfClosedRuns()
		logrus.Infof("checkRepoRuns, complete run number: %d", complete_runs.TotalCount)
		wf.checkQueuedCompleteRun(complete_runs, rep)
	}
}

func (wf WorkFlow) getUrl(op string) string {
	pref := Proto + GitDM
	if wf.ent {
		pref = Proto + EntDM + ApiPath
	}
	if op == ListRepoOps {
		return pref + "/orgs/" + wf.org + "/repos"
	} else if op == ListRunsOps {
		return pref + "/repos/" + wf.org + "/" + wf.repo + WfList + "?per_page=100"
	} else if op == ListJobsOps {
		return pref + "/repos/" + wf.org + "/" + wf.repo + WfList + "/" + wf.runid + "/jobs"
	} else if op == GetRepoDetailOps {
		logrus.Infof("pref %s, wf.org %s, wf.repo %s", pref, wf.org, wf.repo)
		return pref + "/repos/" + wf.org + "/" + wf.repo
	}
	return ""
}

func (wf *WorkFlow) getOrg() {
	if len(wf.url) > 0 {
		subp := strings.ReplaceAll(wf.url, Proto+WfGitDM+"/", "")
		subp = strings.ReplaceAll(subp, Proto+EntDM+"/", "")
		subp = strings.ReplaceAll(subp, wf.repo, "")
		wf.org = strings.ReplaceAll(subp, "/", "")
		logrus.Infof("getOrg url %s, wf.org %s", wf.url, wf.org)
	}
}

func (wf WorkFlow) request(op string, para string) interface{} {
	client := http.Client{
		Timeout: time.Duration(60 * time.Second),
	}
	request, _ := http.NewRequest("GET", wf.getUrl(op)+para, nil)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+wf.tk)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "serverless-hosted-runner")
	resp, err := client.Do(request)
	if err != nil || resp == nil || resp.StatusCode != 200 {
		if resp != nil {
			bodyBytes, err := io.ReadAll(resp.Body)
			bodyString := "empty body"
			if err == nil {
				bodyString = string(bodyBytes)
			}
			logrus.Errorf("Unable to get %s, %s, %s", op, err, bodyString)
		} else {
			logrus.Errorf("Unable to get %s, %s", op, err)
		}

		fmt.Println(resp)
		return nil
	}
	defer resp.Body.Close()
	return wf.response(op, resp.Body)
}

func (wf WorkFlow) response(op string, body io.Reader) interface{} {
	data, _ := io.ReadAll(body)
	// body.(io.ReadCloser).Close()
	if op == ListRepoOps {
		reps := common.Repos{}
		json.Unmarshal(data, &reps)
		for idx, item := range reps {
			logrus.Infof("response rep %d, fullname %s, owner %s, name %s, svnurl %s", idx,
				item.FullName, item.Owner.Login, item.Name, item.SvnURL)
		}
		return reps
	} else if op == ListRunsOps {
		runs := common.WorkflowRuns{}
		json.Unmarshal(data, &runs)
		logrus.Infof("response run count: %d", runs.TotalCount)
		for idx, item := range runs.WorkflowRuns {
			logrus.Infof("response run %d, id %d, name %s, status %s, conclusion %s, url %s", idx,
				item.ID, item.Name, item.Status, item.Conclusion, item.HTMLURL)
		}
		return runs
	} else if op == ListJobsOps {
		jobs := common.WorkflowJobs{}
		json.Unmarshal(data, &jobs)
		logrus.Infof("response job count %d", jobs.TotalCount)
		for idx, item := range jobs.Jobs {
			logrus.Infof("response job %d, id %d, runid %d, status %s", idx, item.ID, item.RunID, item.Status)
			for _, label := range item.Labels {
				logrus.Infof("label: %s", label)
			}
		}
		return jobs
	} else if op == GetRepoDetailOps {
		rep := common.Repository{}
		json.Unmarshal(data, &rep)
		logrus.Infof("response repo fullname %s", rep.FullName)
		logrus.Infof("response repo, id %d, name %s, htmlurl %s, owner %s", rep.ID, rep.Name,
			rep.HTMLURL, rep.Owner.Login)
		return rep
	}
	return nil
}

func (wf WorkFlow) getRepo() common.Repository {
	resp := wf.request(GetRepoDetailOps, "")
	return common.Ternary(resp == nil, common.Repository{}, resp).(common.Repository)
}

func (wf WorkFlow) getRepos() common.Repos {
	resp := wf.request(ListRepoOps, "")
	return common.Ternary(resp == nil, common.Repos{}, resp).(common.Repos)
}

func (wf WorkFlow) getWfQueuedRuns() common.WorkflowRuns {
	return wf.getWfRuns("&status=" + WfStatusQueued)
}

func (wf WorkFlow) getWfClosedRuns() common.WorkflowRuns {
	start := time.Now().Add(time.Duration(wf.run_last_h) * time.Hour)
	end := time.Now()
	// "&status=" + WfStatusCompleted
	return wf.getWfRuns("&created=" + fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ..%d-%02d-%02dT%02d:%02d:%02dZ",
		start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute(), start.Second(),
		end.Year(), end.Month(), end.Day(), end.Hour(), end.Minute(), end.Second()))
}

func (wf WorkFlow) getWfRuns(para string) common.WorkflowRuns {
	logrus.Infof("getWfRuns, para: %s", para)
	resp := wf.request(ListRunsOps, para)
	return common.Ternary(resp == nil, common.WorkflowRuns{}, resp).(common.WorkflowRuns)
}

func (wf WorkFlow) getJobs() common.WorkflowJobs {
	resp := wf.request(ListJobsOps, "")
	return common.Ternary(resp == nil, common.WorkflowJobs{}, resp).(common.WorkflowJobs)
}
