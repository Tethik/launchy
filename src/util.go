package main

import (
	"os"
	"strings"
)

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func environMap() map[string]string {
	envs := map[string]string{}
	unmappedEnvs := os.Environ()

	for _, e := range unmappedEnvs {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envs[key] = value
	}
	return envs
}
