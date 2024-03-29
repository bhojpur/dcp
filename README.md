# Bhojpur DCP - Distributed Cloud Platform

The `Bhojpur DCP` is a high-performance, distributed cloud computing engine applied within the
[Bhojpur.NET Platform](https://github.com/bhojpur/platform) for delivery of `applications` and/or
`services` seamlessly in a geo-distributed scenario (e.g., IoT/M2M). It can host a wide range of
workloads and assure reliable execution.

## Key Features

- Integrated Distributed Cloud Orchestration Engine
- Seamless Edge/Cloud cluster management based on Kubernetes
- High-density Unikernel workload management in IaaS
- Pre-integrated with Network Service Mesh
- Web-based dashboard for Distributed Application management

## Building Source Code

After downloading this source code into a local folder, run the following commands to set the
correct version (e.g., v1.23.5) of `Kubernetes` libraries to avoid source code compilation
issues. The `set_k8s_version.sh` either updates or upgrades `go.mod` file accordingly.

```bash
$ go clean --cache
$ go clean --modcache
$ set_k8s_version.sh v1.23.5
$ go mod tidy
$ go get
```

You need either `vagrant`+`virtualbox` or `Docker` instance to build from the source code.

```bash
$ mkdir -p build/data && make download && make generate
```

You can build it locally using `make` on Linux or Windows

```bash
$ make
```

Alternatively, you can use `task build-cloud-tools` and `task build-grid-tools` to get the
binary images locally for development purpose.

### Known Compilation Issues

- `k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65` required to resolve `gnostic` issue
- `github.com/kubernetes-sigs/cri-tools/cmd/crictl` must be made importable by pkg name change
- `github.com/bhojpur/apiserver-network-proxy` must be `v0.0.22`

## Quick-Start - Install Script

The [`install.sh`](https://get.bhojpur.net/dcp/install.sh) script provides a convenient way to
download Bhojpur DCP and add a service to systemd or openrc. To install `Bhojpur DCP` as a
service, run:

```bash
$ curl -sfL https://get.bhojpur.net/dcp/install.sh | sh -
```

A kubeconfig file is written to `/etc/bhojpur/dcp/dcp.yaml` and the service is automatically started
or restarted. The install script will install `Bhojpur DCP` and additional utilities, such as `kubectl`, `crictl`, `dcp-killall.sh`, and `dcp-uninstall.sh`, for example:

```bash
$ sudo kubectl get nodes
```

The `DCP_TOKEN` is created at `/var/lib/bhojpur/dcp/server/node-token` on your server.

To install on worker nodes, pass `DCP_URL` along with `DCP_TOKEN` or `DCP_CLUSTER_SECRET`
environment variables. For example:

```bash
$ curl -sfL https://get.bhojpur.net/dcp/install.sh | DCP_URL=https://myserver:6443 DCP_TOKEN=XXX sh -
```

## Manual Download

1. Download `Bhojpur DCP` from latest [release](https://github.com/bhojpur/dcp/releases/latest),
x86_64, arm64, and arm64 are supported.

2. Run the server.

```bash
$ sudo dcp server &
# Kubeconfig is written to /etc/bhojpur/dcp/dcp.yaml
$ sudo dcp kubectl get nodes

# On a different node run the below. NODE_TOKEN comes from
# /var/lib/bhojpur/dcp/server/node-token on your server

$ sudo dcp agent --server https://myserver:6443 --token ${NODE_TOKEN}
```
