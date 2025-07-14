package common

import (
	"time"
)

// MNS
type MsgRsp struct {
	MessageId        string `xml:"MessageId" json:"message_id"`
	ReceiptHandle    string `xml:"ReceiptHandle" json:"receipt_handle"`
	MessageBodyMD5   string `xml:"MessageBodyMD5" json:"message_body_md5"`
	MessageBody      string `xml:"MessageBody" json:"message_body"`
	EnqueueTime      int64  `xml:"EnqueueTime" json:"enqueue_time"`
	NextVisibleTime  int64  `xml:"NextVisibleTime" json:"next_visible_time"`
	FirstDequeueTime int64  `xml:"FirstDequeueTime" json:"first_dequeue_time"`
	DequeueCount     int64  `xml:"DequeueCount" json:"dequeue_count"`
	Priority         int64  `xml:"Priority" json:"priority"`
}

// Allen
type AllenMsg struct {
	Type                  string `json:"Type"`
	Name                  string `json:"Name"`
	Pat                   string `json:"Pat"`
	Url                   string `json:"Url"`
	Size                  string `json:"Size"`
	Key                   string `json:"Key"`
	Secret                string `json:"Secret"`
	Region                string `json:"Region"`
	SecGpId               string `json:"SecGpId"`
	VSwitchId             string `json:"VSwitchId"`
	Cpu                   string `json:"Cpu"`
	Memory                string `json:"Memory"`
	Repos                 string `json:"Repos"`
	Labels                string `json:"Labels"`
	ChargeLabels          string `json:"ChargeLabels"`
	RunnerGroup           string `json:"runner_group"`
	UserID                string `json:"UserId"`
	ArmClientId           string `json:"ArmClientId"`
	ArmClientSecret       string `json:"ArmClientSecret"`
	ArmSubscriptionId     string `json:"ArmSubscriptionId"`
	ArmTenantId           string `json:"ArmTenantId"`
	ArmEnvironment        string `json:"ArmEnvironment"`
	ArmRPRegistration     string `json:"ArmRPRegistration"`
	ArmResourceGroupName  string `json:"ArmResourceGroupName"`
	ArmSubnetId           string `json:"ArmSubnetId"`
	ArmLogAnaWorkspaceId  string `json:"ArmLogAnaWorkspaceId"`
	ArmLogAnaWorkspaceKey string `json:"ArmLogAnaWorkspaceKey"`
	GcpCredential         string `json:"GcpCredential"`
	GcpProject            string `json:"GcpProject"`
	GcpRegion             string `json:"GcpRegion"`
	GcpSA                 string `json:"GcpSA"`
	GcpApikey             string `json:"GcpApikey"`
	GcpDind               string `json:"GcpDind"`
	GcpVpc                string `json:"GcpVpc"`
	GcpSubnet             string `json:"GcpSubnet"`
	ImageVersion          string `json:"ImageVersion"`
	AciLocation           string `json:"AciLocation"`
	AciSku                string `json:"AciSku"`
	AciNetworkType        string `json:"AciNetworkType"`
}

// pool
type PoolMsg struct {
	Type                  string
	Name                  string
	Pat                   string
	Url                   string
	Size                  string
	Key                   string
	Secret                string
	Region                string
	SecGpId               string
	VSwitchId             string
	Cpu                   string
	Memory                string
	Repos                 string
	PullInterval          string
	Labels                string
	ChargeLabels          string
	RunnerGroup           string
	ArmClientId           string
	ArmClientSecret       string
	ArmSubscriptionId     string
	ArmTenantId           string
	ArmEnvironment        string
	ArmRPRegistration     string
	ArmResourceGroupName  string
	ArmSubnetId           string
	ArmLogAnaWorkspaceId  string
	ArmLogAnaWorkspaceKey string
	GcpCredential         string
	GcpProject            string
	GcpRegion             string
	GcpSA                 string
	GcpApikey             string
	GcpDind               string
	GcpVpc                string
	GcpSubnet             string
	ImageVersion          string
	AciLocation           string
	AciSku                string
	AciNetworkType        string
}
type Mns_Process func(interface{}, interface{}) bool

