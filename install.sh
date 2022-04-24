#!/bin/sh

# Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
#
# Bhojpur DCP installation software

set -e
set -o noglob

# Usage:
#   curl ... | ENV_VAR=... sh -
#       or
#   ENV_VAR=... ./install.sh
#
# Example:
#   Installing a server without traefik:
#     curl ... | INSTALL_DCP_EXEC="--disable=traefik" sh -
#   Installing an agent to point at a server:
#     curl ... | DCP_TOKEN=xxx DCP_URL=https://server-url:6443 sh -
#
# Environment variables:
#   - DCP_*
#     Environment variables which begin with DCP_ will be preserved for the
#     systemd service to use. Setting DCP_URL without explicitly setting
#     a systemd exec command will default the command to "agent", and we
#     enforce that DCP_TOKEN or DCP_CLUSTER_SECRET is also set.
#
#   - INSTALL_DCP_SKIP_DOWNLOAD
#     If set to true will not download Bhojpur DCP hash or binary.
#
#   - INSTALL_DCP_FORCE_RESTART
#     If set to true will always restart the Bhojpur DCP service
#
#   - INSTALL_DCP_SYMLINK
#     If set to 'skip' will not create symlinks, 'force' will overwrite,
#     default will symlink if command does not exist in path.
#
#   - INSTALL_DCP_SKIP_ENABLE
#     If set to true will not enable or start Bhojpur DCP service.
#
#   - INSTALL_DCP_SKIP_START
#     If set to true will not start Bhojpur DCP service.
#
#   - INSTALL_DCP_VERSION
#     Version of Bhojpur DCP to download from github. Will attempt to download from the
#     stable channel if not specified.
#
#   - INSTALL_DCP_COMMIT
#     Commit of Bhojpur DCP to download from temporary cloud storage.
#     * (for developer & QA use)
#
#   - INSTALL_DCP_BIN_DIR
#     Directory to install Bhojpur DCP binary, links, and uninstall script to, or use
#     /usr/local/bin as the default
#
#   - INSTALL_DCP_BIN_DIR_READ_ONLY
#     If set to true will not write files to INSTALL_DCP_BIN_DIR, forces
#     setting INSTALL_DCP_SKIP_DOWNLOAD=true
#
#   - INSTALL_DCP_SYSTEMD_DIR
#     Directory to install systemd service and environment files to, or use
#     /etc/systemd/system as the default
#
#   - INSTALL_DCP_EXEC or script arguments
#     Command with flags to use for launching Bhojpur DCP in the systemd service, if
#     the command is not specified will default to "agent" if DCP_URL is set
#     or "server" if not. The final systemd command resolves to a combination
#     of EXEC and script args ($@).
#
#     The following commands result in the same behavior:
#       curl ... | INSTALL_DCP_EXEC="--disable=traefik" sh -s -
#       curl ... | INSTALL_DCP_EXEC="server --disable=traefik" sh -s -
#       curl ... | INSTALL_DCP_EXEC="server" sh -s - --disable=traefik
#       curl ... | sh -s - server --disable=traefik
#       curl ... | sh -s - --disable=traefik
#
#   - INSTALL_DCP_NAME
#     Name of systemd service to create, will default from the Bhojpur DCP exec command
#     if not specified. If specified the name will be prefixed with 'DCP-'.
#
#   - INSTALL_DCP_TYPE
#     Type of systemd service to create, will default from the Bhojpur DCP exec command
#     if not specified.
#
#   - INSTALL_DCP_SELINUX_WARN
#     If set to true will continue if dcp-selinux policy is not found.
#
#   - INSTALL_DCP_SKIP_SELINUX_RPM
#     If set to true will skip automatic installation of the Bhojpur DCP RPM.
#
#   - INSTALL_DCP_CHANNEL_URL
#     Channel URL for fetching Bhojpur DCP download URL.
#     Defaults to 'https://update.bhojpur.net/v1-release/channels'.
#
#   - INSTALL_DCP_CHANNEL
#     Channel to use for fetching Bhojpur DCP download URL.
#     Defaults to 'stable'.

GITHUB_URL=https://github.com/bhojpur/dcp/releases
STORAGE_URL=https://storage.googleapis.com/bhojpur-net-platform
DOWNLOADER=

# --- helper functions for logs ---
info()
{
    echo '[INFO] ' "$@"
}
warn()
{
    echo '[WARN] ' "$@" >&2
}
fatal()
{
    echo '[ERROR] ' "$@" >&2
    exit 1
}

