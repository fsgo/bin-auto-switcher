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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xanygo/anygo/cli/xcolor"
)

// FindExec 查找指定文件名，名在其目录下执行执行子命令
type FindExec struct {
	Args     []string
	flagName string
	wd       string
}

func (fe *FindExec) Name() string {
	return Prefix + "find-exec"
}

type ss string

func (s ss) Match(name string) bool {
	arr := strings.Split(string(s), ",")
	for _, a := range arr {
		a = strings.TrimSpace(a)
		if len(a) == 0 {
			continue
		}
		if strings.Contains(name, a) {
			return true
		}
	}
	return false
}

func (fe *FindExec) Run(ctx context.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fe.wd = wd

	var useRegular bool
	var notInDirs string
	var rootDir string
	fset := flag.NewFlagSet(fe.Name(), flag.ContinueOnError)
	fset.StringVar(&rootDir, "root", ".git,go.mod", "search up root dir")
	fset.StringVar(&fe.flagName, "name", "go.mod", "find file name")
	fset.BoolVar(&useRegular, "e", false, "name as regular expression")
	fset.StringVar(&notInDirs, "dir_not", "", "not in these dir names, multiple are connected with ','")
	if err = fset.Parse(fe.Args); err != nil {
		return err
	}

	if len(fe.flagName) == 0 {
		return errors.New("-name is empty")
	}

	var reg *regexp.Regexp
	if useRegular {
		r, err := regexp.Compile(fe.flagName)
		if err != nil {
			return fmt.Errorf("regexp.Compile(%q): %v", fe.flagName, err)
		}
		reg = r
	}

	match := func(fileName string) bool {
		if len(notInDirs) > 0 {
			ap, err := filepath.Abs(fileName)
			if err != nil {
				log.Printf("filepath.Abs(%q) failed: %v", fileName, err)
				return false
			}
			if ss(notInDirs).Match(ap) {
				return false
			}
		}
		if useRegular {
			return reg.MatchString(fileName)
		}
		return fileName == fe.flagName
	}

	cmdName := fset.Arg(0)
	if len(cmdName) == 0 {
		return errors.New("cmd is empty")
	}

	rd, err := fe.findRootDir(strings.Split(rootDir, ","))
	if err != nil {
		return err
	}

	if Trace.Load() {
		rl, err := filepath.Rel(fe.wd, rd)
		if err == nil {
			rl += string(filepath.Separator)
		}
		log.Println("FindRootDir:", rl, ", Rel.err:", err)
	}

	return fe.run(ctx, rd, match, cmdName, fset.Args()[1:])
}

func (fe *FindExec) findRootDir(names []string) (string, error) {
	names = stringsTrim(names)
	if len(names) == 0 {
		return "./", nil
	}
	wd := fe.wd

	hasFile := func() bool {
		for _, name := range names {
			_, err := os.Stat(filepath.Join(wd, name))
			if err == nil {
				return true
			}
		}
		return false
	}

	for i := 0; i < 128; i++ {
		if hasFile() {
			return wd, nil
		}
		wdn := filepath.Dir(wd)
		if wdn == wd {
			return wd, nil
		}
		wd = wdn
	}
	return "./", nil
}

func (fe *FindExec) run(ctx context.Context, rootDir string, match func(fileName string) bool, cmdName string, args []string) error {
	var index int
	var fail int

	dirs := map[string]bool{}

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
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

		if Trace.Load() {
			s0 := xcolor.GreenString("%3d.", index)
			rl, _ := filepath.Rel(fe.wd, dir)
			s1 := fmt.Sprintf("Dir: %s, MatchFile: %s", rl, fileName)
			s2 := xcolor.YellowString("Exec: %s", rr.String())
			log.Println(s0, s1, s2)
		}

		if e1 := rr.Run(ctx); e1 != nil {
			fail++
			if Trace.Load() {
				log.Println(xcolor.RedString(e1.Error()))
			}
		}
		return fs.SkipDir
	})
	if err != nil {
		return err
	}
	if fail > 0 {
		return fmt.Errorf("total %d/%d tasks failed", fail, index)
	}

	if index == 0 && Trace.Load() {
		log.Printf("file %q not found, skipped", fe.flagName)
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
