/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package main

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	MAX_DEVICE_ID_LEN   = 200
	MAX_DEVNODE_ID_LEN  = MAX_DEVICE_ID_LEN
	MAX_GUID_STRING_LEN = 39 // 38 chars + terminator null
	MAX_CLASS_NAME_LEN  = 32
	MAX_PROFILE_LEN     = 80
	MAX_CONFIG_VALUE    = 9999
	MAX_INSTANCE_VALUE  = 9999
	CONFIGMG_VERSION    = 0x0400
)

// Define maximum string length constants
const (
	LINE_LEN                    = 256  // Windows 9x-compatible maximum for displayable strings coming from a device INF.
	MAX_INF_STRING_LENGTH       = 4096 // Actual maximum size of an INF string (including string substitutions).
	MAX_INF_SECTION_NAME_LENGTH = 255  // For Windows 9x compatibility, INF section names should be constrained to 32 characters.
	MAX_TITLE_LEN               = 60
	MAX_INSTRUCTION_LEN         = 256
	MAX_LABEL_LEN               = 30
	MAX_SERVICE_NAME_LEN        = 256
	MAX_SUBTITLE_LEN            = 256
)

const (
	// SP_MAX_MACHINENAME_LENGTH defines maximum length of a machine name in the format expected by ConfigMgr32 CM_Connect_Machine (i.e., "\\\\MachineName\0").
	SP_MAX_MACHINENAME_LENGTH = windows.MAX_PATH + 3
)

// HSPFILEQ is type for setup file queue
type HSPFILEQ uintptr

// DevInfo holds reference to device information set
type DevInfo windows.Handle

// DevInfoData is a device information structure (references a device instance that is a member of a device information set)
type DevInfoData struct {
	size      uint32
	ClassGUID windows.GUID
	DevInst   uint32 // DEVINST handle
	_         uintptr
}

// DevInfoListDetailData is a structure for detailed information on a device information set (used for SetupDiGetDeviceInfoListDetail which supercedes the functionality of SetupDiGetDeviceInfoListClass).
type DevInfoListDetailData struct {
	size                uint32
	ClassGUID           windows.GUID
	RemoteMachineHandle windows.Handle
	remoteMachineName   [SP_MAX_MACHINENAME_LENGTH]uint16
}

func (data *DevInfoListDetailData) GetRemoteMachineName() string {
	return windows.UTF16ToString(data.remoteMachineName[:])
}

func (data *DevInfoListDetailData) SetRemoteMachineName(remoteMachineName string) error {
	str, err := syscall.UTF16FromString(remoteMachineName)
	if err != nil {
		return err
	}
	copy(data.remoteMachineName[:], str)
	return nil
}

// DI_FUNCTION is function type for device installer
type DI_FUNCTION uint32

