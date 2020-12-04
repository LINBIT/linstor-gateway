%define name linstor-gateway
%define release 1
%define version 0.7.0
#%define buildroot %{_topdir}/BUILD/%{name}-%{version}

%define _firewalldir /usr/lib/firewalld

%if 0%{?suse_version}
%define firewall_macros_package firewall-macros
%else
%define firewall_macros_package firewalld-filesystem
%endif

BuildRoot: %{buildroot}
BuildRequires: %{firewall_macros_package}
Summary: linstor-gateway manages higly available iSCSI targets and NFS shares using LINSTOR and Pacemaker
License: GPLv3+
ExclusiveOS: linux
Name: %{name}
Version: %{version}
Release: %{release}
Source: %{name}-%{version}.tar.gz

%description
linstor-gateway manages higly available iSCSI targets and NFS shares using LINSTOR and Pacemaker

%prep
%setup -q

%build

%install
mkdir -p %{buildroot}/%{_sbindir}/
cp %{_builddir}/%{name}-%{version}/%{name} %{buildroot}/%{_sbindir}/
ln -s %{_sbindir}/%{name} %{buildroot}/%{_sbindir}/linstor-iscsi
ln -s %{_sbindir}/%{name} %{buildroot}/%{_sbindir}/linstor-nfs
install -D -m 644 linstor-iscsi.service %{buildroot}%{_unitdir}/linstor-iscsi.service
install -D -m 644 linstor-nfs.service %{buildroot}%{_unitdir}/linstor-nfs.service
install -D -m 644 linstor-iscsi.xml %{buildroot}%{_firewalldir}/services/linstor-iscsi.xml
install -D -m 644 linstor-nfs.xml %{buildroot}%{_firewalldir}/services/linstor-nfs.xml

%post
%systemd_post linstor-iscsi.service
%systemd_post linstor-nfs.service
%firewalld_reload

%preun
%systemd_preun linstor-iscsi.service
%systemd_preun linstor-nfs.service

%postun
%systemd_postun

%files
%defattr(-,root,root)
	%{_sbindir}/%{name}
	%{_sbindir}/linstor-iscsi
	%{_sbindir}/linstor-nfs
	%{_unitdir}/linstor-iscsi.service
	%{_unitdir}/linstor-nfs.service
	%dir %{_firewalldir}
	%dir %{_firewalldir}/services
	%{_firewalldir}/services/linstor-iscsi.xml
	%{_firewalldir}/services/linstor-nfs.xml
