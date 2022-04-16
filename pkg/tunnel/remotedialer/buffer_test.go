package remotedialer

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
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExceedBuffer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	producerAddress, err := newTestProducer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	serverAddress, server, err := newTestServer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := newTestClient(ctx, "ws://"+serverAddress); err != nil {
		t.Fatal(err)
	}

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, proto, address string) (net.Conn, error) {
				return server.Dialer("client")(ctx, proto, address)
			},
		},
	}

	producerURL := "http://" + producerAddress

	resp, err := client.Get(producerURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	resp2, err := client.Get(producerURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	resp2Body, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 4096*4096, len(resp2Body))
	assert.Equal(t, 4096*4096, len(respBody))
}

func newTestServer(ctx context.Context) (string, *Server, error) {
	auth := func(req *http.Request) (clientKey string, authed bool, err error) {
		return "client", true, nil
	}

	server := New(auth, DefaultErrorWriter)
	address, err := newServer(ctx, server)
	return address, server, err
}

func newTestClient(ctx context.Context, url string) error {
	result := make(chan error, 2)
	go func() {
		err := ConnectToProxy(ctx, url, nil, func(proto, address string) bool {
			return true
		}, nil, func(ctx context.Context, session *Session) error {
			result <- nil
			return nil
		})
		result <- err
	}()
	return <-result
}

func newServer(ctx context.Context, handler http.Handler) (string, error) {
	server := http.Server{
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		Handler: handler,
	}
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	go func() {
		<-ctx.Done()
		listener.Close()
		server.Shutdown(context.Background())
	}()
	go server.Serve(listener)
	return listener.Addr().String(), nil
}

func newTestProducer(ctx context.Context) (string, error) {
	buffer := make([]byte, 4096)
	return newServer(ctx, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		for i := 0; i < 4096; i++ {
			if _, err := resp.Write(buffer); err != nil {
				panic(err)
			}
		}
	}))
}