const (
	DIF_SELECTDEVICE                   DI_FUNCTION = 0x00000001
	DIF_INSTALLDEVICE                  DI_FUNCTION = 0x00000002
	DIF_ASSIGNRESOURCES                DI_FUNCTION = 0x00000003
	DIF_PROPERTIES                     DI_FUNCTION = 0x00000004
	DIF_REMOVE                         DI_FUNCTION = 0x00000005
	DIF_FIRSTTIMESETUP                 DI_FUNCTION = 0x00000006
	DIF_FOUNDDEVICE                    DI_FUNCTION = 0x00000007
	DIF_SELECTCLASSDRIVERS             DI_FUNCTION = 0x00000008
	DIF_VALIDATECLASSDRIVERS           DI_FUNCTION = 0x00000009
	DIF_INSTALLCLASSDRIVERS            DI_FUNCTION = 0x0000000A
	DIF_CALCDISKSPACE                  DI_FUNCTION = 0x0000000B
	DIF_DESTROYPRIVATEDATA             DI_FUNCTION = 0x0000000C
	DIF_VALIDATEDRIVER                 DI_FUNCTION = 0x0000000D
	DIF_DETECT                         DI_FUNCTION = 0x0000000F
	DIF_INSTALLWIZARD                  DI_FUNCTION = 0x00000010
	DIF_DESTROYWIZARDDATA              DI_FUNCTION = 0x00000011
	DIF_PROPERTYCHANGE                 DI_FUNCTION = 0x00000012
	DIF_ENABLECLASS                    DI_FUNCTION = 0x00000013
	DIF_DETECTVERIFY                   DI_FUNCTION = 0x00000014
	DIF_INSTALLDEVICEFILES             DI_FUNCTION = 0x00000015
	DIF_UNREMOVE                       DI_FUNCTION = 0x00000016
	DIF_SELECTBESTCOMPATDRV            DI_FUNCTION = 0x00000017
	DIF_ALLOW_INSTALL                  DI_FUNCTION = 0x00000018
	DIF_REGISTERDEVICE                 DI_FUNCTION = 0x00000019
	DIF_NEWDEVICEWIZARD_PRESELECT      DI_FUNCTION = 0x0000001A
	DIF_NEWDEVICEWIZARD_SELECT         DI_FUNCTION = 0x0000001B
	DIF_NEWDEVICEWIZARD_PREANALYZE     DI_FUNCTION = 0x0000001C
	DIF_NEWDEVICEWIZARD_POSTANALYZE    DI_FUNCTION = 0x0000001D
	DIF_NEWDEVICEWIZARD_FINISHINSTALL  DI_FUNCTION = 0x0000001E
	DIF_INSTALLINTERFACES              DI_FUNCTION = 0x00000020
	DIF_DETECTCANCEL                   DI_FUNCTION = 0x00000021
	DIF_REGISTER_COINSTALLERS          DI_FUNCTION = 0x00000022
	DIF_ADDPROPERTYPAGE_ADVANCED       DI_FUNCTION = 0x00000023
	DIF_ADDPROPERTYPAGE_BASIC          DI_FUNCTION = 0x00000024
	DIF_TROUBLESHOOTER                 DI_FUNCTION = 0x00000026
	DIF_POWERMESSAGEWAKE               DI_FUNCTION = 0x00000027
	DIF_ADDREMOTEPROPERTYPAGE_ADVANCED DI_FUNCTION = 0x00000028
	DIF_UPDATEDRIVER_UI                DI_FUNCTION = 0x00000029
	DIF_FINISHINSTALL_ACTION           DI_FUNCTION = 0x0000002A
)

// DevInstallParams is device installation parameters structure (associated with a particular device information element, or globally with a device information set)
type DevInstallParams struct {
	size                     uint32
	Flags                    DI_FLAGS
	FlagsEx                  DI_FLAGSEX
	hwndParent               uintptr
	InstallMsgHandler        uintptr
	InstallMsgHandlerContext uintptr
	FileQueue                HSPFILEQ
	_                        uintptr
	_                        uint32
	driverPath               [windows.MAX_PATH]uint16
}

func (params *DevInstallParams) GetDriverPath() string {
	return windows.UTF16ToString(params.driverPath[:])
}

func (params *DevInstallParams) SetDriverPath(driverPath string) error {
	str, err := syscall.UTF16FromString(driverPath)
	if err != nil {
		return err
	}
	copy(params.driverPath[:], str)
	return nil
}

// DI_FLAGS is SP_DEVINSTALL_PARAMS.Flags values
type DI_FLAGS uint32

