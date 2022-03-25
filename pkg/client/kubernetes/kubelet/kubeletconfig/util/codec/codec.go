package codec

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

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/component-base/codec"
	"k8s.io/klog/v2"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"

	kubeletconfig "github.com/bhojpur/dcp/pkg/client/kubernetes/kubelet/apis/config"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubelet/apis/config/scheme"
)

// EncodeKubeletConfig encodes an internal KubeletConfiguration to an external YAML representation.
func EncodeKubeletConfig(internal *kubeletconfig.KubeletConfiguration, targetVersion schema.GroupVersion) ([]byte, error) {
	encoder, err := NewKubeletconfigYAMLEncoder(targetVersion)
	if err != nil {
		return nil, err
	}
	// encoder will convert to external version
	data, err := runtime.Encode(encoder, internal)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// NewKubeletconfigYAMLEncoder returns an encoder that can write objects in the kubeletconfig API group to YAML.
func NewKubeletconfigYAMLEncoder(targetVersion schema.GroupVersion) (runtime.Encoder, error) {
	_, codecs, err := scheme.NewSchemeAndCodecs()
	if err != nil {
		return nil, err
	}
	mediaType := "application/yaml"
	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return nil, fmt.Errorf("unsupported media type %q", mediaType)
	}
	return codecs.EncoderForVersion(info.Serializer, targetVersion), nil
}

// DecodeKubeletConfiguration decodes a serialized KubeletConfiguration to the internal type.
func DecodeKubeletConfiguration(kubeletCodecs *serializer.CodecFactory, data []byte) (*kubeletconfig.KubeletConfiguration, error) {
	var (
		obj runtime.Object
		gvk *schema.GroupVersionKind
	)

	// The UniversalDecoder runs defaulting and returns the internal type by default.
	obj, gvk, err := kubeletCodecs.UniversalDecoder().Decode(data, nil, nil)
	if err != nil {
		// Try strict decoding first. If that fails decode with a lenient
		// decoder, which has only v1beta1 registered, and log a warning.
		// The lenient path is to be dropped when support for v1beta1 is dropped.
		if !runtime.IsStrictDecodingError(err) {
			return nil, errors.Wrap(err, "failed to decode")
		}

		var lenientErr error
		_, lenientCodecs, lenientErr := codec.NewLenientSchemeAndCodecs(
			kubeletconfig.AddToScheme,
			kubeletconfigv1beta1.AddToScheme,
		)

		if lenientErr != nil {
			return nil, lenientErr
		}

		obj, gvk, lenientErr = lenientCodecs.UniversalDecoder().Decode(data, nil, nil)
		if lenientErr != nil {
			// Lenient decoding failed with the current version, return the
			// original strict error.
			return nil, fmt.Errorf("failed lenient decoding: %v", err)
		}
		// Continue with the v1beta1 object that was decoded leniently, but emit a warning.
		klog.Warningf("using lenient decoding as strict decoding failed: %v", err)
	}

	internalKC, ok := obj.(*kubeletconfig.KubeletConfiguration)
	if !ok {
		return nil, fmt.Errorf("failed to cast object to KubeletConfiguration, unexpected type: %v", gvk)
	}

	return internalKC, nil
}
