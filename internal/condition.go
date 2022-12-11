// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/11/1

package internal

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	"exec":   condExec,
	"in_dir": condInDir,
	"not_in_dir": func(v string) bool {
		return !condInDir(v)
	},
}

func inGoModule() bool {
	return hasFile("go.mod")
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
		if !os.IsNotExist(err) {
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
	dirs := strings.Fields(v)
	for i := 0; i < len(dirs); i++ {
		dir := filepath.Clean(dirs[i]) + string(filepath.Separator)
		if strings.HasPrefix(dir, pwd) {
			return true
		}
	}
	return false
}
