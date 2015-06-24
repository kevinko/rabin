// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !appengine

// func update32SSE2(f1, f2, uint32, rawTables *[4][256]uint64, p []byte, numWords int) (newF1, newF2 uint32) {
TEXT ·update32SSE2(SB),7,$0
	// 0(FP) f1
	// 4(FP) f2
	// 8(FP) rawTables
	// 16(FP) p
	// 24(FP) len(p)
	// 32(FP) cap(p)
	// 40(FP) numWords
	// 48(FP) newF1
	// 52(FP) newF2
	MOVL f1+0(FP), AX
	SHLQ $32, AX
	MOVL f2+4(FP), BX
	// AX = (f1, f2)
	XORQ BX, AX

	MOVQ rawTables+8(FP), R8  // t64

	MOVQ p+16(FP), SI
	MOVQ numWords+40(FP), CX

	/* Process each 32-bit word at a time. */
loop:
	CMPL CX, $0
	JE done

	// t64[fprint >> 32]
	MOVQ AX, BX
	SHRQ $32, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm0[2] = t64
	MOVLPS (R8)(DX*8), X0

	// t72[fprint >> 40]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm0[1] = t72
	MOVHPS (8*256)(R8)(DX*8), X0
	// xmm0 = (t72, t64)

	// t80[fprint >> 48]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm1[2] = t80
	MOVLPS (2*8*256)(R8)(DX*8), X1

	// t88[fprint >> 56]
	SHRQ $8, BX
	// DX = BL
	MOVBQZX BX, DX
	// xmm1[1] = t88
	MOVHPS (3*8*256)(R8)(DX*8), X1
	// xmm1 = (t88, t80)

	// xmm0 = (t72 ^ t88, t64 ^ t80)
	PXOR X1, X0

	// xmm1[2] = xmm0[1] = t72 ^ t88
	MOVHLPS X0, X1

	// xmm0[2] ^= xmm1[2] => t64 ^ t80 ^ t72 ^ t88
	PXOR X1, X0

	// AH = fprint[2]
	SHLQ $32, AX

	// We can safely switch out of SSE mode at this point, since we'll be
	// working with 64-bit values.
	// BX = t64 ^ t72 ^ t80 ^ t88
	MOVQ X0, BX

	// AX = t64 ^ t72 ^ t80 ^ t88 ^ fprint[2]t^{32}
	XORQ BX, AX

	// BL = inWord
	MOVL 0(SI), BX

	// This is processed in big-endian order.
	BSWAPL BX

	// AX = t64 ^ t72 ^ t80 ^ t88 ^ (fprint[2], inWord)
	// This is the new fingerprint.
	XORQ BX, AX

	// Upkeep ii++
	DECL CX
	// Processing 32-bit words.
	ADDQ $4, SI
	JMP loop

done:
	// f2
	MOVL AX, newF2+52(FP)
	// f1
	SHRQ $32, AX
	MOVL AX, newF1+48(FP)
	RET

TEXT ·haveSSE2(SB),7,$0
	XORQ AX, AX
	INCL AX
	CPUID
	SHRQ $26, DX
	ANDQ $1, DX
	MOVB DX, ret+0(FP)
	RET
