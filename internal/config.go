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

func (c *Config) Format() error {
	for _, r := range c.Rules {
		if e := r.Format(); e != nil {
			return e
		}
	}
	return nil
}

func (c *Config) Rule() (*Rule, error) {
	if len(c.Rules) == 0 {
		return nil, fmt.Errorf("bin-auto-switcher has no rules")
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	wd = wd + string(filepath.Separator)

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
	Dir  []string
	Cmd  string
	Args []string
	Env  []string
}

func (r *Rule) Match(wd string) int {
	if len(r.Dir) == 0 {
		return 1
	}

	for _, dir := range r.Dir {
		if len(dir) == 0 {
			return 1
		}

		if strings.HasPrefix(wd, dir) {
			return len(dir) * 5
		}
	}
	return 0
}

func (r *Rule) Format() error {
	if len(r.Dir) == 0 {
		return nil
	}
	for i := 0; i < len(r.Dir); i++ {
		dir := r.Dir[i]
		if len(dir) == 0 {
			continue
		}
		dir = filepath.Clean(dir)
		if strings.HasPrefix(dir, "~") {
			dir = filepath.Join(homeDir, dir[1:])
		}
		dir = dir + string(filepath.Separator)
		r.Dir[i] = dir
	}
	return nil
}

var signalsToIgnore = []os.Signal{os.Interrupt, syscall.SIGQUIT}

const caseInsensitiveEnv = runtime.GOOS == "windows"

func (r *Rule) Run(args []string) {
	ss := strings.Fields(r.Cmd)
	cmdName := ss[0]
	cmdArgs := append(ss[1:], r.Args...)
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(cmdName, cmdArgs...)
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
	return filepath.Join(homeDir, ".config", "bin-auto-switcher", name+".toml")
}

func LoadConfig(name string) (*Config, error) {
	fp := ConfigPath(name)
	if _, err := os.Stat(fp); err != nil && os.IsNotExist(err) {
		tpl := cmdTPl(name + "_xxx")
		return nil, fmt.Errorf("config %q not exists, you can create it like this:\n %s", fp, tpl)
	}
	var cfg *Config
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Format(); err != nil {
		return nil, err
	}
	return cfg, nil
}

var configTpl = `
# The default rules
[[Rules]]
Cmd = "{CMD}"                  # Required
# Args = [""]                  # Optional, extra args for command
# Env = ["k1=v1","k2=v2"]      # Optional, extra env variable for command

# Rules for some dirs
#[[Rules]]
#Dir = ["/home/work/dir_1/"]   # Required
#Cmd = "{CMD}_v1"              # Required
# Args = [""]                  # Optional, extra args for command
# Env = ["k1=v1","k2=v2"]      # Optional, extra env variable for command


# Rules for other dirs
#[[Rules]]
#Dir = ["/home/work/dir_2/"]   # Required
#Cmd = "{CMD}_v2"              # Required

`

func cmdTPl(name string) string {
	return strings.ReplaceAll(configTpl, "{CMD}", name)
}
