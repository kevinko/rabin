// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This implements Rabin fingerprinting using a fixed irreducible 64-bit
// polynomial.
package rabin

import (
	"fmt"
	"hash"
)

type RollingHash interface {
	hash.Hash64

	// Drains the oldest bytes oldData from the hash and appends newData.
	// This is used for rolling windows.
	//
	// len(oldData) MUST equal len(newData).  This may not be checked,
	// and errant behavior may result if this does not hold.
	Roll(oldData, newData []byte) (int, error)
}

const (
	// 64-bit fingerprints
	kNumBytes = 8
)

// x^64 + x^62 + x^60 + x^59 + x^56 + x^55 + x^54 + x^51
// + x^50 + x^48 + x^47 + x^43 + x^34 + x^33 + x^32 + x^31
// + x^29 + x^27 + x^26 + x^21 + x^20 + x^19 + x^18 + x^17
// + x^14 + x^4 + x^2 + x^1 + 1
//
// This holds the coefficents of degree < 64.
const (
	kIrreduciblePolyCoeffs = 0x59cd8807ac3e4017
	kIrreduciblePolyDegree = 64
)

// These are tables for the 32-bit approach.
var kTables *rabinTables32

type digest struct {
	// The fingerprint is (f1 f2) = (f1 << 32) | f2
	f1 uint32
	f2 uint32

	// The following are only defined if a rolling window is specified.
	windowSize    int
	rollingTables *rabinRollingTables32
}

func init() {
	kTables = makeRabinTables32()
}

func New() hash.Hash64 {
	hash := new(digest)
	return hash
}

// windowSize in bytes.  A table will be pre-computed, so a non-negligible setup
// cost occurs for each rolling hash construction.
func NewRolling(windowSize int) RollingHash {
	hash := new(digest)
	hash.windowSize = windowSize
	hash.rollingTables = makeRabinRollingTables32(windowSize)
	return hash
}

func (d *digest) BlockSize() int {
	return 4
}

func (d *digest) Reset() {
	d.f1 = 0
	d.f2 = 0
}

// Rolling is similar to writing new bytes.  For each step, we need only
// subtract out a corresponding amount of oldData.  (See rabin.tex.)
func (d *digest) Roll(oldData, newData []byte) (int, error) {
	if len(oldData) != len(newData) {
		panic("len(oldData) != len(newData)")
	}

	// Number of 32-bit words
	numWords := len(newData) >> 2

	// f = f1 f2  // f1 is the high word
	f1 := d.f1
	f2 := d.f2
	for ii := 0; ii < numWords; ii++ {
		offset := 4 * ii
		inWord := (uint32(newData[offset]) << 24) |
			(uint32(newData[offset+1]) << 16) |
			(uint32(newData[offset+2]) << 8) |
			(uint32(newData[offset+3]))

		ta := kTables.t88[uint8(f1>>24)]
		tb := kTables.t80[uint8(f1>>16)]
		tc := kTables.t72[uint8(f1>>8)]
		td := kTables.t64[uint8(f1)]

		f1 = uint32(ta>>32) ^ uint32(tb>>32) ^
			uint32(tc>>32) ^ uint32(td>>32) ^ f2
		f2 = uint32(ta) ^ uint32(tb) ^
			uint32(tc) ^ uint32(td) ^ inWord

		// Subtract the old data.  Maintain big-endian order.
		t8m0 := d.rollingTables.t8m0[oldData[offset+3]]
		t8m8 := d.rollingTables.t8m8[oldData[offset+2]]
		t8m16 := d.rollingTables.t8m16[oldData[offset+1]]
		t8m24 := d.rollingTables.t8m24[oldData[offset]]
		f1 ^= uint32(t8m0>>32) ^ uint32(t8m8>>32) ^
			uint32(t8m16>>32) ^ uint32(t8m24>>32)
		f2 ^= uint32(t8m0) ^ uint32(t8m8) ^
			uint32(t8m16) ^ uint32(t8m24)
	}

	// Process the remainder.
	offset := numWords * 4
	f1, f2 = updateSubword(f1, f2, newData[offset:])

	// Fix up the remainder.
	switch len(oldData) - offset {
	case 3:
		t8m0 := d.rollingTables.t8m0[oldData[offset+2]]
		t8m8 := d.rollingTables.t8m8[oldData[offset+1]]
		t8m16 := d.rollingTables.t8m16[oldData[offset]]

		f1 ^= uint32(t8m0>>32) ^ uint32(t8m8>>32) ^ uint32(t8m16>>32)
		f2 ^= uint32(t8m0) ^ uint32(t8m8) ^ uint32(t8m16)
		break
	case 2:
		t8m0 := d.rollingTables.t8m0[oldData[offset+1]]
		t8m8 := d.rollingTables.t8m8[oldData[offset]]

		f1 ^= uint32(t8m0>>32) ^ uint32(t8m8>>32)
		f2 ^= uint32(t8m0) ^ uint32(t8m8)
		break
	case 1:
		t8m0 := d.rollingTables.t8m0[oldData[offset]]

		f1 ^= uint32(t8m0 >> 32)
		f2 ^= uint32(t8m0)
		break
	case 0:
		break
	}

	// Store the updated fingerprints.
	d.f1, d.f2 = f1, f2

	return len(newData), nil
}

