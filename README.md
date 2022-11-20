# Bin-Auto-Switcher

Auto switch binary in different directories by rules.  


## 1. Install

```bash
go install github.com/fsgo/bin-auto-switcher@latest
```

## 2. Config
config file:
```
~/.config/bin-auto-switcher/{cmd}.toml
```

## 3. Example
### 3.1 Auto Switch Go versions
you should already [install multiple Go versions](https://github.com/fsgo/smart-go-dl)

#### 1.create Symlink for `go`:
```bash
bin-auto-switcher ln go1.16.7 go
```

or

```bash
# cd ~/go/bin/
# ln -s bin-auto-switcher go
```
#### 2. edit config file `go.toml` (`~/.config/bin-auto-switcher/go.toml`):
```toml
# config for 'go' command

# default rule
[[Rules]]
Cmd = "go1.19.3"           # command, Required
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
Dir = ["~/workspace/fsgo/myserver"]
Cmd = "go1.17"

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