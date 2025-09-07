package common

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func MustSetEnv(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		log.Fatalf("failed to set env for key %s", key)
	}
}

func MustSetEnvs(envs map[string]string) {
	for k, v := range envs {
		MustSetEnv(k, v)
	}
}
