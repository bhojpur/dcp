package configuration

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
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	webhookutil "github.com/bhojpur/dcp/pkg/appmanager/webhook/util"
	"k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	mutatingWebhookConfigurationName   = "app-mutating-webhook-configuration"
	validatingWebhookConfigurationName = "app-validating-webhook-configuration"
)

func Ensure(c client.Client, handlers map[string]webhookutil.Handler, caBundle []byte) error {
	mutatingConfig := &v1beta1.MutatingWebhookConfiguration{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: mutatingWebhookConfigurationName}, mutatingConfig); err != nil {
		return fmt.Errorf("not found MutatingWebhookConfiguration %s", mutatingWebhookConfigurationName)
	}
	validatingConfig := &v1beta1.ValidatingWebhookConfiguration{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: validatingWebhookConfigurationName}, validatingConfig); err != nil {
		return fmt.Errorf("not found ValidatingWebhookConfiguration %s", validatingWebhookConfigurationName)
	}

	mutatingTemplate, err := parseMutatingTemplate(mutatingConfig)
	if err != nil {
		return err
	}
	validatingTemplate, err := parseValidatingTemplate(validatingConfig)
	if err != nil {
		return err
	}

	var mutatingWHs []v1beta1.MutatingWebhook
	for i := range mutatingTemplate {
		wh := &mutatingTemplate[i]
		wh.ClientConfig.CABundle = caBundle
		path, err := getPath(&wh.ClientConfig)
		if err != nil {
			return err
		}
		if _, ok := handlers[path]; !ok {
			klog.Warningf("Find path %s not in handlers %v", path, handlers)
			continue
		}
		if wh.ClientConfig.Service != nil {
			wh.ClientConfig.Service.Namespace = webhookutil.GetNamespace()
			wh.ClientConfig.Service.Name = webhookutil.GetServiceName()
		}
		if host := webhookutil.GetHost(); len(host) > 0 && wh.ClientConfig.Service != nil {
			convertClientConfig(&wh.ClientConfig, host, webhookutil.GetPort())
		}
		mutatingWHs = append(mutatingWHs, *wh)
	}
	mutatingConfig.Webhooks = mutatingWHs

	var validatingWHs []v1beta1.ValidatingWebhook
	for i := range validatingTemplate {
		wh := &validatingTemplate[i]
		wh.ClientConfig.CABundle = caBundle
		path, err := getPath(&wh.ClientConfig)
		if err != nil {
			return err
		}
		if _, ok := handlers[path]; !ok {
			klog.Warningf("Find path %s not in handlers %v", path, handlers)
			continue
		}
		if wh.ClientConfig.Service != nil {
			wh.ClientConfig.Service.Namespace = webhookutil.GetNamespace()
			wh.ClientConfig.Service.Name = webhookutil.GetServiceName()
		}
		if host := webhookutil.GetHost(); len(host) > 0 && wh.ClientConfig.Service != nil {
			convertClientConfig(&wh.ClientConfig, host, webhookutil.GetPort())
		}
		validatingWHs = append(validatingWHs, *wh)
	}
	validatingConfig.Webhooks = validatingWHs

	if err := c.Update(context.TODO(), validatingConfig); err != nil {
		return fmt.Errorf("failed to update %s: %v", validatingWebhookConfigurationName, err)
	}
	if err := c.Update(context.TODO(), mutatingConfig); err != nil {
		return fmt.Errorf("failed to update %s: %v", mutatingWebhookConfigurationName, err)
	}

	return nil
}

func getPath(clientConfig *v1beta1.WebhookClientConfig) (string, error) {
	if clientConfig.Service != nil {
		return *clientConfig.Service.Path, nil
	} else if clientConfig.URL != nil {
		u, err := url.Parse(*clientConfig.URL)
		if err != nil {
			return "", err
		}
		return u.Path, nil
	}
	return "", fmt.Errorf("invalid clientConfig: %+v", clientConfig)
}

func convertClientConfig(clientConfig *v1beta1.WebhookClientConfig, host string, port int) {
	url := fmt.Sprintf("https://%s:%d%s", host, port, *clientConfig.Service.Path)
	clientConfig.URL = &url
	clientConfig.Service = nil
}

func parseMutatingTemplate(mutatingConfig *v1beta1.MutatingWebhookConfiguration) ([]v1beta1.MutatingWebhook, error) {
	if templateStr := mutatingConfig.Annotations["template"]; len(templateStr) > 0 {
		var mutatingWHs []v1beta1.MutatingWebhook
		if err := json.Unmarshal([]byte(templateStr), &mutatingWHs); err != nil {
			return nil, err
		}
		return mutatingWHs, nil
	}

	templateBytes, err := json.Marshal(mutatingConfig.Webhooks)
	if err != nil {
		return nil, err
	}
	if mutatingConfig.Annotations == nil {
		mutatingConfig.Annotations = make(map[string]string, 1)
	}
	mutatingConfig.Annotations["template"] = string(templateBytes)
	return mutatingConfig.Webhooks, nil
}

func parseValidatingTemplate(validatingConfig *v1beta1.ValidatingWebhookConfiguration) ([]v1beta1.ValidatingWebhook, error) {
	if templateStr := validatingConfig.Annotations["template"]; len(templateStr) > 0 {
		var validatingWHs []v1beta1.ValidatingWebhook
		if err := json.Unmarshal([]byte(templateStr), &validatingWHs); err != nil {
			return nil, err
		}
		return validatingWHs, nil
	}

	templateBytes, err := json.Marshal(validatingConfig.Webhooks)
	if err != nil {
		return nil, err
	}
	if validatingConfig.Annotations == nil {
		validatingConfig.Annotations = make(map[string]string, 1)
	}
	validatingConfig.Annotations["template"] = string(templateBytes)
	return validatingConfig.Webhooks, nil
}
