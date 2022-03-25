package iptables

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
	"net"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	coreinformer "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"

	"github.com/bhojpur/dcp/pkg/tunnel/constants"
	"github.com/bhojpur/dcp/pkg/tunnel/util"
	"github.com/bhojpur/dcp/pkg/utils/iptables"
)

var (
	ListenAddrForMaster         = net.JoinHostPort("0.0.0.0", constants.TunnelServerMasterPort)
	ListenInsecureAddrForMaster = net.JoinHostPort("127.0.0.1", constants.TunnelServerMasterInsecurePort)
	IptablesSyncPeriod          = 60
)

func newFakeIptablesManager(client clientset.Interface,
	nodeInformer coreinformer.NodeInformer,
	listenAddr string,
	listenInsecureAddr string,
	syncPeriod int,
	execer exec.Interface) *iptablesManager {

	protocol := iptables.ProtocolIpv4
	iptInterface := iptables.New(execer, protocol)

	if syncPeriod < defaultSyncPeriod {
		syncPeriod = defaultSyncPeriod
	}

	im := &iptablesManager{
		kubeClient:       client,
		iptables:         iptInterface,
		execer:           execer,
		nodeInformer:     nodeInformer,
		secureDnatDest:   listenAddr,
		insecureDnatDest: listenInsecureAddr,
		lastNodesIP:      make([]string, 0),
		lastDnatPorts:    make([]string, 0),
		syncPeriod:       syncPeriod,
	}
	return im
}

func TestCleanupIptableSettingAllExists(t *testing.T) {
	//1. create iptabeleMgr
	fakeClient := &fake.Clientset{}
	fakeInformerFactory := informers.NewSharedInformerFactory(fakeClient, 0*time.Second)
	fcmd := fakeexec.FakeCmd{
		CombinedOutputScript: []fakeexec.FakeAction{
			// iptables version check
			func() ([]byte, []byte, error) { return []byte("iptables v1.9.22"), nil, nil },
			// DeleteRule Success
			func() ([]byte, []byte, error) { return []byte{}, nil, nil }, // success on the first call
			func() ([]byte, []byte, error) { return []byte{}, nil, nil }, // success on the second call

			// FlushChain Success
			func() ([]byte, []byte, error) { return []byte{}, nil, nil },
			// DeleteChain Success
			func() ([]byte, []byte, error) { return []byte{}, nil, nil },

			// FlushChain Success
			func() ([]byte, []byte, error) { return []byte{}, nil, nil },
			// DeleteChain Success
			func() ([]byte, []byte, error) { return []byte{}, nil, nil },

			// FlushChain Success
			func() ([]byte, []byte, error) { return []byte{}, nil, nil },
			// DeleteChain Success
			func() ([]byte, []byte, error) { return []byte{}, nil, nil },
		},
	}
	fexec := fakeexec.FakeExec{
		CommandScript: []fakeexec.FakeCommandAction{
			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },

			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },

			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },

			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },

			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
			func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(&fcmd, cmd, args...) },
		},
	}

	iptablesMgr := newFakeIptablesManager(fakeClient,
		fakeInformerFactory.Core().V1().Nodes(),
		ListenAddrForMaster,
		ListenInsecureAddrForMaster,
		IptablesSyncPeriod,
		&fexec)

	if iptablesMgr == nil {
		t.Errorf("fail to create a new IptableManager")
	}

	//2. create configmap
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.TunnelServerDnatConfigMapName,
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"dnat-ports-pair": "",
		},
	}
	fakeInformerFactory.Core().V1().ConfigMaps().Informer().GetStore().Add(configmap)

	//3. call cleanupIptableSetting
	iptablesMgr.cleanupIptableSetting()
}
