#!/usr/bin/env bash

[[ -v _ACCESS ]] && return  
_ACCESS="$(realpath "${BASH_SOURCE[0]}")"; declare -rg _ACCESS 

_SRC_PATH="$(dirname $_ACCESS)"

function access::_url {
    printf "https://www.googleapis.com/oauth2/v4/token"
}

function access::_base64 {
    printf "$1" | access::_base64stream
}

function access::_base64stream {
    base64 | tr '/+' '_-' | tr -d '=\n'
}

function access::jwt {
    valid_for_sec="${2:-3600}"  
    # private_key=$(echo $GOOGLE_CREDENTIALS | jq .private_key | sed 's/"//g')
    private_key=$GOOGLE_CREDENTIALS_PRIVATEKEY
    sa_email=$(echo $GOOGLE_CREDENTIALS | jq .client_email)
    header='{"alg":"RS256","typ":"JWT"}'
claim=$(cat <<EOF
    {
        "iss": $sa_email,
        "scope": "$1",
        "aud": "$(access::_url)",
        "exp": $(($(date +%s) + $valid_for_sec)),
        "iat": $(date +%s)
    }
EOF
)
    claim_c=$(echo $claim | jq -c . | sed 's/"/\\"/g')
    request_body="$(access::_base64 "$header").$(access::_base64 "$claim_c")"
    signature=$(openssl dgst -sha256 -sign <(echo "$private_key") <(printf "$request_body") | access::_base64stream)
    # echo "${private_key//'\n'/$'\n'}" > privatekey.pem
    # echo "$request_body" > body.txt
    # signature=$(openssl dgst -sha256 -sign privatekey.pem body.txt | access::_base64stream)
    # rm privatekey.pem body.txt
    # echo "the request_body is '$(echo $request_body)'"
    # echo "the signature is '$(echo $signature)'"

    printf "$request_body.$signature"
}

function access::token {
    jwt_token=$(access::jwt "$1")
    # echo $jwt_token

    curl -s -X POST $(access::_url) \
        --data-urlencode 'grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer' \
        --data-urlencode "assertion=$jwt_token" | jq -r .access_token

    # resp=$(curl -s -X POST $(access::_url) \
    #     --data-urlencode 'grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer' \
    #     --data-urlencode "assertion=$request_body.$signature")
    # echo "resp is $resp"
    # echo "access_token is $(echo $resp | jq -r .access_token)"
}

# export GOOGLE_CREDENTIALS=$(cat ~/ingka-ccoe-gcptaskforce-dev-61954b0eccc8.json | tr -s '\n' ' ')
# export GOOGLE_CREDENTIALS_PRIVATEKEY_BASE64=$(jq -r .private_key ~/ingka-ccoe-gcptaskforce-dev-61954b0eccc8.json | base64)
# export GOOGLE_CREDENTIALS_PRIVATEKEY=$(echo ${GOOGLE_CREDENTIALS_PRIVATEKEY_BASE64} | base64 -d)

# ## Scope
# # TODO: there is no scope only for batch service, using cloud platform instead. ref: 
# # https://developers.google.com/identity/protocols/oauth2/scopes
# # https://cloud.google.com/docs/authentication#authorization-gcp
# scope=https://www.googleapis.com/auth/cloud-platform
# access::token $scope
