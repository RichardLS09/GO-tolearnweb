package log

import (
	"log"
	"os"
)

type nilLogger struct{}

var NilLogger *nilLogger

func init() {
	NilLogger = &nilLogger{}
}

func (lg *nilLogger) Debug(msg ...interface{}) {
}

func (lg *nilLogger) Info(msg ...interface{}) {
}

func (lg *nilLogger) Warn(msg ...interface{}) {
}

func (lg *nilLogger) Error(msg ...interface{}) {
}

func (lg *nilLogger) Fatal(msg ...interface{}) {
	log.Println(msg...)
	os.Exit(1)
}
