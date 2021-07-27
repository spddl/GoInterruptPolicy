package main

import (
	"fmt"
	"log"
	"sort"

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
	if _, err := (MainWindow{
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
					{Name: "Index", Frozen: true, Title: " ", Width: 30, Hidden: true},
					{Name: "DeviceDesc", Title: "Name", Width: 150},
					{Name: "FriendlyName", Title: "Friendly Name"},
					{Name: "LocationInformation", Title: "Location Info", Width: 150},
					{
						Name:      "MsiSupported",
						Title:     "MSI Mode",
						Alignment: AlignCenter,
						FormatFunc: func(value interface{}) string {
							if value.(int32) == 0 {
								return "✖"
							} else {
								return "✔"
							}
						},
					},
					{
						Name: "DevicePolicy",
						FormatFunc: func(value interface{}) string {
							// https://docs.microsoft.com/en-us/windows-hardware/drivers/kernel/interrupt-affinity-and-priority
							switch value.(int32) {
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
								return fmt.Sprintf("%d", value.(int))
							}
						},
					},
					{
						Name: "DevicePriority",
						FormatFunc: func(value interface{}) string {
							switch value.(int32) {
							case 0:
								return "Undefined"
							case 1:
								return "Low"
							case 2:
								return "Normal"
							case 3:
								return "High"
							default:
								return fmt.Sprintf("%d", value.(int))
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
	}.Run()); err != nil {
		log.Fatal(err)
	}
}

type MyMainWindow struct {
	*walk.MainWindow
	tv    *walk.TableView
	model *Model
	sbi   *walk.StatusBarItem
}

func (mw *MyMainWindow) lb_ItemActivated() {
	item := &mw.model.items[mw.tv.CurrentIndex()]
	orgItem := *item
	_, err := RunDialog(mw, item)
	if err != nil {
		log.Print(err)
	}

	if orgItem.MsiSupported != item.MsiSupported {
		setMSIMode(item)
		mw.sbi.SetText("Restart required")
	}

	if orgItem.DevicePolicy != item.DevicePolicy || orgItem.DevicePriority != item.DevicePriority || orgItem.AssignmentSetOverride != item.AssignmentSetOverride {
		setAffinityPolicy(item)
		mw.sbi.SetText("Restart required")
	}
}

type Model struct {
	walk.SortedReflectTableModelBase
	items []Device
}

func (m *Model) Items() interface{} {
	return m.items
}
