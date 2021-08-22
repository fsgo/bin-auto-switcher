// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"log"
)

func Execute(args []string) {
	if len(args) == 0 {
		panic("min args is 1, got 0")
	}
	execute(args[0], args[1:])
}

func execute(name string, args []string) {
	cfg, err := LoadConfig(name)
	if err != nil {
		log.Fatalln(err.Error())
	}
	rule, err := cfg.Rule()
	if err != nil {
		log.Fatalln(err.Error())
	}
	rule.Run(args)
}
