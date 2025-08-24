package file

import (
	"os"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type FileLock struct {
	lockFile string
	fd       *os.File
}

func (l *FileLock) Lock() error {
	lockFile, err := os.Create(l.lockFile)
	if err != nil {
		log.Errorf("failed to create lock file: %v", err)
		return err
	}
	l.fd = lockFile
	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX)
	if err != nil {
		l.close()
		log.Errorf("failed to acquire lock: %v", err)
		return err
	}
	return nil
}

func (l *FileLock) Unlock() error {
	if l.fd != nil {
		err := syscall.Flock(int(l.fd.Fd()), syscall.LOCK_UN)
		if err != nil {
			log.Errorf("failed to release lock: %v", err)
		}
		return err
		l.close()
	}
	return nil
}

func (l *FileLock) close() error {
	if l.fd != nil {
		err := l.fd.Close()
		l.fd = nil
		return err
	}
	return nil
}

func NewFileLock(lockFile string) *FileLock {
	return &FileLock{
		lockFile: lockFile,
	}
}
