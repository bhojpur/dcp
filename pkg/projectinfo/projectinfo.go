package projectinfo

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
	"runtime"
	"strings"
)

var (
	projectPrefix = "dcp"
	labelPrefix   = "bhojpur.net"
	gitVersion    = "v0.0.0"
	gitCommit     = "unknown"
	buildDate     = "2018-03-26T09:00:00Z"
)

func ShortAgentVersion() string {
	commit := gitCommit
	if len(gitCommit) > 7 {
		commit = gitCommit[:7]
	}
	return GetAgentName() + "/" + gitVersion + "-" + commit
}

func ShortServerVersion() string {
	commit := gitCommit
	if len(gitCommit) > 7 {
		commit = gitCommit[:7]
	}
	return GetServerName() + "/" + gitVersion + "-" + commit
}

// The project prefix is: dcp
func GetProjectPrefix() string {
	return projectPrefix
}

// Server name: tunnel-server
func GetServerName() string {
	return projectPrefix + "tunnel-server"
}

// tunnel server label: dcp-tunnel-server
func TunnelServerLabel() string {
	return strings.TrimSuffix(projectPrefix, "-") + "-tunnel-server"
}

// Agent name: tunnel-agent
func GetAgentName() string {
	return projectPrefix + "tunnel-agent"
}

// GetEdgeWorkerLabelKey returns the edge-worker label ("bhojpur.net/is-edge-worker"),
// which is used to identify if a node is a edge node ("true") or a cloud node ("false")
func GetEdgeWorkerLabelKey() string {
	return labelPrefix + "/is-edge-worker"
}

// GetEngineName returns name of Bhojpur DCP server engine agent: dcpsvr
func GetEngineName() string {
	return projectPrefix + "svr"
}

// GetEdgeEnableTunnelLabelKey returns the tunnel agent label ("bhojpur.net/edge-enable-reverseTunnel-client"),
// which is used to identify if tunnel agent is running on the node or not.
func GetEdgeEnableTunnelLabelKey() string {
	return labelPrefix + "/edge-enable-reverseTunnel-client"
}

// GetTunnelName returns name of tunnel: tunnel
func GetTunnelName() string {
	return projectPrefix + "tunnel"
}

// GetControllerManagerName returns name of Bhojpur DCP controller-manager: controller-manager
func GetControllerManagerName() string {
	return projectPrefix + "controller-manager"
}

// GetAppManagerName returns name of tunnel: dcpapp-manager
func GetAppManagerName() string {
	return projectPrefix + "app-manager"
}

// GetAutonomyAnnotation returns annotation key for node autonomy
func GetAutonomyAnnotation() string {
	return fmt.Sprintf("node.beta.%s/autonomy", labelPrefix)
}

// normalizeGitCommit reserve 7 characters for gitCommit
func normalizeGitCommit(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}

	return commit
}

// Info contains version information.
type Info struct {
	GitVersion string `json:"gitVersion"`
	GitCommit  string `json:"gitCommit"`
	BuildDate  string `json:"buildDate"`
	GoVersion  string `json:"goVersion"`
	Compiler   string `json:"compiler"`
	Platform   string `json:"platform"`
}

// Get returns the overall codebase version.
func Get() Info {
	return Info{
		GitVersion: gitVersion,
		GitCommit:  normalizeGitCommit(gitCommit),
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
