package common

import "os"

func Ternary(cond bool, true_val, false_val interface{}) interface{} {
	if cond {
		return true_val
	} else {
		return false_val
	}
}

func SetContextLogLevel(level string) {
	os.Setenv("CTX_LOG_LEVEL", level)
}

func SetLogursLogLevel(level string) {
	os.Setenv("LOGRUS_LOG_LEVEL", level)
}

func SetFmtLogLevel(level string) {
	os.Setenv("FMT_LOG_LEVEL", level)
}
