// Copyright 2020 Longxiao Zhang <zhanglongx@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a GPLv3-style
// license that can be found in the LICENSE file.

package web

import (
	"reflect"
	"testing"
)

func Test_selectStr(t *testing.T) {
	type args struct {
		list []string
		s    string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{
				list: []string{""},
				s:    " ",
			},
			want: []string{""},
		},
		{
			args: args{
				list: []string{"aaa", "bbb", "ccc"},
				s:    "ddd",
			},
			want: []string{"aaa", "bbb", "ccc"},
		},
		{
			args: args{
				list: []string{"aaa", "bbb", "ccc"},
				s:    "aaa",
			},
			want: []string{"aaa", "bbb", "ccc"},
		},
		{
			args: args{
				list: []string{"aaa", "bbb", "ccc"},
				s:    "bbb",
			},
			want: []string{"bbb", "aaa", "ccc"},
		},
		{
			args: args{
				list: nil,
				s:    "",
			},
			want: nil,
		},
		{
			args: args{
				list: []string{"aaa"},
				s:    "",
			},
			want: []string{"aaa"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectStr(tt.args.list, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
