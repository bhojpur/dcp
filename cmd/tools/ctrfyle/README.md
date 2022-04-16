# Bhojpur DCP - Container Image Untar

It is a utility library to provide additional functionality for users of go-containerregistry.
Also, it includes a basic command-line app demonstrating use of the library code.

## Command Line Tool

```console
NAME:
   ctrfyle - pulls and unpacks a container image to the local filesystem

USAGE:
   ctrfyle [global options] command [command options] <image> <destination>

VERSION:
   v0.3.1

DESCRIPTION:
   Supports DCP/UKE style repository rewrites, endpoint overrides, and auth configuration.
   Supports optional loading from local image tarballs or layer cache.
   Supports Kubelet credential provider plugins.

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --private-registry value                   Private registry configuration file (default: "/etc/bhojpur/common/registries.yaml")
   --images-dir value                         Images tarball directory
   --cache                                    Enable layer cache when image is not available locally
   --cache-dir value                          Layer cache directory (default: "$XDG_CACHE_HOME/bhojpur/ctrfyle")
   --image-credential-provider-config value   Image credential provider configuration file
   --image-credential-provider-bin-dir value  Image credential provider binary directory
   --debug                                    Enable debug logging
   --help, -h                                 show help
   --version, -v                              print the version
```

## Image Credential Providers

The [kubelet image credential providers](https://kubernetes.io/docs/tasks/kubelet-credential-provider/kubelet-credential-provider/)
are supported. At the time of this writing, none of out-of-tree cloud providers offer standalone binaries.
The `ctrfyle` docker image (available by running `make package-image`) bundles provider plugins at
`/bin/plugins`, with a sample config file at `/etc/config.yaml`.
