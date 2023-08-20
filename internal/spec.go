// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/8/20

package internal

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

func parserSpecial(name string, r *Rule) error {
	if len(r.Spec) == 0 {
		return nil
	}
	switch name {
	case "go":
		s := &specGo{}
		return s.Parser(r)
	}
	return nil
}

type specGo struct {
	// GoVersionFile  定义 go 版本的文件，目前支持 go.mod、no
	GoVersionFile string

	// GoWork 是否修订当前目录未在 go.work 中定义不能运行的问题
	// 目前支持:
	// 1 auto: 若模块不在 go.work，则设置环境变量 GOWORK=off
	// 2 no: 跳过
	GoWork string
}

func (s *specGo) Parser(r *Rule) error {
	if err := convertByJSON(r.Spec, s); err != nil {
		return err
	}

	if err := s.goVersionFile(r); err != nil {
		return err
	}

	if err := s.goWork(r); err != nil {
		return err
	}

	return nil
}

func (s *specGo) goVersionFile(r *Rule) error {
	if s.GoVersionFile == "" || s.GoVersionFile == "no" {
		return nil
	}
	if s.GoVersionFile != "go.mod" {
		return fmt.Errorf("not support GoVersionFile=%q, now support 'go.mod'", s.GoVersionFile)
	}
	fp, err := findFileUpper(s.GoVersionFile, 128)
	if err != nil {
		if errors.Is(err, errFileNotFound) {
			return nil
		}
		return err
	}
	f, err := parserGoModFile(fp)
	if err != nil {
		return err
	}
	if f.Go != nil && f.Go.Version != "" {
		cmd := "go" + f.Go.Version
		filePath, err := exec.LookPath(cmd)
		if err != nil {
			if r.Trace {
				log.Printf("LookPath %q with error, ignore it", cmd)
			}
			// 当不存在的时候，忽略错误
			return nil
		}
		r.Cmd = filePath
	}
	return nil
}

func (s *specGo) goWork(r *Rule) error {
	if s.GoWork == "" || s.GoWork == "no" {
		return nil
	}
	if s.GoWork != "auto" {
		return fmt.Errorf("not support GoWork=%q", s.GoWork)
	}

	fp, err := findFileUpper("go.work", 128)
	if err != nil {
		if errors.Is(err, errFileNotFound) {
			return nil
		}
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	wd += string(filepath.Separator)

	wfDir := filepath.Dir(fp)

	code, err := os.ReadFile(fp)
	if err != nil {
		return err
	}
	wf, err := modfile.ParseWork(fp, code, nil)
	if err != nil {
		return err
	}
	for _, m := range wf.Use {
		fullPath := filepath.Join(wfDir, m.Path) + string(filepath.Separator)
		if strings.HasPrefix(wd, fullPath) {
			return nil
		}
	}
	r.Env = append([]string{"GOWORK=off"}, r.Env...)
	if r.Trace {
		log.Println("module not in go.work, set GOWORK=off")
	}
	return nil
}
