package config

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmconfig "k8s.io/controller-manager/config"
	nodelifecycleconfig "k8s.io/kube-controller-manager/config/v1alpha1"
)

// ControllerManagerConfiguration contains elements describing controller manager.
type ControllerManagerConfiguration struct {
	metav1.TypeMeta

	// Generic holds configuration for a generic controller-manager
	Generic cmconfig.GenericControllerManagerConfiguration

	// NodeLifecycleControllerConfiguration holds configuration for
	// NodeLifecycleController related features.
	NodeLifecycleController nodelifecycleconfig.NodeLifecycleControllerConfiguration
}
