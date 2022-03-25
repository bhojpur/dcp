package node_servant

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

	batchv1 "k8s.io/api/batch/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"

	tmplutil "github.com/bhojpur/dcp/pkg/client/util/templates"
)

// RenderNodeServantJob return k8s job
// to start k8s job to run convert/revert on specific node
func RenderNodeServantJob(action string, tmplCtx map[string]string, nodeName string) (*batchv1.Job, error) {
	if err := validate(action, tmplCtx, nodeName); err != nil {
		return nil, err
	}

	var servantJobTemplate, jobBaseName string
	switch action {
	case "convert":
		servantJobTemplate = ConvertServantJobTemplate
		jobBaseName = ConvertJobNameBase
	case "revert":
		servantJobTemplate = RevertServantJobTemplate
		jobBaseName = RevertJobNameBase
	case "preflight-convert":
		servantJobTemplate = ConvertPreflightJobTemplate
		jobBaseName = ConvertPreflightJobNameBase
	}

	tmplCtx["jobName"] = jobBaseName + "-" + nodeName
	tmplCtx["nodeName"] = nodeName
	jobYaml, err := tmplutil.SubsituteTemplate(servantJobTemplate, tmplCtx)
	if err != nil {
		return nil, err
	}

	srvJobObj, err := YamlToObject([]byte(jobYaml))
	if err != nil {
		return nil, err
	}
	srvJob, ok := srvJobObj.(*batchv1.Job)
	if !ok {
		return nil, fmt.Errorf("fail to assert dcpctl-servant job")
	}

	return srvJob, nil
}

// YamlToObject deserializes object in yaml format to a runtime.Object
func YamlToObject(yamlContent []byte) (k8sruntime.Object, error) {
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, _, err := decode(yamlContent, nil, nil)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func validate(action string, tmplCtx map[string]string, nodeName string) error {
	if nodeName == "" {
		return fmt.Errorf("nodeName empty")
	}

	switch action {
	case "convert":
		keysMustHave := []string{"node_servant_image", "dcpsvr_image", "joinToken"}
		return checkKeys(keysMustHave, tmplCtx)
	case "revert":
		keysMustHave := []string{"node_servant_image"}
		return checkKeys(keysMustHave, tmplCtx)
	case "preflight-convert":
		keysMustHave := []string{"node_servant_image"}
		return checkKeys(keysMustHave, tmplCtx)
	default:
		return fmt.Errorf("action invalied: %s ", action)
	}
}

func checkKeys(arr []string, tmplCtx map[string]string) error {
	for _, k := range arr {
		if _, ok := tmplCtx[k]; !ok {
			return fmt.Errorf("key %s not found", k)
		}
	}
	return nil
}
