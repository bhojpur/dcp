package v1alpha1

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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/kubernetes/pkg/apis/core/v1"
	utilpointer "k8s.io/utils/pointer"
)

// SetDefaultsDcpAppDaemon set default values for DcpAppDaemon.
func SetDefaultsDcpAppDaemon(obj *DcpAppDaemon) {

	if obj.Spec.RevisionHistoryLimit == nil {
		obj.Spec.RevisionHistoryLimit = utilpointer.Int32Ptr(10)
	}

	if obj.Spec.WorkloadTemplate.StatefulSetTemplate != nil {
		SetDefaultPodSpec(&obj.Spec.WorkloadTemplate.StatefulSetTemplate.Spec.Template.Spec)
		for i := range obj.Spec.WorkloadTemplate.StatefulSetTemplate.Spec.VolumeClaimTemplates {
			a := &obj.Spec.WorkloadTemplate.StatefulSetTemplate.Spec.VolumeClaimTemplates[i]
			v1.SetDefaults_PersistentVolumeClaim(a)
			v1.SetDefaults_ResourceList(&a.Spec.Resources.Limits)
			v1.SetDefaults_ResourceList(&a.Spec.Resources.Requests)
			v1.SetDefaults_ResourceList(&a.Status.Capacity)
		}
	}
	if obj.Spec.WorkloadTemplate.DeploymentTemplate != nil {
		SetDefaultPodSpec(&obj.Spec.WorkloadTemplate.DeploymentTemplate.Spec.Template.Spec)
	}

}

// SetDefaults_UnitedDeployment set default values for UnitedDeployment.
func SetDefaultsUnitedDeployment(obj *UnitedDeployment) {

	if obj.Spec.RevisionHistoryLimit == nil {
		obj.Spec.RevisionHistoryLimit = utilpointer.Int32Ptr(10)
	}

	if obj.Spec.WorkloadTemplate.StatefulSetTemplate != nil {
		SetDefaultPodSpec(&obj.Spec.WorkloadTemplate.StatefulSetTemplate.Spec.Template.Spec)
		for i := range obj.Spec.WorkloadTemplate.StatefulSetTemplate.Spec.VolumeClaimTemplates {
			a := &obj.Spec.WorkloadTemplate.StatefulSetTemplate.Spec.VolumeClaimTemplates[i]
			v1.SetDefaults_PersistentVolumeClaim(a)
			v1.SetDefaults_ResourceList(&a.Spec.Resources.Limits)
			v1.SetDefaults_ResourceList(&a.Spec.Resources.Requests)
			v1.SetDefaults_ResourceList(&a.Status.Capacity)
		}
	}
	if obj.Spec.WorkloadTemplate.DeploymentTemplate != nil {
		SetDefaultPodSpec(&obj.Spec.WorkloadTemplate.DeploymentTemplate.Spec.Template.Spec)
	}

}

// SetDefaultPod sets default pod
func SetDefaultPod(in *corev1.Pod) {
	SetDefaultPodSpec(&in.Spec)
	if in.Spec.EnableServiceLinks == nil {
		enableServiceLinks := corev1.DefaultEnableServiceLinks
		in.Spec.EnableServiceLinks = &enableServiceLinks
	}
}

