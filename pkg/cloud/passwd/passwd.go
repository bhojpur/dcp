package passwd

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
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bhojpur/dcp/pkg/cloud/token"
	"github.com/bhojpur/dcp/pkg/cloud/util"
)

type entry struct {
	pass string
	role string
}

type Passwd struct {
	changed bool
	names   map[string]entry
}

func Read(file string) (*Passwd, error) {
	result := &Passwd{
		names: map[string]entry{},
	}

	f, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) < 2 {
			return nil, fmt.Errorf("password file '%s' must have at least 2 columns (password, name), found %d", file, len(record))
		}
		e := entry{
			pass: record[0],
		}
		if len(record) > 3 {
			e.role = record[3]
		}
		result.names[record[1]] = e
	}

	return result, nil
}

func (p *Passwd) Check(name, pass string) (matches bool, exists bool) {
	e, ok := p.names[name]
	if !ok {
		return false, false
	}
	return e.pass == pass, true
}

func (p *Passwd) EnsureUser(name, role, passwd string) error {
	tokenPrefix := "::" + name + ":"
	idx := strings.Index(passwd, tokenPrefix)
	if idx > 0 && strings.HasPrefix(passwd, "K10") {
		passwd = passwd[idx+len(tokenPrefix):]
	}

	if e, ok := p.names[name]; ok {
		if passwd != "" && e.pass != passwd {
			p.changed = true
			e.pass = passwd
		}

		if e.role != role {
			p.changed = true
			e.role = role
		}

		p.names[name] = e
		return nil
	}

	if passwd == "" {
		token, err := token.Random(16)
		if err != nil {
			return err
		}
		passwd = token
	}

	p.changed = true
	p.names[name] = entry{
		pass: passwd,
		role: role,
	}

	return nil
}

func (p *Passwd) Users() []string {
	users := []string{}
	for user := range p.names {
		users = append(users, user)
	}
	return users
}

func (p *Passwd) Pass(name string) (string, bool) {
	e, ok := p.names[name]
	if !ok {
		return "", false
	}
	return e.pass, true
}

func (p *Passwd) Write(passwdFile string) error {
	if !p.changed {
		return nil
	}

	var records [][]string
	for name, e := range p.names {
		records = append(records, []string{
			e.pass,
			name,
			name,
			e.role,
		})
	}

	return writePasswords(passwdFile, records)
}

func writePasswords(passwdFile string, records [][]string) error {
	err := func() error {
		// ensure to close tmp file before rename for filesystems like NTFS
		out, err := os.Create(passwdFile + ".tmp")
		if err != nil {
			return err
		}
		defer out.Close()

		if err := util.SetFileModeForFile(out, 0600); err != nil {
			return err
		}

		return csv.NewWriter(out).WriteAll(records)
	}()
	if err != nil {
		return err
	}

	return os.Rename(passwdFile+".tmp", passwdFile)
}
