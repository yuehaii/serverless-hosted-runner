package main

import (
	"fmt"
	"os"
	common "serverless-hosted-runner/common"

	kingpin "github.com/alecthomas/kingpin/v2"
)

var (
	ini       = kingpin.Flag("init", "init safer").Default("true").Bool()
	sec       = kingpin.Flag("secret", "secret to encrypt").Short('s').String()
	plain     = kingpin.Flag("plain", "plain text to encrypt").Short('p').String()
	testsec   = kingpin.Flag("testSec", "TODO: test features, will disable later").String()
	testplain = kingpin.Flag("testPlain", "TODO: test features, will disable later").String()
)

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	if *ini {
		handler := common.RSACryptography("")
		handler.GenKeys()
	} else if len(*plain) > 0 && len(os.Getenv("SLS_ENC_KEY")) > 0 {
		handler := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))
		fmt.Println(handler.EncryptMsg(*plain))
	} else if len(*sec) > 0 {
		handler := common.RSACryptography("")
		fmt.Println(handler.EncryptMsg(*sec))
	} else if len(*testsec) > 0 && len(os.Getenv("SLS_ENC_KEY")) > 0 {
		handler := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))
		fmt.Println(handler.DecryptMsg(*testsec))
	} else if len(*testplain) > 0 {
		handler := common.RSACryptography("")
		fmt.Println(handler.DecryptMsg(*testplain))
	}
}
