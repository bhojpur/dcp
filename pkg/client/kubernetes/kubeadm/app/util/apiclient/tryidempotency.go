package apiclient

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
	"time"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

// CreateOrUpdateConfigMapWithTry runs CreateOrUpdateSecret with try.
func CreateOrUpdateConfigMapWithTry(client clientset.Interface, cm *v1.ConfigMap) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrUpdateConfigMap(client, cm)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrRetainConfigMapWithTry runs CreateOrRetainConfigMap with try.
func CreateOrRetainConfigMapWithTry(client clientset.Interface, cm *v1.ConfigMap, configMapName string) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrRetainConfigMap(client, cm, configMapName)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrUpdateSecretWithTry runs CreateOrUpdateSecret with try.
func CreateOrUpdateSecretWithTry(client clientset.Interface, secret *v1.Secret) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrUpdateSecret(client, secret)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrUpdateServiceAccountWithTry runs CreateOrUpdateServiceAccount with try.
func CreateOrUpdateServiceAccountWithTry(client clientset.Interface, sa *v1.ServiceAccount) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrUpdateServiceAccount(client, sa)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrUpdateDeploymentWithTry runs CreateOrUpdateDeployment with try.
func CreateOrUpdateDeploymentWithTry(client clientset.Interface, deploy *apps.Deployment) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrUpdateDeployment(client, deploy)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrUpdateDaemonSetWithTry runs CreateOrUpdateDaemonSet with try.
func CreateOrUpdateDaemonSetWithTry(client clientset.Interface, ds *apps.DaemonSet) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrUpdateDaemonSet(client, ds)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// DeleteDaemonSetForegroundWithTry runs DeleteDaemonSetForeground with try.
func DeleteDaemonSetForegroundWithTry(client clientset.Interface, namespace, name string) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := DeleteDaemonSetForeground(client, namespace, name)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// DeleteDeploymentForegroundWithTry runs DeleteDeploymentForeground with try.
func DeleteDeploymentForegroundWithTry(client clientset.Interface, namespace, name string) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := DeleteDeploymentForeground(client, namespace, name)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrUpdateRoleWithTry runs CreateOrUpdateRole with try.
func CreateOrUpdateRoleWithTry(client clientset.Interface, role *rbac.Role) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrUpdateRole(client, role)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrUpdateRoleBindingWithTry runs CreateOrUpdateRoleBinding with try.
func CreateOrUpdateRoleBindingWithTry(client clientset.Interface, roleBinding *rbac.RoleBinding) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrUpdateRoleBinding(client, roleBinding)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrUpdateClusterRoleWithTry runs CreateOrUpdateClusterRole with try.
func CreateOrUpdateClusterRoleWithTry(client clientset.Interface, clusterRole *rbac.ClusterRole) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrUpdateClusterRole(client, clusterRole)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrUpdateClusterRoleBindingWithTry runs CreateOrUpdateClusterRoleBinding with try.
func CreateOrUpdateClusterRoleBindingWithTry(client clientset.Interface, clusterRoleBinding *rbac.ClusterRoleBinding) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrUpdateClusterRoleBinding(client, clusterRoleBinding)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// CreateOrMutateConfigMapWithTry runs CreateOrUpdateClusterRoleBinding with try.
func CreateOrMutateConfigMapWithTry(client clientset.Interface, cm *v1.ConfigMap, mutator ConfigMapMutator) error {
	backoff := getBackOff()

	return wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := CreateOrMutateConfigMap(client, cm, mutator)
		if err != nil {
			// Retry until the timeout
			return false, nil
		}
		// The last f() call was a success, return cleanly
		return true, nil
	})
}

// try 200 times, the interval is three seconds.
func getBackOff() wait.Backoff {
	backoff := wait.Backoff{
		Duration: 3 * time.Second,
		Factor:   1,
		Steps:    200,
	}
	return backoff
}
