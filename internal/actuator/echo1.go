// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/11/19

package actuator

import (
	"context"
	"log"
	"strings"

	"github.com/fatih/color"
)

var _ Actuator = (*Echo1)(nil)

// Echo1 测试用的命令
type Echo1 struct {
	Args []string
}

func (e *Echo1) Name() string {
	return Prefix + "echo1"
}

func (e *Echo1) Run(ctx context.Context) error {
	log.Println(color.YellowString("Call Echo1: %s", e.String()))
	return nil
}

func (e *Echo1) String() string {
	return e.Name() + " " + strings.Join(e.Args, " ")
}

func init() {
	register(func(args []string) Actuator {
		return &Echo1{
			Args: args,
		}
	})
}
