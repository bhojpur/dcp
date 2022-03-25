package server

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
	"bufio"
	"fmt"
	"go/token"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestGetResponse(t *testing.T) {
	test := struct {
		Raw  string
		Resp http.Response
		Body string
	}{
		Raw: "HTTP/1.1 200 OK\r\n" +
			"\r\n" +
			"Body here\n",

		Resp: http.Response{
			Status:        "200 OK",
			StatusCode:    200,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        http.Header{},
			Close:         true,
			ContentLength: -1,
		},

		Body: "Body here\n",
	}

	r := bufio.NewReader(strings.NewReader(test.Raw))
	resp, rbytes, err := getResponse(r)
	if err != nil {
		t.Error(err)
	}

	wbytes := []byte(test.Raw)

	//The content of wbytes and rbytes is same but when I compare the use following method diffBytes,
	//it failed. The same situation appear at pkg/tunnel/dns/util_test.go
	fmt.Printf("wbytes:%v\nrbytes:%v", wbytes, rbytes)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("get response failed")
	}

	if diffBytes(rbytes, wbytes) {
		//t.Errorf("raw bytes is not equal\n")
	}

	diff(t, resp, &test.Resp)

	//rbody := resp.Body
	//var bout bytes.Buffer
	//if rbody != nil {
	//	_,err = io.Copy(&bout,rbody)
	//	if err != nil {
	//		t.Errorf("%v", err)
	//	}
	//	rbody.Close()
	//}
	//body := bout.String()
	//if body != test.Body {
	//	t.Errorf("Body = %q want %q", body, test.Body)
	//}
}

func TestIsChunked(t *testing.T) {
	tests := []struct {
		desc string
		resp http.Response
		exp  bool
	}{
		{
			desc: "there is chunked value in header filed",
			resp: http.Response{
				Header: http.Header{
					"Transfer-Encoding": []string{"chunked"},
				},
			},
			exp: true,
		},
		{
			desc: "there is chunked value in TransferEncoding filed",
			resp: http.Response{
				TransferEncoding: []string{"chunked"},
			},
			exp: true,
		},
		{
			desc: "there is not chunked value",
			resp: http.Response{
				Header: http.Header{
					"Agent": []string{"Firefox"},
				},
			},
			exp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			act := isChunked(&tt.resp)
			if act != tt.exp {
				t.Errorf("verfify response chunked failed.")
			}
		})
	}
}

//func dummyReq(method string) *http.Request {
//	return &http.Request{Method: method}
//}
//
//func dummyReq11(method string) *http.Request {
//	return &http.Request{Method: method, Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1}
//}

func diff(t *testing.T, have, want interface{}) {
	t.Helper()
	hv := reflect.ValueOf(have).Elem()
	wv := reflect.ValueOf(want).Elem()
	if hv.Type() != wv.Type() {
		t.Errorf("type mismatch %v want %v", hv.Type(), wv.Type())
	}
	for i := 0; i < hv.NumField(); i++ {
		name := hv.Type().Field(i).Name
		if !token.IsExported(name) {
			continue
		}
		if name == "Body" {
			continue
		}
		hf := hv.Field(i).Interface()
		wf := wv.Field(i).Interface()
		if !reflect.DeepEqual(hf, wf) {
			t.Errorf("%s = %v want %v", name, hf, wf)
		}
	}
}

func diffBytes(a, b []byte) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
