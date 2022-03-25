package util

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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"
	"k8s.io/klog/v2"
)

// DownloadFile try to download file from URL and save to savePath multiple times.
func DownloadFile(URL, savePath string, retry int) error {
	var count int
	var err error
	for count = 1; count <= retry; count++ {
		if _, err := os.Stat(savePath); os.IsExist(err) {
			if err := os.Remove(savePath); err != nil {
				return err
			}
		}

		if err = downloadFileOnce(URL, savePath); err != nil {
			continue
		} else {
			break
		}
	}
	if count > retry {
		return err
	}
	return nil
}

//downloadFileOnce download file from URL and save to savePath.
func downloadFileOnce(URL, savePath string) error {
	client := http.Client{Timeout: time.Minute * 10}
	resp, err := client.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server return non-200 status: %v\n", resp.Status)
	}
	len, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	bar := pb.New(int(len)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Second * 1)
	bar.ShowSpeed = true
	bar.Start()

	reader := bar.NewProxyReader(resp.Body)
	if _, err = io.Copy(out, reader); err != nil {
		return err
	}

	bar.Finish()
	return nil
}

//Untar unzip the file to the target directory.
func Untar(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		destFile := filepath.Join(dest, hdr.Name)
		if hdr.Typeflag == tar.TypeDir {
			if _, err := os.Stat(destFile); err != nil {
				if os.IsNotExist(err) {
					if err := os.MkdirAll(destFile, 0775); err != nil {
						return err
					}
					continue
				}
				return err
			}
		} else if hdr.Typeflag == tar.TypeReg {
			file, err := os.OpenFile(destFile, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				klog.Errorf("open file %s error: %v", destFile, err)
				return err
			}
			defer file.Close()
			if _, err := io.Copy(file, tr); err != nil {
				return err
			}
		}
	}
}
