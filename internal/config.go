// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const envKeyPrefix = "BAS_"

type Config struct {
	filePath string
	Rules    []*Rule

	// Trace 是否调试模式，默认 true
	// 目前只控制是否打印过程日志
	Trace bool
}

func (c *Config) Format() error {
	for _, r := range c.Rules {
		if e := r.Format(); e != nil {
			return e
		}
	}
	return nil
}

type tmpRule struct {
	Rule  *Rule
	Score int
	Index int
}

func (c *Config) Rule() (*Rule, error) {
	if len(c.Rules) == 0 {
		return nil, errors.New("bin-auto-switcher has no rules")
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	wd = wd + string(filepath.Separator)

	var ms []tmpRule
	for idx, rule := range c.Rules {
		score := rule.Match(wd)
		if score > 0 {
			item := tmpRule{
				Rule:  rule,
				Score: score,
				Index: idx,
			}
			if c.Trace {
				item.Rule.Trace = true
			}
			ms = append(ms, item)
		}
	}

	var using int
	if c.Trace {
		defer func() {
			log.Printf("Total %d rules, using Rule %d\n", len(ms), using)
		}()
	}
	if len(ms) < 2 {
		return c.Rules[0], nil
	}

	sort.SliceStable(ms, func(i, j int) bool {
		return ms[i].Score > ms[j].Score
	})
	using = ms[0].Index
	return ms[0].Rule, nil
}

type Rule struct {
	Cmd  string
	Dir  []string
	Args []string
	Env  []string

	Pre   []*Command
	Post  []*Command
	Trace bool
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

const caseInsensitiveEnv = runtime.GOOS == "windows"

// var signalsToIgnore = []os.Signal{os.Interrupt, syscall.SIGQUIT}

func (r *Rule) Run(args []string) {
	ss := strings.Fields(r.Cmd)
	cmdName := ss[0]
	cmdArgs := append(ss[1:], r.Args...)
	cmdArgs = append(cmdArgs, args...)
	cmdArgsStr := strings.Join(cmdArgs, " ")

	env := dedupEnv(caseInsensitiveEnv, append(os.Environ(), r.Env...))
	env = append(env, fmt.Sprintf(envKeyPrefix+"CMD=%s", cmdName))
	env = append(env, fmt.Sprintf(envKeyPrefix+"ARGS=%q", cmdArgsStr))

	// signal.Notify(make(chan os.Signal), signalsToIgnore...)

	setLogPrefix("Before")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	noHooks := len(os.Getenv(envKeyPrefix+"NoHook")) != 0

	if !noHooks {
		r.execCmds(ctx, r.Pre, cmdArgsStr, env)
	}

	setLogPrefix("Main")

	mc := &Command{
		Cmd:   cmdName,
		Args:  cmdArgs,
		Trace: r.Trace,
	}
	mc.Exec(ctx, env)

	setLogPrefix("After")
	if !noHooks {
		r.execCmds(ctx, r.Post, cmdArgsStr, env)
	}
	os.Exit(0)
}

func (r *Rule) execCmds(ctx context.Context, cmds []*Command, argsStr string, env []string) {
	if len(cmds) == 0 {
		return
	}

	for _, pc := range cmds {
		if len(pc.Cmd) == 0 {
			continue
		}
		m, err := pc.IsMatch(argsStr)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			os.Exit(1)
		}
		if !m {
			continue
		}

		if !pc.CanRun() {
			continue
		}

		if err = ctx.Err(); err != nil {
			log.Println("context canceled:", err.Error())
			break
		}

		pc.Trace = r.Trace

		func() {
			timeout := pc.getTimeout()
			ctx1, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			pc.Exec(ctx1, env)
		}()
	}
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
	cfg := &Config{
		Trace: true,
	}
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Format(); err != nil {
		return nil, err
	}
	cfg.filePath = fp

	if len(os.Getenv(envKeyPrefix+"_Trace")) != 0 {
		cfg.Trace = true
	}

	return cfg, nil
}

var configTpl = `
[[Rules]]
Cmd = "{CMD}"                  # Required
# Args = [""]                  # Optional, extra args for command
# Env = ["k1=v1","k2=v2"]      # Optional, extra env variable for command

# [[Rules.Pre]]                # Optional, pre command
# Match = ""                   # Optional, regexp to match Args,eg "^add\\s" will match "git add ."

# Cond Optional, extra conditions
# eg. "go_module","has_file app.toml", "exec hello.sh"
# Cond  = [""]                

# Cmd   = ""                   # Required
# Args  = [""]                 # Optional
# AllowFail = true/false       # Optional
# Timeout = "2m"               # Optional, exec timeout, default 1 min

# [[Rules.Post]]               # Optional, post command
# Cmd  = ""
# Args = [""]

# Rules for some dirs
# [[Rules]]
# Dir = ["/home/work/dir_1/"]   # Required
# Cmd = "{CMD}_v1"              # Required
# Args = [""]                   # Optional, extra args for command
# Env = ["k1=v1","k2=v2"]       # Optional, extra env variable for command


# Rules for other dirs
#[[Rules]]
#Dir = ["/home/work/dir_2/"]   # Required
#Cmd = "{CMD}_v2"              # Required

`

func cmdTPl(name string) string {
	return strings.ReplaceAll(configTpl, "{CMD}", name)
}

func setLogPrefix(msg string) {
	log.SetPrefix(fmt.Sprintf("%6s | ", msg))
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
}
