// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"debug/buildinfo"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"golang.org/x/mod/modfile"
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

var rawBinNames sync.Map

// 尝试从环境变量中找到真正要执行的命令
func getRawBinName(binName string) string {
	v, ok := rawBinNames.Load(binName)
	if ok {
		return v.(string)
	}

	namePath := "PATH"
	if isWindows() {
		namePath = "path"
	}
	p := os.Getenv(namePath)
	for _, dir := range filepath.SplitList(p) {
		p1, err1 := exec.LookPath(filepath.Join(dir, binName))
		if err1 == nil && !isSelfBin(p1) {
			rawBinNames.Store(binName, p1)
			if enableTrace {
				log.Println("RawBinName:", p1)
			}
			return p1
		}
	}
	rawBinNames.Store(binName, "")
	return ""
}

func isSelfBin(p string) bool {
	info, err := buildinfo.ReadFile(p)
	if err != nil {
		return false
	}
	main, ok := debug.ReadBuildInfo()
	if !ok {
		return false
	}
	return info.Path == main.Path
}

func convertByJSON(data any, to any) error {
	bf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bf, to)
}

var errFileNotFound = errors.New("file not found")

func findFileUpper(name string, max int) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := wd
	for i := 0; i < max; i++ {
		fp := filepath.Join(current, name)
		st, err1 := os.Stat(fp)
		if err1 == nil && !st.IsDir() {
			return fp, nil
		}
		next := filepath.Dir(current)
		if next == current {
			break
		}
		current = next
	}
	return "", fmt.Errorf("%w: %s", errFileNotFound, name)
}

func parserGoModFile(fp string) (*modfile.File, error) {
	content, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	return modfile.Parse(fp, content, nil)
}

func disableHooks() bool {
	// 环境变量 BAS_NoHook=true 或者 bas=off
	return os.Getenv(envKey("NoHook")) != "" || os.Getenv("bas") == "off"
}
