package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	modSetupapi = windows.NewLazyDLL("setupapi.dll")

	procSetupDiGetClassDevsW = modSetupapi.NewProc("SetupDiGetClassDevsW")
)

const CONFIG_FLAG_DISABLED uint32 = 1

func FindAllDevices() ([]Device, DevInfo) {
	var allDevices []Device
	handle, err := SetupDiGetClassDevs(nil, nil, 0, uint32(DIGCF_ALLCLASSES|DIGCF_PRESENT))
	if err != nil {
		panic(err)
	}

	var index = 0
	for {
		idata, err := SetupDiEnumDeviceInfo(handle, index)
		if err != nil { // ERROR_NO_MORE_ITEMS
			break
		}
		index++

		dev := Device{
			Idata: *idata,
		}

		val, err := SetupDiGetDeviceRegistryProperty(handle, idata, SPDRP_CONFIGFLAGS)
		if err == nil {
			if val.(uint32)&CONFIG_FLAG_DISABLED != 0 {
				// Sorts out deactivated devices
				continue
			}
		}

		val, err = SetupDiGetDeviceRegistryProperty(handle, idata, SPDRP_DEVICEDESC)
		if err == nil {
			if val.(string) == "" {
				continue
			}
			dev.DeviceDesc = val.(string)
		} else {
			continue
		}

		valProp, err := GetDeviceProperty(handle, idata, DEVPKEY_PciDevice_InterruptSupport)
		if err == nil {
			dev.InterruptTypeMap = Bits(btoi16(valProp))
		}

		valProp, err = GetDeviceProperty(handle, idata, DEVPKEY_PciDevice_InterruptMessageMaximum)
		if err == nil {
			dev.MaxMSILimit = btoi32(valProp)
		}

		val, err = SetupDiGetDeviceRegistryProperty(handle, idata, SPDRP_FRIENDLYNAME)
		if err == nil {
			dev.FriendlyName = val.(string)
		}

		val, err = SetupDiGetDeviceRegistryProperty(handle, idata, SPDRP_PHYSICAL_DEVICE_OBJECT_NAME)
		if err == nil {
			dev.DevObjName = val.(string)
		}

		val, err = SetupDiGetDeviceRegistryProperty(handle, idata, SPDRP_LOCATION_INFORMATION)
		if err == nil {
			dev.LocationInformation = val.(string)
		}

		dev.reg, _ = SetupDiOpenDevRegKey(handle, idata, DICS_FLAG_GLOBAL, 0, DIREG_DEV, windows.KEY_SET_VALUE)

		keyinfo, err := dev.reg.Stat()
		if err == nil {
			dev.LastChange = keyinfo.ModTime()
		}

		affinityPolicyKey, _ := registry.OpenKey(dev.reg, `Interrupt Management\Affinity Policy`, registry.QUERY_VALUE)
		dev.DevicePolicy = GetDWORDuint32Value(affinityPolicyKey, "DevicePolicy")               // REG_DWORD
		dev.DevicePriority = GetDWORDuint32Value(affinityPolicyKey, "DevicePriority")           // REG_DWORD
		AssignmentSetOverrideByte := GetBinaryValue(affinityPolicyKey, "AssignmentSetOverride") // REG_BINARY
		affinityPolicyKey.Close()

		if len(AssignmentSetOverrideByte) != 0 {
			AssignmentSetOverrideBytes := make([]byte, 8)
			copy(AssignmentSetOverrideBytes, AssignmentSetOverrideByte)
			dev.AssignmentSetOverride = Bits(btoi64(AssignmentSetOverrideBytes))
		}

		if dev.InterruptTypeMap != ZeroBit {
			messageSignaledInterruptPropertiesKey, _ := registry.OpenKey(dev.reg, `Interrupt Management\MessageSignaledInterruptProperties`, registry.QUERY_VALUE)
			dev.MessageNumberLimit = GetDWORDuint32Value(messageSignaledInterruptPropertiesKey, "MessageNumberLimit") // REG_DWORD https://docs.microsoft.com/de-de/windows-hardware/drivers/kernel/enabling-message-signaled-interrupts-in-the-registry
			dev.MsiSupported = GetDWORDuint32Value(messageSignaledInterruptPropertiesKey, "MSISupported")             // REG_DWORD
			messageSignaledInterruptPropertiesKey.Close()
		} else {
			dev.MsiSupported = 2 // invalid
		}

		allDevices = append(allDevices, dev)
	}
	return allDevices, handle
}

func SetupDiGetClassDevs(classGuid *windows.GUID, enumerator *uint16, hwndParent uintptr, flags uint32) (handle DevInfo, err error) {
	r0, _, e1 := syscall.SyscallN(procSetupDiGetClassDevsW.Addr(), uintptr(unsafe.Pointer(classGuid)), uintptr(unsafe.Pointer(enumerator)), uintptr(hwndParent), uintptr(flags))
	handle = DevInfo(r0)
	if handle == DevInfo(windows.InvalidHandle) {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func GetDeviceProperty(dis DevInfo, devInfoData *DevInfoData, devPropKey DEVPROPKEY) ([]byte, error) {
	var propt, size uint32
	buf := make([]byte, 16)
	run := true
	for run {
		err := SetupDiGetDeviceProperty(dis, devInfoData, &devPropKey, &propt, &buf[0], uint32(len(buf)), &size, 0)
		switch {
		case size > uint32(len(buf)):
			buf = make([]byte, size+16)
		case err != nil:
			return buf, err
		default:
			run = false
		}
	}

	return buf, nil
}