const (
	// Flags for choosing a device
	DI_SHOWOEM       DI_FLAGS = 0x00000001 // support Other... button
	DI_SHOWCOMPAT    DI_FLAGS = 0x00000002 // show compatibility list
	DI_SHOWCLASS     DI_FLAGS = 0x00000004 // show class list
	DI_SHOWALL       DI_FLAGS = 0x00000007 // both class & compat list shown
	DI_NOVCP         DI_FLAGS = 0x00000008 // don't create a new copy queue--use caller-supplied FileQueue
	DI_DIDCOMPAT     DI_FLAGS = 0x00000010 // Searched for compatible devices
	DI_DIDCLASS      DI_FLAGS = 0x00000020 // Searched for class devices
	DI_AUTOASSIGNRES DI_FLAGS = 0x00000040 // No UI for resources if possible

	// Flags returned by DiInstallDevice to indicate need to reboot/restart
	DI_NEEDRESTART DI_FLAGS = 0x00000080 // Reboot required to take effect
	DI_NEEDREBOOT  DI_FLAGS = 0x00000100 // ""

	// Flags for device installation
	DI_NOBROWSE DI_FLAGS = 0x00000200 // no Browse... in InsertDisk

	// Flags set by DiBuildDriverInfoList
	DI_MULTMFGS DI_FLAGS = 0x00000400 // Set if multiple manufacturers in class driver list

	// Flag indicates that device is disabled
	DI_DISABLED DI_FLAGS = 0x00000800 // Set if device disabled

	// Flags for Device/Class Properties
	DI_GENERALPAGE_ADDED  DI_FLAGS = 0x00001000
	DI_RESOURCEPAGE_ADDED DI_FLAGS = 0x00002000

	// Flag to indicate the setting properties for this Device (or class) caused a change so the Dev Mgr UI probably needs to be updated.
	DI_PROPERTIES_CHANGE DI_FLAGS = 0x00004000

	// Flag to indicate that the sorting from the INF file should be used.
	DI_INF_IS_SORTED DI_FLAGS = 0x00008000

	// Flag to indicate that only the the INF specified by SP_DEVINSTALL_PARAMS.DriverPath should be searched.
	DI_ENUMSINGLEINF DI_FLAGS = 0x00010000

	// Flag that prevents ConfigMgr from removing/re-enumerating devices during device
	// registration, installation, and deletion.
	DI_DONOTCALLCONFIGMG DI_FLAGS = 0x00020000

	// The following flag can be used to install a device disabled
	DI_INSTALLDISABLED DI_FLAGS = 0x00040000

	// Flag that causes SetupDiBuildDriverInfoList to build a device's compatible driver
	// list from its existing class driver list, instead of the normal INF search.
	DI_COMPAT_FROM_CLASS DI_FLAGS = 0x00080000

	// This flag is set if the Class Install params should be used.
	DI_CLASSINSTALLPARAMS DI_FLAGS = 0x00100000

	// This flag is set if the caller of DiCallClassInstaller does NOT want the internal default action performed if the Class installer returns ERROR_DI_DO_DEFAULT.
	DI_NODI_DEFAULTACTION DI_FLAGS = 0x00200000

	// Flags for device installation
	DI_QUIETINSTALL        DI_FLAGS = 0x00800000 // don't confuse the user with questions or excess info
	DI_NOFILECOPY          DI_FLAGS = 0x01000000 // No file Copy necessary
	DI_FORCECOPY           DI_FLAGS = 0x02000000 // Force files to be copied from install path
	DI_DRIVERPAGE_ADDED    DI_FLAGS = 0x04000000 // Prop provider added Driver page.
	DI_USECI_SELECTSTRINGS DI_FLAGS = 0x08000000 // Use Class Installer Provided strings in the Select Device Dlg
	DI_OVERRIDE_INFFLAGS   DI_FLAGS = 0x10000000 // Override INF flags
	DI_PROPS_NOCHANGEUSAGE DI_FLAGS = 0x20000000 // No Enable/Disable in General Props

	DI_NOSELECTICONS DI_FLAGS = 0x40000000 // No small icons in select device dialogs

	DI_NOWRITE_IDS DI_FLAGS = 0x80000000 // Don't write HW & Compat IDs on install
)

// DI_FLAGSEX is SP_DEVINSTALL_PARAMS.FlagsEx values
type DI_FLAGSEX uint32

