package basicauth

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
	"errors"
	"net/http"

	localAuthenticator "github.com/bhojpur/dcp/pkg/cloud/authenticator"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

// Authenticator authenticates requests using basic auth
type Authenticator struct {
	auth localAuthenticator.Password
}

// New returns a request authenticator that validates credentials using the provided password authenticator
func New(auth localAuthenticator.Password) *Authenticator {
	return &Authenticator{auth}
}

var errInvalidAuth = errors.New("invalid username/password combination")

// AuthenticateRequest authenticates the request using the "Authorization: Basic" header in the request
func (a *Authenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	username, password, found := req.BasicAuth()
	if !found {
		return nil, false, nil
	}

	resp, ok, err := a.auth.AuthenticatePassword(req.Context(), username, password)

	// If the password authenticator didn't error, provide a default error
	if !ok && err == nil {
		err = errInvalidAuth
	}

	return resp, ok, err
}
