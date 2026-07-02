%define debug_package %{nil}

Name:           dcvix-director
Version:        %{pkg_version}
Release:        1%{?dist}
Summary:        Dcvix Director - Sessions and Servers Management Service for Amazon DCV
License:        MIT
URL:            https://github.com/dcvix/dcvix-director
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.22
BuildRequires:  systemd
BuildRequires:  systemd-rpm-macros
Requires:       systemd openssl

%description
dcvix is a session broker and server-pool manager
for Amazon DCV. It provides centralized authentication,
desktop session lifecycle management,
and automatic allocation of DCV servers.
This package provides the director, the central management service.
It runs on a server and provides the REST API,
admin web UI, CA authority, and user authentication
for the dcvix session broker.

%prep
%setup -q

%build
make build

%install
rm -rf $RPM_BUILD_ROOT
# Create directories
install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_sysconfdir}/%{name}
install -d %{buildroot}%{_unitdir}
install -d %{buildroot}%{_localstatedir}/log/%{name}
install -d %{_docdir}/%{name}
# Install binary
install -p -m 755 dist/dcvix-director-v%{version}-linux-amd64/%{name} %{buildroot}%{_bindir}/%{name}
# Install config
install -p -m 644 %{name}.conf %{buildroot}%{_sysconfdir}/%{name}/%{name}.conf
# Install systemd service
install -p -m 644 contrib/systemd/%{name}.service %{buildroot}%{_unitdir}/%{name}.service

%clean
rm -rf $RPM_BUILD_ROOT

%files
%{_bindir}/%{name}
%dir %{_sysconfdir}/%{name}
%config(noreplace) %{_sysconfdir}/%{name}/%{name}.conf
%{_unitdir}/%{name}.service
%dir %{_localstatedir}/log/%{name}
%license LICENSE.md
%doc README.md

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun_with_restart %{name}.service

%changelog
* Fri Jul 02 2026 Diego Cortassa <diego@cortassa.net> - %{version}-%{release}
- Initial RPM release