# --- fatal if no systemd or openrc ---
verify_system() {
    if [ -x /sbin/openrc-run ]; then
        HAS_OPENRC=true
        return
    fi
    if [ -x /bin/systemctl ] || type systemctl > /dev/null 2>&1; then
        HAS_SYSTEMD=true
        return
    fi
    fatal 'Can not find systemd or openrc to use as a process supervisor for Bhojpur DCP'
}

# --- add quotes to command arguments ---
quote() {
    for arg in "$@"; do
        printf '%s\n' "$arg" | sed "s/'/'\\\\''/g;1s/^/'/;\$s/\$/'/"
    done
}

# --- add indentation and trailing slash to quoted args ---
quote_indent() {
    printf ' \\\n'
    for arg in "$@"; do
        printf '\t%s \\\n' "$(quote "$arg")"
    done
}

# --- escape most punctuation characters, except quotes, forward slash, and space ---
escape() {
    printf '%s' "$@" | sed -e 's/\([][!#$%&()*;<=>?\_`{|}]\)/\\\1/g;'
}

# --- escape double quotes ---
escape_dq() {
    printf '%s' "$@" | sed -e 's/"/\\"/g'
}

# --- ensures $DCP_URL is empty or begins with https://, exiting fatally otherwise ---
verify_dcp_url() {
    case "${DCP_URL}" in
        "")
            ;;
        https://*)
            ;;
        *)
            fatal "Only https:// URLs are supported for Bhojpur DCP_URL (have ${DCP_URL})"
            ;;
    esac
}

# --- define needed environment variables ---
setup_env() {
    # --- use command args if passed or create default ---
    case "$1" in
        # --- if we only have flags discover if command should be server or agent ---
        (-*|"")
            if [ -z "${DCP_URL}" ]; then
                CMD_DCP=server
            else
                if [ -z "${DCP_TOKEN}" ] && [ -z "${DCP_TOKEN_FILE}" ] && [ -z "${DCP_CLUSTER_SECRET}" ]; then
                    fatal "Defaulted Bhojpur DCP exec command to 'agent' because DCP_URL is defined, but DCP_TOKEN, DCP_TOKEN_FILE or DCP_CLUSTER_SECRET is not defined."
                fi
                CMD_DCP=agent
            fi
        ;;
        # --- command is provided ---
        (*)
            CMD_DCP=$1
            shift
        ;;
    esac

    verify_dcp_url

    CMD_DCP_EXEC="${CMD_DCP}$(quote_indent "$@")"

    # --- use systemd name if defined or create default ---
    if [ -n "${INSTALL_DCP_NAME}" ]; then
        SYSTEM_NAME=dcp-${INSTALL_DCP_NAME}
    else
        if [ "${CMD_DCP}" = server ]; then
            SYSTEM_NAME=dcp
        else
            SYSTEM_NAME=dcp-${CMD_DCP}
        fi
    fi

    # --- check for invalid characters in system name ---
    valid_chars=$(printf '%s' "${SYSTEM_NAME}" | sed -e 's/[][!#$%&()*;<=>?\_`{|}/[:space:]]/^/g;' )
    if [ "${SYSTEM_NAME}" != "${valid_chars}"  ]; then
        invalid_chars=$(printf '%s' "${valid_chars}" | sed -e 's/[^^]/ /g')
        fatal "Invalid characters for system name:
            ${SYSTEM_NAME}
            ${invalid_chars}"
    fi

    # --- use sudo if we are not already root ---
    SUDO=sudo
    if [ $(id -u) -eq 0 ]; then
        SUDO=
    fi

    # --- use systemd type if defined or create default ---
    if [ -n "${INSTALL_DCP_TYPE}" ]; then
        SYSTEMD_TYPE=${INSTALL_DCP_TYPE}
    else
        if [ "${CMD_DCP}" = server ]; then
            SYSTEMD_TYPE=notify
        else
            SYSTEMD_TYPE=exec
        fi
    fi

    # --- use binary install directory if defined or create default ---
    if [ -n "${INSTALL_DCP_BIN_DIR}" ]; then
        BIN_DIR=${INSTALL_DCP_BIN_DIR}
    else
        # --- use /usr/local/bin if root can write to it, otherwise use /opt/bin if it exists
        BIN_DIR=/usr/local/bin
        if ! $SUDO sh -c "touch ${BIN_DIR}/dcp-ro-test && rm -rf ${BIN_DIR}/dcp-ro-test"; then
            if [ -d /opt/bin ]; then
                BIN_DIR=/opt/bin
            fi
        fi
    fi

    # --- use systemd directory if defined or create default ---
    if [ -n "${INSTALL_DCP_SYSTEMD_DIR}" ]; then
        SYSTEMD_DIR="${INSTALL_DCP_SYSTEMD_DIR}"
    else
        SYSTEMD_DIR=/etc/systemd/system
    fi

    # --- set related files from system name ---
    SERVICE_DCP=${SYSTEM_NAME}.service
    UNINSTALL_DCP_SH=${UNINSTALL_DCP_SH:-${BIN_DIR}/${SYSTEM_NAME}-uninstall.sh}
    KILLALL_DCP_SH=${KILLALL_DCP_SH:-${BIN_DIR}/dcp-killall.sh}

    # --- use service or environment location depending on systemd/openrc ---
    if [ "${HAS_SYSTEMD}" = true ]; then
        FILE_DCP_SERVICE=${SYSTEMD_DIR}/${SERVICE_DCP}
        FILE_DCP_ENV=${SYSTEMD_DIR}/${SERVICE_DCP}.env
    elif [ "${HAS_OPENRC}" = true ]; then
        $SUDO mkdir -p /etc/bhojpur/dcp
        FILE_DCP_SERVICE=/etc/init.d/${SYSTEM_NAME}
        FILE_DCP_ENV=/etc/bhojpur/dcp/${SYSTEM_NAME}.env
    fi

    # --- get hash of config & exec for currently installed Bhojpur DCP ---
    PRE_INSTALL_HASHES=$(get_installed_hashes)

    # --- if bin directory is read only skip download ---
    if [ "${INSTALL_DCP_BIN_DIR_READ_ONLY}" = true ]; then
        INSTALL_DCP_SKIP_DOWNLOAD=true
    fi

    # --- setup channel values
    INSTALL_DCP_CHANNEL_URL=${INSTALL_DCP_CHANNEL_URL:-'https://update.bhojpur.net/v1-release/channels'}
    INSTALL_DCP_CHANNEL=${INSTALL_DCP_CHANNEL:-'stable'}
}

