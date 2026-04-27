//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-30

package actuator

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var _ Actuator = (*GitAddModify)(nil)

type GitAddModify struct {
	Args []string
}

func (gm *GitAddModify) Name() string {
	return Prefix + "git-am"
}

// Run
//
//	git status -su
//	 XY 文件路径
//	 MM file.html   ->  已修改（modified），且已 git add,而且工作区有修改（未 add）
//
// 第一列（X）的含义: 暂存区（index）状态
//
//	 标志	含义
//	    （空格）	暂存区无变化
//	   M	已修改（modified），且已 git add
//	   A	已新增（added），已加入暂存区
//	   D	已删除（deleted），已暂存
//	   R	重命名（renamed）
//	   C	复制（copied）
//	   U	冲突（unmerged）
//
//	第二列（Y）的含义:工作区（working tree）状态
//	   标志	含义
//	   （空格）	工作区无变化
//	   M	工作区有修改（未 add）
//	   D	工作区已删除
//	   ?	未跟踪文件（配合 -u）
//	   U	冲突
//
//	?? 是一个整体，表示“未跟踪文件（untracked）”,既不在暂存区，也不在版本库中 —— 完全是 Git 不认识的新文件
func (gm *GitAddModify) Run(ctx context.Context) error {
	gitBin := GetRawBinName("git")
	cmd := exec.CommandContext(ctx, gitBin, "status", "-su")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("exec:", cmd.String(), err)
		return err
	}
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil
	}
	var flagName string
	var useRegular bool
	fset := flag.NewFlagSet(gm.Name(), flag.ContinueOnError)
	fset.StringVar(&flagName, "name", "", "find file name")
	fset.BoolVar(&useRegular, "e", false, "name as regular expression")
	if err = fset.Parse(gm.Args); err != nil {
		return err
	}
	if flagName == "" {
		return errors.New("flag -name is required")
	}

	var reg *regexp.Regexp
	if useRegular {
		reg, err = regexp.Compile(flagName)
		if err != nil {
			return fmt.Errorf("regexp.Compile(%q): %v", flagName, err)
		}
	}

	match := func(fileName string) bool {
		if flagName == "*" {
			return true
		}
		name := filepath.Base(fileName)
		if reg != nil {
			return reg.MatchString(name)
		}
		return name == flagName
	}

	lines := strings.Split(string(out), "\n")
	var errs []error
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		xy := line[:2]
		// D: 删除状态,U:冲突
		if strings.ContainsAny(xy, "DU") {
			continue
		}
		filename := strings.TrimSpace(line[2:])
		if filename == "" || !match(filename) {
			continue
		}

		var args []string
		for _, a := range fset.Args() {
			a = strings.ReplaceAll(a, "{name}", filename)
			args = append(args, a)
		}
		var extArgs []string
		if len(args) > 1 {
			extArgs = args[1:]
		}
		sub := exec.CommandContext(ctx, args[0], extArgs...)
		if Trace.Load() {
			log.Println("Exec:", sub.String())
		}
		err1 := sub.Run()
		if err1 != nil {
			errs = append(errs, err1)
		}
	}
	return errors.Join(errs...)
}

func (gm *GitAddModify) String() string {
	return gm.Name() + " " + strings.Join(gm.Args, " ")
}

func init() {
	register(func(args []string) Actuator {
		return &GitAddModify{
			Args: args,
		}
	})
}
