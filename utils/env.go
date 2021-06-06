package utils

import (
	"os"
	"strings"
)

func GetEnv() (envList map[string]string) {
	envList = make(map[string]string)
	for _, s := range os.Environ() {
		pair := strings.SplitN(s, "=", 2)
		envList[pair[0]] = pair[1]
	}
	return
}
