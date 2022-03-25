package node

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

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"

	kubeadmapi "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/apis/kubeadm"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util/apiclient"
)

// TODO(mattmoyer): Move CreateNewTokens, UpdateOrCreateTokens out of this package to client-go for a generic abstraction and client for a Bootstrap Token

// CreateNewTokens tries to create a token and fails if one with the same ID already exists
func CreateNewTokens(client clientset.Interface, tokens []kubeadmapi.BootstrapToken) error {
	return UpdateOrCreateTokens(client, true, tokens)
}

// UpdateOrCreateTokens attempts to update a token with the given ID, or create if it does not already exist.
func UpdateOrCreateTokens(client clientset.Interface, failIfExists bool, tokens []kubeadmapi.BootstrapToken) error {

	for _, token := range tokens {

		secretName := bootstraputil.BootstrapTokenSecretName(token.Token.ID)
		secret, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(context.TODO(), secretName, metav1.GetOptions{})
		if secret != nil && err == nil && failIfExists {
			return errors.Errorf("a token with id %q already exists", token.Token.ID)
		}

		updatedOrNewSecret := token.ToSecret()
		// Try to create or update the token with an exponential backoff
		err = apiclient.TryRunCommand(func() error {
			if err := apiclient.CreateOrUpdateSecret(client, updatedOrNewSecret); err != nil {
				return errors.Wrapf(err, "failed to create or update bootstrap token with name %s", secretName)
			}
			return nil
		}, 5)
		if err != nil {
			return err
		}
	}
	return nil
}
