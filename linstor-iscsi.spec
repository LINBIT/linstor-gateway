%define name linstor-iscsi
%define release 1
%define version 0.1.0
#%define buildroot %{_topdir}/BUILD/%{name}-%{version}

BuildRoot: %{buildroot}
# Requires: drbd-utils >= 9.0.0
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

%files
%defattr(-,root,root)
	%{_sbindir}/%{name}

