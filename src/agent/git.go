package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	common "serverless-hosted-runner/common"
	"strings"
	"time"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type IGit interface {
	GetRegistrationToken(GitRegToken) (GitRegToken, bool)
	IsTokenValid(GitRegToken) bool
}

type Git struct {
	hubToken     string
	entToken     string
	hubAPIServer string
	entAPIServer string
	regTokenPath string
	dmEnt        string
	dmHub        string
}

type GitRegToken struct {
	IsOrg bool
	Repo  string
	URL   string
	Token string
	Exp   string
}

func CreateGitAgent() IGit {
	return &Git{os.Getenv("SLS_GITHUB_TK"), os.Getenv("SLS_GITENT_TK"),
		"https://api.github.com/", "https://git.build.ingka.ikea.com/api/v3/",
		"/actions/runners/registration-token", "https://git.build.ingka.ikea.com/", "https://github.com/"}
}

func (git Git) GetRegistrationToken(token GitRegToken) (GitRegToken, bool) {
	if git.IsTokenValid(token) {
		return token, true
	} else {
		return git.genRegToken(token), false
	}
}

func (git Git) IsTokenValid(token GitRegToken) bool {
	if len(token.Exp) > 0 {
		logrus.Warnf("isTokenValid token.exp: %v", token.Exp)
		exp, err := common.ParseTimeLocation(token.Exp)
		if err != nil {
			logrus.Errorf("fail to parse git token exp: %v", err)
			return false
		}
		logrus.Warnf("isTokenValid time.Now(): %s, exp: %s", time.Now().String(), exp.In(time.UTC).String())
		if time.Now().After(exp.In(time.UTC)) {
			logrus.Warnf("now is after exp, invalid exp")
			return false
		} else {
			logrus.Warnf("now is before exp, valid exp")
			return true
		}
	}
	return false
}

func (git Git) genRegToken(token GitRegToken) GitRegToken {
	org := strings.ReplaceAll(strings.ReplaceAll(token.URL, git.dmHub, ""), git.dmEnt, "")
	url := common.Ternary(strings.Contains(token.URL, git.dmEnt), git.entAPIServer, git.hubAPIServer).(string) +
		common.Ternary(token.IsOrg, "orgs/"+org+git.regTokenPath, "repos/"+org+"/"+token.Repo+git.regTokenPath).(string)
	client := http.Client{
		Timeout: time.Duration(60 * time.Second),
	}
	request, _ := http.NewRequest("POST", url, nil)
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+common.Ternary(strings.Contains(token.URL, git.dmEnt), git.entToken, git.hubToken).(string))
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "serverless-hosted-runner")
	resp, err := client.Do(request)
	parseBody := func() []byte {
		bodyClose := func() {
			if err := resp.Body.Close(); err != nil {
				logrus.Errorf("fail to close body in genRegToken, %v", err)
			}
		}
		defer bodyClose()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Errorf("fail to read git reg resp body, %v", err)
		}
		return data
	}
	if err != nil {
		logrus.Errorf("fail to do request for git registration %v", err)
		return GitRegToken{}
	} else if resp != nil && resp.StatusCode != 201 {
		logrus.Errorf("request for git registration got incorrect response, url %s, code %v, resp %s", url, resp.StatusCode, string(parseBody()))
		return GitRegToken{}
	}
	regToken := common.RunnerToken{}
	if err := json.Unmarshal(parseBody(), &regToken); err != nil {
		logrus.Errorf("fail to unmarshal git reg token, %v", err)
		return GitRegToken{}
	} else {
		return GitRegToken{token.IsOrg, token.Repo, token.URL, regToken.Token, regToken.ExpiresAt}
	}
}
