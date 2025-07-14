# co360 team runner registration
WF_SLS_TEAM := "cnisdp,cdc-preprocesor,soim-monorepo"
WF_SLS_DOMAIN := github.com
WF_SLS_ORG_NAME := ingka-group-digital
WF_SLS_REPOS_NAME := "sp-co360-preprocessor-cdc,soim-monorepo"
WF_SLS_TYPE := "Org"  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "2.0"
WF_SLS_SIZE_MEMORY := "2.0"
WF_SLS_CHARGE_LABELS := "co360-runner" #label should not contains blank
WF_SLS_LABELS := "sp-co360-preprocessor-cdc,soim-monorepo" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_COMMON_CO360 := "\"Pat\":\"${WF_SLS_PAT_CO360}\",\"Key\":\"${WF_SLS_KEY}\",\"Secret\":\"${WF_SLS_SECRET}\",\"Region\":\"${WF_SLS_REGION}\",\"SecGpId\":\"${WF_SLS_SECGROUP_ID}\",\"VSwitchId\":\"${WF_SLS_VSWITCH_ID}\",\"PullInterval\":\"${WF_SLS_PULL_INTERVAL_Sec}\""
WF_SLS_REG_ITEM := "{${WF_SLS_REG_COMMON_CO360}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
REG := $(WF_SLS_REG_ITEM)
WF_SLS_REGS := "[${REG}]"