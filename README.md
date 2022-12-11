# Bin-Auto-Switcher

1. Auto switch binary in different directories by rules.  
2. Add pre- and post- Hooks.

## 1. Install

```bash
go install github.com/fsgo/bin-auto-switcher@latest
```

## 2. Config
Local Config File（1st priority）:
```
{CurrentDir}/.bin-auto-switcher/{cmd}.toml
```

Global Config File（2nd priority）:
```
~/.config/bin-auto-switcher/{cmd}.toml
```

## 3. Example
### 3.1 Auto Switch Go versions
you should already [install multiple Go versions](https://github.com/fsgo/smart-go-dl)

#### 1.create Symlink for `go`:
```bash
bin-auto-switcher ln go.latest go
```

or

```bash
# cd ~/go/bin/
# ln -s bin-auto-switcher go
```
#### 2. edit config file `go.toml` (`~/.config/bin-auto-switcher/go.toml`):
```toml
# config for 'go' command

# 1th is the default rule
[[Rules]]
Cmd = "go.latest"          # command, Required
# Env =["k1=v1","k2=v2"]   # extra env variable, Optional
# Args = ["-k","-v"]       # extra cmd args, Optional

# [[Rules.Pre]]                # Optional, pre command
# Match = ""                   # Optional, regexp to match Args,eg "^add\\s" will match "git add ."
# Cmd   = ""                   # Required
# Args  = [""]                 # Optional
# AllowFail = true/false       # Optional
# Timeout = "2m"               # Optional, exec timeout, default 1 min

# [[Rules.Post]]               # Optional, post command
# Cmd  = ""
# Args = [""]

# rule for some dir
[[Rules]]
# when in these dirs, this rule can be match
Dir = ["~/workspace/fsgo/myserver"]
Cmd = "go1.19"

# rule for other dir
# [[Rules]]
# Dir = ["~/workspace/job"]
# Cmd = "go1.18"
```

#### 3. Check It:

----------
① At  `~/workspace/fsgo/myserver`: 
```bash
# go version
go version go1.17 darwin/amd64
```
-----------
②  At other dirs (e.g.: `~/workspace/`):
```bash
# go version
go version go1.19.3 darwin/amd64
```

### 3.2 git hooks
1. create Symlink for `git`:
```bash
bin-auto-switcher ln /usr/local/bin/git git
```

2. edit config: `~/.config/bin-auto-switcher/git.toml`
```toml
[[Rules]]
Cmd = "/usr/local/bin/git"              

# with env "BAS_NoHook=true" to disable Pre and Post Hooks
[[Rules.Pre]]               
Match = "^add\\s"       # when exec "git add" subCommand
Cond  = ["go_module"]   # condition: in go module dir
Cmd   = "gorgeous"      # https://github.com/fsgo/go_fmt

[[Rules.Pre]]               
Match = "^add\\s"
# use inner command 'find-exec' to find filename 'go.mod' 
# and then exec "staticcheck ./..." in the dir
Cmd   = "inner:find-exec"
Args  = ["-name","go.mod","staticcheck","./..."]
AllowFail = true        # allow cmd fail
```

### 3.3 inner:find-exec
find a filename and exec another Command in this dir
```bash
Usage of find-exec:
  -e	name as regular expression( default false)
  -name string
    	find file name (default "go.mod")
  -dir_not string
    	not in these dir names, multiple are connected with ','"
```

Examples:
```
# exec: gorgeous (https://github.com/fsgo/go_fmt)
inner:find-exec -name go.mod gorgeous

# exec: staticcheck ./...
inner:find-exec -name go.mod staticcheck ./...
```

### 3.4 Condition
When `Cond` success, exec `Cmd`.
```toml
[[Rules.Pre]]               
Match = "^add\\s"       # when exec "git add" subCommand
Cond  = ["go_module"]   # condition: in go module dir
Cmd   = "gorgeous"      # https://github.com/fsgo/go_fmt
```
| Condition                   | Note                                                        |
|-----------------------------|-------------------------------------------------------------|
| `go_module`                 | in Go module dir, with a file named "go.mod"                |
| `has_file` xyz              | in some dirs with a file named "xyz"                        | 
| `not_has_file` xyz          | in some dirs without a file named "xyz"                     | 
| `not_has_file` xyz          | in some dirs without a file named "xyz"                     | 
| `exec` xyz.sh               | in some dirs without a file named "xyz.sh" and exec success | 
| `in_dir` xyz/abc [dir2]     | in "xyz/abc" dir or in "dir2"                               | 
| `not_in_dir` xyz/abc [dir2] | not in "xyz/abc" and "dir2" dir                             | 