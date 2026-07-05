%global debug_package %{nil}
%global _missing_build_ids_terminate_build 0

Name:           miburi
Version:        0.1.0
Release:        1%{?dist}
Summary:        mouse gestures for gnome using evdev

License:        MIT
URL:            https://github.com/addidotlol/miburi

BuildRequires:  golang
BuildRequires:  systemd-rpm-macros

%description
mouse gesture daemon that turns forward button plus mouse movement into
key presses

%build
go build -mod=vendor -trimpath -o miburi .

%install
install -Dm755 miburi %{buildroot}%{_bindir}/miburi
install -Dm644 packaging/miburi.service %{buildroot}%{_unitdir}/miburi.service
install -Dm644 packaging/miburi.sysconfig %{buildroot}%{_sysconfdir}/sysconfig/miburi

%post
%systemd_post miburi.service

%preun
%systemd_preun miburi.service

%postun
%systemd_postun_with_restart miburi.service

%files
%{_bindir}/miburi
%{_unitdir}/miburi.service
%config(noreplace) %{_sysconfdir}/sysconfig/miburi

%changelog
* Sat Jul 04 2026 addison <addidotlol@gmail.com> - 0.1.0-1
- initial package
