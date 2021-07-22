package main

// import (
// 	"bytes"
// 	"log"
// 	"strings"
// 	"text/template"
// )

// func createInfoText(item *Device) string {

// 	// fmt.Sprintf("item.DeviceIDs: %s\r\nDevicePolicy: %s\r\n", item.DeviceIDs, item.AffinityPolicy.DevicePolicy)

// 	const templateText = `IrqPolicy: {{.IrqPolicy}}
// DeviceDesc: {{.DeviceDesc}}
// DeviceIDs: {{.DeviceIDs}}
// DevObjName: {{.DevObjName}}
// Driver: {{.Driver}}
// LocationInformation: {{checkEmpty .LocationInformation}}
// Regpath: {{.Regpath}}
// FriendlyName: {{.FriendlyName}}

// // AffinityPolicy
// DevicePolicy: {{.DevicePolicy}}
// DevicePriority: {{.DevicePriority}}
// AssignmentSetOverride: {{.AssignmentSetOverride}}
// AssignmentSetOverrideBits: {{.AssignmentSetOverrideBits}}

// // MessageSignaledInterruptProperties
// MessageNumberLimit: {{checkEmpty .MessageNumberLimit}}
// MsiSupported: {{.MsiSupported}}`

// 	t, err := template.New("").Funcs(template.FuncMap{
// 		"checkEmpty": checkEmpty,
// 	}).Parse(templateText)
// 	if err != nil {
// 		log.Fatalf("parsing: %s", err)
// 	}

// 	var tpl bytes.Buffer
// 	if err := t.Execute(&tpl, item); err != nil {
// 		log.Fatalf("Execute: %s", err)
// 	}

// 	return strings.ReplaceAll(tpl.String(), "\n", "\r\n")
// }

// func checkEmpty(value string) string { // https://golang.org/pkg/html/template/#Template.Funcs
// 	if value == "" {
// 		return "N/A"
// 	}
// 	return value
// }
