package main

import (
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

// https://stackoverflow.com/a/27834860/4509632
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// https://github.com/golang/go/blob/master/src/encoding/binary/binary.go#L130
func littleEndian_PutUint64(v uint64) []byte {
	b := make([]byte, 8)
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
	return b
}

func littleEndian_Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[0]) | uint16(b[1])<<8
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
	return uint32(littleEndian_Uint16(buf))
}

func GetDWORDHexValue(key registry.Key, name string) string {
	buf := make([]byte, 4)
	key.GetValue(name, buf)
	return fmt.Sprintf("%#02x", littleEndian_Uint16(buf))
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

		if item.MaxMSILimit != 0 {
			if err := k.SetDWordValue("MessageNumberLimit", uint32(item.MaxMSILimit)); err != nil {
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

		AssignmentSetOverrideByte := littleEndian_PutUint64(uint64(item.AssignmentSetOverride))
		if err := k.SetBinaryValue("AssignmentSetOverride", AssignmentSetOverrideByte[:clen(AssignmentSetOverrideByte)]); err != nil {
			log.Println(err)
		}

	}
	if err := k.Close(); err != nil {
		log.Println(err)
	}
}
