// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64, !appengine

package rabin

var hasSSE2 = haveSSE2()

// Implemented in rabin_amd64.s
func haveSSE2() bool

func update32(f1, f2 uint32, rawTables *[4][256]uint64, p []byte, numWords int) (newF1, newF2 uint32) {
	if hasSSE2 {
		return update32SSE2(f1, f2, rawTables, p, numWords)
	}
	return update32Generic(f1, f2, rawTables, p, numWords)
}

func update32SSE2(f1, f2 uint32, rawTables *[4][256]uint64, p []byte, numWords int) (newF1, newF2 uint32)
