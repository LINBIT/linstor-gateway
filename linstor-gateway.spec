%define _firewalldir /usr/lib/firewalld

%if 0%{?suse_version}
%define firewall_macros_package firewall-macros
%else
%define firewall_macros_package firewalld-filesystem
%endif

Name: linstor-gateway
Version: 1.1.0
Release: 1
Summary: LINSTOR Gateway exposes highly available LINSTOR storage via iSCSI, NFS, or NVMe-OF.
%global tarball_version %(echo "%{version}" | sed -e 's/~rc/-rc/' -e 's/~alpha/-alpha/')

URL: https://www.github.com/LINBIT/linstor-gateway
Source: %{name}-%{tarball_version}.tar.gz
BuildRoot: %{buildroot}
BuildRequires: %{firewall_macros_package}
License: GPLv3+
ExclusiveOS: linux

%description
LINSTOR Gateway exposes highly available LINSTOR storage via iSCSI, NFS, or NVMe-OF.

%prep
%setup -q -n %{name}-%{tarball_version}

%build
make

%install
install -D -m 755 %{_builddir}/%{name}-%{tarball_version}/%{name} %{buildroot}/%{_sbindir}/%{name}
install -D -m 644 %{name}.service %{buildroot}%{_unitdir}/%{name}.service
install -D -m 644 %{name}.xml %{buildroot}%{_firewalldir}/services/%{name}.xml

%post
%systemd_post %{name}.service
%firewalld_reload

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun %{name}.service

%files
%defattr(-,root,root)
	%{_sbindir}/%{name}
	%{_unitdir}/%{name}.service
	%dir %{_firewalldir}
	%dir %{_firewalldir}/services
	%{_firewalldir}/services/%{name}.xml

%changelog
* Fri Feb 24 2023 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 1.1.0-1
-  New upstream release

* Mon Nov 21 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 1.0.0-1
-  New upstream release

* Fri Nov 4 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 1.0.0~rc.1-1
-  New upstream release

* Tue Jul 26 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.13.1-1
-  New upstream release

* Mon Jun 27 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.13.0-1
-  New upstream release

* Tue May 3 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.12.1-1
-  New upstream release

* Sun Apr 3 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.12.0-1
-  New upstream release

* Thu Mar 17 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.12.0~rc.1-1
-  New upstream release

* Mon Feb 14 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.11.0-1
-  New upstream release

* Tue Feb 8 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.11.0~rc.2-1
-  New upstream release

* Mon Jan 31 2022 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.11.0~rc.1-1
-  New upstream release

* Wed Nov 24 2021 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.10.0-1
-  New upstream release

* Wed Nov 17 2021 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.10.0~rc.1-1
-  New upstream release

* Tue Sep 28 2021 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.9.0-1
-  New upstream release

* Thu Sep 23 2021 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.9.0~rc.3-1
-  New upstream release

* Wed Sep 15 2021 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.9.0~rc.2-1
-  New upstream release

* Wed Sep 1 2021 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.9.0~rc.1-1
-  New upstream release

* Tue Mar 23 2021 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.8.0-1
-  New upstream release

* Fri Dec 04 2020 Christoph Böhmwalder <christoph.boehmwalder@linbit.com> - 0.7.0-1
-  Rename to linstor-gateway

* Wed Oct 09 2019 Roland Kammerer <roland.kammerer@linbit.com> - 0.1.0-1
-  Initial Release
