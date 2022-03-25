package secretsencrypt

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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
	"github.com/bhojpur/dcp/pkg/cloud/version"
	corev1 "k8s.io/api/core/v1"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiserverconfigv1 "k8s.io/apiserver/pkg/apis/config/v1"
)

const (
	EncryptionStart             string = "start"
	EncryptionPrepare           string = "prepare"
	EncryptionRotate            string = "rotate"
	EncryptionReencryptRequest  string = "reencrypt_request"
	EncryptionReencryptActive   string = "reencrypt_active"
	EncryptionReencryptFinished string = "reencrypt_finished"
)

var EncryptionHashAnnotation = version.Program + ".io/encryption-config-hash"

func GetEncryptionProviders(runtime *config.ControlRuntime) ([]apiserverconfigv1.ProviderConfiguration, error) {
	curEncryptionByte, err := ioutil.ReadFile(runtime.EncryptionConfig)
	if err != nil {
		return nil, err
	}

	curEncryption := apiserverconfigv1.EncryptionConfiguration{}
	if err = json.Unmarshal(curEncryptionByte, &curEncryption); err != nil {
		return nil, err
	}
	return curEncryption.Resources[0].Providers, nil
}

func GetEncryptionKeys(runtime *config.ControlRuntime) ([]apiserverconfigv1.Key, error) {

	providers, err := GetEncryptionProviders(runtime)
	if err != nil {
		return nil, err
	}
	if len(providers) > 2 {
		return nil, fmt.Errorf("more than 2 providers (%d) found in secrets encryption", len(providers))
	}

	var curKeys []apiserverconfigv1.Key
	for _, p := range providers {
		if p.AESCBC != nil {
			curKeys = append(curKeys, p.AESCBC.Keys...)
		}
		if p.AESGCM != nil || p.KMS != nil || p.Secretbox != nil {
			return nil, fmt.Errorf("non-standard encryption keys found")
		}
	}
	return curKeys, nil
}

func WriteEncryptionConfig(runtime *config.ControlRuntime, keys []apiserverconfigv1.Key, enable bool) error {

	// Placing the identity provider first disables encryption
	var providers []apiserverconfigv1.ProviderConfiguration
	if enable {
		providers = []apiserverconfigv1.ProviderConfiguration{
			{
				AESCBC: &apiserverconfigv1.AESConfiguration{
					Keys: keys,
				},
			},
			{
				Identity: &apiserverconfigv1.IdentityConfiguration{},
			},
		}
	} else {
		providers = []apiserverconfigv1.ProviderConfiguration{
			{
				Identity: &apiserverconfigv1.IdentityConfiguration{},
			},
			{
				AESCBC: &apiserverconfigv1.AESConfiguration{
					Keys: keys,
				},
			},
		}
	}

	encConfig := apiserverconfigv1.EncryptionConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EncryptionConfiguration",
			APIVersion: "apiserver.config.k8s.io/v1",
		},
		Resources: []apiserverconfigv1.ResourceConfiguration{
			{
				Resources: []string{"secrets"},
				Providers: providers,
			},
		},
	}
	jsonfile, err := json.Marshal(encConfig)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(runtime.EncryptionConfig, jsonfile, 0600)
}

func GenEncryptionConfigHash(runtime *config.ControlRuntime) (string, error) {
	curEncryptionByte, err := ioutil.ReadFile(runtime.EncryptionConfig)
	if err != nil {
		return "", err
	}
	encryptionConfigHash := sha256.Sum256(curEncryptionByte)
	return hex.EncodeToString(encryptionConfigHash[:]), nil
}

// GenReencryptHash generates a sha256 hash fom the existing secrets keys and
// a new key based on the input arguments.
func GenReencryptHash(runtime *config.ControlRuntime, keyName string) (string, error) {

	keys, err := GetEncryptionKeys(runtime)
	if err != nil {
		return "", err
	}
	newKey := apiserverconfigv1.Key{
		Name:   keyName,
		Secret: "12345",
	}
	keys = append(keys, newKey)
	b, err := json.Marshal(keys)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
}

func getEncryptionHashFile(runtime *config.ControlRuntime) (string, error) {
	curEncryptionByte, err := ioutil.ReadFile(runtime.EncryptionHash)
	if err != nil {
		return "", err
	}
	return string(curEncryptionByte), nil
}

func BootstrapEncryptionHashAnnotation(node *corev1.Node, runtime *config.ControlRuntime) error {
	existingAnn, err := getEncryptionHashFile(runtime)
	if err != nil {
		return err
	}
	node.Annotations[EncryptionHashAnnotation] = existingAnn
	return nil
}

func WriteEncryptionHashAnnotation(runtime *config.ControlRuntime, node *corev1.Node, stage string) error {
	encryptionConfigHash, err := GenEncryptionConfigHash(runtime)
	if err != nil {
		return err
	}
	if node.Annotations == nil {
		return fmt.Errorf("node annotations do not exist for %s", node.ObjectMeta.Name)
	}
	ann := stage + "-" + encryptionConfigHash
	node.Annotations[EncryptionHashAnnotation] = ann
	if _, err = runtime.Core.Core().V1().Node().Update(node); err != nil {
		return err
	}
	logrus.Debugf("encryption hash annotation set successfully on node: %s\n", node.ObjectMeta.Name)
	return ioutil.WriteFile(runtime.EncryptionHash, []byte(ann), 0600)
}
