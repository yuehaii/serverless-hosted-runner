package common

import (
	"os"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

func Ternary(cond bool, trueVal, falseVal interface{}) interface{} {
	if cond {
		return trueVal
	} else {
		return falseVal
	}
}

func TernaryComparable[T comparable](cond bool, trueval T, falseval T, typeIdentify ...T) T {
	if cond {
		return trueval
	} else {
		return falseval
	}
}

func SetContextLogLevel(level string) {
	if err := os.Setenv("CTX_LOG_LEVEL", level); err != nil {
		logrus.Errorf("fail to set context log level to env: %v", err)
	}
}

func SetLogursLogLevel(level string) {
	if err := os.Setenv("LOGRUS_LOG_LEVEL", level); err != nil {
		logrus.Errorf("fail to set logrus log level to env: %v", err)
	}
}

func SetFmtLogLevel(level string) {
	if err := os.Setenv("FMT_LOG_LEVEL", level); err != nil {
		logrus.Errorf("fail to set fmt log level to env: %v", err)
	}
}
