package main

import (
	"log"
	"strings"

	"github.com/lxn/walk"
	//lint:ignore ST1001 standard behavior lxn/walk
	. "github.com/lxn/walk/declarative"
)

type IrqPolicys struct {
	Enums float64
	Name  string
}

type IrqPrioritys struct {
	Enums float64
	Name  string
}

func IrqPolicy() []*IrqPolicys {
	return []*IrqPolicys{
		{IrqPolicyMachineDefault, "IrqPolicyMachineDefault"},
		{IrqPolicyAllCloseProcessors, "IrqPolicyAllCloseProcessors"},
		{IrqPolicyOneCloseProcessor, "IrqPolicyOneCloseProcessor"},
		{IrqPolicyAllProcessorsInMachine, "IrqPolicyAllProcessorsInMachine"},
		{IrqPolicySpecifiedProcessors, "IrqPolicySpecifiedProcessors"},
		{IrqPolicySpreadMessagesAcrossAllProcessors, "IrqPolicySpreadMessagesAcrossAllProcessors"},
	}
}

func IrqPriority() []*IrqPrioritys {
	return []*IrqPrioritys{
		{0, "Undefined"},
		{1, "Low"},
		{2, "Normal"},
		{3, "High"},
	}
}

func RunDialog(owner walk.Form, device *Device) (int, error) {
	var dlg *walk.Dialog
	var db *walk.DataBinder
	var acceptPB, cancelPB *walk.PushButton
	var cpuArrayCom *walk.Composite
	var devicePolicyCB, devicePriorityCB *walk.ComboBox

	return Dialog{
		AssignTo:      &dlg,
		Title:         Bind("'Device Policy' + (device.DeviceDesc == '' ? '' : ' - ' + device.DeviceDesc)"),
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		DataBinder: DataBinder{
			AssignTo:       &db,
			Name:           "device",
			DataSource:     device,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
		MinSize: Size{
			Width:  300,
			Height: 300,
		},
		Layout: VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 2},
						Children: []Widget{
							Label{
								Text: "Name:",
							},
							Label{
								Text: Bind("device.DeviceDesc == '' ? 'N/A' : device.DeviceDesc"),
							},

							Label{
								Text: "Location Info:",
							},
							Label{
								Text: Bind("device.LocationInformation == '' ? 'N/A' : device.LocationInformation"),
							},

							Label{
								Text: "DevObj Name:",
							},
							Label{
								Text: Bind("device.DevObjName == '' ? 'N/A' : device.DevObjName"),
							},

							CheckBox{
								ColumnSpan:     2,
								Name:           "MsiSupported",
								Text:           "MSI Mode:",
								TextOnLeftSide: true,
								Checked:        device.MsiSupported == 1,
								// Checked: Bind("eq(device.MsiSupported, 1, 'MsiSupported')"), // dont work, idk
								OnClicked: func() {
									if device.MsiSupported == 0 {
										device.MsiSupported = 1
									} else {
										device.MsiSupported = 0
									}
								},
							},
						},
					},

					GroupBox{
						Title:  "Advanced Policies",
						Layout: Grid{Rows: 3},
						Children: []Widget{

							ComboBox{
								AssignTo:      &devicePriorityCB,
								Value:         Bind("device.DevicePriority"),
								BindingMember: "Enums",
								DisplayMember: "Name",
								Model:         IrqPriority(),
								OnCurrentIndexChanged: func() {
									device.DevicePriority = devicePriorityCB.CurrentIndex()
								},
							},

							ComboBox{
								AssignTo:      &devicePolicyCB,
								Value:         Bind("device.DevicePolicy"),
								BindingMember: "Enums",
								DisplayMember: "Name",
								Model:         IrqPolicy(),
								OnCurrentIndexChanged: func() {
									device.DevicePolicy = devicePolicyCB.CurrentIndex()
									if device.DevicePolicy == 4 {
										cpuArrayCom.SetEnabled(true)
									} else {
										cpuArrayCom.SetEnabled(false)
									}
								},
							},

							Composite{
								AssignTo: &cpuArrayCom,
								// Enabled: Bind("eq(device.DevicePolicy, 4)"),
								Enabled:  Bind("device.DevicePolicy == 4"),
								Layout:   Grid{Columns: 2, Alignment: AlignHFarVNear},
								Children: CheckBoxList(CPUArray, &device.AssignmentSetOverrideBits),
								// Invalid Option
								// TODO: The affinity mask must contain at least one processor.
							},
						},
					},
				},
				Functions: map[string]func(args ...interface{}) (interface{}, error){
					"checkIrqPolicy": func(args ...interface{}) (interface{}, error) {
						for _, v := range IrqPolicy() {
							if v.Enums == args[0].(float64) {
								return v.Name, nil
							}
						}
						return "", nil
					},
					"viewAsHex": func(args ...interface{}) (interface{}, error) {
						if args[0].(Bits) == Bits(0) {
							return "N/A", nil
						}
						bits := args[0].(Bits)
						var result []string
						for bit, cpu := range CPUMap {
							if Has(bit, bits) {
								result = append(result, cpu)
							}
						}
						return strings.Join(result, ", "), nil
					},
					"eq": func(args ...interface{}) (interface{}, error) {
						if len(args) != 2 {
							log.Println("len(args) != 2")
							return false, nil
						}
						switch v := args[0].(type) {
						case float64:
							if v == args[1].(float64) {
								return true, nil
							}
						case Bits:
							if v == Bits(args[1].(float64)) {
								return true, nil
							}
						default:
							log.Printf("I don't know about type %T!\n", v)
						}

						return false, nil
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							if err := db.Submit(); err != nil {
								return
							}
							dlg.Accept()
						},
					},
					PushButton{
						AssignTo:  &cancelPB,
						Text:      "Cancel",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Run(owner)
}

func CheckBoxList(names []string, bits *Bits) []Widget {
	bs := make([]*walk.CheckBox, len(names))
	children := []Widget{}
	for i, name := range names {
		bs[i] = new(walk.CheckBox)
		local_CPUBits := CPUBits[i]
		children = append(children, CheckBox{
			AssignTo: &bs[i],
			Text:     name,
			Checked:  Has(*bits, local_CPUBits),
			OnClicked: func() {
				*bits = Toggle(local_CPUBits, *bits)
			},
		})
	}
	return children
}
