// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/10/30

package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Command struct {
	Cmd     string
	Args    []string
	Match   string // 正则表达式
	Timeout time.Duration
	Trace   bool
}

func (c *Command) IsMatch(str string) (bool, error) {
	if len(c.Match) == 0 {
		return true, nil
	}
	return regexp.MatchString(c.Match, str)
}

func (c *Command) getTimeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return time.Minute
}

func (c *Command) Exec(ctx context.Context, env []string) {
	ss := strings.Fields(c.Cmd)
	args := append(ss[1:], c.Args...)
	cmd := exec.CommandContext(ctx, ss[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	if c.Trace {
		var timeout string
		if dl, ok := ctx.Deadline(); ok {
			timeout = fmt.Sprintf("%.1fs", time.Until(dl).Seconds())
		}
		log.Println("Exec:", cmd.String(), ", Timeout:", timeout)
	}
	if err := cmd.Run(); err != nil {
		msg := fmt.Sprintf("Exec %s failed: %s", c.Cmd, err.Error())
		fmt.Fprint(os.Stderr, ConsoleRed(msg))
		os.Exit(1)
	}
}