const (
	DI_FLAGSEX_CI_FAILED                DI_FLAGSEX = 0x00000004 // Failed to Load/Call class installer
	DI_FLAGSEX_FINISHINSTALL_ACTION     DI_FLAGSEX = 0x00000008 // Class/co-installer wants to get a DIF_FINISH_INSTALL action in client context.
	DI_FLAGSEX_DIDINFOLIST              DI_FLAGSEX = 0x00000010 // Did the Class Info List
	DI_FLAGSEX_DIDCOMPATINFO            DI_FLAGSEX = 0x00000020 // Did the Compat Info List
	DI_FLAGSEX_FILTERCLASSES            DI_FLAGSEX = 0x00000040
	DI_FLAGSEX_SETFAILEDINSTALL         DI_FLAGSEX = 0x00000080
	DI_FLAGSEX_DEVICECHANGE             DI_FLAGSEX = 0x00000100
	DI_FLAGSEX_ALWAYSWRITEIDS           DI_FLAGSEX = 0x00000200
	DI_FLAGSEX_PROPCHANGE_PENDING       DI_FLAGSEX = 0x00000400 // One or more device property sheets have had changes made to them, and need to have a DIF_PROPERTYCHANGE occur.
	DI_FLAGSEX_ALLOWEXCLUDEDDRVS        DI_FLAGSEX = 0x00000800
	DI_FLAGSEX_NOUIONQUERYREMOVE        DI_FLAGSEX = 0x00001000
	DI_FLAGSEX_USECLASSFORCOMPAT        DI_FLAGSEX = 0x00002000 // Use the device's class when building compat drv list. (Ignored if DI_COMPAT_FROM_CLASS flag is specified.)
	DI_FLAGSEX_NO_DRVREG_MODIFY         DI_FLAGSEX = 0x00008000 // Don't run AddReg and DelReg for device's software (driver) key.
	DI_FLAGSEX_IN_SYSTEM_SETUP          DI_FLAGSEX = 0x00010000 // Installation is occurring during initial system setup.
	DI_FLAGSEX_INET_DRIVER              DI_FLAGSEX = 0x00020000 // Driver came from Windows Update
	DI_FLAGSEX_APPENDDRIVERLIST         DI_FLAGSEX = 0x00040000 // Cause SetupDiBuildDriverInfoList to append a new driver list to an existing list.
	DI_FLAGSEX_PREINSTALLBACKUP         DI_FLAGSEX = 0x00080000 // not used
	DI_FLAGSEX_BACKUPONREPLACE          DI_FLAGSEX = 0x00100000 // not used
	DI_FLAGSEX_DRIVERLIST_FROM_URL      DI_FLAGSEX = 0x00200000 // build driver list from INF(s) retrieved from URL specified in SP_DEVINSTALL_PARAMS.DriverPath (empty string means Windows Update website)
	DI_FLAGSEX_EXCLUDE_OLD_INET_DRIVERS DI_FLAGSEX = 0x00800000 // Don't include old Internet drivers when building a driver list. Ignored on Windows Vista and later.
	DI_FLAGSEX_POWERPAGE_ADDED          DI_FLAGSEX = 0x01000000 // class installer added their own power page
	DI_FLAGSEX_FILTERSIMILARDRIVERS     DI_FLAGSEX = 0x02000000 // only include similar drivers in class list
	DI_FLAGSEX_INSTALLEDDRIVER          DI_FLAGSEX = 0x04000000 // only add the installed driver to the class or compat driver list.  Used in calls to SetupDiBuildDriverInfoList
	DI_FLAGSEX_NO_CLASSLIST_NODE_MERGE  DI_FLAGSEX = 0x08000000 // Don't remove identical driver nodes from the class list
	DI_FLAGSEX_ALTPLATFORM_DRVSEARCH    DI_FLAGSEX = 0x10000000 // Build driver list based on alternate platform information specified in associated file queue
	DI_FLAGSEX_RESTART_DEVICE_ONLY      DI_FLAGSEX = 0x20000000 // only restart the device drivers are being installed on as opposed to restarting all devices using those drivers.
	DI_FLAGSEX_RECURSIVESEARCH          DI_FLAGSEX = 0x40000000 // Tell SetupDiBuildDriverInfoList to do a recursive search
	DI_FLAGSEX_SEARCH_PUBLISHED_INFS    DI_FLAGSEX = 0x80000000 // Tell SetupDiBuildDriverInfoList to do a "published INF" search
)

// ClassInstallHeader is the first member of any class install parameters structure. It contains the device installation request code that defines the format of the rest of the install parameters structure.
type ClassInstallHeader struct {
	size            uint32
	InstallFunction DI_FUNCTION
}

func MakeClassInstallHeader(installFunction DI_FUNCTION) *ClassInstallHeader {
	hdr := &ClassInstallHeader{InstallFunction: installFunction}
	hdr.size = uint32(unsafe.Sizeof(*hdr))
	return hdr
}

// DICS_STATE specifies values indicating a change in a device's state
type DICS_STATE uint32

const (
	DICS_ENABLE     DICS_STATE = 0x00000001 // The device is being enabled.
	DICS_DISABLE    DICS_STATE = 0x00000002 // The device is being disabled.
	DICS_PROPCHANGE DICS_STATE = 0x00000003 // The properties of the device have changed.
	DICS_START      DICS_STATE = 0x00000004 // The device is being started (if the request is for the currently active hardware profile).
	DICS_STOP       DICS_STATE = 0x00000005 // The device is being stopped. The driver stack will be unloaded and the CSCONFIGFLAG_DO_NOT_START flag will be set for the device.
)

// DICS_FLAG specifies the scope of a device property change
type DICS_FLAG uint32

