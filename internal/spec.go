// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/8/20

package internal

import (
	"errors"
	"fmt"
	"os/exec"
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
	// GoVersionFile  定义 go 版本的文件，目前支持 go.mod
	GoVersionFile string
}

func (s *specGo) Parser(r *Rule) error {
	if err := convertByJSON(r.Spec, s); err != nil {
		return err
	}
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
		fp, err := exec.LookPath(cmd)
		if err != nil {
			return err
		}
		r.Cmd = fp
	}
	return nil
}
