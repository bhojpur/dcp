package lock

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
	"errors"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/constants"
)

const (
	AnnotationAcquireTime = "bhojpur.net/dcpctllock.acquire.time"
	AnnotationIsLocked    = "bhojpur.net/dcpctllock.locked"

	LockTimeoutMin = 5
)

var (
	ErrAcquireLock error = errors.New("fail to acquire lock configmap/dcpctl-lock")
	ErrReleaseLock error = errors.New("fail to release lock configmap/dcpctl-lock")
)

// AcquireLock tries to acquire the lock lock configmap/dcpctl-lock
func AcquireLock(cli *kubernetes.Clientset) error {
	lockCm, err := cli.CoreV1().ConfigMaps("kube-system").
		Get(context.Background(), constants.DcpctlLockConfigMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// the lock is not exist, create one
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.DcpctlLockConfigMapName,
					Namespace: "kube-system",
					Annotations: map[string]string{
						AnnotationAcquireTime: strconv.FormatInt(time.Now().Unix(), 10),
						AnnotationIsLocked:    "true",
					},
				},
			}
			if _, err := cli.CoreV1().ConfigMaps("kube-system").
				Create(context.Background(), cm, metav1.CreateOptions{}); err != nil {
				klog.Error("the lock configmap/dcpctl-lock is not found, " +
					"but fail to create a new one")
				return ErrAcquireLock
			}
			return nil
		}
		return ErrAcquireLock
	}

	if lockCm.Annotations[AnnotationIsLocked] == "true" {
		// check if the lock expired
		//
		// TODO The timeout mechanism is just a short-term workaround, in the
		// future version, we will use a CRD and controller to manage the
		// Bhojpur DCP cluster, which also prevents the contention between users.
		old, err := strconv.ParseInt(lockCm.Annotations[AnnotationAcquireTime], 10, 64)
		if err != nil {
			return ErrAcquireLock
		}
		if isTimeout(old) {
			// if the lock is expired, acquire it
			if err := acquireLockAndUpdateCm(cli, lockCm); err != nil {
				return err
			}
			return nil
		}

		// lock has been acquired by others
		klog.Errorf("the lock is held by others, it was being acquired at %s",
			lockCm.Annotations[AnnotationAcquireTime])
		return ErrAcquireLock
	}

	if lockCm.Annotations[AnnotationIsLocked] == "false" {
		if err := acquireLockAndUpdateCm(cli, lockCm); err != nil {
			return err
		}
	}

	return nil
}

func isTimeout(old int64) bool {
	deadline := old + LockTimeoutMin*60
	return time.Now().Unix() > deadline
}

func acquireLockAndUpdateCm(cli kubernetes.Interface, lockCm *v1.ConfigMap) error {
	lockCm.Annotations[AnnotationIsLocked] = "true"
	lockCm.Annotations[AnnotationAcquireTime] = strconv.FormatInt(time.Now().Unix(), 10)
	if _, err := cli.CoreV1().ConfigMaps("kube-system").
		Update(context.Background(), lockCm, metav1.UpdateOptions{}); err != nil {
		if apierrors.IsResourceExpired(err) {
			klog.Error("the lock is held by others")
			return ErrAcquireLock
		}
		klog.Error("successfully acquire the lock but fail to update it")
		return ErrAcquireLock
	}
	return nil
}

// ReleaseLock releases the lock configmap/dcpctl-lock
func ReleaseLock(cli *kubernetes.Clientset) error {
	lockCm, err := cli.CoreV1().ConfigMaps("kube-system").
		Get(context.Background(), constants.DcpctlLockConfigMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Error("lock is not found when try to release, " +
				"please check if the configmap/dcpctl-lock " +
				"is being deleted manually")
			return ErrReleaseLock
		}
		klog.Error("fail to get lock configmap/dcpctl-lock, " +
			"when try to release it")
		return ErrReleaseLock
	}
	if lockCm.Annotations[AnnotationIsLocked] == "false" {
		klog.Error("lock has already been released, " +
			"please check if the configmap/dcpctl-lock " +
			"is being updated manually")
		return ErrReleaseLock
	}

	// release the lock
	lockCm.Annotations[AnnotationIsLocked] = "false"
	delete(lockCm.Annotations, AnnotationAcquireTime)

	_, err = cli.CoreV1().ConfigMaps("kube-system").Update(context.Background(), lockCm, metav1.UpdateOptions{})
	if err != nil {
		if apierrors.IsResourceExpired(err) {
			klog.Error("lock has been touched by others during release, " +
				"which is not supposed to happen. " +
				"Please check if lock is being updated manually.")
			return ErrReleaseLock

		}
		klog.Error("fail to update lock configmap/dcpctl-lock, " +
			"when try to release it")
		return ErrReleaseLock
	}

	return nil
}

// DeleteLock should only be called when you've achieved the lock.
// It will delete the dcpctl-lock configmap.
func DeleteLock(cli *kubernetes.Clientset) error {
	if err := cli.CoreV1().ConfigMaps("kube-system").
		Delete(context.Background(), constants.DcpctlLockConfigMapName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		klog.Error("fail to delete the dcpctl lock", err)
		return err
	}
	return nil
}