const (
	DICS_FLAG_GLOBAL         DICS_FLAG = 0x00000001 // make change in all hardware profiles
	DICS_FLAG_CONFIGSPECIFIC DICS_FLAG = 0x00000002 // make change in specified profile only
	DICS_FLAG_CONFIGGENERAL  DICS_FLAG = 0x00000004 // 1 or more hardware profile-specific changes to follow (obsolete)
)

// PropChangeParams is a structure corresponding to a DIF_PROPERTYCHANGE install function.
type PropChangeParams struct {
	ClassInstallHeader ClassInstallHeader
	StateChange        DICS_STATE
	Scope              DICS_FLAG
	HwProfile          uint32
}

// DI_REMOVEDEVICE specifies the scope of the device removal
type DI_REMOVEDEVICE uint32

const (
	DI_REMOVEDEVICE_GLOBAL         DI_REMOVEDEVICE = 0x00000001 // Make this change in all hardware profiles. Remove information about the device from the registry.
	DI_REMOVEDEVICE_CONFIGSPECIFIC DI_REMOVEDEVICE = 0x00000002 // Make this change to only the hardware profile specified by HwProfile. this flag only applies to root-enumerated devices. When Windows removes the device from the last hardware profile in which it was configured, Windows performs a global removal.
)

// RemoveDeviceParams is a structure corresponding to a DIF_REMOVE install function.
type RemoveDeviceParams struct {
	ClassInstallHeader ClassInstallHeader
	Scope              DI_REMOVEDEVICE
	HwProfile          uint32
}

// DrvInfoData is driver information structure (member of a driver info list that may be associated with a particular device instance, or (globally) with a device information set)
type DrvInfoData struct {
	size          uint32
	DriverType    uint32
	_             uintptr
	description   [LINE_LEN]uint16
	mfgName       [LINE_LEN]uint16
	providerName  [LINE_LEN]uint16
	DriverDate    windows.Filetime
	DriverVersion uint64
}

func (data *DrvInfoData) GetDescription() string {
	return windows.UTF16ToString(data.description[:])
}

func (data *DrvInfoData) SetDescription(description string) error {
	str, err := syscall.UTF16FromString(description)
	if err != nil {
		return err
	}
	copy(data.description[:], str)
	return nil
}

func (data *DrvInfoData) GetMfgName() string {
	return windows.UTF16ToString(data.mfgName[:])
}

func (data *DrvInfoData) SetMfgName(mfgName string) error {
	str, err := syscall.UTF16FromString(mfgName)
	if err != nil {
		return err
	}
	copy(data.mfgName[:], str)
	return nil
}

func (data *DrvInfoData) GetProviderName() string {
	return windows.UTF16ToString(data.providerName[:])
}

func (data *DrvInfoData) SetProviderName(providerName string) error {
	str, err := syscall.UTF16FromString(providerName)
	if err != nil {
		return err
	}
	copy(data.providerName[:], str)
	return nil
}

// IsNewer method returns true if DrvInfoData date and version is newer than supplied parameters.
func (data *DrvInfoData) IsNewer(driverDate windows.Filetime, driverVersion uint64) bool {
	if data.DriverDate.HighDateTime > driverDate.HighDateTime {
		return true
	}
	if data.DriverDate.HighDateTime < driverDate.HighDateTime {
		return false
	}

	if data.DriverDate.LowDateTime > driverDate.LowDateTime {
		return true
	}
	if data.DriverDate.LowDateTime < driverDate.LowDateTime {
		return false
	}

	if data.DriverVersion > driverVersion {
		return true
	}
	if data.DriverVersion < driverVersion {
		return false
	}

	return false
}

// DrvInfoDetailData is driver information details structure (provides detailed information about a particular driver information structure)
type DrvInfoDetailData struct {
	size            uint32 // On input, this must be exactly the sizeof(DrvInfoDetailData). On output, we set this member to the actual size of structure data.
	InfDate         windows.Filetime
	compatIDsOffset uint32
	compatIDsLength uint32
	_               uintptr
	sectionName     [LINE_LEN]uint16
	infFileName     [windows.MAX_PATH]uint16
	drvDescription  [LINE_LEN]uint16
	hardwareID      [1]uint16
}

func (data *DrvInfoDetailData) GetSectionName() string {
	return windows.UTF16ToString(data.sectionName[:])
}

func (data *DrvInfoDetailData) GetInfFileName() string {
	return windows.UTF16ToString(data.infFileName[:])
}

