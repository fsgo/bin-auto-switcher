// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/11/19

package actuator

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// FindExec 查找指定文件名，名在其目录下执行执行子命令
type FindExec struct {
	Args []string
}

func (fe *FindExec) Name() string {
	return Prefix + "find-exec"
}

func (fe *FindExec) Run(ctx context.Context) error {
	var name string
	var useRegular bool
	fset := flag.NewFlagSet(fe.Name(), flag.ContinueOnError)
	fset.StringVar(&name, "name", "go.mod", "find file name")
	fset.BoolVar(&useRegular, "e", false, "name as regular expression")
	if err := fset.Parse(fe.Args); err != nil {
		return err
	}

	if len(name) == 0 {
		return errors.New("-name is empty")
	}

	var reg *regexp.Regexp
	if useRegular {
		r, err := regexp.Compile(name)
		if err != nil {
			return fmt.Errorf("regexp.Compile(%q): %v", name, err)
		}
		reg = r
	}

	match := func(fileName string) bool {
		if useRegular {
			return reg.MatchString(fileName)
		}
		return fileName == name
	}

	cmdName := fset.Arg(0)
	if len(cmdName) == 0 {
		return errors.New("cmd is empty")
	}

	return fe.run(ctx, match, cmdName, fset.Args()[1:])
}

func (fe *FindExec) run(ctx context.Context, match func(fileName string) bool, cmdName string, args []string) error {
	var index int
	var fail int

	dirs := map[string]bool{}

	err := filepath.Walk("./", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if err1 := ctx.Err(); err1 != nil {
			return err1
		}

		fileName := filepath.Base(path)

		if !match(fileName) {
			return nil
		}

		dir := filepath.Dir(path)
		// 一个目录只执行一次命令
		if dirs[dir] {
			return nil
		}

		dirs[dir] = true

		index++

		rr := &Config{
			Name: cmdName,
			Args: args,
			Dir:  dir,
		}

		s0 := color.GreenString("%3d.", index)
		s1 := color.CyanString("Dir: %s, MatchFile: %s", dir, fileName)
		s2 := color.YellowString("Exec: %s", rr.String())
		log.Println(s0, s1, s2)

		if e1 := rr.Run(ctx); e1 != nil {
			fail++
			color.Red(e1.Error())
		}
		return fs.SkipDir
	})
	if err != nil {
		return err
	}
	if fail > 0 {
		return fmt.Errorf("total %d/%d tasks failed", fail, index)
	}

	if index == 0 {
		log.Printf("file not found, skipped for %s", cmdName)
	}

	return nil
}

func (fe *FindExec) String() string {
	return fe.Name() + " " + strings.Join(fe.Args, " ")
}

func init() {
	register(func(args []string) Actuator {
		return &FindExec{
			Args: args,
		}
	})
}
