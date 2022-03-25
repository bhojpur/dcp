package kubernetes

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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
)

const testDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
`

func TestYamlToObject(t *testing.T) {
	obj, err := YamlToObject([]byte(testDeployment))
	if err != nil {
		t.Fatalf("YamlToObj failed: %s", err)
	}

	nd, ok := obj.(*appsv1.Deployment)
	if !ok {
		t.Fatalf("Fail to assert deployment: %s", err)
	}

	if nd.GetName() != "nginx-deployment" {
		t.Fatalf("YamlToObj failed: want \"nginx-deployment\" get \"%s\"", nd.GetName())
	}

	val, exist := nd.GetLabels()["app"]
	if !exist {
		t.Fatal("YamlToObj failed: label \"app\" doesnot exist")
	}
	if val != "nginx" {
		t.Fatalf("YamlToObj failed: want \"nginx\" get %s", val)
	}

	if *nd.Spec.Replicas != 3 {
		t.Fatalf("YamlToObj failed: want 3 get %d", *nd.Spec.Replicas)
	}
}
