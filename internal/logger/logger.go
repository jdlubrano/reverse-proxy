package logger

import (
	"log"
	"os"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
	Infof(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

type NullLogger struct{}

func (logger *NullLogger) Info(msg string)                        {}
func (logger *NullLogger) Infof(format string, a ...interface{})  {}
func (logger *NullLogger) Error(msg string)                       {}
func (logger *NullLogger) Errorf(format string, a ...interface{}) {}

type StdoutLogger struct {
	logger *log.Logger
}

func NewStdoutLogger() *StdoutLogger {
	l := log.New(os.Stdout, "[REVERSE PROXY] ", log.LstdFlags)
	return &StdoutLogger{logger: l}
}

func (logger *StdoutLogger) Close() error {
	return os.Stdout.Close()
}

func (logger *StdoutLogger) Info(msg string) {
	logger.logger.Println(msg)
}

func (logger *StdoutLogger) Infof(format string, a ...interface{}) {
	logger.logger.Printf(format, a...)
}

func (logger *StdoutLogger) Error(msg string) {
	logger.logger.Println(msg)
}

func (logger *StdoutLogger) Errorf(format string, a ...interface{}) {
	logger.logger.Printf(format, a...)
}
