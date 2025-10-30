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

func (gm *GitAddModify) Run(ctx context.Context) error {
	gitBin := GetRawBinName("git")
	cmd := exec.CommandContext(ctx, gitBin, "ls-files", "--others", "-m")
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
	for _, filename := range lines {
		filename = strings.TrimSpace(filename)
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
