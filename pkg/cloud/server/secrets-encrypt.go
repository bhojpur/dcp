package server

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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bhojpur/dcp/pkg/cloud/cluster"
	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
	"github.com/bhojpur/dcp/pkg/cloud/secretsencrypt"
	"github.com/rancher/wrangler/pkg/cloud/generated/controllers/core"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiserverconfigv1 "k8s.io/apiserver/pkg/apis/config/v1"
	"k8s.io/utils/pointer"
)

const aescbcKeySize = 32

type EncryptionState struct {
	Stage        string   `json:"stage"`
	ActiveKey    string   `json:"activekey"`
	Enable       *bool    `json:"enable,omitempty"`
	HashMatch    bool     `json:"hashmatch,omitempty"`
	HashError    string   `json:"hasherror,omitempty"`
	InactiveKeys []string `json:"inactivekeys,omitempty"`
}

type EncryptionRequest struct {
	Stage  *string `json:"stage,omitempty"`
	Enable *bool   `json:"enable,omitempty"`
	Force  bool    `json:"force"`
	Skip   bool    `json:"skip"`
}

func getEncryptionRequest(req *http.Request) (EncryptionRequest, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return EncryptionRequest{}, err
	}
	result := EncryptionRequest{}
	err = json.Unmarshal(b, &result)
	return result, err
}

func encryptionStatusHandler(server *config.Control) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.TLS == nil {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
		status, err := encryptionStatus(server)
		if err != nil {
			genErrorMessage(resp, http.StatusInternalServerError, err)
			return
		}
		b, err := json.Marshal(status)
		if err != nil {
			genErrorMessage(resp, http.StatusInternalServerError, err)
			return
		}
		resp.Write(b)
	})
}

func encryptionStatus(server *config.Control) (EncryptionState, error) {
	state := EncryptionState{}
	providers, err := secretsencrypt.GetEncryptionProviders(server.Runtime)
	if os.IsNotExist(err) {
		return state, nil
	} else if err != nil {
		return state, err
	}
	if providers[1].Identity != nil && providers[0].AESCBC != nil {
		state.Enable = pointer.Bool(true)
	} else if providers[0].Identity != nil && providers[1].AESCBC != nil || !server.EncryptSecrets {
		state.Enable = pointer.Bool(false)
	}

	if err := verifyEncryptionHashAnnotation(server.Runtime, server.Runtime.Core.Core(), ""); err != nil {
		state.HashMatch = false
		state.HashError = err.Error()
	} else {
		state.HashMatch = true
	}
	stage, _, err := getEncryptionHashAnnotation(server.Runtime.Core.Core())
	if err != nil {
		return state, err
	}
	state.Stage = stage
	active := true
	for _, p := range providers {
		if p.AESCBC != nil {
			for _, aesKey := range p.AESCBC.Keys {
				if active {
					active = false
					state.ActiveKey = aesKey.Name
				} else {
					state.InactiveKeys = append(state.InactiveKeys, aesKey.Name)
				}
			}
		}
		if p.Identity != nil {
			active = false
		}
	}

	return state, nil
}

func encryptionEnable(ctx context.Context, server *config.Control, enable bool) error {
	providers, err := secretsencrypt.GetEncryptionProviders(server.Runtime)
	if err != nil {
		return err
	}
	if len(providers) > 2 {
		return fmt.Errorf("more than 2 providers (%d) found in secrets encryption", len(providers))
	}
	curKeys, err := secretsencrypt.GetEncryptionKeys(server.Runtime)
	if err != nil {
		return err
	}
	if providers[1].Identity != nil && providers[0].AESCBC != nil && !enable {
		logrus.Infoln("Disabling secrets encryption")
		if err := secretsencrypt.WriteEncryptionConfig(server.Runtime, curKeys, enable); err != nil {
			return err
		}
	} else if !enable {
		logrus.Infoln("Secrets encryption already disabled")
		return nil
	} else if providers[0].Identity != nil && providers[1].AESCBC != nil && enable {
		logrus.Infoln("Enabling secrets encryption")
		if err := secretsencrypt.WriteEncryptionConfig(server.Runtime, curKeys, enable); err != nil {
			return err
		}
	} else if enable {
		logrus.Infoln("Secrets encryption already enabled")
		return nil
	} else {
		return fmt.Errorf("unable to enable/disable secrets encryption, unknown configuration")
	}
	return cluster.Save(ctx, server, true)
}

