// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/10/30

package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/fsgo/bin-auto-switcher/internal/actuator"
)

type Command struct {
	// Match 用于匹配执行命令的正则表达式，可选
	// 如命令为 "git add ." 则，"add ." 会交给此正则来匹配
	// 若不匹配，当前这组命令将不会执行
	Match string

	// Cond 额外的执行条件，可选
	// 如：
	// go_module: 当前命令在 go module 里，即当前目录或者上级目录有 go.mod 文件
	// exec xx.sh : 执行 xx.sh 并执行成功
	// has_file app.toml: 当前目录或者上级目录有 app.toml 文件
	Cond []Condition

	// Cmd 命令，必填
	Cmd string

	Args []string

	// Timeout 超时时间，默认 1 分钟
	Timeout time.Duration

	Trace bool

	// AllowFail 是否允许执行失败，默认否
	// 默认情况下，当此命令执行失败后，后续命令也不会执行，程序将退出
	AllowFail bool
}

func (c *Command) IsMatch(str string) (bool, error) {
	if len(c.Match) == 0 {
		return true, nil
	}
	return regexp.MatchString(c.Match, str)
}

func (c *Command) CanRun() bool {
	if len(c.Cond) == 0 {
		return true
	}
	for _, item := range c.Cond {
		if !item.Allow() {
			return false
		}
	}
	return true
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

	co := &actuator.Config{
		Name: ss[0],
		Args: args,
		Env:  env,
	}

	if c.Trace {
		var timeout string
		if dl, ok := ctx.Deadline(); ok {
			timeout = fmt.Sprintf("%.1fs", time.Until(dl).Seconds())
		}
		log.Println("Exec:", color.CyanString(co.String()), ", Timeout:", timeout)
	}

	err := co.Run(ctx)
	if err == nil {
		return
	}
	msg := fmt.Sprintf("Exec %s failed: %s", c.Cmd, err.Error())
	if c.AllowFail {
		msg += ", skipped"
	}
	fmt.Fprintln(os.Stderr, ConsoleRed(msg))
	if !c.AllowFail {
		exitCode := co.ExitCode()
		os.Exit(exitCode)
	}
}