# --- check if skip download environment variable set ---
can_skip_download() {
    if [ "${INSTALL_DCP_SKIP_DOWNLOAD}" != true ]; then
        return 1
    fi
}

# --- verify an executable Bhojpur DCP binary is installed ---
verify_dcp_is_executable() {
    if [ ! -x ${BIN_DIR}/dcp ]; then
        fatal "Executable Bhojpur DCP binary not found at ${BIN_DIR}/dcp"
    fi
}

# --- set arch and suffix, fatal if architecture not supported ---
setup_verify_arch() {
    if [ -z "$ARCH" ]; then
        ARCH=$(uname -m)
    fi
    case $ARCH in
        amd64)
            ARCH=amd64
            SUFFIX=
            ;;
        x86_64)
            ARCH=amd64
            SUFFIX=
            ;;
        arm64)
            ARCH=arm64
            SUFFIX=-${ARCH}
            ;;
        aarch64)
            ARCH=arm64
            SUFFIX=-${ARCH}
            ;;
        arm*)
            ARCH=arm
            SUFFIX=-${ARCH}hf
            ;;
        *)
            fatal "Unsupported architecture $ARCH"
    esac
}

# --- verify existence of network downloader executable ---
verify_downloader() {
    # Return failure if it doesn't exist or is no executable
    [ -x "$(command -v $1)" ] || return 1

    # Set verified executable as our downloader program and return success
    DOWNLOADER=$1
    return 0
}

# --- create temporary directory and cleanup when done ---
setup_tmp() {
    TMP_DIR=$(mktemp -d -t dcp-install.XXXXXXXXXX)
    TMP_HASH=${TMP_DIR}/dcp.hash
    TMP_BIN=${TMP_DIR}/dcp.bin
    cleanup() {
        code=$?
        set +e
        trap - EXIT
        rm -rf ${TMP_DIR}
        exit $code
    }
    trap cleanup INT EXIT
}

