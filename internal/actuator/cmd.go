// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/11/19

package actuator

import (
	"context"
	"os"
	"os/exec"
	"sync/atomic"
)

var _ Actuator = (*Cmd)(nil)

type Cmd struct {
	Setup    func(cmd *exec.Cmd)
	CmdName  string
	Args     []string
	exitCode atomic.Int32
}

func (c *Cmd) Name() string {
	return c.CmdName
}

func (c *Cmd) Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.CmdName, c.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if c.Setup != nil {
		c.Setup(cmd)
	}
	err := cmd.Run()
	if cmd.ProcessState != nil {
		c.exitCode.Store(int32(cmd.ProcessState.ExitCode()))
	}
	return err
}

func (c *Cmd) String() string {
	cmd := exec.Command(c.CmdName, c.Args...)
	return cmd.String()
}

func (c *Cmd) ExitCode() int {
	return int(c.exitCode.Load())
}