func encryptionConfigHandler(ctx context.Context, server *config.Control) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.TLS == nil {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
		if req.Method != http.MethodPut {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		encryptReq, err := getEncryptionRequest(req)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			return
		}
		if encryptReq.Stage != nil {
			switch *encryptReq.Stage {
			case secretsencrypt.EncryptionPrepare:
				err = encryptionPrepare(ctx, server, encryptReq.Force)
			case secretsencrypt.EncryptionRotate:
				err = encryptionRotate(ctx, server, encryptReq.Force)
			case secretsencrypt.EncryptionReencryptActive:
				err = encryptionReencrypt(ctx, server, encryptReq.Force, encryptReq.Skip)
			default:
				err = fmt.Errorf("unknown stage %s requested", *encryptReq.Stage)
			}
		} else if encryptReq.Enable != nil {
			err = encryptionEnable(ctx, server, *encryptReq.Enable)
		}

		if err != nil {
			genErrorMessage(resp, http.StatusBadRequest, err)
			return
		}
		resp.WriteHeader(http.StatusOK)
	})
}

func encryptionPrepare(ctx context.Context, server *config.Control, force bool) error {
	states := secretsencrypt.EncryptionStart + "-" + secretsencrypt.EncryptionReencryptFinished
	if err := verifyEncryptionHashAnnotation(server.Runtime, server.Runtime.Core.Core(), states); err != nil && !force {
		return err
	}

	curKeys, err := secretsencrypt.GetEncryptionKeys(server.Runtime)
	if err != nil {
		return err
	}

	if err := AppendNewEncryptionKey(&curKeys); err != nil {
		return err
	}
	logrus.Infoln("Adding secrets-encryption key: ", curKeys[len(curKeys)-1])

	if err := secretsencrypt.WriteEncryptionConfig(server.Runtime, curKeys, true); err != nil {
		return err
	}
	nodeName := os.Getenv("NODE_NAME")
	node, err := server.Runtime.Core.Core().V1().Node().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if err = secretsencrypt.WriteEncryptionHashAnnotation(server.Runtime, node, secretsencrypt.EncryptionPrepare); err != nil {
		return err
	}
	return cluster.Save(ctx, server, true)
}

func encryptionRotate(ctx context.Context, server *config.Control, force bool) error {

	if err := verifyEncryptionHashAnnotation(server.Runtime, server.Runtime.Core.Core(), secretsencrypt.EncryptionPrepare); err != nil && !force {
		return err
	}

	curKeys, err := secretsencrypt.GetEncryptionKeys(server.Runtime)
	if err != nil {
		return err
	}

	// Right rotate elements
	rotatedKeys := append(curKeys[len(curKeys)-1:], curKeys[:len(curKeys)-1]...)

	if err = secretsencrypt.WriteEncryptionConfig(server.Runtime, rotatedKeys, true); err != nil {
		return err
	}
	logrus.Infoln("Encryption keys right rotated")
	nodeName := os.Getenv("NODE_NAME")
	node, err := server.Runtime.Core.Core().V1().Node().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if err := secretsencrypt.WriteEncryptionHashAnnotation(server.Runtime, node, secretsencrypt.EncryptionRotate); err != nil {
		return err
	}
	return cluster.Save(ctx, server, true)
}

