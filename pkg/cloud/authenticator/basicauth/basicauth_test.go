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
	"context"
	"errors"
	"net/http"
	"testing"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
)

type testPassword struct {
	Username string
	Password string
	Called   bool

	User user.Info
	OK   bool
	Err  error
}

func (t *testPassword) AuthenticatePassword(ctx context.Context, user, password string) (*authenticator.Response, bool, error) {
	t.Called = true
	t.Username = user
	t.Password = password
	return &authenticator.Response{User: t.User}, t.OK, t.Err
}

func Test_UnitBasicAuth(t *testing.T) {
	testCases := map[string]struct {
		Header   string
		Password testPassword

		ExpectedCalled   bool
		ExpectedUsername string
		ExpectedPassword string

		ExpectedUser string
		ExpectedOK   bool
		ExpectedErr  bool
	}{
		"no auth": {},
		"empty password basic header": {
			ExpectedCalled:   true,
			ExpectedUsername: "user_with_empty_password",
			ExpectedPassword: "",
			ExpectedErr:      true,
		},
		"valid basic header": {
			ExpectedCalled:   true,
			ExpectedUsername: "myuser",
			ExpectedPassword: "mypassword:withcolon",
			ExpectedErr:      true,
		},
		"password auth returned user": {
			Password:         testPassword{User: &user.DefaultInfo{Name: "returneduser"}, OK: true},
			ExpectedCalled:   true,
			ExpectedUsername: "myuser",
			ExpectedPassword: "mypw",
			ExpectedUser:     "returneduser",
			ExpectedOK:       true,
		},
		"password auth returned error": {
			Password:         testPassword{Err: errors.New("auth error")},
			ExpectedCalled:   true,
			ExpectedUsername: "myuser",
			ExpectedPassword: "mypw",
			ExpectedErr:      true,
		},
	}

	for k, testCase := range testCases {
		password := testCase.Password
		auth := authenticator.Request(New(&password))

		req, _ := http.NewRequest("GET", "/", nil)
		if testCase.ExpectedUsername != "" || testCase.ExpectedPassword != "" {
			req.SetBasicAuth(testCase.ExpectedUsername, testCase.ExpectedPassword)
		}

		resp, ok, err := auth.AuthenticateRequest(req)

		if testCase.ExpectedCalled != password.Called {
			t.Errorf("%s: Expected called=%v, got %v", k, testCase.ExpectedCalled, password.Called)
			continue
		}
		if testCase.ExpectedUsername != password.Username {
			t.Errorf("%s: Expected called with username=%v, got %v", k, testCase.ExpectedUsername, password.Username)
			continue
		}
		if testCase.ExpectedPassword != password.Password {
			t.Errorf("%s: Expected called with password=%v, got %v", k, testCase.ExpectedPassword, password.Password)
			continue
		}

		if testCase.ExpectedErr != (err != nil) {
			t.Errorf("%s: Expected err=%v, got err=%v", k, testCase.ExpectedErr, err)
			continue
		}
		if testCase.ExpectedOK != ok {
			t.Errorf("%s: Expected ok=%v, got ok=%v", k, testCase.ExpectedOK, ok)
			continue
		}
		if testCase.ExpectedUser != "" && testCase.ExpectedUser != resp.User.GetName() {
			t.Errorf("%s: Expected user.GetName()=%v, got %v", k, testCase.ExpectedUser, resp.User.GetName())
			continue
		}
	}
}
