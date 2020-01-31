%define name linstor-iscsi
%define release 1
%define version 0.4.1
#%define buildroot %{_topdir}/BUILD/%{name}-%{version}

%define _firewalldir /usr/lib/firewalld

%if 0%{?suse_version}
%define firewall_macros_package firewall-macros
%else
%define firewall_macros_package firewalld-filesystem
%endif

BuildRoot: %{buildroot}
# Requires: drbd-utils >= 9.0.0
BuildRequires: %{firewall_macros_package}
Summary: linstor-iscsi manages higly available iSCSI targets by leveraging on linstor
License: GPLv3+
ExclusiveOS: linux
Name: %{name}
Version: %{version}
Release: %{release}
Source: %{name}-%{version}.tar.gz

%description
linstor-iscsi manages higly available iSCSI targets by leveraging on linstor

%prep
%setup -q

%build

%install
mkdir -p %{buildroot}/%{_sbindir}/
cp %{_builddir}/%{name}-%{version}/%{name} %{buildroot}/%{_sbindir}/
install -D -m 644 linstor-iscsi.service %{buildroot}%{_unitdir}/linstor-iscsi.service
install -D -m 644 linstor-iscsi.xml %{buildroot}%{_firewalldir}/services/linstor-iscsi.xml

%post
%systemd_post linstor-iscsi.service
%firewalld_reload

%preun
%systemd_preun linstor-iscsi.service

%postun
%systemd_postun

%files
%defattr(-,root,root)
	%{_sbindir}/%{name}
	%{_unitdir}/linstor-iscsi.service
	%dir %{_firewalldir}
	%dir %{_firewalldir}/services
	%{_firewalldir}/services/linstor-iscsi.xml
