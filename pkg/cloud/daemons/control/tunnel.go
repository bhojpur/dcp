package control

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
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rancher/remotedialer"
	"github.com/rancher/wrangler/pkg/kv"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
)

func setupTunnel() http.Handler {
	tunnelServer := remotedialer.New(authorizer, remotedialer.DefaultErrorWriter)
	setupProxyDialer(tunnelServer)
	return tunnelServer
}

func setupProxyDialer(tunnelServer *remotedialer.Server) {
	app.DefaultProxyDialerFn = utilnet.DialFunc(func(ctx context.Context, network, address string) (net.Conn, error) {
		_, port, _ := net.SplitHostPort(address)
		addr := "127.0.0.1"
		if port != "" {
			addr += ":" + port
		}
		nodeName, _ := kv.Split(address, ":")
		if tunnelServer.HasSession(nodeName) {
			return tunnelServer.Dial(nodeName, 15*time.Second, "tcp", addr)
		}
		var d net.Dialer
		return d.DialContext(ctx, network, address)
	})
}

func authorizer(req *http.Request) (clientKey string, authed bool, err error) {
	user, ok := request.UserFrom(req.Context())
	if !ok {
		return "", false, nil
	}

	if strings.HasPrefix(user.GetName(), "system:node:") {
		return strings.TrimPrefix(user.GetName(), "system:node:"), true, nil
	}

	return "", false, nil
}
