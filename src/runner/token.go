package main

type RunnerToken struct {
	Token string `json:"token"`
}

type JitToken struct {
	Runner struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Os     string `json:"os"`
		Status string `json:"status"`
		Busy   bool   `json:"busy"`
		Labels []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"labels"`
		RunnerGroupID int `json:"runner_group_id"`
	} `json:"runner"`
	EncodedJitConfig string `json:"encoded_jit_config"`
}