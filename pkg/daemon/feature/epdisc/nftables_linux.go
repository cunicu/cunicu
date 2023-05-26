// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import "encoding/binary"

type NftablesUserDataType byte

const (
	NftablesUserDataTypeComment NftablesUserDataType = iota
	NftablesUserDataTypeRuleID  NftablesUserDataType = 100 // custom extension
)

func NftablesUserDataPut(udata []byte, typ NftablesUserDataType, data []byte) []byte {
	udata = append(udata, byte(typ), byte(len(data)))
	udata = append(udata, data...)

	return udata
}

func NftablesUserDataGet(udata []byte, styp NftablesUserDataType) []byte {
	for {
		if len(udata) < 2 {
			break
		}

		typ := NftablesUserDataType(udata[0])
		length := int(udata[1])
		data := udata[2 : 2+length]

		if styp == typ {
			return data
		}

		if len(udata) < 2+length {
			break
		}

		udata = udata[2+length:]
	}

	return nil
}

func NftablesUserDataPutInt(udata []byte, typ NftablesUserDataType, num uint32) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, num)

	return NftablesUserDataPut(udata, typ, data)
}

func NftablesUserDataGetInt(udata []byte, typ NftablesUserDataType) (uint32, bool) {
	data := NftablesUserDataGet(udata, typ)
	if data == nil {
		return 0, false
	}

	return binary.LittleEndian.Uint32(data), true
}

func NftablesUserDataPutString(udata []byte, typ NftablesUserDataType, str string) []byte {
	data := append([]byte(str), 0)
	return NftablesUserDataPut(udata, typ, data)
}

func NftablesUserDataGetString(udata []byte, typ NftablesUserDataType) (string, bool) {
	data := NftablesUserDataGet(udata, typ)
	if data == nil {
		return "", false
	}

	return string(data), true
}
