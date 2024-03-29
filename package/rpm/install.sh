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

# Usage:
#   curl ... | ENV_VAR=... sh -
#       or
#   ENV_VAR=... ./install.sh
#
# Example:
#   Installing a server without traefik:
#     curl ... | INSTALL_DCP_EXEC="--no-deploy=traefik" sh -
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
#   - INSTALL_DCP_UPGRADE
#     If set to true will not overwrite env file.
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
#     Version of Bhojpur DCP to download from github. Will attempt to download the
#     latest version if not specified.
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
#     Directory to install systemd service files to, or use
#     /etc/systemd/system as the default
#
#   - INSTALL_DCP_CONFIG_DIR
#     Directory to install environment and flag files to, or use
#     /etc/bhojpur/dcp as the default
#
#   - INSTALL_DCP_EXEC or script arguments
#     Command with flags to use for launching Bhojpur DCP in the systemd service, if
#     the command is not specified will default to "agent" if DCP_URL is set
#     or "server" if not. The final systemd command resolves to a combination
#     of EXEC and script args ($@).
#
#     The following commands result in the same behavior:
#       curl ... | INSTALL_DCP_EXEC="--no-deploy=traefik" sh -s -
#       curl ... | INSTALL_DCP_EXEC="server --no-deploy=traefik" sh -s -
#       curl ... | INSTALL_DCP_EXEC="server" sh -s - --no-deploy=traefik
#       curl ... | sh -s - server --no-deploy=traefik
#       curl ... | sh -s - --no-deploy=traefik
#
#   - INSTALL_DCP_NAME
#     Name of systemd service to create, will default from the Bhojpur DCP exec command
#     if not specified. If specified the name will be prefixed with 'dcp-'.
#
#   - INSTALL_DCP_TYPE
#     Type of systemd service to create, will default from the Bhojpur DCP exec command
#     if not specified.
#
#   - INSTALL_DCP_DEBUG
#     Set to true for debug shell scripit output (set -x).
#
#   - KILLALL_DCP_SH
#     Full path of killall script to create.
#
#   - UNINSTALL_DCP_SH
#     Full path of uninstall script to create.

if [ "$INSTALL_DCP_DEBUG" = "true" ]; then
    set -x
fi

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
    if [ -d /run/systemd ]; then
        HAS_SYSTEMD=true
        return
    fi
    fatal 'Cannot find systemd or openrc to use as a process supervisor for Bhojpur DCP'
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
    printf ' \n'
}

# --- escape most punctuation characters, except quotes, forward slash, and space ---
escape() {
    printf '%s' "$@" | sed -e 's/\([][!#$%&()*;<=>?\_`{|}]\)/\\\1/g;'
}

# --- escape double quotes ---
escape_dq() {
    printf '%s' "$@" | sed -e 's/"/\\"/g'
}

# --- define needed environment variables ---
setup_env() {
    case "$1" in
        (--debug)
            export DCP_DEBUG=true
            shift
        ;;
    esac
    # --- use command args if passed or create default ---
    case "$1" in
        # --- if we only have flags discover if command should be server or agent ---
        (-*|"")
            if [ -z "${DCP_URL}" ]; then
                CMD_DCP=server
            else
                if [ -z "${DCP_TOKEN}" ] && [ -z "${DCP_CLUSTER_SECRET}" ]; then
                    fatal "Defaulted Bhojpur DCP exec command to 'agent' because DCP_URL is defined, but DCP_TOKEN or DCP_CLUSTER_SECRET is not defined."
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
    CMD_DCP_ARGS=$(quote "$@")
    CMD_DCP_ARGS_VAR=DCP_$(echo "${CMD_DCP}" | tr [a-z] [A-Z])_ARGS

    # --- use systemd name if defined or create default ---
    if [ -n "${INSTALL_DCP_NAME}" ]; then
        SERVICE_NAME=dcp-${INSTALL_DCP_NAME}
    else
        if [ "${CMD_DCP}" = server ]; then
            SERVICE_NAME=dcp
        else
            SERVICE_NAME=dcp-${CMD_DCP}
        fi
    fi

    # --- check for invalid characters in system name ---
    valid_chars=$(printf '%s' "${SERVICE_NAME}" | sed -e 's/[][!#$%&()*;<=>?\_`{|}/[:space:]]/^/g;' )
    if [ "${SERVICE_NAME}" != "${valid_chars}"  ]; then
        invalid_chars=$(printf '%s' "${valid_chars}" | sed -e 's/[^^]/ /g')
        fatal "Invalid characters for system name:
            ${SERVICE_NAME}
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
    BIN_DIR=${INSTALL_DCP_BIN_DIR:-/usr/local/bin}
    DATA_DIR=/var/lib/bhojpur/dcp

    # --- set related files from system name ---
    UNINSTALL_DCP_SH=${UNINSTALL_DCP_SH:-${BIN_DIR}/DCP-uninstall.sh}
    KILLALL_DCP_SH=${KILLALL_DCP_SH:-${BIN_DIR}/dcp-killall.sh}

    # --- use systemd directory if defined or create default ---
    if [ -n "${INSTALL_DCP_SYSTEMD_DIR}" ]; then
        warn "INSTALL_DCP_SYSTEMD_DIR is deprecated, use INSTALL_DCP_SERVICE_DIR instead"
        [ -n "${INSTALL_DCP_SERVICE_DIR}" ] && \
            error "INSTALL_DCP_SERVICE_DIR is defined, but so is INSTALL_DCP_SYSTEMD_DIR"
        SERVICE_DIR="${INSTALL_DCP_SYSTEMD_DIR}"
    fi
    # --- use service directory if defined or create default ---
    if [ -n "${INSTALL_DCP_SERVICE_DIR}" ]; then
        SERVICE_DIR="${INSTALL_DCP_SERVICE_DIR}"
    fi

    # --- use service or environment location depending on systemd/openrc ---
    CONFIG_DIR=${INSTALL_DCP_CONFIG_DIR:-/etc/bhojpur/dcp}
    if [ "${HAS_SYSTEMD}" = true ]; then
        SERVICE_FILE=${SERVICE_NAME}.service
        SERVICE_DIR=${SERVICE_DIR:-/etc/systemd/system}
    elif [ "${HAS_OPENRC}" = true ]; then
        SERVICE_FILE=${SERVICE_NAME}
        SERVICE_DIR=${SERVICE_DIR:-/etc/init.d}
    fi

    $SUDO mkdir -p ${CONFIG_DIR}
    FILE_DCP_SERVICE=${SERVICE_DIR}/${SERVICE_FILE}
    FILE_DCP_ENV=${CONFIG_DIR}/${SERVICE_NAME}.env
    #  FILE_DCP_FLAGS=${CONFIG_DIR}/${SERVICE_NAME}.flags

    # --- get hash of config & exec for currently installed Bhojpur DCP ---
    PRE_INSTALL_HASHES=$(get_installed_hashes)

    # --- if bin directory is read only skip download ---
    if [ "${INSTALL_DCP_BIN_DIR_READ_ONLY}" = true ]; then
        INSTALL_DCP_SKIP_DOWNLOAD=true
    fi
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
    [ -x "$(which $1)" ] || return 1

    # Set verified executable as our downloader program and return success
    DOWNLOADER=$1
    return 0
}

