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
	"net/http"

	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func hasRole(mustRoles []string, roles []string) bool {
	for _, check := range roles {
		for _, role := range mustRoles {
			if role == check {
				return true
			}
		}
	}
	return false
}

func doAuth(roles []string, serverConfig *config.Control, next http.Handler, rw http.ResponseWriter, req *http.Request) {
	switch {
	case serverConfig == nil:
		logrus.Errorf("Authenticate not initialized: serverConfig is nil")
		rw.WriteHeader(http.StatusUnauthorized)
		return
	case serverConfig.Runtime.Authenticator == nil:
		logrus.Errorf("Authenticate not initialized: serverConfig.Runtime.Authenticator is nil")
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp, ok, err := serverConfig.Runtime.Authenticator.AuthenticateRequest(req)
	if err != nil {
		logrus.Errorf("Failed to authenticate request from %s: %v", req.RemoteAddr, err)
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !ok || !hasRole(roles, resp.User.GetGroups()) {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	ctx := request.WithUser(req.Context(), resp.User)
	req = req.WithContext(ctx)
	next.ServeHTTP(rw, req)
}

func authMiddleware(serverConfig *config.Control, roles ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			doAuth(roles, serverConfig, next, rw, req)
		})
	}
}
