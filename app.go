package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/lxn/walk"

	//lint:ignore ST1001 standard behavior lxn/walk
	. "github.com/lxn/walk/declarative"
)

func main() {
	devices := FindAllDevices()

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
		Layout: VBox{
			MarginsZero: true,
			SpacingZero: true,
		},
		Children: []Widget{

			TableView{
				OnItemActivated:  mw.lb_ItemActivated,
				Name:             "tableView", // Name is needed for settings persistence
				AlternatingRowBG: true,
				ColumnsOrderable: true,
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
						Name:  "InterrupTypeMap",
						Title: "Interrup Type",
						FormatFunc: func(value interface{}) string {
							return interrupType(value.(Bits))

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
		log.Fatal(err)
	}

	var maxDeviceDesc int
	for i := range devices {
		newDeviceDesc := mw.TextWidthSize(devices[i].DeviceDesc)
		if maxDeviceDesc < newDeviceDesc {
			maxDeviceDesc = newDeviceDesc
		}
	}
	if 150 < maxDeviceDesc {
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
	if orgItem.MsiSupported != newItem.MsiSupported {
		setMSIMode(newItem)
		mw.sbi.SetText("Restart required")
	}

	if orgItem.DevicePolicy != newItem.DevicePolicy || orgItem.DevicePriority != newItem.DevicePriority || orgItem.AssignmentSetOverride != newItem.AssignmentSetOverride {
		setAffinityPolicy(newItem)
		mw.sbi.SetText("Restart required")
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
