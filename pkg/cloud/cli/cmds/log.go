package cmds

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
	"flag"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Log struct {
	VLevel          int
	VModule         string
	LogFile         string
	AlsoLogToStderr bool
}

var (
	LogConfig Log

	VLevel = cli.IntFlag{
		Name:        "v",
		Usage:       "(logging) Number for the log level verbosity",
		Destination: &LogConfig.VLevel,
	}
	VModule = cli.StringFlag{
		Name:        "vmodule",
		Usage:       "(logging) Comma-separated list of pattern=N settings for file-filtered logging",
		Destination: &LogConfig.VModule,
	}
	LogFile = cli.StringFlag{
		Name:        "log,l",
		Usage:       "(logging) Log to file",
		Destination: &LogConfig.LogFile,
	}
	AlsoLogToStderr = cli.BoolFlag{
		Name:        "alsologtostderr",
		Usage:       "(logging) Log to standard error as well as file (if set)",
		Destination: &LogConfig.AlsoLogToStderr,
	}

	logSetupOnce sync.Once
)

func InitLogging() error {
	var rErr error
	logSetupOnce.Do(func() {
		if err := forkIfLoggingOrReaping(); err != nil {
			rErr = err
			return
		}

		if err := checkUnixTimestamp(); err != nil {
			rErr = err
			return
		}

		setupLogging()
	})
	return rErr
}

func checkUnixTimestamp() error {
	timeNow := time.Now()
	// check if time before 01/01/1980
	if timeNow.Before(time.Unix(315532800, 0)) {
		return fmt.Errorf("server time isn't set properly: %v", timeNow)
	}
	return nil
}

func setupLogging() {
	flag.Set("v", strconv.Itoa(LogConfig.VLevel))
	flag.Set("vmodule", LogConfig.VModule)
	flag.Set("alsologtostderr", strconv.FormatBool(Debug))
	flag.Set("logtostderr", strconv.FormatBool(!Debug))
	if Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
