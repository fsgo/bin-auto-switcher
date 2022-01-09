# Bin-Auto-Switcher

Auto switch binary in different directories by rules.  


## 1. Install

```bash
go install github.com/fsgo/bin-auto-switcher@main
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
# cd /home/work/go/bin/
# ln -s bin-auto-switcher go
```
#### 2. edit config file `go.toml` (`~/.config/bin-auto-switcher/go.toml`):
```toml
# config for 'go' command

# default rule
[[Rules]]
Cmd = "go1.16.7"           # command, Required
# Env =["k1=v1","k2=v2"]   # extra env variable, Optional
# Args = ["-k","-v"]       # extra cmd args, Optional

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
actual was executed the `go1.17`

-----------
②  At other dirs (eg: `~/workspace/`):
```bash
# go version
go version go1.16.7 darwin/amd64
```
actual was executed the `go1.16.7`