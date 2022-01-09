// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"os"
	"runtime"
	"strings"
)

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func dedupEnv(caseInsensitive bool, env []string) []string {
	out := make([]string, 0, len(env))
	saw := map[string]int{} // to index in the array
	for _, kv := range env {
		eq := strings.Index(kv, "=")
		if eq < 1 {
			out = append(out, kv)
			continue
		}
		k := kv[:eq]
		if caseInsensitive {
			k = strings.ToLower(k)
		}
		if dupIdx, isDup := saw[k]; isDup {
			out[dupIdx] = kv
		} else {
			saw[k] = len(out)
			out = append(out, kv)
		}
	}
	return out
}

var homeDir string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	homeDir = home
}