# --- use desired Bhojpur DCP version if defined or find version from channel ---
get_release_version() {
    if [ -n "${INSTALL_DCP_COMMIT}" ]; then
        VERSION_DCP="commit ${INSTALL_DCP_COMMIT}"
    elif [ -n "${INSTALL_DCP_VERSION}" ]; then
        VERSION_DCP=${INSTALL_DCP_VERSION}
    else
        info "Finding release for channel ${INSTALL_DCP_CHANNEL}"
        version_url="${INSTALL_DCP_CHANNEL_URL}/${INSTALL_DCP_CHANNEL}"
        case $DOWNLOADER in
            curl)
                VERSION_DCP=$(curl -w '%{url_effective}' -L -s -S ${version_url} -o /dev/null | sed -e 's|.*/||')
                ;;
            wget)
                VERSION_DCP=$(wget -SqO /dev/null ${version_url} 2>&1 | grep -i Location | sed -e 's|.*/||')
                ;;
            *)
                fatal "Incorrect downloader executable '$DOWNLOADER'"
                ;;
        esac
    fi
    info "Using ${VERSION_DCP} as release"
}

# --- download from github url ---
download() {
    [ $# -eq 2 ] || fatal 'download needs exactly 2 arguments'

    case $DOWNLOADER in
        curl)
            curl -o $1 -sfL $2
            ;;
        wget)
            wget -qO $1 $2
            ;;
        *)
            fatal "Incorrect executable '$DOWNLOADER'"
            ;;
    esac

    # Abort if download command failed
    [ $? -eq 0 ] || fatal 'Download failed'
}

# --- download hash from GitHub URL ---
download_hash() {
    if [ -n "${INSTALL_DCP_COMMIT}" ]; then
        HASH_URL=${STORAGE_URL}/dcp${SUFFIX}-${INSTALL_DCP_COMMIT}.sha256sum
    else
        # HASH_URL=${GITHUB_URL}/download/${VERSION_DCP}/sha256sum-${ARCH}.txt
        HASH_URL=${GITHUB_URL}/download/${VERSION_DCP}/checksums.txt
    fi
    info "Downloading Bhojpur DCP release hash from ${HASH_URL}"
    download ${TMP_HASH} ${HASH_URL}
    HASH_EXPECTED=$(grep " dcp${SUFFIX}$" ${TMP_HASH})
    HASH_EXPECTED=${HASH_EXPECTED%%[[:blank:]]*}
}

# --- check hash against installed version ---
installed_hash_matches() {
    if [ -x ${BIN_DIR}/dcp ]; then
        HASH_INSTALLED=$(sha256sum ${BIN_DIR}/dcp)
        HASH_INSTALLED=${HASH_INSTALLED%%[[:blank:]]*}
        if [ "${HASH_EXPECTED}" = "${HASH_INSTALLED}" ]; then
            return
        fi
    fi
    return 1
}

# --- download binary from github url ---
download_binary() {
    if [ -n "${INSTALL_DCP_COMMIT}" ]; then
        BIN_URL=${STORAGE_URL}/dcp${SUFFIX}-${INSTALL_DCP_COMMIT}
    else
        BIN_URL=${GITHUB_URL}/download/${VERSION_DCP}/dcp${SUFFIX}
    fi
    info "Downloading binary ${BIN_URL}"
    download ${TMP_BIN} ${BIN_URL}
}

# --- verify downloaded binary hash ---
verify_binary() {
    info "Verifying binary download"
    HASH_BIN=$(sha256sum ${TMP_BIN})
    HASH_BIN=${HASH_BIN%%[[:blank:]]*}
    if [ "${HASH_EXPECTED}" != "${HASH_BIN}" ]; then
        fatal "Download sha256 does not match ${HASH_EXPECTED}, got ${HASH_BIN}"
    fi
}

# --- setup permissions and move binary to system directory ---
setup_binary() {
    chmod 755 ${TMP_BIN}
    info "Installing Bhojpur DCP to ${BIN_DIR}/dcp"
    $SUDO chown root:root ${TMP_BIN}
    $SUDO mv -f ${TMP_BIN} ${BIN_DIR}/dcp
}

