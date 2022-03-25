//go:build windows
// +build windows

package preflight

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
	"os/user"

	"github.com/pkg/errors"
)

// The "Well-known SID" of Administrator group
// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
const administratorSID = "S-1-5-32-544"

// Check validates if a user has elevated (administrator) privileges.
func (ipuc IsPrivilegedUserCheck) Check() (warnings, errorList []error) {
	currUser, err := user.Current()
	if err != nil {
		return nil, []error{errors.Wrap(err, "cannot get current user")}
	}

	groupIds, err := currUser.GroupIds()
	if err != nil {
		return nil, []error{errors.Wrap(err, "cannot get group IDs for current user")}
	}

	for _, sid := range groupIds {
		if sid == administratorSID {
			return nil, nil
		}
	}

	return nil, []error{errors.New("user is not running as administrator")}
}
