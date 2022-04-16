package extract

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
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestFindPathFromExtract(t *testing.T) {
	tempdir := t.TempDir()
	testImageRef := "docker.io/bhojpur/uke-runtime:v1.22.4-uker1"
	ref, err := name.ParseReference(testImageRef)
	if err != nil {
		t.Fatalf("Failed to parse image reference: %v", err)
	}

	testOperatingSystems := map[string]string{
		"linux":   "containerd",
		"windows": "containerd.exe",
	}

	// https://github.com/google/go-containerregistry/commit/f9a1886f3df0e2b00d6c62715114fe1093ab1ad7
	// changed go-containerregistry behavior; tar paths are now platform-specific and will have forward
	// slashes on Linux and backslashes on Windows.
	for operatingSystem, pauseBin := range testOperatingSystems {
		image, err := remote.Image(ref, remote.WithPlatform(v1.Platform{Architecture: "amd64", OS: operatingSystem}))
		if err != nil {
			t.Fatalf("Failed to pull remote image: %v", err)
		}

		extractMap := map[string]string{
			"/bin":    filepath.Join(tempdir, "bin"),
			"/charts": filepath.Join(tempdir, "charts"),
		}

		t.Logf("Testing ExtractDirs with map %#v for %s", extractMap, operatingSystem)
		if err := ExtractDirs(image, extractMap); err != nil {
			t.Errorf("Failed to extract containerd binary for %s: %v", operatingSystem, err)
			continue
		}

		i, err := os.Stat(filepath.Join(tempdir, "bin", pauseBin))
		if err != nil {
			t.Errorf("containerd binary for %s not found: %v", operatingSystem, err)
			continue
		}

		t.Logf("containerd binary for %s extracted successfully: %s", operatingSystem, i.Name())
	}
}

func TestFindPath(t *testing.T) {
	type mss map[string]string
	type testPath struct {
		in  string
		out string
		err error
	}
	temp := os.TempDir()
	findPathTests := []struct {
		dirs  mss
		paths []testPath
	}{
		{
			// test a simple root directory mapping with various valid and invalid paths
			dirs: mss{"/": temp},
			paths: []testPath{
				{
					in:  "/test.txt",
					out: filepath.Join(temp, "test.txt"),
					err: nil,
				}, {
					in:  "///test.txt",
					out: filepath.Join(temp, "test.txt"),
					err: nil,
				}, {
					in:  "/etc/../test.txt",
					out: filepath.Join(temp, "test.txt"),
					err: nil,
				}, {
					in:  "test.txt",
					out: filepath.Join(temp, "test.txt"),
					err: nil,
				}, {
					in:  "/etc/hosts",
					out: filepath.Join(temp, "etc", "hosts"),
					err: nil,
				}, {
					in:  "/var/lib/bhojpur",
					out: filepath.Join(temp, "var", "lib", "bhojpur"),
					err: nil,
				}, {
					in:  "../../etc/passwd",
					out: "",
					err: ErrIllegalPath,
				},
			},
		}, {
			// test no mapping at all
			dirs: mss{},
			paths: []testPath{
				{
					in:  "/text.txt",
					out: "",
					err: nil,
				},
			},
		}, {
			// test mapping various nested paths
			dirs: mss{
				"/Files/bin": filepath.Join(temp, "Files-bin"),
				"/Files":     filepath.Join(temp, "Files"),
				"/etc":       filepath.Join(temp, "etc"),
			},
			paths: []testPath{
				{
					in:  "Files/bin",
					out: filepath.Join(temp, "Files-bin"),
					err: nil,
				}, {
					in:  "Files/bin/test.txt",
					out: filepath.Join(temp, "Files-bin", "test.txt"),
					err: nil,
				}, {
					in:  "Files/bin/aux",
					out: filepath.Join(temp, "Files-bin", "aux"),
					err: nil,
				}, {
					in:  "Files/bin/aux/mount",
					out: filepath.Join(temp, "Files-bin", "aux", "mount"),
					err: nil,
				}, {
					in:  "Files",
					out: filepath.Join(temp, "Files"),
					err: nil,
				}, {
					in:  "Files/test.txt",
					out: filepath.Join(temp, "Files", "test.txt"),
					err: nil,
				}, {
					in:  "Files/opt",
					out: filepath.Join(temp, "Files", "opt"),
					err: nil,
				}, {
					in:  "Files/opt/other.txt",
					out: filepath.Join(temp, "Files", "opt", "other.txt"),
					err: nil,
				}, {
					in:  "etc",
					out: filepath.Join(temp, "etc"),
					err: nil,
				}, {
					in:  "etc/hosts",
					out: filepath.Join(temp, "etc", "hosts"),
					err: nil,
				}, {
					in:  "etc/shadow/passwd",
					out: filepath.Join(temp, "etc", "shadow", "passwd"),
					err: nil,
				}, {
					in:  "sbin",
					out: "",
					err: nil,
				}, {
					in:  "sbin/ip",
					out: "",
					err: nil,
				}, {
					in:  "Files/bin/../../../../etc/passwd",
					out: "",
					err: ErrIllegalPath,
				},
			},
		},
	}

	for _, test := range findPathTests {
		t.Logf("Testing paths with dirs %#v", test.dirs)
		for _, testPath := range test.paths {
			dirs, err := cleanExtractDirs(test.dirs)
			if err != nil {
				t.Errorf("Failed to clean extracted dirs: %v", err)
				continue
			}
			// as of recent go-containerruntime versions, tar file paths are pre-processed with filepath.Clean
			in := filepath.Clean(testPath.in)
			destination, err := findPath(dirs, in)
			t.Logf("Got mapped path %q, err %v for image path %q", destination, err, in)
			if destination != testPath.out {
				t.Errorf("Expected path %q but got path %q for image path %q", testPath.out, destination, in)
			}
			if err != testPath.err {
				t.Errorf("Expected error %v but got error %v for image path %q", testPath.err, err, in)
			}
		}
	}
}