# --- setup selinux policy ---
setup_selinux() {
    case ${INSTALL_DCP_CHANNEL} in 
        *testing)
            rpm_channel=testing
            ;;
        *latest)
            rpm_channel=latest
            ;;
        *)
            rpm_channel=stable
            ;;
    esac

    rpm_site="rpm.bhojpur.net"
    if [ "${rpm_channel}" = "testing" ]; then
        rpm_site="rpm-testing.bhojpur.net"
    fi

    [ -r /etc/os-release ] && . /etc/os-release
    if [ "${ID_LIKE%%[ ]*}" = "suse" ]; then
        rpm_target=sle
        rpm_site_infix=microos
        package_installer=zypper
    elif [ "${VERSION_ID%%.*}" = "7" ]; then
        rpm_target=el7
        rpm_site_infix=centos/7
        package_installer=yum
    else
        rpm_target=el8
        rpm_site_infix=centos/8
        package_installer=yum
    fi

    if [ "${package_installer}" = "yum" ] && [ -x /usr/bin/dnf ]; then
        package_installer=dnf
    fi

    policy_hint="please install:
    ${package_installer} install -y container-selinux
    ${package_installer} install -y https://${rpm_site}/dcp/${rpm_channel}/common/${rpm_site_infix}/noarch/dcp-selinux-0.4-1.${rpm_target}.noarch.rpm
"

    if [ "$INSTALL_DCP_SKIP_SELINUX_RPM" = true ] || can_skip_download || [ ! -d /usr/share/selinux ]; then
        info "Skipping installation of SELinux RPM"
    elif  [ "${ID_LIKE:-}" != coreos ] && [ "${VARIANT_ID:-}" != coreos ]; then
        install_selinux_rpm ${rpm_site} ${rpm_channel} ${rpm_target} ${rpm_site_infix}
    fi

    policy_error=fatal
    if [ "$INSTALL_DCP_SELINUX_WARN" = true ] || [ "${ID_LIKE:-}" = coreos ] || [ "${VARIANT_ID:-}" = coreos ]; then
        policy_error=warn
    fi

    if ! $SUDO chcon -u system_u -r object_r -t container_runtime_exec_t ${BIN_DIR}/dcp >/dev/null 2>&1; then
        if $SUDO grep '^\s*SELINUX=enforcing' /etc/selinux/config >/dev/null 2>&1; then
            $policy_error "Failed to apply container_runtime_exec_t to ${BIN_DIR}/dcp, ${policy_hint}"
        fi
    elif [ ! -f /usr/share/selinux/packages/dcp.pp ]; then
        if [ -x /usr/sbin/transactional-update ]; then
            warn "Please reboot your machine to activate the changes and avoid data loss."
        else
            $policy_error "Failed to find the dcp-selinux policy, ${policy_hint}"
        fi
    fi
}

install_selinux_rpm() {
    if [ -r /etc/redhat-release ] || [ -r /etc/centos-release ] || [ -r /etc/oracle-release ] || [ "${ID_LIKE%%[ ]*}" = "suse" ]; then
        repodir=/etc/yum.repos.d
        if [ -d /etc/zypp/repos.d ]; then
            repodir=/etc/zypp/repos.d
        fi
        set +o noglob
        $SUDO rm -f ${repodir}/bhojpur-dcp-common*.repo
        set -o noglob
        if [ -r /etc/redhat-release ] && [ "${3}" = "el7" ]; then
            $SUDO yum install -y yum-utils
            $SUDO yum-config-manager --enable rhel-7-server-extras-rpms
        fi
        $SUDO tee ${repodir}/bhojpur-dcp-common.repo >/dev/null << EOF
[bhojpur-dcp-common-${2}]
name=Bhojpur DCP Common (${2})
baseurl=https://${1}/dcp/${2}/common/${4}/noarch
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://${1}/public.key
EOF
        case ${3} in
        sle)
            rpm_installer="zypper --gpg-auto-import-keys"
            if [ "${TRANSACTIONAL_UPDATE=false}" != "true" ] && [ -x /usr/sbin/transactional-update ]; then
                rpm_installer="transactional-update --no-selfupdate -d run ${rpm_installer}"
                : "${INSTALL_DCP_SKIP_START:=true}"
            fi
            ;;
        *)
            rpm_installer="yum"
            ;;
        esac
        if [ "${rpm_installer}" = "yum" ] && [ -x /usr/bin/dnf ]; then
            rpm_installer=dnf
        fi
        # shellcheck disable=SC2086
        $SUDO ${rpm_installer} install -y "dcp-selinux"
    fi
    return
}

