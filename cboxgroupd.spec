# 
# cboxgroupd spec file
#

Name: cboxgroupd
Summary: A server that allows the resolution of e-groups belonging to an user and viceversa.
Version: 1.3.1
Release: 1%{?dist}
License: AGPLv3
BuildRoot: %{_tmppath}/%{name}-buildroot
Group: CERN-IT/ST
BuildArch: x86_64
Source: %{name}-%{version}.tar.gz

%description
This RPM provides a golang webserver that performs e-group resolutions for CERN AD users

# Don't do any post-install weirdness, especially compiling .py files
%define __os_install_post %{nil}

%prep
%setup -n %{name}-%{version}

%install
# server versioning

# installation
rm -rf %buildroot/
mkdir -p %buildroot/usr/local/bin
mkdir -p %buildroot/etc/cboxgroupd
mkdir -p %buildroot/etc/logrotate.d
mkdir -p %buildroot/usr/lib/systemd/system
mkdir -p %buildroot/var/log/cboxgroupd
install -m 755 cboxgroupd	     %buildroot/usr/local/bin/cboxgroupd
install -m 644 cboxgroupd.service    %buildroot/usr/lib/systemd/system/cboxgroupd.service
install -m 644 cboxgroupd.yaml       %buildroot/etc/cboxgroupd/cboxgroupd.yaml
install -m 644 cboxgroupd.logrotate  %buildroot/etc/logrotate.d/cboxgroupd

%clean
rm -rf %buildroot/

%preun

%post

%files
%defattr(-,root,root,-)
/etc/cboxgroupd
/etc/logrotate.d/cboxgroupd
/var/log/cboxgroupd
/usr/lib/systemd/system/cboxgroupd.service
/usr/local/bin/*


%changelog
* Thu Nov 27 2017 Hugo Gonzalez Labrador <hugo.gonzalez.labrador@cern.ch> 1.3.1
- Add --config flag to use custom configuration file
* Thu Nov 27 2017 Hugo Gonzalez Labrador <hugo.gonzalez.labrador@cern.ch> 1.3.0
- RPMization of the sofware