func (aln AllenMsg) ConvertPoolMsg() PoolMsg {
	msg := PoolMsg{}
	msg.Type = aln.Type
	msg.Name = aln.Name
	msg.Pat = aln.Pat
	msg.Url = aln.Url
	msg.Size = aln.Size
	msg.Key = aln.Key
	msg.Secret = aln.Secret
	msg.Region = aln.Region
	msg.SecGpId = aln.SecGpId
	msg.VSwitchId = aln.VSwitchId
	msg.Cpu = aln.Cpu
	msg.Memory = aln.Memory
	msg.Repos = aln.Repos
	msg.Labels = aln.Labels
	msg.ChargeLabels = aln.ChargeLabels
	msg.ArmClientId = aln.ArmClientId
	msg.ArmClientSecret = aln.ArmClientSecret
	msg.ArmSubscriptionId = aln.ArmSubscriptionId
	msg.ArmTenantId = aln.ArmTenantId
	msg.ArmEnvironment = aln.ArmEnvironment
	msg.ArmRPRegistration = aln.ArmRPRegistration
	msg.ArmResourceGroupName = aln.ArmResourceGroupName
	msg.ArmSubnetId = aln.ArmSubnetId
	msg.ArmLogAnaWorkspaceId = aln.ArmLogAnaWorkspaceId
	msg.ArmLogAnaWorkspaceKey = aln.ArmLogAnaWorkspaceKey
	msg.GcpCredential = aln.GcpCredential
	msg.GcpProject = aln.GcpProject
	msg.GcpRegion = aln.GcpRegion
	msg.GcpSA = aln.GcpSA
	msg.GcpApikey = aln.GcpApikey
	msg.GcpDind = aln.GcpDind
	msg.GcpVpc = aln.GcpVpc
	msg.GcpSubnet = aln.GcpSubnet
	msg.ImageVersion = aln.ImageVersion
	msg.AciLocation = aln.AciLocation
	msg.AciSku = aln.AciSku
	msg.AciNetworkType = aln.AciNetworkType
	return msg
}