# --- download and verify Bhojpur DCP ---
download_and_verify() {
    if can_skip_download; then
       info 'Skipping Bhojpur DCP download and verify'
       verify_dcp_is_executable
       return
    fi

    setup_verify_arch
    verify_downloader curl || verify_downloader wget || fatal 'Can not find curl or wget for downloading files'
    setup_tmp
    get_release_version
    download_hash

    if installed_hash_matches; then
        info 'Skipping binary downloaded, installed Bhojpur DCP matches hash'
        return
    fi

    download_binary
    verify_binary
    setup_binary
}

# --- add additional utility links ---
create_symlinks() {
    [ "${INSTALL_DCP_BIN_DIR_READ_ONLY}" = true ] && return
    [ "${INSTALL_DCP_SYMLINK}" = skip ] && return

    for cmd in kubectl crictl ctr; do
        if [ ! -e ${BIN_DIR}/${cmd} ] || [ "${INSTALL_DCP_SYMLINK}" = force ]; then
            which_cmd=$(command -v ${cmd} 2>/dev/null || true)
            if [ -z "${which_cmd}" ] || [ "${INSTALL_DCP_SYMLINK}" = force ]; then
                info "Creating ${BIN_DIR}/${cmd} symlink to Bhojpur DCP"
                $SUDO ln -sf dcp ${BIN_DIR}/${cmd}
            else
                info "Skipping ${BIN_DIR}/${cmd} symlink to Bhojpur DCP, command exists in PATH at ${which_cmd}"
            fi
        else
            info "Skipping ${BIN_DIR}/${cmd} symlink to Bhojpur DCP, already exists"
        fi
    done
}

# --- create killall script ---
create_killall() {
    [ "${INSTALL_DCP_BIN_DIR_READ_ONLY}" = true ] && return
    info "Creating killall script ${KILLALL_DCP_SH}"
    $SUDO tee ${KILLALL_DCP_SH} >/dev/null << \EOF
#!/bin/sh
[ $(id -u) -eq 0 ] || exec sudo $0 $@

for bin in /var/lib/bhojpur/dcp/data/**/bin/; do
    [ -d $bin ] && export PATH=$PATH:$bin:$bin/aux
done

set -x

for service in /etc/systemd/system/dcp*.service; do
    [ -s $service ] && systemctl stop $(basename $service)
done

for service in /etc/init.d/dcp*; do
    [ -x $service ] && $service stop
done

pschildren() {
    ps -e -o ppid= -o pid= | \
    sed -e 's/^\s*//g; s/\s\s*/\t/g;' | \
    grep -w "^$1" | \
    cut -f2
}

pstree() {
    for pid in $@; do
        echo $pid
        for child in $(pschildren $pid); do
            pstree $child
        done
    done
}

killtree() {
    kill -9 $(
        { set +x; } 2>/dev/null;
        pstree $@;
        set -x;
    ) 2>/dev/null
}

getshims() {
    ps -e -o pid= -o args= | sed -e 's/^ *//; s/\s\s*/\t/;' | grep -w 'dcp/data/[^/]*/bin/containerd-shim' | cut -f1
}

killtree $({ set +x; } 2>/dev/null; getshims; set -x)

do_unmount_and_remove() {
    set +x
    while read -r _ path _; do
        case "$path" in $1*) echo "$path" ;; esac
    done < /proc/self/mounts | sort -r | xargs -r -t -n 1 sh -c 'umount "$0" && rm -rf "$0"'
    set -x
}

do_unmount_and_remove '/run/dcp'
do_unmount_and_remove '/var/lib/bhojpur/dcp'
do_unmount_and_remove '/var/lib/kubelet/pods'
do_unmount_and_remove '/var/lib/kubelet/plugins'
do_unmount_and_remove '/run/netns/cni-'

# Remove CNI namespaces
ip netns show 2>/dev/null | grep cni- | xargs -r -t -n 1 ip netns delete

# Delete network interface(s) that match 'master cni0'
ip link show 2>/dev/null | grep 'master cni0' | while read ignore iface ignore; do
    iface=${iface%%@*}
    [ -z "$iface" ] || ip link delete $iface
done
ip link delete cni0
ip link delete flannel.1
ip link delete flannel-v6.1
rm -rf /var/lib/cni/
iptables-save | grep -v KUBE- | grep -v CNI- | grep -v flannel | iptables-restore
ip6tables-save | grep -v KUBE- | grep -v CNI- | grep -v flannel | ip6tables-restore
EOF
    $SUDO chmod 755 ${KILLALL_DCP_SH}
    $SUDO chown root:root ${KILLALL_DCP_SH}
}

