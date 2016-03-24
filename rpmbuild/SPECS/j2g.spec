%define		systemddir	/usr/lib/systemd/system		
Name:		j2g
Version:	0.0.1
Release:	1%{?dist}
Summary:	j2g - journald forwarder to gelf endpoint

Group:		extras
License:	WTFPL
URL:		https://github.com/xytis/j2g
Source:		%{name}-%{version}.tar.gz	
Packager: 	Random Folk @ N4L

%description
Simple golang process to tail journald and forward events to gelf

For futher info try j2g help

%prep
%setup

%install
mkdir -p %{buildroot}/%{_sbindir}
mkdir -p %{buildroot}/%{systemddir}

cp %{_builddir}/%{name}-%{version}/%{name} %{buildroot}/%{_sbindir}
cp %{_sourcedir}/%{name}.service %{buildroot}/%{systemddir}/%{name}.service

%files
%{systemddir}/j2g.service
%{_sbindir}/j2g

%post
/bin/systemctl --system daemon-reload &> /dev/null || :

%postun
/bin/systemctl --system daemon-reload &> /dev/null || :

%changelog
* Thu Mar 24 2016 Vytis Valentinavičius <xytis@nolife4life.org> - 0.0.1-1
- Initial release

