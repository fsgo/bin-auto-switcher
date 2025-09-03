// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	"github.com/fsgo/fsenv"
)

const (
	selfBinName      = "bin-auto-switcher"
	selfBinNameShort = "bas"
)

func Execute(args []string) {
	if len(args) == 0 {
		panic("min args is 1, got 0")
	}
	setup()

	app := getApp(filepath.Base(args[0]))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if app == selfBinName || app == selfBinNameShort {
		executeSelf(ctx, args[1:])
		return
	}

	execute(ctx, app, args[1:])
}

func setup() {
	fsenv.SetConfDir(configDir())
	fsenv.SetRootDir(filepath.Join(configDir(), "app_data"))
}

func getApp(name string) string {
	if !isWindows() {
		return name
	}
	return strings.TrimRight(name, ".ex")
}

func execute(ctx context.Context, name string, args []string) {
	setLogPrefix("Load")
	cfg, err := LoadConfig(name)
	if err != nil {
		log.Fatalln("LoadConfig failed:", err)
	}
	rule, err := cfg.Rule()
	if err != nil {
		log.Fatalln("Pick Rule failed:", err)
	}
	if err = rule.BeforeExec(ctx, name); err != nil {
		log.Fatalln("BeforeExec failed:", err)
	}
	rule.Run(ctx, args)
}
