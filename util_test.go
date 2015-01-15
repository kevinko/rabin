// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rabin

import (
	"testing"
)

import (
	Crand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"
)

func Test_CalcPrimes(t *testing.T) {
	cmp := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for ii, v := range a {
			if v != b[ii] {
				return false
			}
		}
		return true
	}

	if !cmp(CalcPrimes(2), []int{2}) {
		t.Error("2")
	}
	if !cmp(CalcPrimes(11), []int{2, 3, 5, 7, 11}) {
		t.Error("11")
	}
}

func Test_IsPowerOfTwo(t *testing.T) {
	if !IsPowerOfTwo(2) {
		t.Error("2")
	}
	if IsPowerOfTwo(3) {
		t.Error("3")
	}
	if !IsPowerOfTwo(4) {
		t.Error("4")
	}
	if IsPowerOfTwo(5) {
		t.Error("5")
	}
}

func Test_MakeRandom(t *testing.T) {
	// Seed the source with a strongly random seed (crypto/rand).
	maxSeed := big.NewInt(math.MaxInt64)
	seed, err := Crand.Int(Crand.Reader, maxSeed)
	if err != nil {
		panic(err)
	}
	source := rand.NewSource(seed.Int64())
	randGen := rand.New(source)

	// This is a crude test.  There are 2^60 possible degree 61 polynomials.
	// The probability that we draw 5 identical ones in sequence
	// uniformly at random is (1/2^60)^4 = 1/2^64, which should never
	// happen.

	degree := 61
	p := MakeRandom(randGen, degree)
	if p.degree != 61 {
		t.Error("wrong degree")
	}

	numDuplicates := 0
	for ii := 0; ii < 4; ii++ {
		cmp := MakeRandom(randGen, degree)
		if p.Cmp(cmp) == 0 {
			numDuplicates++
		}
	}
	if numDuplicates >= 4 {
		t.Error("improbable match")
	}
}

func Test_FindIrreducible(t *testing.T) {
	// NOTE: this is very crude.
	p := FindIrreducible(32)
	if p.degree != 32 {
		t.Error("p.degree != 32")
	}
	if !p.Irreducible() {
		t.Error("p not irreducible")
	}

	p = FindIrreducible(64)
	if !p.Irreducible() {
		t.Error("p not irreducible")
	}
}

func Test_IrreducibleUint64(t *testing.T) {
	p := FindIrreducible(64)
	deg, coeffs := p.Uint64()

	cmp := NewPolynomialFromUint64(deg, coeffs)
	if p.Cmp(cmp) != 0 {
		t.Error("equality failed")
	}
}

func Test_PowerTable(t *testing.T) {
	p := NewPolynomialFromUint64(kIrreduciblePolyDegree, kIrreduciblePolyCoeffs)
	// Offset by a little bit to test forwarding.
	pt := makePowerTable(70)
	for ii := 0; ii < len(pt); ii++ {
		coeffs := new(big.Int)
		coeffs.SetBit(coeffs, 70+ii, 1)
		powerPoly := NewPolynomialFromBigInt(coeffs)
		powerPoly.Mod(powerPoly, p)
		_, cmpCoeffs := powerPoly.Uint64()
		if cmpCoeffs != pt[ii] {
			t.Error(fmt.Sprintf("mismatch term %d (0x%x, 0x%x)",
				ii, cmpCoeffs, pt[ii]))
		}
	}
}

func Test_MakeTables32(t *testing.T) {
	p := NewPolynomialFromUint64(kIrreduciblePolyDegree, kIrreduciblePolyCoeffs)
	cmp := MakeRabinTables32FromPoly(p)
	tables := makeRabinTables32Raw()
	for ii := 0; ii < 4; ii++ {
		for jj := 0; jj < 256; jj++ {
			vCmp := cmp[ii][jj]
			v := tables[ii][jj]
			if vCmp != v {
				t.Error(fmt.Sprintf("[%d][%d]: 0x%x != 0x%x", ii, jj, v, vCmp))
			}
		}
	}
}

func Test_RabinRollingTables32(t *testing.T) {
	p := NewPolynomialFromUint64(kIrreduciblePolyDegree, kIrreduciblePolyCoeffs)

	checkCoeffs := func(cmp uint64, power int, b byte) {
		coeffs := big.NewInt(int64(b))
		coeffs.Lsh(coeffs, uint(power))
		powerPoly := NewPolynomialFromBigInt(coeffs)
		powerPoly.Mod(powerPoly, p)
		_, cmpCoeffs := powerPoly.Uint64()
		if cmpCoeffs != cmp {
			t.Error(fmt.Sprintf("mismatch term (0x%x, 0x%x)",
				cmpCoeffs, cmp))
		}
	}

	basePower := 128 * 8

	tables := makeRabinRollingTables32(128)
	for ii := 0; ii < 256; ii++ {
		checkCoeffs(tables.t8m0[ii], basePower, byte(ii))
		checkCoeffs(tables.t8m8[ii], basePower+8, byte(ii))
		checkCoeffs(tables.t8m16[ii], basePower+16, byte(ii))
		checkCoeffs(tables.t8m24[ii], basePower+24, byte(ii))
	}
}

func Test_MakeTables64(t *testing.T) {
	p := NewPolynomialFromUint64(kIrreduciblePolyDegree, kIrreduciblePolyCoeffs)
	cmp := MakeRabinTables64FromPoly(p)
	tables := makeRabinTables64Raw()
	for ii := 0; ii < 8; ii++ {
		for jj := 0; jj < 256; jj++ {
			vCmp := cmp[ii][jj]
			v := tables[ii][jj]
			if vCmp != v {
				t.Error(fmt.Sprintf("[%d][%d]: 0x%x != 0x%x", ii, jj, v, vCmp))
			}
		}
	}
}

func Benchmark_MakeTables32(b *testing.B) {
	for ii := 0; ii < b.N; ii++ {
		makeRabinTables32()
	}
}

func Benchmark_MakeTables32Naive(b *testing.B) {
	p := NewPolynomialFromUint64(kIrreduciblePolyDegree, kIrreduciblePolyCoeffs)
	for ii := 0; ii < b.N; ii++ {
		MakeRabinTables32FromPoly(p)
	}
}
