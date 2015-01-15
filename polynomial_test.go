// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rabin

import (
	"testing"
)

var (
	ZERO = new(Polynomial)
	ONE  = NewPolynomialFromInt(1)
)

func Test_Cmp(t *testing.T) {
	x := new(Polynomial)
	y := new(Polynomial)
	if x.Cmp(y) != 0 {
		t.Error("x == y failed")
	}

	x.SetCoefficient(0, 1)
	if x.Cmp(y) != 1 {
		t.Error("x > y failed")
	}

	y.SetCoefficient(1, 1)
	if x.Cmp(y) != -1 {
		t.Error("x < y failed")
	}

	y.SetCoefficient(1, 0)
	if x.Cmp(y) != 1 {
		t.Error("x > y failed")
	}
}

func Test_Add(t *testing.T) {
	zero := new(Polynomial)

	x := new(Polynomial)

	z := new(Polynomial)
	z.Add(x, zero)
	if z.Cmp(zero) != 0 {
		t.Error("x + y == 0 failed")
	}

	x.SetCoefficient(1, 1)
	z.Add(x, zero)
	if z.Cmp(zero) != 1 {
		t.Error("x + y > 0 failed")
	}

	z.Add(x, x)
	if z.Cmp(zero) != 0 {
		t.Error("x + x == 0 failed")
	}
}

func Test_Div(t *testing.T) {
	// f(x) = x + 1
	f := NewPolynomialFromInt(0x3)
	// g(x) = x^2 + x + 1
	g := NewPolynomialFromInt(0x7)

	// q(x) = x, r(x) = 1
	q := new(Polynomial)
	q, r := q.Div(g, f)
	qExpected := NewPolynomialFromInt(0x2)
	if q.Cmp(qExpected) != 0 {
		t.Error("q != x")
	}
	if r.Cmp(ONE) != 0 {
		t.Error("r != 1")
	}

	// g(x) = x^5 + x^2 + 1
	g = NewPolynomialFromInt(37)
	q, r = q.Div(g, f)
	// q(x) = x^4 + x^3 + x^2
	// r(x) = 1
	qExpected = NewPolynomialFromInt(28)
	if q.Cmp(qExpected) != 0 {
		t.Error("q != x^4 + x^3 + x^2")
	}
	if r.Cmp(ONE) != 0 {
		t.Error("r != 1")
	}

	// Test re-using the parameter.
	_, r = f.Div(g, f)
	if f.Cmp(qExpected) != 0 {
		t.Error("f != x^4 + x^3 + x^2")
	}
}

func Test_Mod(t *testing.T) {
	r := new(Polynomial)
	r.Mod(ZERO, ONE)
	if r.Cmp(ZERO) != 0 {
		t.Error("r != 0")
	}

	// p(x) = x + 1
	p := NewPolynomialFromInt(0x3)
	// f(x) = x^5 + x^2 + 1
	f := NewPolynomialFromInt(37)

	r = r.Mod(f, p)

	// q(x) = x^4 + x^3 + x^2
	// r(x) = 1
	if r.Cmp(ONE) != 0 {
		t.Error("r != 1")
	}
}

func Test_Square(t *testing.T) {
	x := new(Polynomial).Square(ONE)
	if x.Cmp(ONE) != 0 {
		t.Error("1^2 != 1")
	}

	x.SetCoefficient(1, 1)
	x.Square(x)
	xCmp := NewPolynomialFromInt(5)
	if x.Cmp(xCmp) != 0 {
		t.Error("(x + 1)^2 != x^2 + 1")
	}
}

func Test_Mul(t *testing.T) {
	// p(x) = x + 1
	p := NewPolynomialFromInt(0x3)
	p.Mul(p, p)
	z := NewPolynomialFromInt(0x5)
	if p.Cmp(z) != 0 {
		t.Error("(x+1)^2 != x^2 + 1")
	}
}

func Test_Gcd(t *testing.T) {
	a := NewPolynomialFromInt(0x3)
	b := NewPolynomialFromInt(0x5)
	a.Gcd(a, b)
	if a.Cmp(NewPolynomialFromInt(0x3)) != 0 {
		t.Error("GCD(x + 1, x^2 + 1) != x + 1")
	}

	c := new(Polynomial).Mul(a, b)
	gcd := new(Polynomial).Gcd(b, c)
	if gcd.Cmp(b) != 0 {
		t.Error("GCD(x^3 + x^2 + x + 1, x^2 + 1) != x^2 + 1")
	}

	d := new(Polynomial).Mul(a, NewPolynomialFromInt(0x7))
	gcd = new(Polynomial).Gcd(c, d)
	if gcd.Cmp(a) != 0 {
		t.Error("GCD(x^3 + x^2 + x + 1, x^2 + x + 1) != x + 1")
	}
}

func Test_Irreducible(t *testing.T) {
	f := NewPolynomialFromCoeffs([]uint{2, 0})
	if f.Irreducible() {
		t.Error("x^2 + 1 should not be irreducible")
	}

	f = NewPolynomialFromCoeffs([]uint{1, 0})
	if !f.Irreducible() {
		t.Error("x + 1 should be irreducible")
	}

	f = NewPolynomialFromCoeffs([]uint{1, 0})
	if !f.Irreducible() {
		t.Error("x + 1 should be irreducible")
	}

	f = NewPolynomialFromCoeffs([]uint{26, 22, 21, 19, 18, 14, 12, 11, 10, 7, 5, 1, 0})
	if !f.Irreducible() {
		t.Error("X26+X22+X21+X19+X18+X14+X12+X11+X10+X7+X5+X+1 should be irreducible")
	}

	f = NewPolynomialFromCoeffs([]uint{2, 1, 0})
	if !f.Irreducible() {
		t.Error("x^2 + x + 1 should be irreducible")
	}

	f = NewPolynomialFromCoeffs([]uint{15, 1, 0})
	if !f.Irreducible() {
		t.Error("x^15 + x + 1 should be irreducible")
	}

	// (x + 1) ^3
	f = NewPolynomialFromCoeffs([]uint{3, 2, 1, 0})
	if f.Irreducible() {
		t.Error("x^3 + x^2 + x + 1 should not be irreducible")
	}

	f = NewPolynomialFromCoeffs([]uint{4, 3, 2, 1})
	if f.Irreducible() {
		t.Error("x^4 + x^3 + x^2 + x should not be irreducible")
	}
}

func Test_Uint64(t *testing.T) {
	cmp := uint64(0xdeadbeefdeadbeef)
	p := NewPolynomialFromUint64(63, cmp)
	if p.degree != 63 {
		t.Error("p.degree != 63")
	}
	deg, coeffs := p.Uint64()
	if deg != 63 {
		t.Error("deg != 63")
	}
	if coeffs != cmp {
		t.Error("coeffs != cmp")
	}
	p = NewPolynomialFromUint64(64, cmp)
	if p.degree != 64 {
		t.Error("p.degree != 64")
	}
	deg, coeffs = p.Uint64()
	if deg != 64 {
		t.Error("deg != 64")
	}
	if coeffs != cmp {
		t.Error("coeffs != cmp")
	}
}
