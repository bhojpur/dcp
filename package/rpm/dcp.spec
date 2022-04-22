# vim: sw=4:ts=4:et

%define install_path  /usr/bin
%define util_path     %{_datadir}/dcp
%define install_sh    %{util_path}/setup/install.sh
%define uninstall_sh  %{util_path}/setup/uninstall.sh

Name:    dcp
Version: %{dcp_version}
Release: %{dcp_release}%{?dist}
Summary: Bhojpur Kubernetes

Group:   System Environment/Base		
License: ASL 2.0
URL:     https://bhojpur.net

BuildRequires: systemd
Requires(post): dcp-selinux >= %{dcp_policyver}

%description
The certified Kubernetes distribution built for IoT & Edge computing.

%install
install -d %{buildroot}%{install_path}
install dist/artifacts/%{dcp_binary} %{buildroot}%{install_path}/dcp
install -d %{buildroot}%{util_path}/setup
install package/rpm/install.sh %{buildroot}%{install_sh}

%post
# do not overwrite env file if present
export INSTALL_DCP_UPGRADE=true
export INSTALL_DCP_BIN_DIR=%{install_path}
export INSTALL_DCP_SKIP_DOWNLOAD=true
export INSTALL_DCP_SKIP_ENABLE=true
export INSTALL_DCP_DEBUG=true
export UNINSTALL_DCP_SH=%{uninstall_sh}

(
    # install server service
    INSTALL_DCP_NAME=server \
        %{install_sh}

    # install agent service
    INSTALL_DCP_SYMLINK=skip \
    INSTALL_DCP_BIN_DIR_READ_ONLY=true \
    DCP_TOKEN=example-token \
    DCP_URL=https://example-dcp-server:6443/ \
        %{install_sh} agent

# save debug log of the install
) >%{util_path}/setup/install.log 2>&1

%systemd_post dcp-server.service
%systemd_post dcp-agent.service
exit 0

%postun
# do not run uninstall script on upgrade
if [ $1 = 0 ]; then
    %{uninstall_sh}
    rm -rf %{util_path}
fi
exit 0

%files
%{install_path}/dcp
%{install_sh}

%changelog
* Mon Mar 26 2018 Shashi Bhushan Rai <info@bhojpur-consulting.com> 0.1-1
- Initial version