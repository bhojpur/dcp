module github.com/bhojpur/dcp

go 1.17

require (
	github.com/containerd/containerd v1.6.1
	github.com/containerd/fuse-overlayfs-snapshotter v1.0.4
	github.com/containerd/stargz-snapshotter v0.11.3
	github.com/kubernetes-sigs/cri-tools v1.23.0
	github.com/lib/pq v1.10.4
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.4.0
	go.etcd.io/etcd/api/v3 v3.5.2
	go.etcd.io/etcd/server/v3 v3.5.1
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.28.0
	k8s.io/cri-api v0.24.0-alpha.3
	k8s.io/kube-controller-manager v0.20.6
	k8s.io/kubelet v0.20.6
	k8s.io/kubernetes v1.21.0
	k8s.io/system-validators v1.7.0
	sigs.k8s.io/apiserver-network-proxy v0.0.4
)

require (
	github.com/Azure/azure-sdk-for-go v43.0.0+incompatible // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.5 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/mocks v0.4.1 // indirect
	github.com/Azure/go-autorest/autorest/to v0.2.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.1.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.0 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/GoogleCloudPlatform/k8s-cloud-provider v0.0.0-20200415212048-7901bc822317 // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/armon/circbuf v0.0.0-20150827004946-bbbad097214e // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/aws/aws-sdk-go v1.42.27 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/checkpoint-restore/go-criu/v5 v5.3.0 // indirect
	github.com/clusterhq/flocker-go v0.0.0-20160920122132-2b8b7259d313 // indirect
	github.com/container-storage-interface/spec v1.2.0 // indirect
	github.com/containerd/btrfs v1.0.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/go-cni v1.0.2 // indirect
	github.com/containerd/go-runc v1.0.0 // indirect
	github.com/containerd/imgcrypt v1.1.1 // indirect
	github.com/containerd/nri v0.1.0 // indirect
	github.com/containers/ocicrypt v1.1.1 // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/go-ozzo/ozzo-validation v3.5.0+incompatible // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hanwen/go-fuse/v2 v2.1.1-0.20220112183258-f57e95bda82d // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0 // indirect
	github.com/heketi/heketi v9.0.1-0.20190917153846-c2e2a4ab7ab9+incompatible // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/libopenstorage/openstorage v1.0.0 // indirect
	github.com/miekg/dns v1.1.41 // indirect
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/mindprince/gonvml v0.0.0-20190828220739-9ebdce4bb989 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/moby/ipvs v1.0.1 // indirect
	github.com/moby/sys/symlink v0.1.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mrunalp/fileutils v0.5.0 // indirect
	github.com/onsi/gomega v1.18.1 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20171018203845-0dec1b30a021 // indirect
	github.com/quobyte/api v0.1.8 // indirect
	github.com/robfig/cron v1.1.0 // indirect
	github.com/rubiojr/go-vhd v0.0.0-20200706105327-02e210299021 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/seccomp/libseccomp-golang v0.9.2-0.20210429002308-3879420cc921 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/stefanberger/go-pkcs11uri v0.0.0-20201008174630-78d3cae3a980 // indirect
	github.com/storageos/go-api v2.2.0+incompatible // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/tchap/go-patricia v2.2.6+incompatible // indirect
	github.com/thecodeteam/goscaleio v0.1.0 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/vmware/govmomi v0.20.3 // indirect
	go.mozilla.org/pkcs7 v0.0.0-20200128120323-432b2356ecb1 // indirect
	go.opentelemetry.io/otel/exporters/otlp v0.20.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.27.0 // indirect
	go.opentelemetry.io/otel/metric v0.27.0 // indirect
	go.opentelemetry.io/otel/sdk v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk/export/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.20.0 // indirect
	go.opentelemetry.io/proto/otlp v0.7.0 // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	gonum.org/v1/gonum v0.6.2 // indirect
	google.golang.org/api v0.62.0 // indirect
	gopkg.in/gcfg.v1 v1.2.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/warnings.v0 v0.1.1 // indirect
	k8s.io/csi-translation-lib v0.20.6 // indirect
	k8s.io/heapster v1.2.0-beta.1 // indirect
	k8s.io/kube-aggregator v0.18.0 // indirect
	k8s.io/kube-proxy v0.0.0 // indirect
	k8s.io/kube-scheduler v0.0.0 // indirect
	k8s.io/legacy-cloud-providers v0.0.0 // indirect
	k8s.io/mount-utils v0.0.0 // indirect
)

