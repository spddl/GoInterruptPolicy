package main

import (
	"encoding/binary"
	"fmt"
	"log"

	"golang.org/x/sys/windows/registry"
)

// https://stackoverflow.com/a/13657822/4509632
func Uvarint(buf []byte) (x uint64) {
	for i, b := range buf {
		x = x<<8 + uint64(b)
		if i == 7 {
			return
		}
	}
	return
}

func intToByte(i int) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
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

func GetDWORDuint16Value(key registry.Key, name string) uint16 {
	buf := make([]byte, 4)
	key.GetValue(name, buf)
	return binary.LittleEndian.Uint16(buf)
}

func GetDWORDHexValue(key registry.Key, name string) string {
	buf := make([]byte, 4)
	key.GetValue(name, buf)
	return fmt.Sprintf("%#02x", binary.LittleEndian.Uint16(buf))
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
	} else {
		k, err = registry.OpenKey(item.reg, `Interrupt Management\MessageSignaledInterruptProperties`, registry.ALL_ACCESS)
		if err != nil {
			log.Println(err)
		}
		err = registry.DeleteKey(item.reg, `Interrupt Management\MessageSignaledInterruptProperties`)
		if err != nil {
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
	if item.DevicePolicy == -1 {
		k, err = registry.OpenKey(item.reg, `Interrupt Management\Affinity Policy`, registry.ALL_ACCESS)
		if err != nil {
			log.Println(err)
		}
		err = registry.DeleteKey(item.reg, `Interrupt Management\Affinity Policy`)
		if err != nil {
			log.Println(err)
		}
	} else {
		k, _, err = registry.CreateKey(item.reg, `Interrupt Management\Affinity Policy`, registry.ALL_ACCESS)
		if err != nil {
			log.Println(err)
		}

		if err := k.SetDWordValue("DevicePolicy", uint32(item.DevicePolicy)); err != nil {
			log.Println(err)
		}
		if err := k.SetDWordValue("DevicePriority", uint32(item.DevicePriority)); err != nil {
			log.Println(err)
		}

		if err := k.SetBinaryValue("AssignmentSetOverride", intToByte(int(item.AssignmentSetOverrideBits))); err != nil {
			log.Println(err)
		}

	}
	if err := k.Close(); err != nil {
		log.Println(err)
	}
}