# --- create uninstall script ---
create_uninstall() {
    [ "${INSTALL_DCP_BIN_DIR_READ_ONLY}" = true ] && return
    info "Creating uninstall script ${UNINSTALL_DCP_SH}"
    $SUDO tee ${UNINSTALL_DCP_SH} >/dev/null << EOF
#!/bin/sh
set -x
[ \$(id -u) -eq 0 ] || exec sudo \$0 \$@

${KILLALL_DCP_SH}

if command -v systemctl; then
    systemctl disable ${SYSTEM_NAME}
    systemctl reset-failed ${SYSTEM_NAME}
    systemctl daemon-reload
fi
if command -v rc-update; then
    rc-update delete ${SYSTEM_NAME} default
fi

rm -f ${FILE_DCP_SERVICE}
rm -f ${FILE_DCP_ENV}

remove_uninstall() {
    rm -f ${UNINSTALL_DCP_SH}
}
trap remove_uninstall EXIT

if (ls ${SYSTEMD_DIR}/dcp*.service || ls /etc/init.d/dcp*) >/dev/null 2>&1; then
    set +x; echo 'Additional Bhojpur DCP services installed, skipping uninstall of Bhojpur DCP'; set -x
    exit
fi

for cmd in kubectl crictl ctr; do
    if [ -L ${BIN_DIR}/\$cmd ]; then
        rm -f ${BIN_DIR}/\$cmd
    fi
done

rm -rf /etc/bhojpur/dcp
rm -rf /run/dcp
rm -rf /run/flannel
rm -rf /var/lib/bhojpur/dcp
rm -rf /var/lib/kubelet
rm -f ${BIN_DIR}/dcp
rm -f ${KILLALL_DCP_SH}

if type yum >/dev/null 2>&1; then
    yum remove -y dcp-selinux
    rm -f /etc/yum.repos.d/bhojpur-dcp-common*.repo
elif type zypper >/dev/null 2>&1; then
    uninstall_cmd="zypper remove -y dcp-selinux"
    if [ "\${TRANSACTIONAL_UPDATE=false}" != "true" ] && [ -x /usr/sbin/transactional-update ]; then
        uninstall_cmd="transactional-update --no-selfupdate -d run \$uninstall_cmd"
    fi
    \$uninstall_cmd
    rm -f /etc/zypp/repos.d/bhojpur-dcp-common*.repo
fi
EOF
    $SUDO chmod 755 ${UNINSTALL_DCP_SH}
    $SUDO chown root:root ${UNINSTALL_DCP_SH}
}

# --- disable current service if loaded --
systemd_disable() {
    $SUDO systemctl disable ${SYSTEM_NAME} >/dev/null 2>&1 || true
    $SUDO rm -f /etc/systemd/system/${SERVICE_DCP} || true
    $SUDO rm -f /etc/systemd/system/${SERVICE_DCP}.env || true
}

# --- capture current env and create file containing DCP_ variables ---
create_env_file() {
    info "env: Creating environment file ${FILE_DCP_ENV}"
    $SUDO touch ${FILE_DCP_ENV}
    $SUDO chmod 0600 ${FILE_DCP_ENV}
    sh -c export | while read x v; do echo $v; done | grep -E '^(DCP|CONTAINERD)_' | $SUDO tee ${FILE_DCP_ENV} >/dev/null
    sh -c export | while read x v; do echo $v; done | grep -Ei '^(NO|HTTP|HTTPS)_PROXY' | $SUDO tee -a ${FILE_DCP_ENV} >/dev/null
}

# --- write systemd service file ---
create_systemd_service_file() {
    info "systemd: Creating service file ${FILE_DCP_SERVICE}"
    $SUDO tee ${FILE_DCP_SERVICE} >/dev/null << EOF
[Unit]
Description=Bhojpur Kubernetes
Documentation=https://bhojpur.net
Wants=network-online.target
After=network-online.target

[Install]
WantedBy=multi-user.target

[Service]
Type=${SYSTEMD_TYPE}
EnvironmentFile=-/etc/default/%N
EnvironmentFile=-/etc/sysconfig/%N
EnvironmentFile=-${FILE_DCP_ENV}
KillMode=process
Delegate=yes
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
Restart=always
RestartSec=5s
ExecStartPre=/bin/sh -xc '! /usr/bin/systemctl is-enabled --quiet nm-cloud-setup.service'
ExecStartPre=-/sbin/modprobe br_netfilter
ExecStartPre=-/sbin/modprobe overlay
ExecStart=${BIN_DIR}/dcp \\
    ${CMD_DCP_EXEC}

EOF
}

