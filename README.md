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

when `cmd=go`,it's config is `go.toml`:
```toml
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


then in `~/workspace/fsgo/bin-auto-switcher`:
```
# go version
go version go1.17 darwin/amd64
```

in other dir:
```
# go version
go version go1.16.7 darwin/amd64
```

## 3. create Symlink for `mycmd`

step 1: find `bin-auto-switcher`
```bash
# which bin-auto-switcher
/home/work/go/bin-auto-switcher
```

step 2: create Symlink for `mycmd`
```
# ln -s /home/work/go/bin-auto-switcher /home/works/go/mycmd
```

step 3: edit config file
```
# vim ~/.config/bin-auto-switcher/mycmd.toml
```

Make sure `mycmd` can find in the `$PATH`:
```
PATH:=xxx/go/:$PATH
export PATH
```

then you can try it:
```
# mycmd
```
