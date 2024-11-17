# 1.1.0 
```text
First version release. It includes the basic function like multiple tenant mode and centrlized mode
```

# 1.2.0
```test
This version contains the lazy installation mode. We don't need to register the runner info with workflow. It simplified the installation. And can install the dispatcher and runner at same tenant.
```

# 1.3.0
```text
This feature included the pull mode to retrieve the workflow information. It dose not rely on SLB to retrieve request from git's webhook. 
It solved the random timeout issue (caused by GFW) when the dispatcher been created in Shanghai region. Git webhook dose not has the retry mechanism to retry the failed delivery of the webhook. 
```

# 1.4.0 
```text
This version integrated with Allen portal. It also optimized the pull mode and solved the tf lock lost and ali tf provide issue in smoke test. 
```

# 1.5.0
```text
This version introduced dynamic labels. User can define the runner CPU and memory size on their workflow's run-on labels. 
```

# 1.6.0
```text
This version solved the dns poision issue. Runner can visit git and runner site normally.
Integration with Context via app monitor agent. If the app monitor agent version update, please change the ver in go.mod and run below command to update go.sum before building the image.  
```
```bash
cd src; go get serverless-hosted-runner/dispatcher
```