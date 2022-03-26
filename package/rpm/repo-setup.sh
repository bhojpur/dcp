cat <<EOF >/etc/yum.repos.d/bhojpur.repo
[bhojpur]
name=Bhojpur Repository
baseurl=https://rpm.bhojpur.net/
enabled=1
gpgcheck=1
gpgkey=https://rpm.bhojpur.net/public.key
EOF