# --- verify existence of semanage when SELinux is enabled ---
verify_semanage() {
    if [ -x "$(which getenforce)" ]; then
        if [ "Disabled" != $(getenforce) ] && [ ! -x "$(which semanage)" ]; then
            fatal 'SELinux is enabled but semanage is not found'
        fi
    fi
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

# --- use desired Bhojpur DCP version if defined or find latest ---
get_release_version() {
    if [ -n "${INSTALL_DCP_COMMIT}" ]; then
        VERSION_DCP="commit ${INSTALL_DCP_COMMIT}"
    elif [ -n "${INSTALL_DCP_VERSION}" ]; then
        VERSION_DCP=${INSTALL_DCP_VERSION}
    else
        info "Finding latest release"
        case $DOWNLOADER in
            curl)
                VERSION_DCP=$(curl -w '%{url_effective}' -I -L -s -S ${GITHUB_URL}/latest -o /dev/null | sed -e 's|.*/||')
                ;;
            wget)
                VERSION_DCP=$(wget -SqO /dev/null ${GITHUB_URL}/latest 2>&1 | grep -i Location | sed -e 's|.*/||')
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

    if command -v getenforce >/dev/null 2>&1; then
        if [ "Disabled" != $(getenforce) ]; then
	    info 'SELinux is enabled, setting permissions'
	    if ! $SUDO semanage fcontext -l | grep "${BIN_DIR}/dcp" > /dev/null 2>&1; then
	        $SUDO semanage fcontext -a -t bin_t "${BIN_DIR}/dcp"
	    fi
	    $SUDO restorecon -v ${BIN_DIR}/dcp > /dev/null
        fi
    fi
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
    verify_semanage
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
            which_cmd=$(which ${cmd} 2>/dev/null || true)
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

do_unmount() {
    awk -v path="$1" '$2 ~ ("^" path) { print $2 }' /proc/self/mounts | sort -r | xargs -r -t -n 1 umount
}

do_unmount '/run/dcp'
do_unmount '/var/lib/bhojpur/dcp'
do_unmount '/var/lib/kubelet/pods'
do_unmount '/run/netns/cni-'

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

if which systemctl; then
    for service in ${SERVICE_DIR}/dcp*.service; do
        systemctl disable \$service
        systemctl reset-failed \$service
        rm -f \$service
    done
    systemctl daemon-reload
fi
if which rc-update; then
    for service in ${SERVICE_DIR}/dcp*; do
        rc-update delete \$service default
        rm -f \$service
    done
fi

remove_uninstall() {
    rm -f ${UNINSTALL_DCP_SH}
}
trap remove_uninstall EXIT

for cmd in kubectl crictl ctr; do
    if [ -L ${BIN_DIR}/\$cmd ]; then
        rm -f ${BIN_DIR}/\$cmd
    fi
done