func (data *DrvInfoDetailData) GetDrvDescription() string {
	return windows.UTF16ToString(data.drvDescription[:])
}

func (data *DrvInfoDetailData) GetHardwareID() string {
	if data.compatIDsOffset > 1 {
		bufW := data.getBuf()
		return windows.UTF16ToString(bufW[:wcslen(bufW)])
	}

	return ""
}

func (data *DrvInfoDetailData) GetCompatIDs() []string {
	a := make([]string, 0)

	if data.compatIDsLength > 0 {
		bufW := data.getBuf()
		bufW = bufW[data.compatIDsOffset : data.compatIDsOffset+data.compatIDsLength]
		for i := 0; i < len(bufW); {
			j := i + wcslen(bufW[i:])
			if i < j {
				a = append(a, windows.UTF16ToString(bufW[i:j]))
			}
			i = j + 1
		}
	}

	return a
}

func (data *DrvInfoDetailData) getBuf() []uint16 {
	length := (data.size - uint32(unsafe.Offsetof(data.hardwareID))) / 2
	sl := struct {
		addr *uint16
		len  int
		cap  int
	}{&data.hardwareID[0], int(length), int(length)}
	return *(*[]uint16)(unsafe.Pointer(&sl))
}

// IsCompatible method tests if given hardware ID matches the driver or is listed on the compatible ID list.
func (data *DrvInfoDetailData) IsCompatible(hwid string) bool {
	hwidLC := strings.ToLower(hwid)
	if strings.EqualFold(data.GetHardwareID(), hwidLC) {
		return true
	}
	a := data.GetCompatIDs()
	for i := range a {
		if strings.EqualFold(a[i], hwidLC) {
			return true
		}
	}

	return false
}

// DICD flags control SetupDiCreateDeviceInfo
type DICD uint32

const (
	DICD_GENERATE_ID       DICD = 0x00000001
	DICD_INHERIT_CLASSDRVS DICD = 0x00000002
)

// SPDIT flags to distinguish between class drivers and
// device drivers.
// (Passed in 'DriverType' parameter of driver information list APIs)
type SPDIT uint32

const (
	SPDIT_NODRIVER     SPDIT = 0x00000000
	SPDIT_CLASSDRIVER  SPDIT = 0x00000001
	SPDIT_COMPATDRIVER SPDIT = 0x00000002
)

// DIGCF flags control what is included in the device information set built by SetupDiGetClassDevs
type DIGCF uint32

const (
	DIGCF_DEFAULT         DIGCF = 0x00000001 // only valid with DIGCF_DEVICEINTERFACE
	DIGCF_PRESENT         DIGCF = 0x00000002
	DIGCF_ALLCLASSES      DIGCF = 0x00000004
	DIGCF_PROFILE         DIGCF = 0x00000008
	DIGCF_DEVICEINTERFACE DIGCF = 0x00000010
)

// DIREG specifies values for SetupDiCreateDevRegKey, SetupDiOpenDevRegKey, and SetupDiDeleteDevRegKey.
type DIREG uint32

const (
	DIREG_DEV  DIREG = 0x00000001 // Open/Create/Delete device key
	DIREG_DRV  DIREG = 0x00000002 // Open/Create/Delete driver key
	DIREG_BOTH DIREG = 0x00000004 // Delete both driver and Device key
)

// SPDRP specifies device registry property codes
// (Codes marked as read-only (R) may only be used for
// SetupDiGetDeviceRegistryProperty)
//
// These values should cover the same set of registry properties
// as defined by the CM_DRP codes in cfgmgr32.h.
//
// Note that SPDRP codes are zero based while CM_DRP codes are one based!
type SPDRP uint32

