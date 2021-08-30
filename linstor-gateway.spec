%define name linstor-gateway
%define release 1
%define version 0.8.0
#%define buildroot %{_topdir}/BUILD/%{name}-%{version}

%define _firewalldir /usr/lib/firewalld

%if 0%{?suse_version}
%define firewall_macros_package firewall-macros
%else
%define firewall_macros_package firewalld-filesystem
%endif

BuildRoot: %{buildroot}
BuildRequires: %{firewall_macros_package}
Summary: LINSTOR Gateway exposes highly available LINSTOR storage via iSCSI, NFS, or NVMe-OF.
License: GPLv3+
ExclusiveOS: linux
Name: %{name}
Version: %{version}
Release: %{release}
Source: %{name}-%{version}.tar.gz

%description
LINSTOR Gateway exposes highly available LINSTOR storage via iSCSI, NFS, or NVMe-OF.

%prep
%setup -q

%build

%install
mkdir -p %{buildroot}/%{_sbindir}/
cp %{_builddir}/%{name}-%{version}/%{name} %{buildroot}/%{_sbindir}/
install -D -m 644 %{name}.service %{buildroot}%{_unitdir}/%{name}.service
install -D -m 644 %{name}.xml %{buildroot}%{_firewalldir}/services/%{name}.xml

%post
%systemd_post %{name}.service
%firewalld_reload

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun

%files
%defattr(-,root,root)
	%{_sbindir}/%{name}
	%{_unitdir}/%{name}.service
	%dir %{_firewalldir}
	%dir %{_firewalldir}/services
	%{_firewalldir}/services/%{name}.xml
