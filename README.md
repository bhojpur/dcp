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

You need a `Docker` instance to build from the source code.

```bash
$ mkdir -p build/data && make download && make generate
```