func encryptionReencrypt(ctx context.Context, server *config.Control, force bool, skip bool) error {

	if err := verifyEncryptionHashAnnotation(server.Runtime, server.Runtime.Core.Core(), secretsencrypt.EncryptionRotate); err != nil && !force {
		return err
	}
	server.EncryptForce = force
	server.EncryptSkip = skip
	nodeName := os.Getenv("NODE_NAME")
	node, err := server.Runtime.Core.Core().V1().Node().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	reencryptHash, err := secretsencrypt.GenReencryptHash(server.Runtime, secretsencrypt.EncryptionReencryptRequest)
	if err != nil {
		return err
	}
	ann := secretsencrypt.EncryptionReencryptRequest + "-" + reencryptHash
	node.Annotations[secretsencrypt.EncryptionHashAnnotation] = ann
	if _, err = server.Runtime.Core.Core().V1().Node().Update(node); err != nil {
		return err
	}
	logrus.Debugf("encryption hash annotation set successfully on node: %s\n", node.ObjectMeta.Name)
	return nil
}

func AppendNewEncryptionKey(keys *[]apiserverconfigv1.Key) error {

	aescbcKey := make([]byte, aescbcKeySize)
	_, err := rand.Read(aescbcKey)
	if err != nil {
		return err
	}
	encodedKey := base64.StdEncoding.EncodeToString(aescbcKey)

	newKey := []apiserverconfigv1.Key{
		{
			Name:   "aescbckey-" + time.Now().Format(time.RFC3339),
			Secret: encodedKey,
		},
	}
	*keys = append(*keys, newKey...)
	return nil
}

func getEncryptionHashAnnotation(core core.Interface) (string, string, error) {
	nodeName := os.Getenv("NODE_NAME")
	node, err := core.V1().Node().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	ann := node.Annotations[secretsencrypt.EncryptionHashAnnotation]
	split := strings.Split(ann, "-")
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid annotation %s found on node %s", ann, node.ObjectMeta.Name)
	}
	return split[0], split[1], nil
}

func getServerNodes(core core.Interface) ([]corev1.Node, error) {
	nodes, err := core.V1().Node().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var serverNodes []corev1.Node
	for _, node := range nodes.Items {
		if v, ok := node.Labels[ControlPlaneRoleLabelKey]; ok && v == "true" {
			serverNodes = append(serverNodes, node)
		}
	}
	return serverNodes, nil
}

// verifyEncryptionHashAnnotation checks that all nodes are on the same stage,
// and that a request for new stage is valid
func verifyEncryptionHashAnnotation(runtime *config.ControlRuntime, core core.Interface, prevStage string) error {

	var firstHash string
	var firstNodeName string
	first := true
	serverNodes, err := getServerNodes(core)
	if err != nil {
		return err
	}
	for _, node := range serverNodes {
		hash, ok := node.Annotations[secretsencrypt.EncryptionHashAnnotation]
		if ok && first {
			firstHash = hash
			first = false
			firstNodeName = node.ObjectMeta.Name
		} else if ok && hash != firstHash {
			return fmt.Errorf("hash does not match between %s and %s", firstNodeName, node.ObjectMeta.Name)
		}
	}

	if prevStage == "" {
		return nil
	}

	oldStage, oldHash, err := getEncryptionHashAnnotation(core)
	if err != nil {
		return err
	}

	encryptionConfigHash, err := secretsencrypt.GenEncryptionConfigHash(runtime)
	if err != nil {
		return err
	}
	if !strings.Contains(prevStage, oldStage) {
		return fmt.Errorf("incorrect stage: %s found on node %s", oldStage, serverNodes[0].ObjectMeta.Name)
	} else if oldHash != encryptionConfigHash {
		return fmt.Errorf("invalid hash: %s found on node %s", oldHash, serverNodes[0].ObjectMeta.Name)
	}

	return nil
}

func genErrorMessage(resp http.ResponseWriter, statusCode int, passedErr error) {
	errID, err := rand.Int(rand.Reader, big.NewInt(99999))
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}
	logrus.Warnf("secrets-encrypt-%s: %s", errID.String(), passedErr.Error())
	resp.WriteHeader(statusCode)
	resp.Write([]byte(fmt.Sprintf("error secrets-encrypt-%s: see server logs for more info", errID.String())))
}
