// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build arm

package rabin

func update64(fp uint64, rawTables *[8][256]uint64, p []byte, numWords int) uint64 {
	return update64Generic(fp, rawTables, p, numWords)
}