# --- write openrc service file ---
create_openrc_service_file() {
    LOG_FILE=/var/log/${SYSTEM_NAME}.log

    info "openrc: Creating service file ${FILE_DCP_SERVICE}"
    $SUDO tee ${FILE_DCP_SERVICE} >/dev/null << EOF
#!/sbin/openrc-run

depend() {
    after network-online
    want cgroups
}

start_pre() {
    rm -f /tmp/dcp.*
}

supervisor=supervise-daemon
name=${SYSTEM_NAME}
command="${BIN_DIR}/dcp"
command_args="$(escape_dq "${CMD_DCP_EXEC}")
    >>${LOG_FILE} 2>&1"

output_log=${LOG_FILE}
error_log=${LOG_FILE}

pidfile="/var/run/${SYSTEM_NAME}.pid"
respawn_delay=5
respawn_max=0

set -o allexport
if [ -f /etc/environment ]; then source /etc/environment; fi
if [ -f ${FILE_DCP_ENV} ]; then source ${FILE_DCP_ENV}; fi
set +o allexport
EOF
    $SUDO chmod 0755 ${FILE_DCP_SERVICE}

    $SUDO tee /etc/logrotate.d/${SYSTEM_NAME} >/dev/null << EOF
${LOG_FILE} {
	missingok
	notifempty
	copytruncate
}
EOF
}

# --- write systemd or openrc service file ---
create_service_file() {
    [ "${HAS_SYSTEMD}" = true ] && create_systemd_service_file
    [ "${HAS_OPENRC}" = true ] && create_openrc_service_file
    return 0
}

# --- get hashes of the current Bhojpur DCP bin and service files
get_installed_hashes() {
    $SUDO sha256sum ${BIN_DIR}/dcp ${FILE_DCP_SERVICE} ${FILE_DCP_ENV} 2>&1 || true
}

# --- enable and start systemd service ---
systemd_enable() {
    info "systemd: Enabling ${SYSTEM_NAME} unit"
    $SUDO systemctl enable ${FILE_DCP_SERVICE} >/dev/null
    $SUDO systemctl daemon-reload >/dev/null
}

systemd_start() {
    info "systemd: Starting ${SYSTEM_NAME}"
    $SUDO systemctl restart ${SYSTEM_NAME}
}

# --- enable and start openrc service ---
openrc_enable() {
    info "openrc: Enabling ${SYSTEM_NAME} service for default runlevel"
    $SUDO rc-update add ${SYSTEM_NAME} default >/dev/null
}

openrc_start() {
    info "openrc: Starting ${SYSTEM_NAME}"
    $SUDO ${FILE_DCP_SERVICE} restart
}

# --- startup systemd or openrc service ---
service_enable_and_start() {
    if [ -f "/proc/cgroups" ] && [ "$(grep memory /proc/cgroups | while read -r n n n enabled; do echo $enabled; done)" -eq 0 ];
    then
        info 'Failed to find memory cgroup, you may need to add "cgroup_memory=1 cgroup_enable=memory" to your linux cmdline (/boot/cmdline.txt on a Raspberry Pi)'
    fi

    [ "${INSTALL_DCP_SKIP_ENABLE}" = true ] && return

    [ "${HAS_SYSTEMD}" = true ] && systemd_enable
    [ "${HAS_OPENRC}" = true ] && openrc_enable

    [ "${INSTALL_DCP_SKIP_START}" = true ] && return

    POST_INSTALL_HASHES=$(get_installed_hashes)
    if [ "${PRE_INSTALL_HASHES}" = "${POST_INSTALL_HASHES}" ] && [ "${INSTALL_DCP_FORCE_RESTART}" != true ]; then
        info 'No change detected so skipping service start'
        return
    fi

    [ "${HAS_SYSTEMD}" = true ] && systemd_start
    [ "${HAS_OPENRC}" = true ] && openrc_start
    return 0
}

# --- re-evaluate args to include env command ---
eval set -- $(escape "${INSTALL_DCP_EXEC}") $(quote "$@")

# --- run the install process --
{
    verify_system
    setup_env "$@"
    download_and_verify
    setup_selinux
    create_symlinks
    create_killall
    create_uninstall
    systemd_disable
    create_env_file
    create_service_file
    service_enable_and_start
}