const (
	SPDRP_DEVICEDESC                  SPDRP = 0x00000000 // DeviceDesc (R/W)
	SPDRP_HARDWAREID                  SPDRP = 0x00000001 // HardwareID (R/W)
	SPDRP_COMPATIBLEIDS               SPDRP = 0x00000002 // CompatibleIDs (R/W)
	SPDRP_SERVICE                     SPDRP = 0x00000004 // Service (R/W)
	SPDRP_CLASS                       SPDRP = 0x00000007 // Class (R--tied to ClassGUID)
	SPDRP_CLASSGUID                   SPDRP = 0x00000008 // ClassGUID (R/W)
	SPDRP_DRIVER                      SPDRP = 0x00000009 // Driver (R/W)
	SPDRP_CONFIGFLAGS                 SPDRP = 0x0000000A // ConfigFlags (R/W)
	SPDRP_MFG                         SPDRP = 0x0000000B // Mfg (R/W)
	SPDRP_FRIENDLYNAME                SPDRP = 0x0000000C // FriendlyName (R/W)
	SPDRP_LOCATION_INFORMATION        SPDRP = 0x0000000D // LocationInformation (R/W)
	SPDRP_PHYSICAL_DEVICE_OBJECT_NAME SPDRP = 0x0000000E // PhysicalDeviceObjectName (R)
	SPDRP_CAPABILITIES                SPDRP = 0x0000000F // Capabilities (R)
	SPDRP_UI_NUMBER                   SPDRP = 0x00000010 // UiNumber (R)
	SPDRP_UPPERFILTERS                SPDRP = 0x00000011 // UpperFilters (R/W)
	SPDRP_LOWERFILTERS                SPDRP = 0x00000012 // LowerFilters (R/W)
	SPDRP_BUSTYPEGUID                 SPDRP = 0x00000013 // BusTypeGUID (R)
	SPDRP_LEGACYBUSTYPE               SPDRP = 0x00000014 // LegacyBusType (R)
	SPDRP_BUSNUMBER                   SPDRP = 0x00000015 // BusNumber (R)
	SPDRP_ENUMERATOR_NAME             SPDRP = 0x00000016 // Enumerator Name (R)
	SPDRP_SECURITY                    SPDRP = 0x00000017 // Security (R/W, binary form)
	SPDRP_SECURITY_SDS                SPDRP = 0x00000018 // Security (W, SDS form)
	SPDRP_DEVTYPE                     SPDRP = 0x00000019 // Device Type (R/W)
	SPDRP_EXCLUSIVE                   SPDRP = 0x0000001A // Device is exclusive-access (R/W)
	SPDRP_CHARACTERISTICS             SPDRP = 0x0000001B // Device Characteristics (R/W)
	SPDRP_ADDRESS                     SPDRP = 0x0000001C // Device Address (R)
	SPDRP_UI_NUMBER_DESC_FORMAT       SPDRP = 0x0000001D // UiNumberDescFormat (R/W)
	SPDRP_DEVICE_POWER_DATA           SPDRP = 0x0000001E // Device Power Data (R)
	SPDRP_REMOVAL_POLICY              SPDRP = 0x0000001F // Removal Policy (R)
	SPDRP_REMOVAL_POLICY_HW_DEFAULT   SPDRP = 0x00000020 // Hardware Removal Policy (R)
	SPDRP_REMOVAL_POLICY_OVERRIDE     SPDRP = 0x00000021 // Removal Policy Override (RW)
	SPDRP_INSTALL_STATE               SPDRP = 0x00000022 // Device Install State (R)
	SPDRP_LOCATION_PATHS              SPDRP = 0x00000023 // Device Location Paths (R)
	SPDRP_BASE_CONTAINERID            SPDRP = 0x00000024 // Base ContainerID (R)

	SPDRP_MAXIMUM_PROPERTY SPDRP = 0x00000025 // Upper bound on ordinals
)

type DEVPROPKEY struct {
	fmtid windows.GUID
	pid   uint32
}

//
// This property describes the hardware support for interrupts on the device.
// It is a bitmask of supported interrupt types, with the following values:
//

// #define DevProp_PciDevice_InterruptType_LineBased       1
// #define DevProp_PciDevice_InterruptType_Msi             2
// #define DevProp_PciDevice_InterruptType_MsiX            4

var DEVPKEY_PciDevice_InterruptSupport = DEVPROPKEY{
	windows.GUID{Data1: 0x3ab22e31, Data2: 0x8264, Data3: 0x4b4e, Data4: [8]byte{0x9a, 0xf5, 0xa8, 0xd2, 0xd8, 0xe3, 0x3e, 0x62}},
	14,
}

//
// This property describes the number of message interrupts a device supports
// in hardware.  This property is only valid if the device supports message
// interrupts.
//

var DEVPKEY_PciDevice_InterruptMessageMaximum = DEVPROPKEY{
	windows.GUID{Data1: 0x3ab22e31, Data2: 0x8264, Data3: 0x4b4e, Data4: [8]byte{0x9a, 0xf5, 0xa8, 0xd2, 0xd8, 0xe3, 0x3e, 0x62}},
	15,
}
