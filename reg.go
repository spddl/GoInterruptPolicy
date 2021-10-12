package main

import (
	"log"

	"golang.org/x/sys/windows/registry"
)

func clen(n []byte) int {
	for i := len(n) - 1; i >= 0; i-- {
		if n[i] != 0 {
			return i + 1
		}
	}
	return len(n)
}

func GetStringValue(key registry.Key, name string) string {
	value, _, err := key.GetStringValue(name)
	if err != nil {
		return ""
	}
	return value
}

func GetBinaryValue(key registry.Key, name string) []byte {
	value, _, err := key.GetBinaryValue(name)
	if err != nil {
		return []byte{}
	}
	return value
}

func GetDWORDuint32Value(key registry.Key, name string) uint32 {
	buf := make([]byte, 4)
	key.GetValue(name, buf)
	return btoi32(buf)
}

func setMSIMode(item *Device) {
	var k registry.Key
	var err error
	if item.MsiSupported == 1 {
		k, _, err = registry.CreateKey(item.reg, `Interrupt Management\MessageSignaledInterruptProperties`, registry.ALL_ACCESS)
		if err != nil {
			log.Println(err)
		}
		if err := k.SetDWordValue("MSISupported", 1); err != nil {
			log.Println(err)
		}

		if item.MessageNumberLimit != 0 {
			if err := k.SetDWordValue("MessageNumberLimit", uint32(item.MessageNumberLimit)); err != nil {
				log.Println(err)
			}
		}
	} else {
		k, err = registry.OpenKey(item.reg, `Interrupt Management\MessageSignaledInterruptProperties`, registry.ALL_ACCESS)
		if err != nil {
			log.Println(err)
		}

		if err := registry.DeleteKey(item.reg, `Interrupt Management\MessageSignaledInterruptProperties`); err != nil {
			log.Println(err)
		}
	}
	if err := k.Close(); err != nil {
		log.Println(err)
	}
}

func setAffinityPolicy(item *Device) {
	var k registry.Key
	var err error

	if item.DevicePolicy == 0 && item.DevicePriority == 0 {

		k, err = registry.OpenKey(item.reg, `Interrupt Management\Affinity Policy`, registry.ALL_ACCESS)
		if err != nil {
			log.Println(err)
		}

		if err := registry.DeleteKey(item.reg, `Interrupt Management\Affinity Policy`); err != nil {
			log.Println(err)
		}

	} else {

		k, _, err = registry.CreateKey(item.reg, `Interrupt Management\Affinity Policy`, registry.ALL_ACCESS)
		if err != nil {
			log.Println(err)
		}

		if err := k.SetDWordValue("DevicePolicy", item.DevicePolicy); err != nil {
			log.Println(err)
		}

		if item.DevicePolicy != 4 {
			k.DeleteValue("AssignmentSetOverride")
		}

		if item.DevicePriority == 0 {
			k.DeleteValue("DevicePriority")
		} else if err := k.SetDWordValue("DevicePriority", item.DevicePriority); err != nil {
			log.Println(err)
		}

		AssignmentSetOverrideByte := i64tob(uint64(item.AssignmentSetOverride))
		if err := k.SetBinaryValue("AssignmentSetOverride", AssignmentSetOverrideByte[:clen(AssignmentSetOverrideByte)]); err != nil {
			log.Println(err)
		}

	}
	if err := k.Close(); err != nil {
		log.Println(err)
	}
}
