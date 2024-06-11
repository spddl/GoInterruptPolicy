package main

import (
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/tailscale/walk"
	"golang.org/x/sys/windows/registry"

	//lint:ignore ST1001 standard behavior tailscale/walk
	. "github.com/tailscale/walk/declarative"
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
	var deviceMessageNumberLimitNE *walk.NumberEdit

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
		Layout: VBox{
			MarginsZero: true,
		},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Composite{
						Layout: Grid{
							Columns: 2,
						},
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
							LineEdit{
								Text:     Bind("device.DevObjName == '' ? 'N/A' : device.DevObjName"),
								ReadOnly: true,
							},
						},
					},

					GroupBox{
						Title:   "Message Signaled-Based Interrupts",
						Visible: device.MsiSupported != 2,
						Layout:  Grid{Columns: 1},
						Children: []Widget{

							CheckBox{
								Name:           "MsiSupported",
								Text:           "MSI Mode:",
								TextOnLeftSide: true,
								Checked:        device.MsiSupported == 1,
								// Checked: Bind("eq(device.MsiSupported, 1, 'MsiSupported')"), // dont work, idk
								OnClicked: func() {
									if device.MsiSupported == 0 {
										device.MsiSupported = 1
										deviceMessageNumberLimitNE.SetEnabled(true)
										device.MessageNumberLimit = uint32(deviceMessageNumberLimitNE.Value())
									} else {
										device.MsiSupported = 0
										deviceMessageNumberLimitNE.SetEnabled(false)
									}
								},
							},

							Composite{
								Layout: Grid{
									Columns:     3,
									MarginsZero: true,
								},
								Children: []Widget{
									LinkLabel{
										Text: `MSI Limit: <a href="https://forums.guru3d.com/threads/windows-line-based-vs-message-signaled-based-interrupts-msi-tool.378044/">?</a>`,
										OnLinkActivated: func(link *walk.LinkLabelLink) {
											// https://stackoverflow.com/a/12076082
											exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", link.URL()).Start()
										},
									},
									NumberEdit{
										SpinButtonsVisible: true,
										AssignTo:           &deviceMessageNumberLimitNE,
										Enabled:            device.MsiSupported == 1,
										MinValue:           1,
										MaxValue:           hasMsiX(device.InterruptTypeMap),
										Value:              Bind("device.MessageNumberLimit < 1.0 ? 1.0 : device.MessageNumberLimit"),
										OnValueChanged: func() {
											device.MessageNumberLimit = uint32(deviceMessageNumberLimitNE.Value())
										},
									},
								},
							},

							Label{
								Text: "Interrupt Type: " + interruptType(device.InterruptTypeMap),
							},

							Label{
								Text: Bind("device.MaxMSILimit == 0 ? '' : 'Max MSI Limit: ' + device.MaxMSILimit"),
							},
						},
					},

					GroupBox{
						Title:  "Advanced Policies",
						Layout: Grid{Columns: 2},
						Children: []Widget{
							Composite{
								Layout: Grid{
									Columns:     2,
									MarginsZero: true,
								},
								Children: []Widget{
									LinkLabel{
										Text: `Device Priority: <a href="https://docs.microsoft.com/en-us/windows-hardware/drivers/ddi/miniport/ne-miniport-_irq_priority">?</a>`,
										OnLinkActivated: func(link *walk.LinkLabelLink) {
											// https://stackoverflow.com/a/12076082
											exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", link.URL()).Start()
										},
									},
									ComboBox{
										AssignTo:      &devicePriorityCB,
										Value:         Bind("device.DevicePriority"),
										BindingMember: "Enums",
										DisplayMember: "Name",
										Model:         IrqPriority(),
										OnCurrentIndexChanged: func() {
											device.DevicePriority = uint32(devicePriorityCB.CurrentIndex())
										},
									},

									LinkLabel{
										Text: `Device Policy: <a href="https://docs.microsoft.com/en-us/windows-hardware/drivers/ddi/miniport/ne-miniport-_irq_device_policy">?</a>`,
										OnLinkActivated: func(link *walk.LinkLabelLink) {
											// https://stackoverflow.com/a/12076082
											exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", link.URL()).Start()
										},
									},
									ComboBox{
										AssignTo:      &devicePolicyCB,
										Value:         Bind("device.DevicePolicy"),
										BindingMember: "Enums",
										DisplayMember: "Name",
										Model:         IrqPolicy(),
										OnCurrentIndexChanged: func() {
											device.DevicePolicy = uint32(devicePolicyCB.CurrentIndex())
											if device.DevicePolicy == 4 {
												cpuArrayCom.SetVisible(true)
											} else {
												cpuArrayCom.SetVisible(false)
											}
										},
									},
								},
							},

							Composite{
								Alignment: AlignHCenterVCenter,
								AssignTo:  &cpuArrayCom,
								Visible:   Bind("device.DevicePolicy == 4"),
								Layout:    Grid{Columns: 2},
								Children:  CheckBoxList(CPUArray, &device.AssignmentSetOverride),
							},
						},
					},

					GroupBox{
						Title:  "Registry",
						Layout: HBox{},
						Children: []Widget{
							PushButton{
								Text: "Open Device",
								OnClicked: func() {
									regPath, err := GetRegistryLocation(uintptr(device.reg))
									if err != nil {
										walk.MsgBox(dlg, "NtQueryKey Error", err.Error(), walk.MsgBoxOK)
									}

									k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Applets\Regedit`, registry.SET_VALUE)
									if err != nil {
										log.Fatal(err)
									}
									defer k.Close()

									if err := k.SetStringValue("LastKey", regPath); err == nil {
										exec.Command("regedit", "-m").Start()
									}
								},
							},
							PushButton{
								Text: "Export current settings",
								OnClicked: func() {
									regPath, err := GetRegistryLocation(uintptr(device.reg))
									if err != nil {
										walk.MsgBox(dlg, "NtQueryKey Error", err.Error(), walk.MsgBoxOK)
									}

									path, err := os.Getwd()
									if err != nil {
										log.Println(err)
									}

									filePath, cancel, err := saveFileExplorer(dlg, path, strings.ReplaceAll(device.DeviceDesc, " ", "_")+".reg", "Save current settings", "Registry File (*.reg)|*.reg")
									if !cancel || err != nil {
										file, err := os.Create(filePath)
										if err != nil {
											return
										}
										defer file.Close()

										file.WriteString(createRegFile(regPath, device))
									}
								},
							},
							HSpacer{},
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
						if args[0].(Bits) == ZeroBit {
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
			VSpacer{},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							if device.DevicePolicy == 4 && device.AssignmentSetOverride == ZeroBit {
								walk.MsgBox(dlg, "Invalid Option", "The affinity mask must contain at least one processor.", walk.MsgBoxIconError)
							} else {
								if err := db.Submit(); err != nil {
									return
								}
								dlg.Accept()
							}
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
			Text:     "CPU " + name,
			Checked:  Has(*bits, local_CPUBits),
			OnClicked: func() {
				*bits = Toggle(local_CPUBits, *bits)
			},
		})
	}
	return children
}

// https://docs.microsoft.com/de-de/windows-hardware/drivers/kernel/enabling-message-signaled-interrupts-in-the-registry
func hasMsiX(b Bits) float64 {
	if Has(b, Bits(4)) {
		return 2048 // MSIX
	} else {
		return 16 // MSI
	}
}

func interruptType(b Bits) string {
	if b == ZeroBit {
		return ""
	}
	var types []string
	for bit, name := range InterruptTypeMap {
		if Has(b, bit) {
			types = append(types, name)
		}
	}
	sort.Strings(types)
	return strings.Join(types, ", ")
}
