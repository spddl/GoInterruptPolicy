package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/tailscale/walk"

	//lint:ignore ST1001 standard behavior tailscale/walk
	. "github.com/tailscale/walk/declarative"
)

var cs CpuSets

func main() {
	cs.Init()

	var devices []Device
	devices, handle = FindAllDevices()

	if CLIMode {
		var newItem *Device
		for i := 0; i < len(devices); i++ {
			if devices[i].DevObjName == flagDevObjName {
				newItem = &devices[i]
				break
			}
		}
		if newItem == nil {
			SetupDiDestroyDeviceInfoList(handle)
			os.Exit(1)
		}
		orgItem := *newItem

		var assignmentSetOverride Bits
		if flagCPU != "" {
			cpuarray := strings.Split(flagCPU, ",")
			for _, val := range cpuarray {
				i, err := strconv.Atoi(val)
				if err != nil {
					log.Println(err)
					continue
				}
				assignmentSetOverride = Set(assignmentSetOverride, CPUBits[i])
			}
		}

		if flagMsiSupported != -1 || flagMessageNumberLimit != -1 {
			if flagMsiSupported != -1 {
				newItem.MsiSupported = uint32(flagMsiSupported)
			}
			if flagMessageNumberLimit != -1 {
				newItem.MessageNumberLimit = uint32(flagMessageNumberLimit)
			}
			setMSIMode(newItem)
		}

		if flagDevicePolicy != -1 || flagDevicePriority != -1 || assignmentSetOverride != ZeroBit {
			if flagDevicePolicy != -1 {
				newItem.DevicePolicy = uint32(flagDevicePolicy)
			}
			if flagDevicePriority != -1 {
				newItem.DevicePriority = uint32(flagDevicePriority)
			}
			if assignmentSetOverride != ZeroBit {
				newItem.AssignmentSetOverride = assignmentSetOverride
			}
			setAffinityPolicy(newItem)
		}

		changed := orgItem.MsiSupported != newItem.MsiSupported || orgItem.MessageNumberLimit != newItem.MessageNumberLimit || orgItem.DevicePolicy != newItem.DevicePolicy || orgItem.DevicePriority != newItem.DevicePriority || orgItem.AssignmentSetOverride != newItem.AssignmentSetOverride
		if flagRestart || (flagRestartOnChange && changed) {
			propChangeParams := PropChangeParams{
				ClassInstallHeader: *MakeClassInstallHeader(DIF_PROPERTYCHANGE),
				StateChange:        DICS_PROPCHANGE,
				Scope:              DICS_FLAG_GLOBAL,
			}

			if err := SetupDiSetClassInstallParams(handle, &newItem.Idata, &propChangeParams.ClassInstallHeader, uint32(unsafe.Sizeof(propChangeParams))); err != nil {
				log.Println(err)
				return
			}

			if err := SetupDiCallClassInstaller(DIF_PROPERTYCHANGE, handle, &newItem.Idata); err != nil {
				log.Println(err)
				return
			}

			if err := SetupDiSetClassInstallParams(handle, &newItem.Idata, &propChangeParams.ClassInstallHeader, uint32(unsafe.Sizeof(propChangeParams))); err != nil {
				log.Println(err)
				return
			}

			if err := SetupDiCallClassInstaller(DIF_PROPERTYCHANGE, handle, &newItem.Idata); err != nil {
				log.Println(err)
				return
			}

			DeviceInstallParams, err := SetupDiGetDeviceInstallParams(handle, &newItem.Idata)
			if err != nil {
				log.Println(err)
				return
			}

			if DeviceInstallParams.Flags&DI_NEEDREBOOT != 0 { // DI_NEEDREBOOT
				fmt.Println("Device could not be restarted. Changes will take effect the next time you reboot.")
			} else {
				fmt.Println("Device successfully restarted.")
			}
		}
		SetupDiDestroyDeviceInfoList(handle)
		os.Exit(0)
	}
	defer SetupDiDestroyDeviceInfoList(handle)

	// Sortiert das Array nach Namen
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].DeviceDesc < devices[j].DeviceDesc
	})

	mw := &MyMainWindow{
		model: &Model{items: devices},
		tv:    &walk.TableView{},
	}

	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "GoInterruptPolicy",
		MinSize: Size{
			Width:  240,
			Height: 320,
		},
		Size: Size{
			Width:  750,
			Height: 600,
		},
		Background: SolidColorBrush{Color: walk.RGB(32, 32, 32)},
		Layout: VBox{
			MarginsZero: true,
			SpacingZero: true,
		},
		Children: []Widget{

			TableView{
				OnItemActivated:     mw.lb_ItemActivated,
				Name:                "tableView", // Name is needed for settings persistence
				AlternatingRowBG:    true,
				ColumnsOrderable:    true,
				ColumnsSizable:      true,
				LastColumnStretched: true,
				Columns: []TableViewColumn{
					{
						Name:  "DeviceDesc",
						Title: "Name",
						Width: 150,
					},
					{
						Name:  "FriendlyName",
						Title: "Friendly Name",
					},
					{
						Name:  "LocationInformation",
						Title: "Location Info",
						Width: 150,
					},
					{
						Name:      "MsiSupported",
						Title:     "MSI Mode",
						Width:     60,
						Alignment: AlignCenter,
						FormatFunc: func(value interface{}) string {
							if value.(uint32) == 0 {
								return "✖"
							} else if value.(uint32) == 1 {
								return "✔"
							}
							return ""
						},
					},
					{
						Name:  "DevicePolicy",
						Title: "Device Policy",
						FormatFunc: func(value interface{}) string {
							// https://docs.microsoft.com/en-us/windows-hardware/drivers/kernel/interrupt-affinity-and-priority
							switch value.(uint32) {
							case IrqPolicyMachineDefault: // 0x00
								return "Default"
							case IrqPolicyAllCloseProcessors: // 0x01
								return "All Close Proc"
							case IrqPolicyOneCloseProcessor: // 0x02
								return "One Close Proc"
							case IrqPolicyAllProcessorsInMachine: // 0x03
								return "All Proc in Machine"
							case IrqPolicySpecifiedProcessors: // 0x04
								return "Specified Proc"
							case IrqPolicySpreadMessagesAcrossAllProcessors: // 0x05
								return "Spread Messages Across All Proc"
							default:
								return fmt.Sprintf("%d", value.(uint32))
							}
						},
					},
					{
						Name:  "AssignmentSetOverride",
						Title: "Specified Processor",
						FormatFunc: func(value interface{}) string {
							if value == ZeroBit {
								return ""
							}
							bits := value.(Bits)
							var result []string
							for bit, cpu := range CPUMap {
								if Has(bit, bits) {
									result = append(result, cpu)
								}
							}

							result, err := sortNumbers(result)
							if err != nil {
								log.Println(err)
							}
							return strings.Join(result, ",")
						},
						LessFunc: func(i, j int) bool {
							return mw.model.items[i].AssignmentSetOverride < mw.model.items[j].AssignmentSetOverride
						},
					},
					{
						Name:  "DevicePriority",
						Title: "Device Priority",
						FormatFunc: func(value interface{}) string {
							switch value.(uint32) {
							case 0:
								return "Undefined"
							case 1:
								return "Low"
							case 2:
								return "Normal"
							case 3:
								return "High"
							default:
								return fmt.Sprintf("%d", value.(uint32))
							}
						},
					},
					{
						Name:  "InterruptTypeMap",
						Title: "Interrupt Type",
						Width: 120,
						FormatFunc: func(value interface{}) string {
							return interruptType(value.(Bits))
						},
						LessFunc: func(i, j int) bool {
							return mw.model.items[i].InterruptTypeMap < mw.model.items[j].InterruptTypeMap
						},
					},
					{
						Name:  "MessageNumberLimit",
						Title: "MSI Limit",
						FormatFunc: func(value interface{}) string {
							switch value.(uint32) {
							case 0:
								return ""
							default:
								return fmt.Sprintf("%d", value.(uint32))
							}
						},
					},
					{
						Name:  "MaxMSILimit",
						Title: "Max MSI Limit",
						FormatFunc: func(value interface{}) string {
							switch value.(uint32) {
							case 0:
								return ""
							default:
								return fmt.Sprintf("%d", value.(uint32))
							}
						},
					},
					{
						Name:  "DevObjName",
						Title: "DevObj Name",
					},
					{
						Name:  "LastChange",
						Title: "Last Change",
						Width: 130,
						FormatFunc: func(value interface{}) string {
							if value.(time.Time).IsZero() {
								return "N/A"
							} else {
								return value.(time.Time).Format("2006-01-02 15:04:05")
							}
						},
					},
				},
				Model:    mw.model,
				AssignTo: &mw.tv,
				OnKeyUp: func(key walk.Key) {
					i := mw.tv.CurrentIndex()
					if i == -1 {
						i = 0
					}
					for ; i < len(mw.model.items); i++ {
						item := &mw.model.items[i]
						if item.DeviceDesc != "" && key.String() == item.DeviceDesc[0:1] {
							err := mw.tv.SetCurrentIndex(i)
							if err != nil {
								log.Println(err)
							}
							return
						}
					}
				},
			},
		},
		StatusBarItems: []StatusBarItem{
			{
				AssignTo: &mw.sbi,
			},
		},
	}).Create(); err != nil {
		log.Println(err)
		return
	}

	var maxDeviceDesc int
	for i := range devices {
		newDeviceDesc := mw.TextWidthSize(devices[i].DeviceDesc)
		if maxDeviceDesc < newDeviceDesc {
			maxDeviceDesc = newDeviceDesc
		}
	}
	if maxDeviceDesc < 150 {
		mw.tv.Columns().At(0).SetWidth(maxDeviceDesc)
	}

	mw.Show()
	mw.Run()
}

