package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// 08 ae
// 04 d2 Port 1234
// 14 20 Old Checksum
//const hexStr = "45 00 00 20 | 48 e9 40 00 | 40 11 dd e1 | 0a 00 00 01 | 0a 00 00 02"
const hexStr = "cd f2 04 d2 00 0c 00 00 00 00 00 00"

const pseudoHdr = "0a 00 00 01 0a 00 00 02 00 11 00 0c"

func main() {
	hexBytesPseudoHdr, err := hex.DecodeString(strings.ReplaceAll(pseudoHdr, " ", ""))
	if err != nil {
		panic(err)
	}

	hexBytesUDP, err := hex.DecodeString(strings.ReplaceAll(hexStr, " ", ""))
	if err != nil {
		panic(err)
	}

	allBytes := append(hexBytesPseudoHdr, hexBytesUDP...)

	fmt.Printf("Bytes: %v\n", allBytes)

	if len(allBytes)%2 != 0 {
		hexBytesUDP = append(allBytes, 0)
	}

	csum := uint32(0)
	for i := 0; i < len(allBytes); i = i + 2 {
		csum += uint32(binary.BigEndian.Uint16(allBytes[i:]))
	}
	csum += csum >> 16
	csum &= 0xffff

	csum = ^csum
	csum16 := uint16(csum)

	fmt.Printf("Csum: %#x\n", csum16)

}
