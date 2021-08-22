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

for command `go`:
```
# cd /home/work/go/bin/ && ls -l
-rwxr-xr-x 1 work work  2550992  8 22 21:56 bin-auto-switcher
lrwxr-xr-x 1 work work       17  8 22 20:48 go -> bin-auto-switcher
```
 `go` is the symlink off `bin-auto-switcher`, you should create it manual.


it's config is `go.toml` (`~/.config/bin-auto-switcher/go.toml`):
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
Dir = ["~/workspace/fsgo/bin-auto-switcher"]
Cmd = "go1.17"
```


then check it in `~/workspace/fsgo/bin-auto-switcher`:
```
# go version
go version go1.17 darwin/amd64
```
actual was executed the `go1.17`  

in other dir (eg: `~/workspace/`):
```
# go version
go version go1.16.7 darwin/amd64
```
actual was executed the `go1.16.7`

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