rm -rf /var/lib/kubelet
rm -rf ${CONFIG_DIR}
rm -rf ${DATA_DIR}
rm -f ${BIN_DIR}/dcp
rm -f ${KILLALL_DCP_SH}
EOF
    $SUDO chmod 755 ${UNINSTALL_DCP_SH}
    $SUDO chown root:root ${UNINSTALL_DCP_SH}
}

# --- disable current service if loaded --
systemd_disable() {
    $SUDO rm -f /etc/systemd/system/${SERVICE_NAME} || true
    $SUDO mv /etc/systemd/system/${SERVICE_NAME}.env ${FILE_DCP_ENV} >/dev/null 2>&1 || true
    $SUDO systemctl disable ${SERVICE_NAME} >/dev/null 2>&1 || true
}

# --- capture current env and create file containing dcp_ variables ---
create_env_file() {
    if [ -f "${FILE_DCP_ENV}" ] && [ "${INSTALL_DCP_UPGRADE}" = true ]; then
        warn "env: Skipping env creation, ${FILE_DCP_ENV} exists and INSTALL_DCP_UPGRADE=true"
        return 0
    fi
    info "env: Creating environment file ${FILE_DCP_ENV}"
    UMASK=$(umask)
    umask 0177
    env | grep '^DCP_' | $SUDO tee ${FILE_DCP_ENV} >/dev/null
    env | egrep -i '^(NO|HTTP|HTTPS)_PROXY' | $SUDO tee -a ${FILE_DCP_ENV} >/dev/null
    echo "${CMD_DCP_ARGS_VAR}=\"${CMD_DCP_ARGS}\"" | $SUDO tee -a ${FILE_DCP_ENV} >/dev/null
    umask $UMASK
}

# --- write systemd service file ---
create_systemd_service_file() {
    info "systemd: Creating service file ${FILE_DCP_SERVICE}"
    conflicts=
    if [ "${SERVICE_NAME}" != 'dcp-agent' ]; then
        conflicts='dcp-agent.service'
    fi
    $SUDO tee ${FILE_DCP_SERVICE} >/dev/null << EOF
[Unit]
Description=Bhojpur Kubernetes
Documentation=https://bhojpur.net
Wants=network-online.target
After=network-online.target
Conflicts=${conflicts}

[Install]
WantedBy=multi-user.target

[Service]
Type=${SYSTEMD_TYPE}
EnvironmentFile=${FILE_DCP_ENV}
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
ExecStart=${BIN_DIR}/dcp ${CMD_DCP} \$${CMD_DCP_ARGS_VAR}
EOF
}

# --- write openrc service file ---
create_openrc_service_file() {
    LOG_FILE=/var/log/${SERVICE_NAME}.log

    info "openrc: Creating service file ${FILE_DCP_SERVICE}"
    $SUDO tee ${FILE_DCP_SERVICE} >/dev/null << EOF
#!/sbin/openrc-run

set -o allexport
if [ -f /etc/environment ]; then source /etc/environment; fi
if [ -f ${FILE_DCP_ENV} ]; then source ${FILE_DCP_ENV}; fi
set +o allexport

depend() {
    after network-online
    want cgroups
}

start_pre() {
    rm -f /tmp/dcp.*
}

supervisor=supervise-daemon
name=${SERVICE_NAME}
command="${BIN_DIR}/dcp"
command_args="${CMD_DCP} \$${CMD_DCP_ARGS_VAR} >>${LOG_FILE} 2>&1"

output_log=${LOG_FILE}
error_log=${LOG_FILE}

pidfile="/var/run/${SERVICE_NAME}.pid"
respawn_delay=5
EOF
    $SUDO chmod 0755 ${FILE_DCP_SERVICE}

    $SUDO tee /etc/logrotate.d/${SERVICE_NAME} >/dev/null << EOF
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
    info "systemd: Enabling ${SERVICE_NAME} unit"
    $SUDO systemctl enable ${FILE_DCP_SERVICE} >/dev/null
    $SUDO systemctl daemon-reload >/dev/null
}

systemd_start() {
    info "systemd: Starting ${SERVICE_NAME}"
    $SUDO systemctl restart ${SERVICE_NAME}
}

# --- enable and start openrc service ---
openrc_enable() {
    info "openrc: Enabling ${SERVICE_NAME} service for default runlevel"
    $SUDO rc-update add ${SERVICE_NAME} default >/dev/null
}

openrc_start() {
    info "openrc: Starting ${SERVICE_NAME}"
    $SUDO ${FILE_DCP_SERVICE} restart
}

# --- startup systemd or openrc service ---
service_enable_and_start() {
    [ "${INSTALL_DCP_SKIP_ENABLE}" = true ] && return

    [ "${HAS_SYSTEMD}" = true ] && systemd_enable
    [ "${HAS_OPENRC}" = true ] && openrc_enable

    [ "${INSTALL_DCP_SKIP_START}" = true ] && return

    POST_INSTALL_HASHES=$(get_installed_hashes)
    if [ "${PRE_INSTALL_HASHES}" = "${POST_INSTALL_HASHES}" ]; then
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
    create_symlinks
    create_killall
    create_uninstall
    systemd_disable
    create_env_file
    create_service_file
    service_enable_and_start
}