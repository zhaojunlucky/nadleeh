package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"nadleeh/pkg/env"
	workflow "nadleeh/pkg/workflow/action"
	workflowDef "nadleeh/pkg/workflow/model"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

func setupLog() func() {
	if runtime.GOOS == "windows" {
		panic("Windows is currently not supported.")
	}
	logPath := "/var/log/nadleeh"
	fiInfo, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(logPath, 0755)
		if err != nil {
			panic(err)
		}
	} else if !fiInfo.IsDir() {
		panic(fmt.Sprintf("%s must be a directory.", logPath))
	}
	curTime := time.Now()
	nanoseconds := curTime.Nanosecond()
	formattedTime := curTime.UTC().Format("20060102150405")

	logFilePath := path.Join(logPath, fmt.Sprintf("nadleeh_%s_%d.log", formattedTime, nanoseconds))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Info("Failed to log to file, using default stderr")
		return nil
	}
	mw := io.MultiWriter(logFile, os.Stdout)
	exit := make(chan bool)

	r, w, _ := os.Pipe()
	go func() {
		// copy all reads from pipe to multiwriter, which writes to stdout and file
		_, _ = io.Copy(mw, r)
		// when r or w is closed copy will finish and true will be sent to channel
		exit <- true
	}()
	os.Stdout = w
	os.Stderr = w
	log.SetOutput(w)
	return func() {
		// close writer then block on exit channel | this will let mw finish writing before the program exits
		_ = w.Close()
		<-exit
		// close file after all writes have finished
		_ = logFile.Close()
	}
}

func main() {
	logFunc := setupLog()
	if logFunc != nil {
		defer logFunc()
	}
	if len(os.Args) <= 1 {
		log.Panic("usage: nadleeh workflow.yml")
	}
	wYml := os.Args[1]

	log.Infof("run workflow file: %s", wYml)
	ext := strings.ToLower(path.Ext(wYml))
	if ext != ".yaml" && ext != ".yml" {
		log.Panicf("%s must be a yaml file", wYml)
	}
	fi, err := os.Stat(wYml)
	if err != nil {
		log.Panic(err)
	}
	if fi.IsDir() {
		log.Panicf("%s must be a file", wYml)
	}

	wfDef, err := workflowDef.ParseWorkflow(wYml)
	if err != nil {
		log.Panic(err)
	}

	wfa := workflow.NewWorkflowRunAction(wfDef)
	result := wfa.Run(env.NewOSEnv())
	log.Infof("run workflow end, status %d", result.ReturnCode)

	if result.ReturnCode != 0 {
		log.Panic(result.Err)
	}
}
