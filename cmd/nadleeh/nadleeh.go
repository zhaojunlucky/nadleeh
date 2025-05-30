package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	log "github.com/sirupsen/logrus"
	"io"
	"nadleeh/internal/argument"
	"nadleeh/pkg/encrypt"
	"nadleeh/pkg/env"
	workflow "nadleeh/pkg/workflow/action"
	"os"
	"path"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

var Version = "1.0.1-dev"

func setupLog() {
	if runtime.GOOS == "windows" {
		log.Fatal("Windows is currently not supported.")
	}
	logPath := "/var/log/nadleeh"
	fiInfo, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(logPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
	} else if !fiInfo.IsDir() {
		log.Fatalf("%s must be a directory.", logPath)
	}
	curTime := time.Now()
	nanoseconds := curTime.Nanosecond()
	formattedTime := curTime.UTC().Format("2006-01-02-15_04_05")

	logFilePath := path.Join(logPath, fmt.Sprintf("nadleeh_%s_%d.log", formattedTime, nanoseconds))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Info("Failed to log to file, using default stderr")
		return
	}
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
			//return frame.Function, fileName
			return "", fileName
		},
	})
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))
}

func createArgsMap(args []argparse.Arg, exclude []string) map[string]argparse.Arg {
	argsMap := make(map[string]argparse.Arg, len(args))
	for _, arg := range args {
		if len(exclude) > 0 && slices.Contains(exclude, arg.GetLname()) {
			continue
		}
		argsMap[arg.GetLname()] = arg
	}
	return argsMap
}

func main() {
	setupLog()
	log.Infof("nadleeh %s", Version)

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
			args := createArgsEnv(cmd.GetArgs())
			workflow.RunWorkflow(cmd, createArgsMap(cmd.GetArgs(), []string{"arg"}), args)
		case "keypair":
			encrypt.GenerateKeyPair(cmd, createArgsMap(cmd.GetArgs(), nil))
		case "encrypt":
			encrypt.Encrypt(cmd, createArgsMap(cmd.GetArgs(), nil))
		default:
			log.Fatalf("unknown command: %s", cmd.GetName())
		}
	}
}

func createArgsEnv(args []argparse.Arg) env.Env {
	argMap := make(map[string]string)
	for _, arg := range args {
		if arg.GetLname() == "arg" && arg.GetParsed() {
			argList := arg.GetResult().(*[]string)
			for _, argLine := range *argList {
				key, value, found := strings.Cut(argLine, "=")
				if !found {
					continue
				}
				argMap[strings.TrimSpace(key)] = strings.TrimSpace(value)
			}
		}
	}
	argEnv := env.NewReadEnv(nil, argMap)
	return argEnv
}
