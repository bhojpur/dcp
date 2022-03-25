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
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// BuildArgumentListFromMap takes two string-string maps, one with the base arguments and one
// with optional override arguments. In the return list override arguments will precede base
// arguments
func BuildArgumentListFromMap(baseArguments map[string]string, overrideArguments map[string]string) []string {
	var command []string
	var keys []string

	argsMap := make(map[string]string)

	for k, v := range baseArguments {
		argsMap[k] = v
	}

	for k, v := range overrideArguments {
		argsMap[k] = v
	}

	for k := range argsMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		command = append(command, fmt.Sprintf("--%s=%s", k, argsMap[k]))
	}

	return command
}

// ParseArgumentListToMap parses a CLI argument list in the form "--foo=bar" to a string-string map
func ParseArgumentListToMap(arguments []string) map[string]string {
	resultingMap := map[string]string{}
	for i, arg := range arguments {
		key, val, err := parseArgument(arg)

		// Ignore if the first argument doesn't satisfy the criteria, it's most often the binary name
		// Warn in all other cases, but don't error out. This can happen only if the user has edited the argument list by hand, so they might know what they are doing
		if err != nil {
			if i != 0 {
				fmt.Printf("[kubeadm] WARNING: The component argument %q could not be parsed correctly. The argument must be of the form %q. Skipping...\n", arg, "--")
			}
			continue
		}

		resultingMap[key] = val
	}
	return resultingMap
}

// ReplaceArgument gets a command list; converts it to a map for easier modification, runs the provided function that
// returns a new modified map, and then converts the map back to a command string slice
func ReplaceArgument(command []string, argMutateFunc func(map[string]string) map[string]string) []string {
	argMap := ParseArgumentListToMap(command)

	// Save the first command (the executable) if we're sure it's not an argument (i.e. no --)
	var newCommand []string
	if len(command) > 0 && !strings.HasPrefix(command[0], "--") {
		newCommand = append(newCommand, command[0])
	}
	newArgMap := argMutateFunc(argMap)
	newCommand = append(newCommand, BuildArgumentListFromMap(newArgMap, map[string]string{})...)
	return newCommand
}

// parseArgument parses the argument "--foo=bar" to "foo" and "bar"
func parseArgument(arg string) (string, string, error) {
	if !strings.HasPrefix(arg, "--") {
		return "", "", errors.New("the argument should start with '--'")
	}
	if !strings.Contains(arg, "=") {
		return "", "", errors.New("the argument should have a '=' between the flag and the value")
	}
	// Remove the starting --
	arg = strings.TrimPrefix(arg, "--")
	// Split the string on =. Return only two substrings, since we want only key/value, but the value can include '=' as well
	keyvalSlice := strings.SplitN(arg, "=", 2)

	// Make sure both a key and value is present
	if len(keyvalSlice) != 2 {
		return "", "", errors.New("the argument must have both a key and a value")
	}
	if len(keyvalSlice[0]) == 0 {
		return "", "", errors.New("the argument must have a key")
	}

	return keyvalSlice[0], keyvalSlice[1], nil
}
