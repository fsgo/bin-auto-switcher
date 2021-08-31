# Bin-Auto-Switcher

Auto switch binary in different directories by rules.  
You can use it to automatically switch  different versions.


## 1. install

```bash
go install github.com/fsgo/bin-auto-switcher@latest
```

## 2. Config
config file:
```
~/.config/bin-auto-switcher/{cmd}.toml
```


## 3. create Symlink for `mycmd`

step 1: find `bin-auto-switcher`
```bash
# which bin-auto-switcher
/home/work/go/bin/bin-auto-switcher
```

step 2: create Symlink for `mycmd`
```
# ln -s /home/work/go/bin/bin-auto-switcher /home/works/go/bin/mycmd
```

step 3: edit config file
```
# vim ~/.config/bin-auto-switcher/mycmd.toml
```

Make sure `mycmd` can find in the `$PATH`:
```
PATH:=/home/work/go/bin/:$PATH
export PATH
```

then you can try it:
```
# mycmd
```


## 4. Example
### 4.1 multiple Go versions
you should already [install multiple Go versions](https://golang.google.cn/doc/manage-install)

#### 1.create Symlink for `go`:
```
# cd /home/work/go/bin/
# ln -s bin-auto-switcher go
```

#### 2. edit config file `go.toml` (`~/.config/bin-auto-switcher/go.toml`):
```toml
# config for 'go' command

# default rule
[[Rules]]
# other k-v env, Optional
Env =["k1=v1","k2=v2"]
# command, Required
Cmd = "go1.16.7"
# cmd args, Optional
# Args = []

# rule for dir
[[Rules]]
Dir = ["~/workspace/fsgo/myserver"]
Cmd = "go1.17"
```

#### 3. Check It:
at  `~/workspace/fsgo/myserver`:
```
# go version
go version go1.17 darwin/amd64
```
actual was executed the `go1.17`

at other dirs (eg: `~/workspace/`):
```
# go version
go version go1.16.7 darwin/amd64
```
actual was executed the `go1.16.7`