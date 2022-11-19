// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2022/1/8

package internal

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var helpMessage = `
bin-auto-switcher subCommand [options]

SubCommands:
    ln {target} {link_name} :
        Like GNU's 'ln' command, create link {link_name} from {target},
        The config file is '~/.config/bin-auto-switcher/{link_name}.toml'.
        eg: "bin-auto-switcher ln go1.19.3 go"

    list:
        list all links

Hooks:
    with env "BAS_NoHook=true" to disable Pre and Post Hooks

Self-Update :
          go install github.com/fsgo/bin-auto-switcher@latest

Site    : https://github.com/fsgo/bin-auto-switcher
Version : 0.1.4
Date    : 2022-11-19
`

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage of %s:\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(out, strings.TrimSpace(helpMessage)+"\n")
}

func executeSelf(args stringSlice) {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" {
		usage()
		return
	}
	var err error
	cmd := args.get(0)
	switch cmd {
	case "ln", "link":
		err = cmdLink(args.get(1), args.get(2))
	case "list":
		err = cmdList()
	default:
		err = fmt.Errorf("not support %q", cmd)
	}

	if err != nil {
		log.Fatalln(err.Error())
	}
}

func cmdLink(target string, linkName string) error {
	if isWindows() {
		return errors.New("not support yet")
	}

	if len(target) == 0 || len(linkName) == 0 {
		return errors.New("invalid params")
	}

	if linkName[0] == '.' || filepath.Base(linkName) != linkName {
		return fmt.Errorf("invalid linkName %q", linkName)
	}

	p, err := exec.LookPath(os.Args[0])
	if err != nil {
		return err
	}
	lp := filepath.Join(filepath.Dir(p), linkName)
	log.Printf("[link] %s %s\n", os.Args[0], lp)
	if err = os.Symlink(os.Args[0], lp); err != nil {
		return err
	}
	cp := ConfigPath(linkName)

	log.Println("[config]", cp, ", you can edit it.")

	if _, err = os.Stat(cp); err == nil {
		return nil
	}

	if err = os.WriteFile(cp, []byte(cmdTPl(target)), 0644); err != nil {
		return err
	}
	return nil
}

func cmdList() error {
	dir := filepath.Dir(ConfigPath("go"))
	ms, err := filepath.Glob(filepath.Join(dir, "*.toml"))
	if err != nil {
		return err
	}
	format := "%-12s  %-25s  %s\n"
	fmt.Printf(format, "Name", "Bin", "Config")
	fmt.Println(strings.Repeat("-", 80))
	for _, item := range ms {
		name := filepath.Base(item)
		name = name[0 : len(name)-5]
		bp, _ := exec.LookPath(name)
		fmt.Printf(format, name, bp, item)
	}
	return nil
}

type stringSlice []string

func (s stringSlice) get(index int) string {
	if index >= len(s) {
		return ""
	}
	return s[index]
}