// Workflow
type Repos []Repository
type Repository struct {
	ID       int    `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"owner"`
	HTMLURL          string      `json:"html_url"`
	Description      string      `json:"description"`
	Fork             bool        `json:"fork"`
	URL              string      `json:"url"`
	ForksURL         string      `json:"forks_url"`
	KeysURL          string      `json:"keys_url"`
	CollaboratorsURL string      `json:"collaborators_url"`
	TeamsURL         string      `json:"teams_url"`
	HooksURL         string      `json:"hooks_url"`
	IssueEventsURL   string      `json:"issue_events_url"`
	EventsURL        string      `json:"events_url"`
	AssigneesURL     string      `json:"assignees_url"`
	BranchesURL      string      `json:"branches_url"`
	TagsURL          string      `json:"tags_url"`
	BlobsURL         string      `json:"blobs_url"`
	GitTagsURL       string      `json:"git_tags_url"`
	GitRefsURL       string      `json:"git_refs_url"`
	TreesURL         string      `json:"trees_url"`
	StatusesURL      string      `json:"statuses_url"`
	LanguagesURL     string      `json:"languages_url"`
	StargazersURL    string      `json:"stargazers_url"`
	ContributorsURL  string      `json:"contributors_url"`
	SubscribersURL   string      `json:"subscribers_url"`
	SubscriptionURL  string      `json:"subscription_url"`
	CommitsURL       string      `json:"commits_url"`
	GitCommitsURL    string      `json:"git_commits_url"`
	CommentsURL      string      `json:"comments_url"`
	IssueCommentURL  string      `json:"issue_comment_url"`
	ContentsURL      string      `json:"contents_url"`
	CompareURL       string      `json:"compare_url"`
	MergesURL        string      `json:"merges_url"`
	ArchiveURL       string      `json:"archive_url"`
	DownloadsURL     string      `json:"downloads_url"`
	IssuesURL        string      `json:"issues_url"`
	PullsURL         string      `json:"pulls_url"`
	MilestonesURL    string      `json:"milestones_url"`
	NotificationsURL string      `json:"notifications_url"`
	LabelsURL        string      `json:"labels_url"`
	ReleasesURL      string      `json:"releases_url"`
	DeploymentsURL   string      `json:"deployments_url"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	PushedAt         time.Time   `json:"pushed_at"`
	GitURL           string      `json:"git_url"`
	SSHURL           string      `json:"ssh_url"`
	CloneURL         string      `json:"clone_url"`
	SvnURL           string      `json:"svn_url"`
	Homepage         string      `json:"homepage"`
	Size             int         `json:"size"`
	StargazersCount  int         `json:"stargazers_count"`
	WatchersCount    int         `json:"watchers_count"`
	Language         string      `json:"language"`
	HasIssues        bool        `json:"has_issues"`
	HasProjects      bool        `json:"has_projects"`
	HasDownloads     bool        `json:"has_downloads"`
	HasWiki          bool        `json:"has_wiki"`
	HasPages         bool        `json:"has_pages"`
	HasDiscussions   bool        `json:"has_discussions"`
	ForksCount       int         `json:"forks_count"`
	MirrorURL        interface{} `json:"mirror_url"`
	Archived         bool        `json:"archived"`
	Disabled         bool        `json:"disabled"`
	OpenIssuesCount  int         `json:"open_issues_count"`
	License          struct {
		Key    string      `json:"key"`
		Name   string      `json:"name"`
		SpdxID string      `json:"spdx_id"`
		URL    interface{} `json:"url"`
		NodeID string      `json:"node_id"`
	} `json:"license"`
	AllowForking             bool          `json:"allow_forking"`
	IsTemplate               bool          `json:"is_template"`
	WebCommitSignoffRequired bool          `json:"web_commit_signoff_required"`
	Topics                   []interface{} `json:"topics"`
	Visibility               string        `json:"visibility"`
	Forks                    int           `json:"forks"`
	OpenIssues               int           `json:"open_issues"`
	Watchers                 int           `json:"watchers"`
	DefaultBranch            string        `json:"default_branch"`
	Permissions              struct {
		Admin    bool `json:"admin"`
		Maintain bool `json:"maintain"`
		Push     bool `json:"push"`
		Triage   bool `json:"triage"`
		Pull     bool `json:"pull"`
	} `json:"permissions"`
}
type WorkflowRuns struct {
	TotalCount   int `json:"total_count"`
	WorkflowRuns []struct {
		ID               int64         `json:"id"`
		Name             string        `json:"name"`
		NodeID           string        `json:"node_id"`
		HeadBranch       string        `json:"head_branch"`
		HeadSha          string        `json:"head_sha"`
		Path             string        `json:"path"`
		DisplayTitle     string        `json:"display_title"`
		RunNumber        int           `json:"run_number"`
		Event            string        `json:"event"`
		Status           string        `json:"status"`
		Conclusion       string        `json:"conclusion"`
		WorkflowID       int           `json:"workflow_id"`
		CheckSuiteID     int64         `json:"check_suite_id"`
		CheckSuiteNodeID string        `json:"check_suite_node_id"`
		URL              string        `json:"url"`
		HTMLURL          string        `json:"html_url"`
		PullRequests     []interface{} `json:"pull_requests"`
		CreatedAt        time.Time     `json:"created_at"`
		UpdatedAt        time.Time     `json:"updated_at"`
		Actor            struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"actor"`
		RunAttempt          int `json:"run_attempt"`
		ReferencedWorkflows []struct {
			Path string `json:"path"`
			Sha  string `json:"sha"`
			Ref  string `json:"ref"`
		} `json:"referenced_workflows"`
		RunStartedAt    time.Time `json:"run_started_at"`
		TriggeringActor struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"triggering_actor"`
		JobsURL            string      `json:"jobs_url"`
		LogsURL            string      `json:"logs_url"`
		CheckSuiteURL      string      `json:"check_suite_url"`
		ArtifactsURL       string      `json:"artifacts_url"`
		CancelURL          string      `json:"cancel_url"`
		RerunURL           string      `json:"rerun_url"`
		PreviousAttemptURL interface{} `json:"previous_attempt_url"`
		WorkflowURL        string      `json:"workflow_url"`
		HeadCommit         struct {
			ID        string    `json:"id"`
			TreeID    string    `json:"tree_id"`
			Message   string    `json:"message"`
			Timestamp time.Time `json:"timestamp"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
			Committer struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"committer"`
		} `json:"head_commit"`
		Repository struct {
			ID       int    `json:"id"`
			NodeID   string `json:"node_id"`
			Name     string `json:"name"`
			FullName string `json:"full_name"`
			Private  bool   `json:"private"`
			Owner    struct {
				Login             string `json:"login"`
				ID                int    `json:"id"`
				NodeID            string `json:"node_id"`
				AvatarURL         string `json:"avatar_url"`
				GravatarID        string `json:"gravatar_id"`
				URL               string `json:"url"`
				HTMLURL           string `json:"html_url"`
				FollowersURL      string `json:"followers_url"`
				FollowingURL      string `json:"following_url"`
				GistsURL          string `json:"gists_url"`
				StarredURL        string `json:"starred_url"`
				SubscriptionsURL  string `json:"subscriptions_url"`
				OrganizationsURL  string `json:"organizations_url"`
				ReposURL          string `json:"repos_url"`
				EventsURL         string `json:"events_url"`
				ReceivedEventsURL string `json:"received_events_url"`
				Type              string `json:"type"`
				SiteAdmin         bool   `json:"site_admin"`
			} `json:"owner"`
			HTMLURL          string      `json:"html_url"`
			Description      interface{} `json:"description"`
			Fork             bool        `json:"fork"`
			URL              string      `json:"url"`
			ForksURL         string      `json:"forks_url"`
			KeysURL          string      `json:"keys_url"`
			CollaboratorsURL string      `json:"collaborators_url"`
			TeamsURL         string      `json:"teams_url"`
			HooksURL         string      `json:"hooks_url"`
			IssueEventsURL   string      `json:"issue_events_url"`
			EventsURL        string      `json:"events_url"`
			AssigneesURL     string      `json:"assignees_url"`
			BranchesURL      string      `json:"branches_url"`
			TagsURL          string      `json:"tags_url"`
			BlobsURL         string      `json:"blobs_url"`
			GitTagsURL       string      `json:"git_tags_url"`
			GitRefsURL       string      `json:"git_refs_url"`
			TreesURL         string      `json:"trees_url"`
			StatusesURL      string      `json:"statuses_url"`
			LanguagesURL     string      `json:"languages_url"`
			StargazersURL    string      `json:"stargazers_url"`
			ContributorsURL  string      `json:"contributors_url"`
			SubscribersURL   string      `json:"subscribers_url"`
			SubscriptionURL  string      `json:"subscription_url"`
			CommitsURL       string      `json:"commits_url"`
			GitCommitsURL    string      `json:"git_commits_url"`
			CommentsURL      string      `json:"comments_url"`
			IssueCommentURL  string      `json:"issue_comment_url"`
			ContentsURL      string      `json:"contents_url"`
			CompareURL       string      `json:"compare_url"`
			MergesURL        string      `json:"merges_url"`
			ArchiveURL       string      `json:"archive_url"`
			DownloadsURL     string      `json:"downloads_url"`
			IssuesURL        string      `json:"issues_url"`
			PullsURL         string      `json:"pulls_url"`
			MilestonesURL    string      `json:"milestones_url"`
			NotificationsURL string      `json:"notifications_url"`
			LabelsURL        string      `json:"labels_url"`
			ReleasesURL      string      `json:"releases_url"`
			DeploymentsURL   string      `json:"deployments_url"`
		} `json:"repository"`
		HeadRepository struct {
			ID       int    `json:"id"`
			NodeID   string `json:"node_id"`
			Name     string `json:"name"`
			FullName string `json:"full_name"`
			Private  bool   `json:"private"`
			Owner    struct {
				Login             string `json:"login"`
				ID                int    `json:"id"`
				NodeID            string `json:"node_id"`
				AvatarURL         string `json:"avatar_url"`
				GravatarID        string `json:"gravatar_id"`
				URL               string `json:"url"`
				HTMLURL           string `json:"html_url"`
				FollowersURL      string `json:"followers_url"`
				FollowingURL      string `json:"following_url"`
				GistsURL          string `json:"gists_url"`
				StarredURL        string `json:"starred_url"`
				SubscriptionsURL  string `json:"subscriptions_url"`
				OrganizationsURL  string `json:"organizations_url"`
				ReposURL          string `json:"repos_url"`
				EventsURL         string `json:"events_url"`
				ReceivedEventsURL string `json:"received_events_url"`
				Type              string `json:"type"`
				SiteAdmin         bool   `json:"site_admin"`
			} `json:"owner"`
			HTMLURL          string      `json:"html_url"`
			Description      interface{} `json:"description"`
			Fork             bool        `json:"fork"`
			URL              string      `json:"url"`
			ForksURL         string      `json:"forks_url"`
			KeysURL          string      `json:"keys_url"`
			CollaboratorsURL string      `json:"collaborators_url"`
			TeamsURL         string      `json:"teams_url"`
			HooksURL         string      `json:"hooks_url"`
			IssueEventsURL   string      `json:"issue_events_url"`
			EventsURL        string      `json:"events_url"`
			AssigneesURL     string      `json:"assignees_url"`
			BranchesURL      string      `json:"branches_url"`
			TagsURL          string      `json:"tags_url"`
			BlobsURL         string      `json:"blobs_url"`
			GitTagsURL       string      `json:"git_tags_url"`
			GitRefsURL       string      `json:"git_refs_url"`
			TreesURL         string      `json:"trees_url"`
			StatusesURL      string      `json:"statuses_url"`
			LanguagesURL     string      `json:"languages_url"`
			StargazersURL    string      `json:"stargazers_url"`
			ContributorsURL  string      `json:"contributors_url"`
			SubscribersURL   string      `json:"subscribers_url"`
			SubscriptionURL  string      `json:"subscription_url"`
			CommitsURL       string      `json:"commits_url"`
			GitCommitsURL    string      `json:"git_commits_url"`
			CommentsURL      string      `json:"comments_url"`
			IssueCommentURL  string      `json:"issue_comment_url"`
			ContentsURL      string      `json:"contents_url"`
			CompareURL       string      `json:"compare_url"`
			MergesURL        string      `json:"merges_url"`
			ArchiveURL       string      `json:"archive_url"`
			DownloadsURL     string      `json:"downloads_url"`
			IssuesURL        string      `json:"issues_url"`
			PullsURL         string      `json:"pulls_url"`
			MilestonesURL    string      `json:"milestones_url"`
			NotificationsURL string      `json:"notifications_url"`
			LabelsURL        string      `json:"labels_url"`
			ReleasesURL      string      `json:"releases_url"`
			DeploymentsURL   string      `json:"deployments_url"`
		} `json:"head_repository"`
	} `json:"workflow_runs"`
}
type WorkflowJobs struct {
	TotalCount int `json:"total_count"`
	Jobs       []struct {
		ID              int64         `json:"id"`
		RunID           int64         `json:"run_id"`
		WorkflowName    string        `json:"workflow_name"`
		HeadBranch      string        `json:"head_branch"`
		RunURL          string        `json:"run_url"`
		RunAttempt      int           `json:"run_attempt"`
		NodeID          string        `json:"node_id"`
		HeadSha         string        `json:"head_sha"`
		URL             string        `json:"url"`
		HTMLURL         string        `json:"html_url"`
		Status          string        `json:"status"`
		Conclusion      string        `json:"conclusion"`
		CreatedAt       time.Time     `json:"created_at"`
		StartedAt       time.Time     `json:"started_at"`
		CompletedAt     time.Time     `json:"completed_at"`
		Name            string        `json:"name"`
		Steps           []interface{} `json:"steps"`
		CheckRunURL     string        `json:"check_run_url"`
		Labels          []string      `json:"labels"`
		RunnerID        interface{}   `json:"runner_id"`
		RunnerName      string        `json:"runner_name"`
		RunnerGroupID   interface{}   `json:"runner_group_id"`
		RunnerGroupName string        `json:"runner_group_name"`
	} `json:"jobs"`
}
type CreateRunner func(string, string, string, string, string, string, []string) string
type DestroyRunner func(string, string, string, string, string, []string, string, string) string
type ReleaseRunner func(string)
type ParseRegistration func(PoolMsg)
type CheckRunner func([]string, string, string, string, string) bool
