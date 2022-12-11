// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/fsgo/fsenv"
)

const selfBinName = "bin-auto-switcher"

func Execute(args []string) {
	if len(args) == 0 {
		panic("min args is 1, got 0")
	}
	setup()

	app := getApp(filepath.Base(args[0]))

	if app == selfBinName || strings.HasPrefix(app, selfBinName) {
		executeSelf(args[1:])
		return
	}

	execute(app, args[1:])
}

func setup() {
	fsenv.SetConfRootDir(configDir())
	fsenv.SetRootDir(filepath.Join(configDir(), "app_data"))
}

func getApp(name string) string {
	if !isWindows() {
		return name
	}
	return strings.TrimRight(name, ".ex")
}

func execute(name string, args []string) {
	setLogPrefix("Load")
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