// SetDefaultPodSpec sets default pod spec
func SetDefaultPodSpec(in *corev1.PodSpec) {
	v1.SetDefaults_PodSpec(in)
	for i := range in.Volumes {
		a := &in.Volumes[i]
		v1.SetDefaults_Volume(a)
		if a.VolumeSource.HostPath != nil {
			v1.SetDefaults_HostPathVolumeSource(a.VolumeSource.HostPath)
		}
		if a.VolumeSource.Secret != nil {
			v1.SetDefaults_SecretVolumeSource(a.VolumeSource.Secret)
		}
		if a.VolumeSource.ISCSI != nil {
			v1.SetDefaults_ISCSIVolumeSource(a.VolumeSource.ISCSI)
		}
		if a.VolumeSource.RBD != nil {
			v1.SetDefaults_RBDVolumeSource(a.VolumeSource.RBD)
		}
		if a.VolumeSource.DownwardAPI != nil {
			v1.SetDefaults_DownwardAPIVolumeSource(a.VolumeSource.DownwardAPI)
			for j := range a.VolumeSource.DownwardAPI.Items {
				b := &a.VolumeSource.DownwardAPI.Items[j]
				if b.FieldRef != nil {
					v1.SetDefaults_ObjectFieldSelector(b.FieldRef)
				}
			}
		}
		if a.VolumeSource.ConfigMap != nil {
			v1.SetDefaults_ConfigMapVolumeSource(a.VolumeSource.ConfigMap)
		}
		if a.VolumeSource.AzureDisk != nil {
			v1.SetDefaults_AzureDiskVolumeSource(a.VolumeSource.AzureDisk)
		}
		if a.VolumeSource.Projected != nil {
			v1.SetDefaults_ProjectedVolumeSource(a.VolumeSource.Projected)
			for j := range a.VolumeSource.Projected.Sources {
				b := &a.VolumeSource.Projected.Sources[j]
				if b.DownwardAPI != nil {
					for k := range b.DownwardAPI.Items {
						c := &b.DownwardAPI.Items[k]
						if c.FieldRef != nil {
							v1.SetDefaults_ObjectFieldSelector(c.FieldRef)
						}
					}
				}
				if b.ServiceAccountToken != nil {
					v1.SetDefaults_ServiceAccountTokenProjection(b.ServiceAccountToken)
				}
			}
		}
		if a.VolumeSource.ScaleIO != nil {
			v1.SetDefaults_ScaleIOVolumeSource(a.VolumeSource.ScaleIO)
		}
	}
	for i := range in.InitContainers {
		a := &in.InitContainers[i]
		v1.SetDefaults_Container(a)
		/* for j := range a.Ports {
			b := &a.Ports[j]
			v1.SetDefaults_ContainerPort(b)
		} */
		for j := range a.Env {
			b := &a.Env[j]
			if b.ValueFrom != nil {
				if b.ValueFrom.FieldRef != nil {
					v1.SetDefaults_ObjectFieldSelector(b.ValueFrom.FieldRef)
				}
			}
		}
		v1.SetDefaults_ResourceList(&a.Resources.Limits)
		v1.SetDefaults_ResourceList(&a.Resources.Requests)
		if a.LivenessProbe != nil {
			v1.SetDefaults_Probe(a.LivenessProbe)
			if a.LivenessProbe.Handler.HTTPGet != nil {
				v1.SetDefaults_HTTPGetAction(a.LivenessProbe.Handler.HTTPGet)
			}
		}
		if a.ReadinessProbe != nil {
			v1.SetDefaults_Probe(a.ReadinessProbe)
			if a.ReadinessProbe.Handler.HTTPGet != nil {
				v1.SetDefaults_HTTPGetAction(a.ReadinessProbe.Handler.HTTPGet)
			}
		}
		if a.Lifecycle != nil {
			if a.Lifecycle.PostStart != nil {
				if a.Lifecycle.PostStart.HTTPGet != nil {
					v1.SetDefaults_HTTPGetAction(a.Lifecycle.PostStart.HTTPGet)
				}
			}
			if a.Lifecycle.PreStop != nil {
				if a.Lifecycle.PreStop.HTTPGet != nil {
					v1.SetDefaults_HTTPGetAction(a.Lifecycle.PreStop.HTTPGet)
				}
			}
		}
	}
	for i := range in.Containers {
		a := &in.Containers[i]
		// For in-place update, we set default imagePullPolicy to Always
		if a.ImagePullPolicy == "" {
			a.ImagePullPolicy = corev1.PullAlways
		}
		v1.SetDefaults_Container(a)
		/* for j := range a.Ports {
			b := &a.Ports[j]
			v1.SetDefaults_ContainerPort(b)
		} */
		for j := range a.Env {
			b := &a.Env[j]
			if b.ValueFrom != nil {
				if b.ValueFrom.FieldRef != nil {
					v1.SetDefaults_ObjectFieldSelector(b.ValueFrom.FieldRef)
				}
			}
		}
		v1.SetDefaults_ResourceList(&a.Resources.Limits)
		v1.SetDefaults_ResourceList(&a.Resources.Requests)
		if a.LivenessProbe != nil {
			v1.SetDefaults_Probe(a.LivenessProbe)
			if a.LivenessProbe.Handler.HTTPGet != nil {
				v1.SetDefaults_HTTPGetAction(a.LivenessProbe.Handler.HTTPGet)
			}
		}
		if a.ReadinessProbe != nil {
			v1.SetDefaults_Probe(a.ReadinessProbe)
			if a.ReadinessProbe.Handler.HTTPGet != nil {
				v1.SetDefaults_HTTPGetAction(a.ReadinessProbe.Handler.HTTPGet)
			}
		}
		if a.Lifecycle != nil {
			if a.Lifecycle.PostStart != nil {
				if a.Lifecycle.PostStart.HTTPGet != nil {
					v1.SetDefaults_HTTPGetAction(a.Lifecycle.PostStart.HTTPGet)
				}
			}
			if a.Lifecycle.PreStop != nil {
				if a.Lifecycle.PreStop.HTTPGet != nil {
					v1.SetDefaults_HTTPGetAction(a.Lifecycle.PreStop.HTTPGet)
				}
			}
		}
	}
}
