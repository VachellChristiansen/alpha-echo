package main

import (
	"fmt"
	"log"
	"os"
)

type Logger map[string]*log.Logger

func NewLogger() Logger {
	var (
		err error
	)
	logFilenames := map[string]string{
		"INFO":    "info.log",
		"WARNING": "warning.log",
		"ERROR":   "error.log",
		"MISC":    "misc.log",
		"TASK":    "task.log",
	}
	logFiles := make(map[string]*os.File)
	logs := make(map[string]*log.Logger)

	for k, v := range logFilenames {
		logFiles[k], err = os.OpenFile(fmt.Sprintf("logs/%s", v), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("Failed Opening Log File: %s, %v", v, err)
		}
		logs[k] = log.New(logFiles[k], fmt.Sprintf("%v : ", k), log.Ldate|log.Ltime|log.Lshortfile)
	}

	return logs
}
