package util

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"net"
	"reflect"
	"testing"

	"github.com/urfave/cli"
)

func Test_UnitParseStringSliceToIPs(t *testing.T) {
	tests := []struct {
		name    string
		arg     cli.StringSlice
		want    []net.IP
		wantErr bool
	}{
		{
			name: "nil string slice must return no errors",
			arg:  nil,
			want: nil,
		},
		{
			name: "empty string slice must return no errors",
			arg:  cli.StringSlice{},
			want: nil,
		},
		{
			name: "single element slice with correct IP must succeed",
			arg:  cli.StringSlice{"10.10.10.10"},
			want: []net.IP{net.ParseIP("10.10.10.10")},
		},
		{
			name: "single element slice with correct IP list must succeed",
			arg:  cli.StringSlice{"10.10.10.10,10.10.10.11"},
			want: []net.IP{
				net.ParseIP("10.10.10.10"),
				net.ParseIP("10.10.10.11"),
			},
		},
		{
			name: "multi element slice with correct IP list must succeed",
			arg:  cli.StringSlice{"10.10.10.10,10.10.10.11", "10.10.10.12,10.10.10.13"},
			want: []net.IP{
				net.ParseIP("10.10.10.10"),
				net.ParseIP("10.10.10.11"),
				net.ParseIP("10.10.10.12"),
				net.ParseIP("10.10.10.13"),
			},
		},
		{
			name:    "single element slice with correct IP list with trailing comma must fail",
			arg:     cli.StringSlice{"10.10.10.10,"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "single element slice with incorrect IP (overflow) must fail",
			arg:     cli.StringSlice{"10.10.10.256"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "single element slice with incorrect IP (foreign symbols) must fail",
			arg:     cli.StringSlice{"xxx.yyy.zzz.www"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ParseStringSliceToIPs(tt.arg)
				if (err != nil) != tt.wantErr {
					t.Errorf("ParseStringSliceToIPs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParseStringSliceToIPs() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
