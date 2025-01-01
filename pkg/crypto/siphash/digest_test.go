// SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-FileCopyrightText: 2012-2014 Dmitry Chestnykh
// SPDX-License-Identifier: Apache-2.0

package siphash_test

import (
	"encoding/binary"
	"testing"

	"cunicu.li/cunicu/pkg/crypto/siphash"
	"cunicu.li/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Siphash Suite")
}

var _ = Describe("Siphash", func() {
	zeroKey := make([]byte, 16)

	golden := []struct {
		k []byte
		m []byte
		r uint64
	}{
		{
			[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
			[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e},
			0xa129ca6149be45e5,
		},
		{
			zeroKey,
			[]byte("Hello world"),
			0xc9e8a3021f3822d9,
		},
		{
			zeroKey,
			[]byte{}, // zero-length message
			0x1e924b9d737700d7,
		},
		{
			zeroKey,
			[]byte("12345678123"),
			0xf95d77ccdb0649f,
		},
		{
			zeroKey,
			make([]byte, 8),
			0xe849e8bb6ffe2567,
		},
		{
			zeroKey,
			make([]byte, 1535),
			0xe74d1c0ab64b2afa,
		},
	}

	// Test vectors from reference implementation.
	//
	// SipHash-2-4 output with
	// k = 00 01 02 ...
	// and
	// in = (empty string)
	// in = 00 (1 byte)
	// in = 00 01 (2 bytes)
	// in = 00 01 02 (3 bytes)
	// ...
	// in = 00 01 02 ... 3e (63 bytes)
	goldenRef := [][]byte{
		{0x31, 0x0e, 0x0e, 0xdd, 0x47, 0xdb, 0x6f, 0x72},
		{0xfd, 0x67, 0xdc, 0x93, 0xc5, 0x39, 0xf8, 0x74},
		{0x5a, 0x4f, 0xa9, 0xd9, 0x09, 0x80, 0x6c, 0x0d},
		{0x2d, 0x7e, 0xfb, 0xd7, 0x96, 0x66, 0x67, 0x85},
		{0xb7, 0x87, 0x71, 0x27, 0xe0, 0x94, 0x27, 0xcf},
		{0x8d, 0xa6, 0x99, 0xcd, 0x64, 0x55, 0x76, 0x18},
		{0xce, 0xe3, 0xfe, 0x58, 0x6e, 0x46, 0xc9, 0xcb},
		{0x37, 0xd1, 0x01, 0x8b, 0xf5, 0x00, 0x02, 0xab},
		{0x62, 0x24, 0x93, 0x9a, 0x79, 0xf5, 0xf5, 0x93},
		{0xb0, 0xe4, 0xa9, 0x0b, 0xdf, 0x82, 0x00, 0x9e},
		{0xf3, 0xb9, 0xdd, 0x94, 0xc5, 0xbb, 0x5d, 0x7a},
		{0xa7, 0xad, 0x6b, 0x22, 0x46, 0x2f, 0xb3, 0xf4},
		{0xfb, 0xe5, 0x0e, 0x86, 0xbc, 0x8f, 0x1e, 0x75},
		{0x90, 0x3d, 0x84, 0xc0, 0x27, 0x56, 0xea, 0x14},
		{0xee, 0xf2, 0x7a, 0x8e, 0x90, 0xca, 0x23, 0xf7},
		{0xe5, 0x45, 0xbe, 0x49, 0x61, 0xca, 0x29, 0xa1},
		{0xdb, 0x9b, 0xc2, 0x57, 0x7f, 0xcc, 0x2a, 0x3f},
		{0x94, 0x47, 0xbe, 0x2c, 0xf5, 0xe9, 0x9a, 0x69},
		{0x9c, 0xd3, 0x8d, 0x96, 0xf0, 0xb3, 0xc1, 0x4b},
		{0xbd, 0x61, 0x79, 0xa7, 0x1d, 0xc9, 0x6d, 0xbb},
		{0x98, 0xee, 0xa2, 0x1a, 0xf2, 0x5c, 0xd6, 0xbe},
		{0xc7, 0x67, 0x3b, 0x2e, 0xb0, 0xcb, 0xf2, 0xd0},
		{0x88, 0x3e, 0xa3, 0xe3, 0x95, 0x67, 0x53, 0x93},
		{0xc8, 0xce, 0x5c, 0xcd, 0x8c, 0x03, 0x0c, 0xa8},
		{0x94, 0xaf, 0x49, 0xf6, 0xc6, 0x50, 0xad, 0xb8},
		{0xea, 0xb8, 0x85, 0x8a, 0xde, 0x92, 0xe1, 0xbc},
		{0xf3, 0x15, 0xbb, 0x5b, 0xb8, 0x35, 0xd8, 0x17},
		{0xad, 0xcf, 0x6b, 0x07, 0x63, 0x61, 0x2e, 0x2f},
		{0xa5, 0xc9, 0x1d, 0xa7, 0xac, 0xaa, 0x4d, 0xde},
		{0x71, 0x65, 0x95, 0x87, 0x66, 0x50, 0xa2, 0xa6},
		{0x28, 0xef, 0x49, 0x5c, 0x53, 0xa3, 0x87, 0xad},
		{0x42, 0xc3, 0x41, 0xd8, 0xfa, 0x92, 0xd8, 0x32},
		{0xce, 0x7c, 0xf2, 0x72, 0x2f, 0x51, 0x27, 0x71},
		{0xe3, 0x78, 0x59, 0xf9, 0x46, 0x23, 0xf3, 0xa7},
		{0x38, 0x12, 0x05, 0xbb, 0x1a, 0xb0, 0xe0, 0x12},
		{0xae, 0x97, 0xa1, 0x0f, 0xd4, 0x34, 0xe0, 0x15},
		{0xb4, 0xa3, 0x15, 0x08, 0xbe, 0xff, 0x4d, 0x31},
		{0x81, 0x39, 0x62, 0x29, 0xf0, 0x90, 0x79, 0x02},
		{0x4d, 0x0c, 0xf4, 0x9e, 0xe5, 0xd4, 0xdc, 0xca},
		{0x5c, 0x73, 0x33, 0x6a, 0x76, 0xd8, 0xbf, 0x9a},
		{0xd0, 0xa7, 0x04, 0x53, 0x6b, 0xa9, 0x3e, 0x0e},
		{0x92, 0x59, 0x58, 0xfc, 0xd6, 0x42, 0x0c, 0xad},
		{0xa9, 0x15, 0xc2, 0x9b, 0xc8, 0x06, 0x73, 0x18},
		{0x95, 0x2b, 0x79, 0xf3, 0xbc, 0x0a, 0xa6, 0xd4},
		{0xf2, 0x1d, 0xf2, 0xe4, 0x1d, 0x45, 0x35, 0xf9},
		{0x87, 0x57, 0x75, 0x19, 0x04, 0x8f, 0x53, 0xa9},
		{0x10, 0xa5, 0x6c, 0xf5, 0xdf, 0xcd, 0x9a, 0xdb},
		{0xeb, 0x75, 0x09, 0x5c, 0xcd, 0x98, 0x6c, 0xd0},
		{0x51, 0xa9, 0xcb, 0x9e, 0xcb, 0xa3, 0x12, 0xe6},
		{0x96, 0xaf, 0xad, 0xfc, 0x2c, 0xe6, 0x66, 0xc7},
		{0x72, 0xfe, 0x52, 0x97, 0x5a, 0x43, 0x64, 0xee},
		{0x5a, 0x16, 0x45, 0xb2, 0x76, 0xd5, 0x92, 0xa1},
		{0xb2, 0x74, 0xcb, 0x8e, 0xbf, 0x87, 0x87, 0x0a},
		{0x6f, 0x9b, 0xb4, 0x20, 0x3d, 0xe7, 0xb3, 0x81},
		{0xea, 0xec, 0xb2, 0xa3, 0x0b, 0x22, 0xa8, 0x7f},
		{0x99, 0x24, 0xa4, 0x3c, 0xc1, 0x31, 0x57, 0x24},
		{0xbd, 0x83, 0x8d, 0x3a, 0xaf, 0xbf, 0x8d, 0xb7},
		{0x0b, 0x1a, 0x2a, 0x32, 0x65, 0xd5, 0x1a, 0xea},
		{0x13, 0x50, 0x79, 0xa3, 0x23, 0x1c, 0xe6, 0x60},
		{0x93, 0x2b, 0x28, 0x46, 0xe4, 0xd7, 0x06, 0x66},
		{0xe1, 0x91, 0x5f, 0x5c, 0xb1, 0xec, 0xa4, 0x6c},
		{0xf3, 0x25, 0x96, 0x5c, 0xa1, 0x6d, 0x62, 0x9f},
		{0x57, 0x5f, 0xf2, 0x8e, 0x60, 0x38, 0x1b, 0xe5},
		{0x72, 0x45, 0x06, 0xeb, 0x4c, 0x32, 0x8a, 0x95},
	}

	goldenRef128 := [][]byte{
		{0xa3, 0x81, 0x7f, 0x04, 0xba, 0x25, 0xa8, 0xe6, 0x6d, 0xf6, 0x72, 0x14, 0xc7, 0x55, 0x02, 0x93},
		{0xda, 0x87, 0xc1, 0xd8, 0x6b, 0x99, 0xaf, 0x44, 0x34, 0x76, 0x59, 0x11, 0x9b, 0x22, 0xfc, 0x45},
		{0x81, 0x77, 0x22, 0x8d, 0xa4, 0xa4, 0x5d, 0xc7, 0xfc, 0xa3, 0x8b, 0xde, 0xf6, 0x0a, 0xff, 0xe4},
		{0x9c, 0x70, 0xb6, 0x0c, 0x52, 0x67, 0xa9, 0x4e, 0x5f, 0x33, 0xb6, 0xb0, 0x29, 0x85, 0xed, 0x51},
		{0xf8, 0x81, 0x64, 0xc1, 0x2d, 0x9c, 0x8f, 0xaf, 0x7d, 0x0f, 0x6e, 0x7c, 0x7b, 0xcd, 0x55, 0x79},
		{0x13, 0x68, 0x87, 0x59, 0x80, 0x77, 0x6f, 0x88, 0x54, 0x52, 0x7a, 0x07, 0x69, 0x0e, 0x96, 0x27},
		{0x14, 0xee, 0xca, 0x33, 0x8b, 0x20, 0x86, 0x13, 0x48, 0x5e, 0xa0, 0x30, 0x8f, 0xd7, 0xa1, 0x5e},
		{0xa1, 0xf1, 0xeb, 0xbe, 0xd8, 0xdb, 0xc1, 0x53, 0xc0, 0xb8, 0x4a, 0xa6, 0x1f, 0xf0, 0x82, 0x39},
		{0x3b, 0x62, 0xa9, 0xba, 0x62, 0x58, 0xf5, 0x61, 0x0f, 0x83, 0xe2, 0x64, 0xf3, 0x14, 0x97, 0xb4},
		{0x26, 0x44, 0x99, 0x06, 0x0a, 0xd9, 0xba, 0xab, 0xc4, 0x7f, 0x8b, 0x02, 0xbb, 0x6d, 0x71, 0xed},
		{0x00, 0x11, 0x0d, 0xc3, 0x78, 0x14, 0x69, 0x56, 0xc9, 0x54, 0x47, 0xd3, 0xf3, 0xd0, 0xfb, 0xba},
		{0x01, 0x51, 0xc5, 0x68, 0x38, 0x6b, 0x66, 0x77, 0xa2, 0xb4, 0xdc, 0x6f, 0x81, 0xe5, 0xdc, 0x18},
		{0xd6, 0x26, 0xb2, 0x66, 0x90, 0x5e, 0xf3, 0x58, 0x82, 0x63, 0x4d, 0xf6, 0x85, 0x32, 0xc1, 0x25},
		{0x98, 0x69, 0xe2, 0x47, 0xe9, 0xc0, 0x8b, 0x10, 0xd0, 0x29, 0x93, 0x4f, 0xc4, 0xb9, 0x52, 0xf7},
		{0x31, 0xfc, 0xef, 0xac, 0x66, 0xd7, 0xde, 0x9c, 0x7e, 0xc7, 0x48, 0x5f, 0xe4, 0x49, 0x49, 0x02},
		{0x54, 0x93, 0xe9, 0x99, 0x33, 0xb0, 0xa8, 0x11, 0x7e, 0x08, 0xec, 0x0f, 0x97, 0xcf, 0xc3, 0xd9},
		{0x6e, 0xe2, 0xa4, 0xca, 0x67, 0xb0, 0x54, 0xbb, 0xfd, 0x33, 0x15, 0xbf, 0x85, 0x23, 0x05, 0x77},
		{0x47, 0x3d, 0x06, 0xe8, 0x73, 0x8d, 0xb8, 0x98, 0x54, 0xc0, 0x66, 0xc4, 0x7a, 0xe4, 0x77, 0x40},
		{0xa4, 0x26, 0xe5, 0xe4, 0x23, 0xbf, 0x48, 0x85, 0x29, 0x4d, 0xa4, 0x81, 0xfe, 0xae, 0xf7, 0x23},
		{0x78, 0x01, 0x77, 0x31, 0xcf, 0x65, 0xfa, 0xb0, 0x74, 0xd5, 0x20, 0x89, 0x52, 0x51, 0x2e, 0xb1},
		{0x9e, 0x25, 0xfc, 0x83, 0x3f, 0x22, 0x90, 0x73, 0x3e, 0x93, 0x44, 0xa5, 0xe8, 0x38, 0x39, 0xeb},
		{0x56, 0x8e, 0x49, 0x5a, 0xbe, 0x52, 0x5a, 0x21, 0x8a, 0x22, 0x14, 0xcd, 0x3e, 0x07, 0x1d, 0x12},
		{0x4a, 0x29, 0xb5, 0x45, 0x52, 0xd1, 0x6b, 0x9a, 0x46, 0x9c, 0x10, 0x52, 0x8e, 0xff, 0x0a, 0xae},
		{0xc9, 0xd1, 0x84, 0xdd, 0xd5, 0xa9, 0xf5, 0xe0, 0xcf, 0x8c, 0xe2, 0x9a, 0x9a, 0xbf, 0x69, 0x1c},
		{0x2d, 0xb4, 0x79, 0xae, 0x78, 0xbd, 0x50, 0xd8, 0x88, 0x2a, 0x8a, 0x17, 0x8a, 0x61, 0x32, 0xad},
		{0x8e, 0xce, 0x5f, 0x04, 0x2d, 0x5e, 0x44, 0x7b, 0x50, 0x51, 0xb9, 0xea, 0xcb, 0x8d, 0x8f, 0x6f},
		{0x9c, 0x0b, 0x53, 0xb4, 0xb3, 0xc3, 0x07, 0xe8, 0x7e, 0xae, 0xe0, 0x86, 0x78, 0x14, 0x1f, 0x66},
		{0xab, 0xf2, 0x48, 0xaf, 0x69, 0xa6, 0xea, 0xe4, 0xbf, 0xd3, 0xeb, 0x2f, 0x12, 0x9e, 0xeb, 0x94},
		{0x06, 0x64, 0xda, 0x16, 0x68, 0x57, 0x4b, 0x88, 0xb9, 0x35, 0xf3, 0x02, 0x73, 0x58, 0xae, 0xf4},
		{0xaa, 0x4b, 0x9d, 0xc4, 0xbf, 0x33, 0x7d, 0xe9, 0x0c, 0xd4, 0xfd, 0x3c, 0x46, 0x7c, 0x6a, 0xb7},
		{0xea, 0x5c, 0x7f, 0x47, 0x1f, 0xaf, 0x6b, 0xde, 0x2b, 0x1a, 0xd7, 0xd4, 0x68, 0x6d, 0x22, 0x87},
		{0x29, 0x39, 0xb0, 0x18, 0x32, 0x23, 0xfa, 0xfc, 0x17, 0x23, 0xde, 0x4f, 0x52, 0xc4, 0x3d, 0x35},
		{0x7c, 0x39, 0x56, 0xca, 0x5e, 0xea, 0xfc, 0x3e, 0x36, 0x3e, 0x9d, 0x55, 0x65, 0x46, 0xeb, 0x68},
		{0x77, 0xc6, 0x07, 0x71, 0x46, 0xf0, 0x1c, 0x32, 0xb6, 0xb6, 0x9d, 0x5f, 0x4e, 0xa9, 0xff, 0xcf},
		{0x37, 0xa6, 0x98, 0x6c, 0xb8, 0x84, 0x7e, 0xdf, 0x09, 0x25, 0xf0, 0xf1, 0x30, 0x9b, 0x54, 0xde},
		{0xa7, 0x05, 0xf0, 0xe6, 0x9d, 0xa9, 0xa8, 0xf9, 0x07, 0x24, 0x1a, 0x2e, 0x92, 0x3c, 0x8c, 0xc8},
		{0x3d, 0xc4, 0x7d, 0x1f, 0x29, 0xc4, 0x48, 0x46, 0x1e, 0x9e, 0x76, 0xed, 0x90, 0x4f, 0x67, 0x11},
		{0x0d, 0x62, 0xbf, 0x01, 0xe6, 0xfc, 0x0e, 0x1a, 0x0d, 0x3c, 0x47, 0x51, 0xc5, 0xd3, 0x69, 0x2b},
		{0x8c, 0x03, 0x46, 0x8b, 0xca, 0x7c, 0x66, 0x9e, 0xe4, 0xfd, 0x5e, 0x08, 0x4b, 0xbe, 0xe7, 0xb5},
		{0x52, 0x8a, 0x5b, 0xb9, 0x3b, 0xaf, 0x2c, 0x9c, 0x44, 0x73, 0xcc, 0xe5, 0xd0, 0xd2, 0x2b, 0xd9},
		{0xdf, 0x6a, 0x30, 0x1e, 0x95, 0xc9, 0x5d, 0xad, 0x97, 0xae, 0x0c, 0xc8, 0xc6, 0x91, 0x3b, 0xd8},
		{0x80, 0x11, 0x89, 0x90, 0x2c, 0x85, 0x7f, 0x39, 0xe7, 0x35, 0x91, 0x28, 0x5e, 0x70, 0xb6, 0xdb},
		{0xe6, 0x17, 0x34, 0x6a, 0xc9, 0xc2, 0x31, 0xbb, 0x36, 0x50, 0xae, 0x34, 0xcc, 0xca, 0x0c, 0x5b},
		{0x27, 0xd9, 0x34, 0x37, 0xef, 0xb7, 0x21, 0xaa, 0x40, 0x18, 0x21, 0xdc, 0xec, 0x5a, 0xdf, 0x89},
		{0x89, 0x23, 0x7d, 0x9d, 0xed, 0x9c, 0x5e, 0x78, 0xd8, 0xb1, 0xc9, 0xb1, 0x66, 0xcc, 0x73, 0x42},
		{0x4a, 0x6d, 0x80, 0x91, 0xbf, 0x5e, 0x7d, 0x65, 0x11, 0x89, 0xfa, 0x94, 0xa2, 0x50, 0xb1, 0x4c},
		{0x0e, 0x33, 0xf9, 0x60, 0x55, 0xe7, 0xae, 0x89, 0x3f, 0xfc, 0x0e, 0x3d, 0xcf, 0x49, 0x29, 0x02},
		{0xe6, 0x1c, 0x43, 0x2b, 0x72, 0x0b, 0x19, 0xd1, 0x8e, 0xc8, 0xd8, 0x4b, 0xdc, 0x63, 0x15, 0x1b},
		{0xf7, 0xe5, 0xae, 0xf5, 0x49, 0xf7, 0x82, 0xcf, 0x37, 0x90, 0x55, 0xa6, 0x08, 0x26, 0x9b, 0x16},
		{0x43, 0x8d, 0x03, 0x0f, 0xd0, 0xb7, 0xa5, 0x4f, 0xa8, 0x37, 0xf2, 0xad, 0x20, 0x1a, 0x64, 0x03},
		{0xa5, 0x90, 0xd3, 0xee, 0x4f, 0xbf, 0x04, 0xe3, 0x24, 0x7e, 0x0d, 0x27, 0xf2, 0x86, 0x42, 0x3f},
		{0x5f, 0xe2, 0xc1, 0xa1, 0x72, 0xfe, 0x93, 0xc4, 0xb1, 0x5c, 0xd3, 0x7c, 0xae, 0xf9, 0xf5, 0x38},
		{0x2c, 0x97, 0x32, 0x5c, 0xbd, 0x06, 0xb3, 0x6e, 0xb2, 0x13, 0x3d, 0xd0, 0x8b, 0x3a, 0x01, 0x7c},
		{0x92, 0xc8, 0x14, 0x22, 0x7a, 0x6b, 0xca, 0x94, 0x9f, 0xf0, 0x65, 0x9f, 0x00, 0x2a, 0xd3, 0x9e},
		{0xdc, 0xe8, 0x50, 0x11, 0x0b, 0xd8, 0x32, 0x8c, 0xfb, 0xd5, 0x08, 0x41, 0xd6, 0x91, 0x1d, 0x87},
		{0x67, 0xf1, 0x49, 0x84, 0xc7, 0xda, 0x79, 0x12, 0x48, 0xe3, 0x2b, 0xb5, 0x92, 0x25, 0x83, 0xda},
		{0x19, 0x38, 0xf2, 0xcf, 0x72, 0xd5, 0x4e, 0xe9, 0x7e, 0x94, 0x16, 0x6f, 0xa9, 0x1d, 0x2a, 0x36},
		{0x74, 0x48, 0x1e, 0x96, 0x46, 0xed, 0x49, 0xfe, 0x0f, 0x62, 0x24, 0x30, 0x16, 0x04, 0x69, 0x8e},
		{0x57, 0xfc, 0xa5, 0xde, 0x98, 0xa9, 0xd6, 0xd8, 0x00, 0x64, 0x38, 0xd0, 0x58, 0x3d, 0x8a, 0x1d},
		{0x9f, 0xec, 0xde, 0x1c, 0xef, 0xdc, 0x1c, 0xbe, 0xd4, 0x76, 0x36, 0x74, 0xd9, 0x57, 0x53, 0x59},
		{0xe3, 0x04, 0x0c, 0x00, 0xeb, 0x28, 0xf1, 0x53, 0x66, 0xca, 0x73, 0xcb, 0xd8, 0x72, 0xe7, 0x40},
		{0x76, 0x97, 0x00, 0x9a, 0x6a, 0x83, 0x1d, 0xfe, 0xcc, 0xa9, 0x1c, 0x59, 0x93, 0x67, 0x0f, 0x7a},
		{0x58, 0x53, 0x54, 0x23, 0x21, 0xf5, 0x67, 0xa0, 0x05, 0xd5, 0x47, 0xa4, 0xf0, 0x47, 0x59, 0xbd},
		{0x51, 0x50, 0xd1, 0x77, 0x2f, 0x50, 0x83, 0x4a, 0x50, 0x3e, 0x06, 0x9a, 0x97, 0x3f, 0xbd, 0x7c},
	}

	It("Sum64", func() {
		var r [8]byte
		for _, v := range golden {
			binary.LittleEndian.PutUint64(r[:], v.r)
			h := siphash.New64(v.k)
			h.Write(v.m)
			Expect(h.Sum(nil)).To(Equal(r[:]))
		}

		var k [16]byte
		var in [64]byte
		for i := range k {
			k[i] = byte(i)
		}

		for i := range 64 {
			in[i] = byte(i)
			h := siphash.New64(k[:])
			h.Write(in[:i])
			Expect(h.Sum(nil)).To(Equal(goldenRef[i]))
		}
	})

	It("Sum64 unaligned", func() {
		const align = 8
		var k [16]byte
		var in [64 + align]byte
		for i := range k {
			k[i] = byte(i)
		}

		for a := 1; a < align; a++ {
			for i := range 64 {
				in[a+i] = byte(i)
				h := siphash.New64(k[:])
				h.Write(in[a : a+i])
				Expect(h.Sum(nil)).To(Equal(goldenRef[i]))
			}
		}
	})

	It("Sum128", func() {
		var k [16]byte
		var in [64]byte
		for i := range k {
			k[i] = byte(i)
		}

		for i := range 64 {
			in[i] = byte(i)
			h := siphash.New128(k[:])
			h.Write(in[:i])
			Expect(h.Sum(nil)).To(Equal(goldenRef128[i]))
		}
	})
})
