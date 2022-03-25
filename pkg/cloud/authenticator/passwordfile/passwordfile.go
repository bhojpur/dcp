package passwordfile

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
	"crypto/subtle"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"k8s.io/klog"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
)

// PasswordAuthenticator authenticates users by password
type PasswordAuthenticator struct {
	users map[string]*userPasswordInfo
}

type userPasswordInfo struct {
	info     *user.DefaultInfo
	password string
}

// NewCSV returns a PasswordAuthenticator, populated from a CSV file.
// The CSV file must contain records in the format "password,username,useruid"
func NewCSV(path string) (*PasswordAuthenticator, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	recordNum := 0
	users := make(map[string]*userPasswordInfo)
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) < 3 {
			return nil, fmt.Errorf("password file '%s' must have at least 3 columns (password, user name, user uid), found %d", path, len(record))
		}
		obj := &userPasswordInfo{
			info:     &user.DefaultInfo{Name: record[1], UID: record[2]},
			password: record[0],
		}
		if len(record) >= 4 {
			obj.info.Groups = strings.Split(record[3], ",")
		}
		recordNum++
		if _, exist := users[obj.info.Name]; exist {
			klog.Warningf("duplicate username '%s' has been found in password file '%s', record number '%d'", obj.info.Name, path, recordNum)
		}
		users[obj.info.Name] = obj
	}

	return &PasswordAuthenticator{users}, nil
}

// AuthenticatePassword returns user info if authentication is successful, nil otherwise
func (a *PasswordAuthenticator) AuthenticatePassword(ctx context.Context, username, password string) (*authenticator.Response, bool, error) {
	user, ok := a.users[username]
	if !ok {
		return nil, false, nil
	}
	if subtle.ConstantTimeCompare([]byte(user.password), []byte(password)) == 0 {
		return nil, false, nil
	}
	return &authenticator.Response{User: user.info}, true, nil
}
