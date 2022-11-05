// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/11/3

package internal

import (
	"testing"
)

func TestCondition_Allow(t *testing.T) {
	tests := []struct {
		name string
		c    Condition
		want bool
	}{
		{
			name: "go_module",
			c:    "go_module",
			want: true,
		},
		{
			name: "has file go.module",
			c:    "has_file go.mod",
			want: true,
		},
		{
			name: "has file abc.so",
			c:    "has_file abc.so",
			want: false,
		},
		{
			name: "has_file empty",
			c:    "has_file",
			want: false,
		},
		{
			name: "other",
			c:    "other",
			want: false,
		},
		{
			name: "exec empty",
			c:    "exec ",
			want: false,
		},
		{
			name: "exec echo",
			c:    "exec echo",
			want: true,
		},
		{
			name: "exec echo a",
			c:    "exec echo a",
			want: true,
		},
		{
			name: "exec not found cmd",
			c:    "exec not_found_cmd",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Allow(); got != tt.want {
				t.Errorf("Allow() = %v, want %v", got, tt.want)
			}
		})
	}
}
