// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

const consoleColorTag = 0x1B

// ConsoleRed 控制台红色字符
func ConsoleRed(txt string) string {
	return fmt.Sprintf("%c[31m%s%c[0m", consoleColorTag, txt, consoleColorTag)
}

const envKeyPrefix = "BAS_"

func envKey(name string) string {
	return envKeyPrefix + name
}

// 尝试从环境变量中找到真正要执行的命令
func getRawBinName(binName string) string {
	namePath := "PATH"
	if isWindows() {
		namePath = "path"
	}

	var found int
	p := os.Getenv(namePath)
	for _, dir := range filepath.SplitList(p) {
		p1, err1 := exec.LookPath(filepath.Join(dir, binName))
		if err1 == nil {
			found++
			if found == 2 {
				return p1
			}
		}
	}
	return ""
}
