// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/01/14

package actuator

import "strings"

func stringsTrim(ss []string) []string {
	if len(ss) == 0 {
		return nil
	}
	ns := make([]string, 0, len(ss))
	ms := make(map[string]struct{}, len(ss))
	for _, name := range ss {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, has := ms[name]; has {
			continue
		}
		ms[name] = struct{}{}
		ns = append(ns, name)
	}
	return ns
}

var GetRawBinName func(binName string) string
