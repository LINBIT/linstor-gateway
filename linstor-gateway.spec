%define name linstor-gateway
%define release 1
%define version 0.6.2
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
install -D -m 644 linstor-gateway.service %{buildroot}%{_unitdir}/linstor-gateway.service
install -D -m 644 linstor-gateway.xml %{buildroot}%{_firewalldir}/services/linstor-gateway.xml

%post
%systemd_post linstor-gateway.service
%firewalld_reload

%preun
%systemd_preun linstor-gateway.service

%postun
%systemd_postun

%files
%defattr(-,root,root)
	%{_sbindir}/%{name}
	%{_unitdir}/linstor-gateway.service
	%dir %{_firewalldir}
	%dir %{_firewalldir}/services
	%{_firewalldir}/services/linstor-gateway.xml
