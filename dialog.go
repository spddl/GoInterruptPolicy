package main

import (
	"fmt"
	"log"
	"math"
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

type CheckBoxList struct {
	Widget []Widget
	List   []*walk.CheckBox
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

	var cpuArrayComView *walk.Composite

	var devicePolicyCB, devicePriorityCB *walk.ComboBox
	var deviceMessageNumberLimitNE *walk.NumberEdit
	var checkBoxList = new(CheckBoxList)

	return Dialog{
		AssignTo:      &dlg,
		Title:         Bind("'Device Policy' + (device.DeviceDesc == '' ? '' : ' - ' + device.DeviceDesc)"),
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		FixedSize:     true,
		DataBinder: DataBinder{
			AssignTo:       &db,
			Name:           "device",
			DataSource:     device,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
		Layout: VBox{
			MarginsZero: true,
		},
		Children: []Widget{
			Composite{
				Layout: VBox{},
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
										MinValue:           0,
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
						Layout: VBox{},
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
											currentIndex := uint32(devicePolicyCB.CurrentIndex())
											if device.DevicePolicy == currentIndex {
												return
											}

											device.DevicePolicy = currentIndex
											if device.DevicePolicy == 4 {
												cpuArrayComView.SetVisible(true)
												return
											}

											cpuArrayComView.SetVisible(false)

											if err := dlg.SetSize(walk.Size{Width: 0, Height: 0}); err != nil {
												panic(err)
											}

										},
									},
								},
							},

							Composite{
								AssignTo: &cpuArrayComView,
								Layout:   VBox{MarginsZero: true},
								Visible:  Bind("device.DevicePolicy == 4"),
								Children: []Widget{
									Composite{
										Alignment: AlignHNearVNear,
										Layout: HBox{
											Alignment:   Alignment2D(walk.AlignHNearVNear),
											MarginsZero: true,
										},
										Children: checkBoxList.create(&device.AssignmentSetOverride),
									},
									GroupBox{
										Title:  "Presets for Specified Processors:",
										Layout: HBox{},
										Children: []Widget{
											PushButton{
												Text: "All On",
												OnClicked: func() {
													checkBoxList.allOn(&device.AssignmentSetOverride)
												},
											},

											PushButton{
												Text: "All Off",
												OnClicked: func() {
													checkBoxList.allOff(&device.AssignmentSetOverride)
												},
											},

											PushButton{
												Text:    "HT Off",
												Visible: cs.HyperThreading,
												OnClicked: func() {
													checkBoxList.htOff(&device.AssignmentSetOverride)
												},
											},

											PushButton{
												Text:    "P-Core Only",
												Visible: cs.EfficiencyClass,
												OnClicked: func() {
													checkBoxList.pCoreOnly(&device.AssignmentSetOverride)
												},
											},

											PushButton{
												Text:    "E-Core Only",
												Visible: cs.EfficiencyClass,
												OnClicked: func() {
													checkBoxList.eCoreOnly(&device.AssignmentSetOverride)
												},
											},

											HSpacer{},
										},
									},
								},
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

func (checkboxlist *CheckBoxList) create(bits *Bits) []Widget {
	var children, partThread, partCore, partNUMA, partGroup, partCache []Widget
	var lastEfficiencyClass, lastNumaNodeIndex, lastLastLevelCache byte
	var cpuCount, numaCount, llcCount int

	checkboxlist.List = make([]*walk.CheckBox, len(cs.CPU))
	for i, cpuThread := range cs.CPU {
		// fmt.Printf("\n\n%d - Core: %d, LogicalProc: %d, Effi: %d\n", i,cpu.CoreIndex, cs.CPU[i].LogicalProcessorIndex, cs.CPU[i].EfficiencyClass)

		var isLastItem = i+1 == cs.CoreCount
		if i == 0 { // The EfficiencyClass starts with 1 on the Intel Gen12+
			lastEfficiencyClass = cpuThread.EfficiencyClass
		}

		if len(partThread) != 0 && cpuThread.CoreIndex == cpuThread.LogicalProcessorIndex {
			partCore = append(partCore, GroupBox{
				Title:     fmt.Sprintf("Core %d", cpuCount),
				Alignment: AlignHCenterVNear,
				Layout: VBox{
					Margins: CalculateMargins(len(partThread)),
				},
				Children: partThread,
			})
			cpuCount++
			partThread = nil
		}

		local_CPUBits := CPUBits[i]
		checkboxlist.List[i] = new(walk.CheckBox)

		partThread = append(partThread, CheckBox{
			ColumnSpan:         3,
			RowSpan:            3,
			AlwaysConsumeSpace: true,
			StretchFactor:      2,
			Text:               fmt.Sprintf("Thread %d", cpuThread.LogicalProcessorIndex),
			AssignTo:           &checkboxlist.List[i],
			Checked:            Has(*bits, local_CPUBits),
			OnClicked: func() {
				*bits = Toggle(local_CPUBits, *bits)
			},
		})

		if isLastItem {
			partCore = append(partCore, GroupBox{
				Title:     fmt.Sprintf("Core %d", cpuCount),
				Alignment: AlignHCenterVNear,
				Layout: VBox{
					Margins: CalculateMargins(len(partThread)),
				},
				Children: partThread,
			})
		}

		if i == 0 {
			continue
		}

		if cs.EfficiencyClass && (lastEfficiencyClass != cpuThread.EfficiencyClass || isLastItem) {
			var title string
			// https://learn.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-system_cpu_set_information

			if isIntel() && cs.EfficiencyClass {
				if lastEfficiencyClass == 0 {
					title = "E-Cores"
				} else {
					title = "P-Cores"
				}
			} else { // AMD
				title = fmt.Sprintf("EfficiencyClass %d", lastEfficiencyClass)
			}

			partNUMA = append(partNUMA, GroupBox{
				Title:       title,
				ToolTipText: ToolTipTextEfficiencyClass,
				Layout: Grid{
					Columns: len(partCore) / cs.Layout.Rows,
				},
				Children: partCore,
			})

			partCore = nil
			cpuCount = 0
		}

		if cs.LastLevelCache && (lastLastLevelCache != cpuThread.LastLevelCacheIndex || isLastItem) {
			var title string
			if isAMD() {
				title = fmt.Sprintf("CCD %d", llcCount)
			} else {
				title = fmt.Sprintf("LLC %d", llcCount)
			}

			switch {
			case len(partNUMA) != 0:
				partGroup = append(partGroup, GroupBox{
					Title:       title,
					ToolTipText: ToolTipTextLastLevelCache,
					Layout: Grid{
						Columns: len(partNUMA) / cs.Layout.Rows,
					},
					Children: partNUMA,
				})
				partNUMA = nil
			case len(partCore) != 0:
				partGroup = append(partGroup, GroupBox{
					Title:       title,
					ToolTipText: ToolTipTextLastLevelCache,
					Layout: Grid{
						Columns: len(partCore) / cs.Layout.Rows,
					},
					Children: partCore,
				})
				llcCount++
				partCore = nil
			}
		}

		if cs.NumaNode && (lastNumaNodeIndex != cpuThread.NumaNodeIndex || isLastItem) {
			switch {
			case len(partGroup) != 0:
				partCache = append(partCache, GroupBox{
					Title:       fmt.Sprintf("NUMA %d", numaCount),
					ToolTipText: ToolTipTextNumaNode,
					Layout: Grid{
						Columns: len(partGroup) / cs.Layout.Rows,
					},
					Children: partGroup,
				})
				partGroup = nil
			case len(partNUMA) != 0:
				partGroup = append(partGroup, GroupBox{
					Title:       fmt.Sprintf("NUMA %d", numaCount),
					ToolTipText: ToolTipTextNumaNode,
					Layout: Grid{
						Columns: len(partNUMA) / cs.Layout.Rows,
					},
					Children: partNUMA,
				})
				partNUMA = nil
			case len(partCore) != 0:
				partGroup = append(partGroup, GroupBox{
					Title:       fmt.Sprintf("NUMA %d", numaCount),
					ToolTipText: ToolTipTextNumaNode,
					Layout: Grid{
						Columns: len(partCore) / cs.Layout.Rows,
					},
					Children: partCore,
				})
				partCore = nil
			}
			numaCount++
		}

		lastEfficiencyClass = cpuThread.EfficiencyClass
		lastNumaNodeIndex = cpuThread.NumaNodeIndex
		lastLastLevelCache = cpuThread.LastLevelCacheIndex
	}

	switch {
	case len(partCache) > 0:
		return partCache
	case len(partGroup) > 0:
		return partGroup
	case len(partNUMA) > 0:
		return partNUMA
	case len(partCore) > 0:
		return []Widget{
			Composite{
				Layout: Grid{
					Columns: len(partCore) / cs.Layout.Rows,
				},
				Children: partCore,
			},
		}
	default:
		return children
	}
}

func (checkboxlist *CheckBoxList) allOn(bits *Bits) {
	for i := 0; i < len(checkboxlist.List); i++ {
		*bits = Set(CPUBits[i], *bits)
		checkboxlist.List[i].SetChecked(true)
	}
}
func (checkboxlist *CheckBoxList) allOff(bits *Bits) {
	for i := 0; i < len(checkboxlist.List); i++ {
		checkboxlist.List[i].SetChecked(false)
	}
	*bits = Bits(0)
}

func (checkboxlist *CheckBoxList) htOff(bits *Bits) {
	for i := 0; i < len(cs.CPU); i++ {
		if cs.CPU[i].CoreIndex != cs.CPU[i].LogicalProcessorIndex {
			checkboxlist.List[i].SetChecked(false)
			if Has(CPUBits[i], *bits) {
				*bits = Toggle(CPUBits[i], *bits)
			}
		}
	}
}

func (checkboxlist *CheckBoxList) pCoreOnly(bits *Bits) {
	for i := 0; i < len(cs.CPU); i++ {
		if cs.CPU[i].EfficiencyClass == 1 {
			checkboxlist.List[i].SetChecked(true)
			*bits = Set(CPUBits[i], *bits)
		} else {
			checkboxlist.List[i].SetChecked(false)
			if Has(CPUBits[i], *bits) {
				*bits = Toggle(CPUBits[i], *bits)
			}
		}
	}
}
func (checkboxlist *CheckBoxList) eCoreOnly(bits *Bits) {
	for i := 0; i < len(cs.CPU); i++ {
		if cs.CPU[i].EfficiencyClass == 0 {
			checkboxlist.List[i].SetChecked(true)
			*bits = Set(CPUBits[i], *bits)
		} else {
			checkboxlist.List[i].SetChecked(false)
			if Has(CPUBits[i], *bits) {
				*bits = Toggle(CPUBits[i], *bits)
			}
		}
	}
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

func CalculateMargins(value int) Margins {
	if cs.MaxThreadsPerCore+1 == value {
		return Margins{
			Left:   9,
			Top:    9,
			Right:  9,
			Bottom: 9,
		}
	} else {
		part := (11.75 * float64(cs.MaxThreadsPerCore+1) / float64(value))
		return Margins{
			Left:   9,
			Top:    int(math.Floor(part)),
			Right:  9,
			Bottom: int(math.Ceil(part)),
		}
	}
}
