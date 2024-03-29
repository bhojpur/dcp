package dataverify

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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// Verify will check the sha256sums and links from the files in a given directory
func Verify(dir string) error {
	failed := false
	if err := VerifySums(dir, ".sha256sums"); err != nil {
		logrus.Errorf("Unable to verify sums: %s", err)
		failed = true
	}
	if err := VerifyLinks(dir, ".links"); err != nil {
		logrus.Errorf("Unable to verify links: %s", err)
		failed = true
	}
	if failed {
		return fmt.Errorf("failed to verify directory %s", dir)
	}
	return nil
}

// VerifySums will take a file which contains a list of hash sums for files and verify they match
func VerifySums(root, sumListFile string) error {
	sums, err := fileMapFields(filepath.Join(root, sumListFile), 1, 0)
	if err != nil {
		return err
	}
	if len(sums) == 0 {
		return fmt.Errorf("no entries found in %s", sumListFile)
	}
	numFailed := 0
	for sumFile, sumExpected := range sums {
		file := filepath.Join(root, sumFile)
		sumActual, _ := sha256Sum(file)
		if sumExpected != sumActual {
			logrus.Errorf("Hash for file %s expected to be %s (fail)", sumFile, sumExpected)
			numFailed++
		} else {
			logrus.Debugf("Verified hash %s is correct", sumFile)
		}
	}
	if numFailed != 0 {
		return fmt.Errorf("failed %d hash verifications", numFailed)
	}
	return nil
}

// VerifyLinks will take a file which contains a list of target links for files and verify they match
func VerifyLinks(root, linkListFile string) error {
	links, err := fileMapFields(filepath.Join(root, linkListFile), 0, 1)
	if err != nil {
		return err
	}
	if len(links) == 0 {
		return fmt.Errorf("no entries found in %s", linkListFile)
	}
	numFailed := 0
	for linkFile, linkExpected := range links {
		file := filepath.Join(root, linkFile)
		linkActual, _ := os.Readlink(file)
		if linkExpected != linkActual {
			logrus.Errorf("Link for file %s expected to be %s (fail)", linkFile, linkExpected)
			numFailed++
		} else {
			logrus.Debugf("Verified link %s is correct", linkFile)
		}
	}
	if numFailed != 0 {
		return fmt.Errorf("failed %d link verifications", numFailed)
	}
	return nil
}

func fileMapFields(fileName string, key, val int) (map[string]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) <= key || len(fields) <= val {
			return nil, fmt.Errorf("fields for file %s (%d) smaller than required index (key: %d, val: %d)", fileName, len(fields), key, val)
		}
		result[fields[key]] = fields[val]
	}
	return result, scanner.Err()
}

func sha256Sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