func (d *digest) Size() int {
	return kNumBytes
}

func (d *digest) Sum(b []byte) []byte {
	// Maintain little-endian order.
	b = append(b, byte(d.f1>>24))
	b = append(b, byte(d.f1>>16))
	b = append(b, byte(d.f1>>8))
	b = append(b, byte(d.f1))

	b = append(b, byte(d.f2>>24))
	b = append(b, byte(d.f2>>16))
	b = append(b, byte(d.f2>>8))
	b = append(b, byte(d.f2))
	return b
}

func (d *digest) Sum64() uint64 {
	return (uint64(d.f1) << 32) | uint64(d.f2)
}

// len(p) is a multiple of 4 (32-bit words) = numWords * 32
func update32Generic(f1, f2 uint32, rawTables *[4][256]uint64, p []byte, numWords int) (newF1, newF2 uint32) {
	t64 := &rawTables[0]
	t72 := &rawTables[1]
	t80 := &rawTables[2]
	t88 := &rawTables[3]

	for ii := 0; ii < numWords; ii++ {
		offset := ii << 2
		inWord := (uint32(p[offset]) << 24) |
			(uint32(p[offset+1]) << 16) |
			(uint32(p[offset+2]) << 8) |
			(uint32(p[offset+3]))

		ta := t88[uint8(f1>>24)]
		tb := t80[uint8(f1>>16)]
		tc := t72[uint8(f1>>8)]
		td := t64[uint8(f1)]

		f1 = uint32(ta>>32) ^ uint32(tb>>32) ^
			uint32(tc>>32) ^ uint32(td>>32) ^ f2
		f2 = uint32(ta) ^ uint32(tb) ^
			uint32(tc) ^ uint32(td) ^ inWord
	}
	newF1 = f1
	newF2 = f2
	return
}

// len(p) must be < 4.  This updates the fingerprint based on p.  It is used
// to finish up processing when word-sized updates can no longer be performed.
// Returns (f1, f2)
func updateSubword(f1, f2 uint32, p []byte) (uint32, uint32) {
	switch len(p) {
	case 3:
		j1 := (f1 << 24) | (f2 >> 8)
		j2 := f2 << 24
		bytes := (uint32(p[0]) << 16) |
			(uint32(p[1]) << 8) |
			uint32(p[2])
		tb := kTables.t80[uint8(f1>>24)]
		tc := kTables.t72[uint8(f1>>16)]
		td := kTables.t64[uint8(f1>>8)]

		f1 = uint32(tb>>32) ^ uint32(tc>>32) ^
			uint32(td>>32) ^ j1
		f2 = uint32(tb) ^ uint32(tc) ^ uint32(td) ^ j2 ^ bytes
		break
	case 2:
		j1 := (f1 << 16) | (f2 >> 16)
		j2 := f2 << 16
		bytes := (uint32(p[0]) << 8) | uint32(p[1])
		tc := kTables.t72[uint8(f1>>24)]
		td := kTables.t64[uint8(f1>>16)]

		f1 = uint32(tc>>32) ^ uint32(td>>32) ^ j1
		f2 = uint32(tc) ^ uint32(td) ^ j2 ^ bytes
		break
	case 1:
		j1 := (f1 << 8) | (f2 >> 24)
		j2 := f2 << 8
		td := kTables.t64[uint8(f1>>24)]

		f1 = uint32(td>>32) ^ j1
		f2 = uint32(td) ^ j2 ^ uint32(p[0])
		break
	case 0:
		break
	default:
		panic(fmt.Sprint("unexpected remainder ", len(p)))
	}
	return f1, f2
}

func (d *digest) Write(p []byte) (n int, err error) {
	// Number of 32-bit words
	numWords := len(p) >> 2

	// f = f1 f2  // f1 is the high word
	f1 := d.f1
	f2 := d.f2
	f1, f2 = update32(f1, f2, kTables.raw, p, numWords)

	// Process the remainder.
	offset := numWords * 4

	// Store the result.
	d.f1, d.f2 = updateSubword(f1, f2, p[offset:])

	return len(p), nil
}
