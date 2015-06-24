// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !appengine

// func update64SSE2(fp uint64, rawTables *[8][256]uint64, p []byte, numWords int) (uint64)
TEXT ·update64SSE2(SB),7,$0
	// 0(FP) fp
	// 8(FP) rawTables
	// 16(FP) p
	// 24(FP) len(p)
	// 32(FP) cap(p)
	// 40(FP) numWords
	// 48(FP) ret (newFp)
	MOVQ fp+0(FP), AX

	MOVQ rawTables+8(FP), R8  // t64

	MOVQ p+16(FP), SI
	MOVQ numWords+40(FP), CX

	/* Process each 64-bit word at a time. */
loop:
	CMPL CX, $0
	JE done

	// Set up BX for table indexing.  In the following, we shift BX
	// and index the table of interest with the lower byte value of BX.
	MOVQ AX, BX

	// t64[uint8(fp)]
	MOVBQZX BX, DX
	// xmm0[2] = t64
	MOVLPS (R8)(DX*8), X0

	// t72[uint8(fp >> 8)]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm0[1] = t72
	MOVHPS (8*256)(R8)(DX*8), X0
	// xmm0 = (t72, t64)

	// t80[fprint >> 16]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm1[2] = t80
	MOVLPS (2*8*256)(R8)(DX*8), X1

	// t88[fprint >> 24]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm1[1] = t88
	MOVHPS (3*8*256)(R8)(DX*8), X1
	// xmm1 = (t88, t80)

	// xmm0 = (t72 ^ t88, t64 ^ t80)
	PXOR X1, X0

	// t96[fprint >> 32]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm1[2] = t96
	MOVLPS (4*8*256)(R8)(DX*8), X1

	// t104[fprint >> 40]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm1[1] = t104
	MOVHPS (5*8*256)(R8)(DX*8), X1
	// xmm1 = (t104, t96)

	// xmm0 = (t72 ^ t88 ^ t104, t64 ^ t80 ^ t96)
	PXOR X1, X0

	// t112[fprint >> 48]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm1[2] = t112
	MOVLPS (6*8*256)(R8)(DX*8), X1

	// t120[fprint >> 56]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm1[1] = t120
	MOVHPS (7*8*256)(R8)(DX*8), X1
	// xmm1 = (t120, t112)

	// BL = inWord
	// Start loading here to avoid latency.
	MOVQ 0(SI), BX

	// xmm0 = (t72 ^ t88 ^ t104 ^ t120, t64 ^ t80 ^ t96 ^ t112)
	PXOR X1, X0

	// xmm1[2] = xmm0[1] = t72 ^ t88 ^ t104 ^ t120
	MOVHLPS X0, X1

	// xmm0[2] ^= xmm1[2] => t64 ^ t80 ^ t72 ^ t88 ^ t96 ^t104 ^ t112 ^ t120
	PXOR X1, X0

	// We can safely switch out of SSE mode at this point, since we'll be
	// working with 64-bit values.
	// AX = t64 ^ t80 ^ t72 ^ t88 ^ t96 ^t104 ^ t112 ^ t120
	MOVQ X0, AX

	// NOTE: pshufb should be faster than bswap.
	// p[] is processed in big-endian order.
	BSWAPQ BX

	// AX = t64 ^ t80 ^ t72 ^ t88 ^ t96 ^t104 ^ t112 ^ t120 ^ inWord
	// This is the new fingerprint.
	XORQ BX, AX

	// Upkeep ii.
	DECL CX
	// Processing 64-bit words.
	ADDQ $8, SI
	JMP loop

done:
	MOVQ AX, ret+48(FP)
	RET
