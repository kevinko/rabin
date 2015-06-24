// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64, !appengine

package rabin

func update64(fp uint64, rawTables *[8][256]uint64, p []byte, numWords int) uint64 {
	if hasSSE2 {
		return update64SSE2(fp, rawTables, p, numWords)
	}
	return update64Generic(fp, rawTables, p, numWords)
}

func update64SSE2(fp uint64, rawTables *[8][256]uint64, p []byte, numWords int) uint64
