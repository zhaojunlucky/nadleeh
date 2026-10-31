package main

import (
	"fmt"
	"nadleeh/pkg/common"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/runner"

	log "github.com/sirupsen/logrus"

	"io"
	"nadleeh/internal/argument"
	"nadleeh/pkg/encrypt"

	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

func ensureLogDir() string {
	candidates := []string{"/var/log/nadleeh"}
	if cacheDir, err := os.UserCacheDir(); err == nil && cacheDir != "" {
		candidates = append(candidates, filepath.Join(cacheDir, "nadleeh", "logs"))
	}
	candidates = append(candidates, filepath.Join(os.TempDir(), "nadleeh"))

	for _, logPath := range candidates {
		fiInfo, err := os.Stat(logPath)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(logPath, 0755); err != nil {
				log.Warnf("failed to create log directory %s: %v", logPath, err)
				continue
			}
			return logPath
		}
		if err != nil {
			log.Warnf("failed to access log directory %s: %v", logPath, err)
			continue
		}
		if !fiInfo.IsDir() {
			log.Warnf("%s must be a directory", logPath)
			continue
		}
		return logPath
	}
	return ""
}

func setupLog() *os.File {
	if runtime.GOOS == "windows" {
		log.Fatal("Windows is currently not supported.")
	}
	logPath := ensureLogDir()
	if logPath == "" {
		log.Info("Failed to create log directory, using default stderr")
		return nil
	}
	curTime := time.Now()
	nanoseconds := curTime.Nanosecond()
	formattedTime := curTime.UTC().Format("2006-01-02-15_04_05")

	logFilePath := path.Join(logPath, fmt.Sprintf("nadleeh_%s_%d.log", formattedTime, nanoseconds))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Info("Failed to log to file, using default stderr")
		return nil
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
	log.SetLevel(log.InfoLevel)
	return logFile
}

func main() {
	if logFile := setupLog(); logFile != nil {
		defer logFile.Close()
	}
	log.Infof("nadleeh %s (%s) - https://gundamz.net/nadleeh/", common.Version, common.BuildDate)

	handlers := &argument.CommandHandlers{
		RunHandler: func(args *argument.RunArgs) {
			if argument.Verbose {
				log.SetLevel(log.DebugLevel)
			}
			log.Debug("args: ", args.Args)
			argEnv := argument.CreateArgsEnv(args.Args)
			runner.RunWorkflow(core.NewWorkflowArgsFromRunArgs(args), argEnv)
		},
		WfHandler: func(args *argument.WorkflowArgs) {
			if argument.Verbose {
				log.SetLevel(log.DebugLevel)
			}
			runner.RunWorkflowConfig(args)
		},
		KeypairHandler: func(args *argument.KeypairArgs) {
			if argument.Verbose {
				log.SetLevel(log.DebugLevel)
			}
			encrypt.GenerateKeyPair(args)
		},
		EncryptHandler: func(args *argument.EncryptArgs) {
			if argument.Verbose {
				log.SetLevel(log.DebugLevel)
			}
			encrypt.Encrypt(args)
		},
	}

	rootCmd := argument.NewNadleehCliParser(handlers)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
