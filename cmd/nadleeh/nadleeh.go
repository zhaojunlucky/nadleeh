package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	log "github.com/sirupsen/logrus"
	"io"
	"nadleeh/internal/argument"
	"nadleeh/pkg/encrypt"
	workflow "nadleeh/pkg/workflow/action"
	"os"
	"path"
	"runtime"
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

func createArgsMap(args []argparse.Arg) map[string]argparse.Arg {
	argsMap := make(map[string]argparse.Arg, len(args))
	for _, arg := range args {
		argsMap[arg.GetLname()] = arg
	}
	return argsMap
}

func main() {
	logFunc := setupLog()
	if logFunc != nil {
		defer logFunc()
	}

	parser := argument.NewNadleehCliParser()
	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Println(parser.Usage(err))
		return
	}

	for _, arg := range parser.GetArgs() {
		if arg.GetLname() == "help" && arg.GetParsed() {
			fmt.Println(parser.Usage(nil))
			return
		}
	}

	for _, cmd := range parser.GetCommands() {
		if !cmd.Happened() {
			log.Debugf("comand %s not specified", cmd.GetName())
			continue
		}
		switch cmd.GetName() {
		case "run":
			workflow.RunWorkflow(cmd, createArgsMap(cmd.GetArgs()))
		case "keypair":
			encrypt.GenerateKeyPair(cmd, createArgsMap(cmd.GetArgs()))
		case "encrypt":
			encrypt.Encrypt(cmd, createArgsMap(cmd.GetArgs()))
		default:
			log.Fatalf("unknown command: %s", cmd.GetName())
		}
	}
}
