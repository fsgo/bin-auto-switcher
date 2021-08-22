# Bin-Auto-Switcher

Auto switch binary in different directories by rules.  
You can use it to automatically switch  different versions.


## 1. Config
config file:
```
~/.config/bin-auto-switcher/{cmd}.toml
```

when `cmd=go`,it's config is `go.toml`:
```
# default rule
[[Rules]]
Cmd = "go1.16.7"
# other k-v env, Optional
Env =["k1=v1","k2=v2"]

# rule for dir
[[Rules]]
Dir = ["~/workspace/fsgo/bin-auto-switcher"]
Cmd = "go1.17"
# cmd args, Optional
#Args = []
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

## 2. Symlink for `mycmd`
```
## step 1: find bin-auto-switcher
# which bin-auto-switcher
xxx/go/bin-auto-switcher

## step 2: create Symlink for mycmd
# ln -s xxx/go/bin-auto-switcher xxx/go/mycmd

## step 3: edit config file
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
