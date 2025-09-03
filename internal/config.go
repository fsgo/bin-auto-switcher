// Copyright(C) 2021 github.com/fsgo  All Rights Reserved.
// Author: fsgo
// Date: 2021/8/22

package internal

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/fsgo/fsconf"
	"github.com/fsgo/fsconf/confext"
)

func init() {
	confext.Init()
}

var enableTrace = os.Getenv(envKey("Trace")) == "true"

type Config struct {
	// binName 当前命令，如 go、git 等
	binName string

	// fileName 当前配置对应的配置文件地址
	fileName string

	Rules []*Rule

	// Trace 是否调试模式，默认 true
	// 目前只控制是否打印过程日志
	Trace bool

	// Spec 不同命令，特殊的配置
	// 每种命令的配置都不同,详见 spec.go
	Spec map[string]any
}

// Format 格式化配置内容
// name: 执行的命令，如 go、git 等
func (c *Config) Format() error {
	var rawBinName string
	for idx, r := range c.Rules {
		if len(r.Cmd) == 0 {
			if len(rawBinName) == 0 {
				rawBinName = getRawBinName(c.binName)
			}
			if len(rawBinName) == 0 {
				return fmt.Errorf("%s rule[%d].Cmd is empty, cannot found %q", c.fileName, idx, c.binName)
			}
			r.Cmd = rawBinName
		}
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
	Dir  []string `json:",omitempty"`
	Args []string `json:",omitempty"`
	Env  []string `json:",omitempty"`

	Pre  []*Command `json:",omitempty"`
	Post []*Command `json:",omitempty"`

	Trace bool `json:",omitempty"`

	Spec map[string]any `json:",omitempty"`
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

// Merge 将 b merge 到 r
func (r *Rule) Merge(b *Rule) {
	// 目前只需要 merge Cmd 这一个值
	if b.Cmd != "" {
		r.Cmd = b.Cmd
	}
}

// Format 格式化当前配置
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

func (r *Rule) Run(ctx context.Context, args []string) {
	ss := strings.Fields(r.Cmd)
	cmdName := ss[0]
	cmdArgs := append(ss[1:], r.Args...)
	cmdArgs = append(cmdArgs, args...)
	cmdArgsStr := strings.Join(cmdArgs, " ")

	env := dedupEnv(caseInsensitiveEnv, append(os.Environ(), r.Env...))
	env = append(env, fmt.Sprintf(envKey("CMD")+"=%s", cmdName))
	env = append(env, fmt.Sprintf(envKey("ARGS")+"=%q", cmdArgsStr))

	// signal.Notify(make(chan os.Signal), signalsToIgnore...)

	setLogPrefix("Before")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	noHooks := disableHooks()

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

func (r *Rule) BeforeExec(ctx context.Context, name string) error {
	if r.Trace {
		log.Printf("Cmd = %q, Spec = %v\n", r.Cmd, r.Spec)
	}
	if disableHooks() {
		return nil
	}
	if len(r.Spec) > 0 {
		if err := parserSpecial(name, r); err != nil {
			return err
		}
	}
	return ctx.Err()
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

		if r.Trace {
			pc.Trace = r.Trace
		}

		func() {
			timeout := pc.getTimeout()
			ctx1, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			pc.Exec(ctx1, env)
		}()
	}
}

func configDir() string {
	return filepath.Join(homeDir, ".config", "bas")
}

func globalConfigPath(name string) string {
	return filepath.Join(configDir(), name+".toml")
}

func localConfigPath(name string) (string, error) {
	fp := filepath.Join(".bas", name+".toml")
	np, err := findFileUpper(fp, 128)
	if err != nil && errors.Is(err, errFileNotFound) {
		return "", nil
	}
	return np, err
}

func fileExists(p string) (bool, error) {
	st, err := os.Stat(p)
	if err == nil {
		if st.IsDir() {
			return false, fmt.Errorf("%q is dir, expect is file", p)
		}
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func LoadConfig(name string) (*Config, error) {
	fileName, err := localConfigPath(name)
	if err != nil {
		return nil, err
	}
	if enableTrace {
		log.Printf("Local ConfigPath = %q %v\n", fileName, err)
	}

	if len(fileName) == 0 {
		fp := globalConfigPath(name)
		ok, err := fileExists(fp)
		if enableTrace {
			log.Printf("Global ConfigPath = %q, exists=%v err=%v\n", fp, ok, err)
		}
		if ok {
			fileName = fp
		}
		if err != nil {
			return nil, err
		}
	}

	cfg := &Config{
		binName: name,
		Trace:   enableTrace,
	}

	cfg.fileName = fileName
	if len(fileName) != 0 {
		if err = fsconf.Parse(fileName, &cfg); err != nil {
			return nil, err
		}
	}

	if cfg.Trace {
		log.Printf("Using ConfigPath = %q\n", cfg.fileName)
	}

	if len(cfg.Rules) == 0 {
		defaultRule := &Rule{}
		cfg.Rules = []*Rule{defaultRule}
	}

	if err = cfg.Format(); err != nil {
		return nil, err
	}

	// 最终使用环境变量里的值
	if enableTrace {
		cfg.Trace = true
	}

	return cfg, nil
}

var configTpl = `
# Optional, enable print trace log
# or with env "BAS_Trace=true" to enable it
# Trace = false

[[Rules]]
Cmd = "{CMD}"                  # Optional
# Args = [""]                  # Optional, extra args for command
# Env = ["k1=v1","k2=v2"]      # Optional, extra env variable for command
# Trace = false                # Optional, print trace log

# -----------------------------------------------------------------------------
# with env "BAS_NoHook=true" to disable Pre and Post Hooks
#
# [[Rules.Pre]]                # Optional, prepare hook command
# Match = ""                   # Optional, regexp for args, eg "^add\\s" for "git add ."
# Trace = false                # Optional, print trace log

# Cond Optional, extra conditions
# e.g. "go_module","has_file app.toml", "exec hello.sh"
# Cond  = [""]                

# Cmd   = ""                   # Required
# Args  = [""]                 # Optional
# AllowFail = true/false       # Optional, break when exec failed
# Timeout = "2m"               # Optional, exec timeout, default 1 min
#
# -----------------------------------------------------------------------------
# [[Rules.Post]]               # Optional, Post hook command，same as Rules.Pre
# Cmd  = ""
# Args = [""]
# -----------------------------------------------------------------------------

# =============================================================================
# Rule for some dirs
# =============================================================================
# [[Rules]]
# Dir = ["/home/work/dir_1/"]   # Required
# Cmd = "{CMD}"                 # Optional
# Args = [""]                   # Optional, extra args for command
# Env = ["k1=v1","k2=v2"]       # Optional, extra env variable for command
# ============================================================================
`

func cmdTPl(name string) string {
	str := strings.ReplaceAll(configTpl, "{CMD}", name)
	str = strings.TrimSpace(str) + "\n"
	return str
}

func setLogPrefix(msg string) {
	log.SetPrefix(fmt.Sprintf("%6s | ", msg))
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
}
