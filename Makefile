.PHONY: clean all release build

all: test build

# Build binaries in the host environment
build:
	bash hack/make-rules/build.sh $(WHAT)

# generate yaml files
gen-yaml:
	hack/make-rules/genyaml.sh $(WHAT)

# Run test
test: fmt vet
	go test -v -short ./pkg/... ./cmd/... -coverprofile cover.out
	go test -v  -coverpkg=./pkg/tunnel/...  -coverprofile=tunnel-cover.out ./test/integration/tunnel_test.go

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Build binaries and docker images.
# NOTE: this rule can take time, as we build binaries inside containers
#
# ARGS:
#   WHAT: list of components that will be compiled.
#   ARCH: list of target architectures.
#   REGION: in which region this rule is executed, if in mainland China,
#   	set it as cn.
#
# Examples:
#   # compile engine, controller-manager and client-servant with
#   # architectures arm64 and arm in the mainland China
#   make release WHAT="engine controller-manager client-servant" ARCH="arm64 arm" REGION=cn
#
#   # compile all components with all architectures (i.e., amd64, arm64, arm)
#   make relase
release:
	bash hack/make-rules/release-images.sh

# push generated images during 'make release'
push:
	bash hack/make-rules/push-images.sh

clean:
	-rm -Rf _output
	-rm -Rf dockerbuild

e2e:
	hack/make-rules/build-e2e.sh

e2e-tests:
	bash hack/run-e2e-tests.sh

# create multi-arch manifest
manifest:
	bash hack/make-rules/release-manifest.sh

# push generated manifest during 'make manifest'
push_manifest:
	bash hack/make-rules/push-manifest.sh

install-golint: ## check golint if not exist install golint tools
ifeq (, $(shell which golangci-lint))
	@{ \
	set -e ;\
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.31.0 ;\
	}
GOLINT_BIN=$(shell go env GOPATH)/bin/golangci-lint
else
GOLINT_BIN=$(shell which golangci-lint)
endif

lint: install-golint ## Run go lint against code.
	$(GOLINT_BIN) run -v