Name:		tr-ocp-installer
Version:	1.0.0
Release:	0.0.master%{?release_suffix}%{?dist}
License:	ASL 2.0
Summary:	OCP installer with terraform provider patch
Group:		Virtualization/Management
URL:		https://github.com/openshift/installer
BuildArch:	noarch

%description
OCP installer binary compiled with terraform provider patch

%files
%{_bindir}/openshift-install

%changelog
* THU JAN 13 2022 Eli Mesika <emesika@redhat.com> 1.0.0
- Created
