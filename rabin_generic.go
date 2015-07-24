// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build arm, appengine

package rabin

func update32(f1, f2 uint32, rawTables *[4][256]uint64, p []byte, numWords int) (newF1, newF2 uint32) {
	return update32Generic(f1, f2, rawTables, p, numWords)
}