require (
	cloud.google.com/go v0.99.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.2
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/Rican7/retry v0.3.1
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/bronze1man/goStrongswanVici v0.0.0-20210615051318-35664904df04 // indirect
	github.com/canonical/go-dqlite v1.10.2
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chai2010/gettext-go v0.0.0-20160711120539-c6fed771bfd5 // indirect
	github.com/cilium/ebpf v0.8.1 // indirect
	github.com/cloudnativelabs/kube-router v1.4.0
	github.com/containerd/cgroups v1.0.3
	github.com/containerd/continuity v0.2.2 // indirect
	github.com/containerd/fifo v1.0.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.11.3 // indirect
	github.com/containerd/ttrpc v1.1.0 // indirect
	github.com/containerd/typeurl v1.0.2 // indirect
	github.com/containernetworking/cni v1.0.1 // indirect
	github.com/containernetworking/plugins v1.1.1 // indirect
	github.com/coreos/etcd v3.3.27+incompatible // indirect
	github.com/coreos/go-iptables v0.6.0 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/daviddengcn/go-colortext v1.0.0
	github.com/docker/cli v20.10.13+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.13+incompatible
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/spdystream v0.1.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/emicklei/go-restful v2.15.0+incompatible // indirect
	github.com/erikdubbelboer/gspt v0.0.0-20210805194459-ce36a5128377
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/flannel-io/flannel v0.17.0
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/fvbommel/sortorder v1.0.2 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/cadvisor v0.44.0
	github.com/google/gnostic v0.6.7 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/go-containerregistry v0.8.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/renameio v1.0.1 // indirect
	github.com/google/uuid v1.3.0
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.1
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-sqlite3 v1.14.12
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.23
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/sys/mountinfo v0.6.0 // indirect
	github.com/moby/sys/signal v0.7.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.0
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417 // indirect
	github.com/opencontainers/selinux v1.10.0
	github.com/otiai10/copy v1.7.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rancher/dynamiclistener v0.3.1
	github.com/rancher/lasso v0.0.0-20220303224901-7dc33fe27b1a
	github.com/rancher/remotedialer v0.2.5
	github.com/rancher/wharfie v0.5.2
	github.com/rancher/wrangler v0.8.10
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rootless-containers/rootlesskit v0.14.6
	github.com/rs/xid v1.3.0 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/safchain/ethtool v0.2.0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.1 // indirect
	github.com/urfave/cli v1.22.5
	github.com/vbatts/tar-split v0.11.2 // indirect
	github.com/vishvananda/netlink v1.1.1-0.20210330154013-f5de75959ad5
	github.com/vishvananda/netns v0.0.0-20211101163701-50045581ed74 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.etcd.io/etcd v3.3.27+incompatible
	go.etcd.io/etcd/client/pkg/v3 v3.5.2
	go.etcd.io/etcd/client/v2 v2.305.2 // indirect
	go.etcd.io/etcd/client/v3 v3.5.2
	go.etcd.io/etcd/etcdutl/v3 v3.5.2
	go.etcd.io/etcd/pkg/v3 v3.5.2 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.2 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.30.0 // indirect
	go.opentelemetry.io/otel v1.5.0
	go.opentelemetry.io/otel/trace v1.5.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20220321153916-2c7772ba3064
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220323144105-ec3c684e5b14 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	inet.af/tcpproxy v0.0.0-20210824174053-2e577fef49e2
	k8s.io/api v0.23.4
	k8s.io/apiextensions-apiserver v0.23.4 // indirect
	k8s.io/apimachinery v0.23.4
	k8s.io/apiserver v0.20.6
	k8s.io/cli-runtime v0.20.6 // indirect
	k8s.io/client-go v1.5.2
	k8s.io/cloud-provider v0.20.6
	k8s.io/cluster-bootstrap v0.20.6
	k8s.io/code-generator v0.20.6 // indirect
	k8s.io/component-base v0.23.0
	k8s.io/component-helpers v0.20.6 // indirect
	k8s.io/controller-manager v0.20.6
	k8s.io/gengo v0.0.0-20220307231824-4627b89bbf1b // indirect
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.60.1
	k8s.io/kube-openapi v0.0.0-20220323210520-29d726468e05 // indirect
	k8s.io/kubectl v0.22.0
	k8s.io/metrics v0.20.6 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.30 // indirect
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/kustomize v2.0.3+incompatible // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.8.22
	github.com/benmoss/go-powershell => github.com/bhojpur/go-powershell v0.0.0-20201118222746-51f4c451fbd7
	github.com/containerd/containerd => github.com/bhojpur/containerd v1.5.10 // dcp-release/1.5
	github.com/coreos/go-systemd => github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/docker/distribution => github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker => github.com/docker/docker v20.10.7+incompatible
	github.com/docker/libnetwork => github.com/docker/libnetwork v0.8.0-dev.2.0.20190624125649-f0e46a78ea34
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.2
	github.com/google/cadvisor => github.com/bhojpur/cadvisor v0.43.0
	github.com/googleapis/gax-go/v2 => github.com/googleapis/gax-go/v2 v2.0.5
	github.com/juju/errors => github.com/bhojpur/nocode v0.0.0-20200630202308-cb097102c09f
	github.com/kubernetes-sigs/cri-tools => github.com/bhojpur/cri-tools v1.22.0
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.0
	github.com/opencontainers/runtime-spec => github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/rancher/wrangler => github.com/rancher/wrangler v0.8.11-0.20220211163748-d5a8ee98be5f
	go.etcd.io/etcd/api/v3 => github.com/bhojpur/etcd/api/v3 v3.5.1
	go.etcd.io/etcd/client/v3 => github.com/bhojpur/etcd/client/v3 v3.5.1
	go.etcd.io/etcd/etcdutl/v3 => github.com/bhojpur/etcd/etcdutl/v3 v3.5.1
	go.etcd.io/etcd/server/v3 => github.com/bhojpur/etcd/server/v3 v3.5.1
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/net => golang.org/x/net v0.0.0-20210825183410-e898025ed96a
	golang.org/x/sys => golang.org/x/sys v0.0.0-20210831042530-f4d43177bf5e
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20210831024726-fe130286e0e2
	google.golang.org/grpc => google.golang.org/grpc v1.40.0
	gopkg.in/square/go-jose.v2 => gopkg.in/square/go-jose.v2 v2.2.2
	k8s.io/api => k8s.io/api v0.20.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.7-rc.0
	k8s.io/apiserver => k8s.io/apiserver v0.20.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.6
	k8s.io/client-go => k8s.io/client-go v0.20.6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.6
	k8s.io/code-generator => k8s.io/code-generator v0.20.11-rc.0
	k8s.io/component-base => k8s.io/component-base v0.20.6
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.6
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.6
	k8s.io/cri-api => k8s.io/cri-api v0.20.12-rc.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.6
	k8s.io/klog => github.com/bhojpur/klog v1.0.0 // dcp-release-1.x
	k8s.io/klog/v2 => github.com/bhojpur/klog/v2 v2.30.0 // dcp-main
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.6
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.6
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.6
	k8s.io/kubectl => k8s.io/kubectl v0.20.6
	k8s.io/kubelet => k8s.io/kubelet v0.20.6
	k8s.io/kubernetes => github.com/bhojpur/kubernetes v1.20.6
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.6
	k8s.io/metrics => k8s.io/metrics v0.20.6
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.7-rc.0
	k8s.io/node-api => github.com/bhojpur/kubernetes/staging/src/k8s.io/node-api v1.20.6
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.20.6
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.6
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.6
	k8s.io/sample-controller => k8s.io/sample-controller v0.20.6
	mvdan.cc/unparam => mvdan.cc/unparam v0.0.0-20210104141923-aac4ce9116a7
)
