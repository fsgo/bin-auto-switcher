// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/11/1

package internal

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Condition string

func (c Condition) Allow() bool {
	if len(c) == 0 {
		return true
	}
	str := strings.TrimSpace(string(c))
	if cd, ok := conditions[str]; ok {
		return cd()
	}
	arr := strings.SplitN(str, " ", 2)
	if len(arr) != 2 {
		return false
	}
	if fn, ok := conditionsFuncs[arr[0]]; ok {
		return fn(arr[1])
	}
	return false
}

var conditions = map[string]func() bool{
	"go_module": inGoModule,
}

var conditionsFuncs = map[string]func(v string) bool{
	"has_file": hasFile,
	"not_has_file": func(v string) bool {
		return !hasFile(v)
	},
	"exec":              condExec,
	"git_status_change": gitStatusChange,
	"in_dir":            condInDir,
	"not_in_dir": func(v string) bool {
		return !condInDir(v)
	},
}

func inGoModule() bool {
	return hasFile("go.mod")
}

func gitStatusChange(str string) bool {
	str = strings.TrimSpace(str)
	if str == "" {
		return true
	}
	str = strings.ReplaceAll(str, ";", ",")
	exts := strings.Split(str, ",")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	gitBin := getRawBinName("git")
	cmd := exec.CommandContext(ctx, gitBin, "status", "-s", "--column=always")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("exec:", cmd.String(), err)
		return false
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		ext := filepath.Ext(line)
		if ext != "" && slices.Contains(exts, ext) {
			return true
		}
	}

	return false
}

func hasFile(name string) bool {
	wd, err := os.Getwd()
	if err != nil {
		log.Println("os.Getwd failed:", err)
		return false
	}
	for i := 0; i < strings.Count(wd, string(filepath.Separator)); i++ {
		fp := filepath.Join(wd, name)
		st, err := os.Stat(fp)
		if err == nil && !st.IsDir() {
			return true
		}
		if !errors.Is(err, fs.ErrNotExist) {
			log.Printf("os.Stat(%q) failed: %v", fp, err)
			return false
		}
		wd = filepath.Dir(wd)
	}
	return false
}

func condExec(v string) bool {
	v = strings.TrimSpace(v)
	if len(v) == 0 {
		return false
	}
	arr := strings.Fields(v)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, arr[0], arr[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr
	return cmd.Run() == nil
}

func condInDir(v string) bool {
	pwd, err := os.Getwd()
	if err != nil {
		return false
	}
	dirs := strings.Split(v, ";")
	for i := 0; i < len(dirs); i++ {
		dir := strings.TrimSpace(dirs[i])
		if dir == "" {
			continue
		}
		dir = filepath.Clean(dir) + string(filepath.Separator)
		if strings.HasPrefix(pwd, dir) {
			return true
		}
	}
	return false
}
