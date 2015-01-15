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

// These are tables for the 64-bit approach.
var kTables64 *rabinTables64

type digest64 struct {
	fingerprint uint64
}

func init() {
	kTables64 = makeRabinTables64()
}

func New64() hash.Hash64 {
	hash := new(digest64)
	return hash
}

func (d *digest64) BlockSize() int {
	return 8
}

func (d *digest64) Reset() {
	d.fingerprint = 0
}

func (d *digest64) Size() int {
	return kNumBytes
}

func (d *digest64) Sum(b []byte) []byte {
	// Maintain little-endian order.
	b = append(b, byte(d.fingerprint>>56))
	b = append(b, byte(d.fingerprint>>48))
	b = append(b, byte(d.fingerprint>>40))
	b = append(b, byte(d.fingerprint>>32))
	b = append(b, byte(d.fingerprint>>24))
	b = append(b, byte(d.fingerprint>>16))
	b = append(b, byte(d.fingerprint>>8))
	b = append(b, byte(d.fingerprint))
	return b
}

func (d *digest64) Sum64() uint64 {
	return d.fingerprint
}

// len(p) must be < 4.  This updates the fingerprint based on p.  It is used
// to finish up processing when word-sized updates can no longer be performed.
// Returns (f1, f2)
func updateSubword64(fp uint64, p []byte) uint64 {
	switch len(p) {
	case 7:
		bytes := (uint64(p[0]) << 48) |
			(uint64(p[1]) << 40) |
			(uint64(p[2]) << 32) |
			(uint64(p[3]) << 24) |
			(uint64(p[4]) << 16) |
			(uint64(p[5]) << 8) |
			(uint64(p[6]))
		fp0 := fp << 56
		t112 := kTables64.t112[uint8(fp>>56)]
		t104 := kTables64.t104[uint8(fp>>48)]
		t96 := kTables64.t96[uint8(fp>>40)]
		t88 := kTables64.t88[uint8(fp>>32)]
		t80 := kTables64.t80[uint8(fp>>24)]
		t72 := kTables64.t72[uint8(fp>>16)]
		t64 := kTables64.t64[uint8(fp>>8)]
		fp = t112 ^ t104 ^ t96 ^ t88 ^ t80 ^ t72 ^ t64 ^ fp0 ^ bytes
		break
	case 6:
		bytes := (uint64(p[0]) << 40) |
			(uint64(p[1]) << 32) |
			(uint64(p[2]) << 24) |
			(uint64(p[3]) << 16) |
			(uint64(p[4]) << 8) |
			(uint64(p[5]))
		fp0 := fp << 48
		t104 := kTables64.t104[uint8(fp>>56)]
		t96 := kTables64.t96[uint8(fp>>48)]
		t88 := kTables64.t88[uint8(fp>>40)]
		t80 := kTables64.t80[uint8(fp>>32)]
		t72 := kTables64.t72[uint8(fp>>24)]
		t64 := kTables64.t64[uint8(fp>>16)]
		fp = t104 ^ t96 ^ t88 ^ t80 ^ t72 ^ t64 ^ fp0 ^ bytes
		break
	case 5:
		bytes := (uint64(p[0]) << 32) |
			(uint64(p[1]) << 24) |
			(uint64(p[2]) << 16) |
			(uint64(p[3]) << 8) |
			(uint64(p[4]))
		fp0 := fp << 40
		t96 := kTables64.t96[uint8(fp>>56)]
		t88 := kTables64.t88[uint8(fp>>48)]
		t80 := kTables64.t80[uint8(fp>>40)]
		t72 := kTables64.t72[uint8(fp>>32)]
		t64 := kTables64.t64[uint8(fp>>24)]
		fp = t96 ^ t88 ^ t80 ^ t72 ^ t64 ^ fp0 ^ bytes
		break
	case 4:
		bytes := (uint64(p[0]) << 24) |
			(uint64(p[1]) << 16) |
			(uint64(p[2]) << 8) |
			(uint64(p[3]))
		fp0 := fp << 32
		t88 := kTables64.t88[uint8(fp>>56)]
		t80 := kTables64.t80[uint8(fp>>48)]
		t72 := kTables64.t72[uint8(fp>>40)]
		t64 := kTables64.t64[uint8(fp>>32)]
		fp = t88 ^ t80 ^ t72 ^ t64 ^ fp0 ^ bytes
		break
	case 3:
		bytes := (uint64(p[0]) << 16) |
			(uint64(p[1]) << 8) |
			uint64(p[2])
		fp0 := fp << 24
		t80 := kTables64.t80[uint8(fp>>56)]
		t72 := kTables64.t72[uint8(fp>>48)]
		t64 := kTables64.t64[uint8(fp>>40)]
		fp = t80 ^ t72 ^ t64 ^ fp0 ^ bytes
		break
	case 2:
		bytes := (uint64(p[0]) << 8) | uint64(p[1])
		fp0 := fp << 16
		t72 := kTables64.t72[uint8(fp>>56)]
		t64 := kTables64.t64[uint8(fp>>48)]
		fp = t72 ^ t64 ^ fp0 ^ bytes
		break
	case 1:
		fp0 := fp << 8
		t64 := kTables64.t64[uint8(fp>>56)]
		fp = t64 ^ fp0 ^ uint64(p[0])
		break
	case 0:
		break
	default:
		panic(fmt.Sprint("unexpected remainder ", len(p)))
	}
	return fp
}

func update64Generic(fp uint64, rawTables *[8][256]uint64, p []byte, numWords int) uint64 {
	table64 := &rawTables[0]
	table72 := &rawTables[1]
	table80 := &rawTables[2]
	table88 := &rawTables[3]
	table96 := &rawTables[4]
	table104 := &rawTables[5]
	table112 := &rawTables[6]
	table120 := &rawTables[7]

	for ii := 0; ii < numWords; ii++ {
		offset := 8 * ii
		inWord := (uint64(p[offset]) << 56) |
			(uint64(p[offset+1]) << 48) |
			(uint64(p[offset+2]) << 40) |
			(uint64(p[offset+3]) << 32) |
			(uint64(p[offset+4]) << 24) |
			(uint64(p[offset+5]) << 16) |
			(uint64(p[offset+6]) << 8) |
			(uint64(p[offset+7]))

		t120 := table120[uint8(fp>>56)]
		t112 := table112[uint8(fp>>48)]
		t104 := table104[uint8(fp>>40)]
		t96 := table96[uint8(fp>>32)]
		t88 := table88[uint8(fp>>24)]
		t80 := table80[uint8(fp>>16)]
		t72 := table72[uint8(fp>>8)]
		t64 := table64[uint8(fp)]

		fp = t120 ^ t112 ^ t104 ^ t96 ^ t88 ^ t80 ^ t72 ^ t64 ^ inWord
	}
	return fp
}

func (d *digest64) Write(p []byte) (n int, err error) {
	// Number of 64-bit words
	numWords := len(p) >> 3

	fp := update64(d.fingerprint, kTables64.raw, p, numWords)

	// Process the remainder.
	offset := numWords * 8

	// Store the result.
	d.fingerprint = updateSubword64(fp, p[offset:])

	return len(p), nil
}
