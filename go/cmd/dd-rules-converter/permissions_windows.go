//go:build windows
// +build windows

package main

import (
  "unsafe"
  "syscall"

  "golang.org/x/sys/windows"
)

var (
	advapi32                        = syscall.NewLazyDLL("advapi32.dll")
	procTreeResetNamedSecurityInfoW = advapi32.NewProc("TreeResetNamedSecurityInfoW")
)

// securityInfo holds the security information extracted from a security
// descriptor for use in Windows API calls such as SetNamedSecurityInfo.
type securityInfo struct {
	Flags windows.SECURITY_INFORMATION
	Owner *windows.SID
	Group *windows.SID
	DACL  *windows.ACL
	SACL  *windows.ACL
}

// SetRepositoryPermissions sets the permissions on the repository directory
// It needs to be world readable so that user processes can load installed libraries
func SetPermissions(path string) error {
  // Desired permissions:
  // - OWNER: Administrators
  // - GROUP: Administrators
  // - SYSTEM: Full Control (propagates to children)
  // - Administrators: Full Control (propagates to children)
  // - Everyone: 0x1200A9 Read and execute (propagates to children)
  // - PROTECTED: does not inherit permissions from parent
  sddl := "O:BAG:BAD:PAI(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;0x1200A9;;;WD)"
	sd, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}
	return treeResetNamedSecurityInfoFromSecurityDescriptor(path, sd)
}

// TreeResetNamedSecurityInfo wraps the TreeResetNamedSecurityInfoW Windows API function
//
// https://learn.microsoft.com/en-us/windows/win32/api/aclapi/nf-aclapi-treeresetnamedsecurityinfow
func TreeResetNamedSecurityInfo(
	objectName string,
	objectType windows.SE_OBJECT_TYPE,
	securityInfo windows.SECURITY_INFORMATION,
	owner *windows.SID,
	group *windows.SID,
	dacl *windows.ACL,
	sacl *windows.ACL,
	keepExplicitDacl bool) error {

	utf16ObjectName, err := windows.UTF16PtrFromString(objectName)
	if err != nil {
		return err
	}

	keepExplicitDaclInt := uintptr(0)
	if keepExplicitDacl {
		keepExplicitDaclInt = 1
	}

	r0, _, _ := procTreeResetNamedSecurityInfoW.Call(
		uintptr(unsafe.Pointer(utf16ObjectName)),
		uintptr(objectType),
		uintptr(securityInfo),
		uintptr(unsafe.Pointer(owner)),
		uintptr(unsafe.Pointer(group)),
		uintptr(unsafe.Pointer(dacl)),
		uintptr(unsafe.Pointer(sacl)),
		keepExplicitDaclInt,
		// don't use a progress callback
		0, 0, 0)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

func securityInformationFromControlFlags(control windows.SECURITY_DESCRIPTOR_CONTROL) windows.SECURITY_INFORMATION {
	var flags windows.SECURITY_INFORMATION
	if control&windows.SE_DACL_PROTECTED == 0 {
		flags |= windows.UNPROTECTED_DACL_SECURITY_INFORMATION
	} else {
		flags |= windows.PROTECTED_DACL_SECURITY_INFORMATION
	}
	if control&windows.SE_SACL_PROTECTED == 0 {
		flags |= windows.UNPROTECTED_SACL_SECURITY_INFORMATION
	} else {
		flags |= windows.PROTECTED_SACL_SECURITY_INFORMATION
	}
	return flags
}

func getSecurityInfoFromSecurityDescriptor(sd *windows.SECURITY_DESCRIPTOR) (*securityInfo, error) {
	var flags windows.SECURITY_INFORMATION
	control, _, err := sd.Control()
	if err != nil {
		return nil, err
	}
	flags |= securityInformationFromControlFlags(control)

	owner, _, err := sd.Owner()
	if err != nil {
		return nil, err
	}
	if owner != nil {
		flags |= windows.OWNER_SECURITY_INFORMATION
	}
	group, _, err := sd.Group()
	if err != nil {
		return nil, err
	}
	if group != nil {
		flags |= windows.GROUP_SECURITY_INFORMATION
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		if err != windows.ERROR_OBJECT_NOT_FOUND {
			return nil, err
		}
	} else {
		flags |= windows.DACL_SECURITY_INFORMATION
	}
	sacl, _, err := sd.SACL()
	if err != nil {
		if err != windows.ERROR_OBJECT_NOT_FOUND {
			return nil, err
		}
	} else {
		flags |= windows.SACL_SECURITY_INFORMATION
	}
	return &securityInfo{
		Flags: flags,
		Owner: owner,
		Group: group,
		DACL:  dacl,
		SACL:  sacl,
	}, nil
}

func treeResetNamedSecurityInfoFromSecurityDescriptor(root string, sd *windows.SECURITY_DESCRIPTOR) error {
	info, err := getSecurityInfoFromSecurityDescriptor(sd)
	if err != nil {
		return err
	}
	err = TreeResetNamedSecurityInfo(
		root,
		windows.SE_FILE_OBJECT,
		info.Flags,
		info.Owner,
		info.Group,
		info.DACL,
		info.SACL,
		// Set to false to remove explicit ACEs from the subtree
		false)
	if err != nil {
		return err
	}
	return nil
}
