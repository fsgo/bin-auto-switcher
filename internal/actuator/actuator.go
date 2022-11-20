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

// Trace 是否打印日志
var Trace = atomic.Bool{}

const Prefix = "inner:"

type Actuator interface {
	Name() string
	Run(ctx context.Context) error
	String() string
}

var all = map[string]func([]string) Actuator{}

func register(fn func([]string) Actuator) {
	ins := fn(nil)
	all[ins.Name()] = fn
}

// find 查找一个注册的 Actuator
func find(name string) func([]string) Actuator {
	return all[name]
}

type Config struct {
	ac       Actuator
	Name     string
	Dir      string
	Args     []string
	Env      []string
	exitCode atomic.Int32
}

func (r *Config) String() string {
	return r.getActuator().String()
}

func (r *Config) getActuator() Actuator {
	if r.ac != nil {
		return r.ac
	}

	fn := find(r.Name)
	if fn != nil {
		r.ac = fn(r.Args)
		return r.ac
	}

	r.ac = &Cmd{
		CmdName: r.Name,
		Args:    r.Args,
		Setup: func(cmd *exec.Cmd) {
			cmd.Dir = r.Dir
			if len(r.Env) > 0 {
				cmd.Env = r.Env
			}
		},
	}
	return r.ac
}

func (r *Config) Run(ctx context.Context) (err error) {
	ac := r.getActuator()
	if c, ok := ac.(*Cmd); ok {
		err = ac.Run(ctx)
		r.exitCode.Store(int32(c.ExitCode()))
		return err
	}

	defer func() {
		if err != nil {
			r.exitCode.Store(1)
		}
	}()

	if len(r.Dir) != 0 {
		pwd, e1 := os.Getwd()
		if e1 != nil {
			return e1
		}
		if pwd != r.Dir {
			if e2 := os.Chdir(r.Dir); e2 != nil {
				return e2
			}
		}
		defer func() {
			_ = os.Chdir(pwd)
		}()
	}

	return ac.Run(ctx)
}

func (r *Config) ExitCode() int {
	return int(r.exitCode.Load())
}
