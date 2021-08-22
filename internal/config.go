// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Rules []*Rule
}

func (c *Config) Rule() (*Rule, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if len(c.Rules) == 0 {
		return nil, fmt.Errorf("bin-auto-switcher has no rules")
	}
	var ms []struct {
		Rule  *Rule
		Score int
	}
	for _, rule := range c.Rules {
		score := rule.Match(wd)
		if score > 0 {
			item := struct {
				Rule  *Rule
				Score int
			}{Rule: rule, Score: score}
			ms = append(ms, item)
		}
	}
	if len(ms) < 2 {
		return c.Rules[0], nil
	}

	sort.SliceStable(ms, func(i, j int) bool {
		return ms[i].Score > ms[j].Score
	})
	return ms[0].Rule, nil
}

type Rule struct {
	Dir []string
	Cmd string
	Env []string
}

func (r *Rule) Match(wd string) int {
	if len(r.Dir) == 0 {
		return 1
	}

	for _, dir := range r.Dir {
		if len(dir) == 0 {
			return 1
		}

		if dir == wd {
			return 10000
		}

		if strings.HasPrefix(dir, wd) {
			return (len(dir) - len(wd)) * 5
		}
	}
	return 0
}

var signalsToIgnore = []os.Signal{os.Interrupt, syscall.SIGQUIT}

const caseInsensitiveEnv = runtime.GOOS == "windows"

func (r *Rule) Run(args []string) {
	cmd := exec.Command(r.Cmd, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = dedupEnv(caseInsensitiveEnv, append(os.Environ(), r.Env...))

	signal.Notify(make(chan os.Signal), signalsToIgnore...)

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func ConfigPath(name string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, ".config", "bin-auto-switcher", name+".toml")
}

func LoadConfig(name string) (*Config, error) {
	fp := ConfigPath(name)
	var cfg *Config
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
