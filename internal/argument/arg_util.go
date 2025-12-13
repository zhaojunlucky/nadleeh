package argument

import (
	"fmt"
	"strings"

	"github.com/zhaojunlucky/golib/pkg/env"
)

// CreateArgsEnv creates an env.Env from a slice of argument strings in the format "key=value"
func CreateArgsEnv(args []string) env.Env {
	argMap := make(map[string]string)
	for _, argLine := range args {
		key, value, found := strings.Cut(argLine, "=")
		if !found {
			argMap[strings.TrimSpace(argLine)] = ""
			continue
		}
		key = strings.TrimSpace(key)
		argVal, ok := argMap[key]
		if ok {
			argMap[key] = fmt.Sprintf("%s,%s", argVal, value)
		} else {
			argMap[key] = value
		}
	}
	return env.NewReadEnv(env.NewEmptyReadEnv(), argMap)
}
