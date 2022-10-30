// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"log"
	"path/filepath"
	"strings"
)

const selfBinName = "bin-auto-switcher"

func Execute(args []string) {
	if len(args) == 0 {
		panic("min args is 1, got 0")
	}
	app := getApp(filepath.Base(args[0]))

	if app == selfBinName || strings.HasPrefix(app, selfBinName) {
		executeSelf(args[1:])
		return
	}

	execute(app, args[1:])
}

func getApp(name string) string {
	if !isWindows() {
		return name
	}
	return strings.TrimRight(name, ".ex")
}

func execute(name string, args []string) {
	cfg, err := LoadConfig(name)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if cfg.Trace {
		log.Println("Config:", cfg.filePath)
	}
	rule, err := cfg.Rule()
	if err != nil {
		log.Fatalln(err.Error())
	}
	rule.Run(args)
}
