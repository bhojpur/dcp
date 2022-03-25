package bootstrap

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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func Handler(bootstrap *config.ControlRuntimeBootstrap) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		ReadFromDisk(rw, bootstrap)
	})
}

// ReadFromDisk reads the bootstrap data from the files on disk and
// writes their content in JSON form to the given io.Writer.
func ReadFromDisk(w io.Writer, bootstrap *config.ControlRuntimeBootstrap) error {
	paths, err := ObjToMap(bootstrap)
	if err != nil {
		return nil
	}

	dataMap := make(map[string]File)
	for pathKey, path := range paths {
		if path == "" {
			continue
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			logrus.Warnf("failed to read %s", path)
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		dataMap[pathKey] = File{
			Timestamp: info.ModTime(),
			Content:   data,
		}
	}

	return json.NewEncoder(w).Encode(dataMap)
}

// File is a representation of a certificate
// or key file within the bootstrap context that contains
// the contents of the file as well as a timestamp from
// when the file was last modified.
type File struct {
	Timestamp time.Time
	Content   []byte
}

type PathsDataformat map[string]File

// WriteToDiskFromStorage writes the contents of the given reader to the paths
// derived from within the ControlRuntimeBootstrap.
func WriteToDiskFromStorage(files PathsDataformat, bootstrap *config.ControlRuntimeBootstrap) error {
	paths, err := ObjToMap(bootstrap)
	if err != nil {
		return err
	}

	for pathKey, bsf := range files {
		path, ok := paths[pathKey]
		if !ok {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return errors.Wrapf(err, "failed to mkdir %s", filepath.Dir(path))
		}
		if err := os.WriteFile(path, bsf.Content, 0600); err != nil {
			return errors.Wrapf(err, "failed to write to %s", path)
		}
		if err := os.Chtimes(path, bsf.Timestamp, bsf.Timestamp); err != nil {
			return errors.Wrapf(err, "failed to update modified time on %s", path)
		}
	}

	return nil
}

func ObjToMap(obj interface{}) (map[string]string, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	data := map[string]string{}
	return data, json.Unmarshal(bytes, &data)
}
