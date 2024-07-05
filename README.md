# Bin-Auto-Switcher

1. Auto switch binary in different directories by rules. 
2. Execute pre-hooks and post-hooks.

## 1. Install

```bash
go install github.com/fsgo/bin-auto-switcher/bas@latest
```

## 2. Config
`{CurrentDir}/.bas/{cmd}.toml` or `~/.config/bas/{cmd}.toml`

## 3. Example
### 3.1 Auto Switch Go versions
you should already [install multiple Go versions](https://github.com/fsgo/smart-go-dl)

#### 1.create alias or symlink for `go`:
```bash
bas ln go.latest go
```

or

```bash
cd ~/go/bin/
ln -s bas go
```


#### 2. edit config file `go.toml` (`~/.config/bas/go.toml`):
```toml
# config for 'go' command

# 1th is the default rule
[[Rules]]
Cmd = "go.latest"          # command, Required
# Env =["k1=v1","k2=v2"]   # extra env variable, Optional
# Args = ["-k","-v"]       # extra cmd args, Optional
[Rules.Spec]
# use go version defined in go.mod if ‘go1.xx’( e.g. go1.21) exists
# go1.xx should be found in $PATH
# if value is "","no", skip it
GoVersionFile = "go.mod"
# set env GOWORK=off if module not defined in go.work
# if value is "","no", skip it
GoWork = "auto"

# [[Rules.Pre]]            # Optional, pre command
# Match = ""               # Optional, regexp to match Args. "^add\\s" will match "git add ."
# Cmd   = ""               # Required
# Args  = [""]             # Optional
# Env =["k3=v3","k2=v2"]   # extra env variable, Optional
# AllowFail = true/false   # Optional
# Timeout = "2m"           # Optional, exec timeout, default 1 min

# [[Rules.Post]]           # Optional, post command
# Cmd  = ""
# Args = [""]

# rule for some dir
[[Rules]]
# when in these dirs, this rule can be match
Dir = ["~/workspace/fsgo/myserver"]
Cmd = "go1.19"
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
bas ln /usr/local/bin/git git
```

2. edit config: `~/.config/bas/git.toml`
```toml
# Trace = true # enable trace log global

[[Rules]]
Cmd = "/usr/local/bin/git"    # the raw Cmd Path, or empty it will auto detect

# with env "BAS_NoHook=true" or "bas=off" to disable Pre and Post Hooks

[[Rules.Pre]]
Match = "^add\\s" # when exec "git add" subCommand
# Trace = true
# find pre-ci.sh and execute it if exists
Cmd   = "inner:find-exec"
Args  = ["-name","pre-ci.sh","bash","pre-ci.sh"]

[[Rules.Pre]]               
Match = "^add\\s"       
Cond  = ["go_module"]   # condition: in go module dir
Cmd   = "gorgeous"      # https://github.com/fsgo/go_fmt

[[Rules.Pre]]               
Match = "^add\\s"
# find file "go.mod" and exec "staticcheck ./..." in the dir
Cmd   = "inner:find-exec"
Args  = ["-name","go.mod","staticcheck","./..."]
AllowFail = true        # allow cmd fail
```

### 3.3 inner:find-exec
find a filename and exec another Command in this dir
```bash
Usage of find-exec:
  -root string
       search up root dir(default "go.mod,.git")
  -name string
    	find file name (default "go.mod")
  -e	name as regular expression( default false)
  -dir_not string
    	not in these dir names, multiple are connected with ","
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
| `in_dir` xyz/abc[;dir2]     | in "xyz/abc" dir or in "dir2"                               | 
| `not_in_dir` xyz/abc[;dir2] | not in "xyz/abc" and "dir2" dir                             | 


### 3.5 Eval
eval command without links.
```bash
bas git st
```
it will eval `git st` command and also execute pre-hooks and post-hooks which defined
in config file （e.g. `~/.config/bas/git.toml` or `.bas/git.toml`）.

### 3.6 Disable Hooks
with env "BAS_NoHook=true" or "bas=off" to disable Pre-Hooks and Post-Hooks