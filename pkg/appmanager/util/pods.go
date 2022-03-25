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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// GetPodNames returns names of the given Pods array
func GetPodNames(pods []*v1.Pod) sets.String {
	set := sets.NewString()
	for _, pod := range pods {
		set.Insert(pod.Name)
	}
	return set
}

// MergePods merges two pods arrays
func MergePods(pods1, pods2 []*v1.Pod) []*v1.Pod {
	var ret []*v1.Pod
	names := sets.NewString()

	for _, pod := range pods1 {
		if !names.Has(pod.Name) {
			ret = append(ret, pod)
			names.Insert(pod.Name)
		}
	}
	for _, pod := range pods2 {
		if !names.Has(pod.Name) {
			ret = append(ret, pod)
			names.Insert(pod.Name)
		}
	}
	return ret
}