type MyMainWindow struct {
	*walk.MainWindow
	tv    *walk.TableView
	model *Model
	sbi   *walk.StatusBarItem
}

func (mw *MyMainWindow) lb_ItemActivated() {
	newItem := &mw.model.items[mw.tv.CurrentIndex()]
	orgItem := *newItem
	result, err := RunDialog(mw, newItem)
	if err != nil {
		log.Print(err)
	}
	if result == 0 || result == 2 { // cancel
		mw.model.items[mw.tv.CurrentIndex()] = orgItem
		return
	}

	if orgItem.MsiSupported != newItem.MsiSupported || orgItem.MessageNumberLimit != newItem.MessageNumberLimit {
		setMSIMode(newItem)
	}

	if orgItem.DevicePolicy != newItem.DevicePolicy || orgItem.DevicePriority != newItem.DevicePriority || orgItem.AssignmentSetOverride != newItem.AssignmentSetOverride {
		setAffinityPolicy(newItem)
	}

	if orgItem.MsiSupported != newItem.MsiSupported || orgItem.MessageNumberLimit != newItem.MessageNumberLimit || orgItem.DevicePolicy != newItem.DevicePolicy || orgItem.DevicePriority != newItem.DevicePriority || orgItem.AssignmentSetOverride != newItem.AssignmentSetOverride {
		if walk.MsgBox(mw.WindowBase.Form(), "Restart Device?", `Your changes will not take effect until the device is restarted.

Would you like to attempt to restart the device now?`, walk.MsgBoxYesNo) == 6 {
			propChangeParams := PropChangeParams{
				ClassInstallHeader: *MakeClassInstallHeader(DIF_PROPERTYCHANGE),
				StateChange:        DICS_PROPCHANGE,
				Scope:              DICS_FLAG_GLOBAL,
			}

			if err := SetupDiSetClassInstallParams(handle, &newItem.Idata, &propChangeParams.ClassInstallHeader, uint32(unsafe.Sizeof(propChangeParams))); err != nil {
				log.Println(err)
				return
			}

			if err := SetupDiCallClassInstaller(DIF_PROPERTYCHANGE, handle, &newItem.Idata); err != nil {
				log.Println(err)
				return
			}

			if err := SetupDiSetClassInstallParams(handle, &newItem.Idata, &propChangeParams.ClassInstallHeader, uint32(unsafe.Sizeof(propChangeParams))); err != nil {
				log.Println(err)
				return
			}

			if err := SetupDiCallClassInstaller(DIF_PROPERTYCHANGE, handle, &newItem.Idata); err != nil {
				log.Println(err)
				return
			}

			DeviceInstallParams, err := SetupDiGetDeviceInstallParams(handle, &newItem.Idata)
			if err != nil {
				log.Println(err)
				return
			}

			if DeviceInstallParams.Flags&DI_NEEDREBOOT != 0 {
				walk.MsgBox(mw.WindowBase.Form(), "Notice", "Device could not be restarted. Changes will take effect the next time you reboot.", walk.MsgBoxOK)
			} else {
				walk.MsgBox(mw.WindowBase.Form(), "Notice", "Device successfully restarted.", walk.MsgBoxOK)
			}

		} else {
			mw.sbi.SetText("Restart required")
		}
	}
}

func (mw *MyMainWindow) TextWidthSize(text string) int {
	canvas, err := (*mw.tv).CreateCanvas()
	if err != nil {
		return 0
	}
	defer canvas.Dispose()

	bounds, _, err := canvas.MeasureTextPixels(text, (*mw.tv).Font(), walk.Rectangle{Width: 9999999}, walk.TextCalcRect)
	if err != nil {
		return 0
	}

	return bounds.Size().Width
}

type Model struct {
	walk.SortedReflectTableModelBase
	items []Device
}

func (m *Model) Items() interface{} {
	return m.items
}

func sortNumbers(data []string) ([]string, error) {
	var lastErr error
	sort.Slice(data, func(i, j int) bool {
		a, err := strconv.ParseInt(data[i], 10, 64)
		if err != nil {
			lastErr = err
			return false
		}
		b, err := strconv.ParseInt(data[j], 10, 64)
		if err != nil {
			lastErr = err
			return false
		}
		return a < b
	})
	return data, lastErr
}
