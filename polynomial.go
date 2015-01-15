// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rabin

import (
	"fmt"
	"math/big"
	"strings"
)

// Represents a polynomial of degree n in GF(2).  The zero value is a
// polynomial of degree 0 with value 0.
type Polynomial struct {
	degree uint
	// coefficients are represented as bits.
	// Note that big.Ints are not deep-copied with assignment.
	coeffs big.Int
}

func NewPolynomial(degree uint, coeffs *big.Int) *Polynomial {
	return &Polynomial{degree: degree, coeffs: *new(big.Int).Set(coeffs)}
}

// coeffs must contain the degree of each non-zero element.
func NewPolynomialFromCoeffs(degrees []uint) *Polynomial {
	maxDegree := uint(0)
	coeffs := new(big.Int)
	for _, deg := range degrees {
		coeffs.SetBit(coeffs, int(deg), 1)
		if deg > maxDegree {
			maxDegree = deg
		}
	}
	return &Polynomial{degree: maxDegree, coeffs: *coeffs}
}

func NewPolynomialFromBigInt(coeffs *big.Int) *Polynomial {
	degree := calcDegree(coeffs)
	return NewPolynomial(degree, coeffs)
}

func NewPolynomialFromInt(coeffs int64) *Polynomial {
	bigCoeffs := big.NewInt(coeffs)
	return NewPolynomialFromBigInt(bigCoeffs)
}

// coeffs corresponds to that returned by the Uint64() method.
func NewPolynomialFromUint64(degree uint, coeffs uint64) *Polynomial {
	lower := int64(coeffs & 0xffffffff)
	bigCoeffs := big.NewInt(lower)

	upper := int64((coeffs >> 32) & 0xffffffff)
	bigUpper := big.NewInt(upper)
	bigUpper.Lsh(bigUpper, 32)

	bigCoeffs.Or(bigCoeffs, bigUpper)

	// Finally, set the degree term.
	bigCoeffs.SetBit(bigCoeffs, int(degree), 1)

	return NewPolynomial(degree, bigCoeffs)
}

// Set p to the sum of (x + y) in GF(2) and returns p.
func (p *Polynomial) Add(x, y *Polynomial) *Polynomial {
	p.coeffs.Xor(&x.coeffs, &y.coeffs)
	p.degree = calcDegree(&p.coeffs)
	return p
}

func calcDegree(coeffs *big.Int) uint {
	deg := coeffs.BitLen() - 1
	if deg < 0 {
		// zero polynomial is degree 0.
		deg = 0
	}
	return uint(deg)
}

// Compare x and y and returns:
//   -1 if x < y
//    0 if x == y
//   +1 if x > y
func (x *Polynomial) Cmp(y *Polynomial) int {
	if x.degree > y.degree {
		return 1
	}
	if x.degree < y.degree {
		return -1
	}
	return x.coeffs.Cmp(&y.coeffs)
}

func (p *Polynomial) Degree() uint {
	return p.degree
}

// Sets q = x / y.  Returns q and the remainder r.
func (q *Polynomial) Div(x, y *Polynomial) (*Polynomial, *Polynomial) {
	zero := new(big.Int)
	if y.degree == 0 {
		// y is 0 or 1.
		if y.coeffs.Cmp(zero) == 0 {
			panic("divide by 0")
		}
		q.Set(x)
		r := NewPolynomial(0, zero)
		return q, r
	}
	if x.degree == 0 {
		// 0 / y = 0.
		if x.coeffs.Cmp(zero) == 0 {
			q.degree = 0
			q.coeffs.Set(zero)

			r := NewPolynomial(0, zero)
			return q, r
		}
	}

	var qCoeffs *big.Int
	// Be careful not to modify any of the input parameters.
	if q == x || q == y {
		qCoeffs = new(big.Int)
	} else {
		// Re-use q to minimize allocations.
		qCoeffs = &q.coeffs
		qCoeffs.SetInt64(0)
	}

	// The remainder, which is x initially.
	r := new(Polynomial).Set(x)

	// This is just elementary long division in GF(2).
	for r.degree >= y.degree {
		// s(x) = y(x) * x^{shift_len}
		// Then, r'(x) = r(x) - s(x)

		// Record the new component in the quotient.
		shiftLen := int(r.degree - y.degree)
		qCoeffs.SetBit(qCoeffs, shiftLen, 1)

		sCoeffs := new(big.Int).Lsh(&y.coeffs, uint(shiftLen))

		// Determine r'(x) by subtracting s(x) from the remainder.
		r.coeffs.Xor(&r.coeffs, sCoeffs)
		r.degree = calcDegree(&r.coeffs)
	}
	// r.degree < y.degree at this point.

	q.coeffs = *qCoeffs
	q.degree = calcDegree(&q.coeffs)
	return q, r
}

// Sets p to the GCD of a and b in GF(2) and returns p.
func (p *Polynomial) Gcd(a, b *Polynomial) *Polynomial {
	zero := new(Polynomial)

	prevR := b
	for {
		_, r := new(Polynomial).Div(a, b)
		if r.Cmp(zero) == 0 {
			break
		}
		prevR = r
		a = b
		b = r
	}

	p.Set(prevR)
	return p
}

// Determines all prime divisors of v.  This is optimized for powers of 2.
func calcDivisors(v int) []int {
	if v < 2 {
		return []int{}
	}

	if IsPowerOfTwo(v) {
		return []int{2}
	}
	primes := make([]int, 0)
	for _, p := range CalcPrimes(v) {
		if v%p == 0 {
			primes = append(primes, p)
		}
	}
	return primes
}

// Returns true if r is irreducible.
//
// r (of degree d) is irreducible iff:
//  1)
//
//     x^{2^d} = x mod r(x)
//
//  and
//
//  2)
//
//     GCD(x^{2^{d/m}} - x, r(x)) = 1 \forall prime divisors m of d.
//
func (r *Polynomial) Irreducible() bool {
	// Check condition 1) (repeated squares).

	// f(x) = x
	f := NewPolynomialFromInt(0x2)
	// f(x) mod r(x)
	f.Mod(f, r)

	tmp := NewPolynomialFromInt(0x2)
	// We should not encounter f(x) the first (r.degree - 1) times of
	// squaring.
	for ii := uint(0); ii < r.degree; ii++ {
		tmp.Square(tmp).Mod(tmp, r)
	}
	if tmp.Cmp(f) != 0 {
		return false
	}

	// Check condition 2.  GCD(x^2^{d/m} - x, r(x)) == 1
	one := NewPolynomialFromInt(1)
	for _, div := range calcDivisors(int(r.degree)) {
		tmp.Set(f)

		repeatCount := int(r.degree) / div
		for ii := 0; ii < repeatCount; ii++ {
			tmp.Square(tmp).Mod(tmp, r)
		}

		if tmp.Add(tmp, f).Gcd(tmp, r).Cmp(one) != 0 {
			return false
		}
	}
	return true
}

// Sets r to x mod p in GF(2) and returns r.
func (r *Polynomial) Mod(x, p *Polynomial) *Polynomial {
	if x.degree < p.degree {
		r.Set(x)
		return r
	}
	_, remainder := new(Polynomial).Div(x, p)
	r.Set(remainder)
	return r
}

func (z *Polynomial) Mul(x, y *Polynomial) *Polynomial {
	var zCoeffs *big.Int
	if z == x || z == y {
		// Ensure that we do not modify z if it's a parameter.
		zCoeffs = new(big.Int)
	} else {
		zCoeffs = &z.coeffs
		zCoeffs.SetInt64(0)
	}

	small, large := x, y
	if y.degree < x.degree {
		small, large = y, x
	}

	// Walk through small coeffs, shift large by the corresponding amount,
	// and accumulate in z.
	coeffs := new(big.Int).Set(&small.coeffs)
	zero := new(big.Int)
	for coeffs.Cmp(zero) > 0 {
		deg := calcDegree(coeffs)
		factor := new(big.Int).Lsh(&large.coeffs, deg)
		zCoeffs.Xor(zCoeffs, factor)

		// Prepare for next iteration.
		coeffs.SetBit(coeffs, int(deg), 0)
	}

	z.degree = calcDegree(zCoeffs)
	z.coeffs = *zCoeffs
	return z
}

// Copies other into p and returns p.
func (p *Polynomial) Set(other *Polynomial) *Polynomial {
	if p == other {
		return p
	}
	p.degree = other.degree
	p.coeffs.Set(&other.coeffs)
	return p
}

// Sets the coefficient of given degree to v in p and returns p.
// v must be 0 or 1 or SetCoefficient will panic.
func (p *Polynomial) SetCoefficient(coeffDeg, v uint) *Polynomial {
	if v != 0 && v != 1 {
		panic(fmt.Sprintf("v is not 0 or 1, %d", v))
	}
	if v > 0 {
		// Adjust p's degree if necessary.
		if p.degree < coeffDeg {
			p.degree = coeffDeg
		}
	}

	p.coeffs.SetBit(&p.coeffs, int(coeffDeg), v)

	if v == 0 {
		// Reduce p's degree, since v cleared the term.
		if p.degree == coeffDeg {
			p.degree = calcDegree(&p.coeffs)
		}
	}
	return p
}

// Sets p to f^2 and returns p.
func (p *Polynomial) Square(f *Polynomial) *Polynomial {
	// (\sum_i a_i x_i)^2 = \sum_i (a_i x_i \sum_j a_j x_j)
	// = \sum_i (a_i x_i * (\sum_{j \ne i} a_j x_j + a_i x_i))
	// = \sum_i (a_i^2 x_i^2) + \sum_i(a_i x_i \sum_{j \ne i} a_j x_j)
	// = \sum_i (a_i^2 x_i^2) \in GF(2)
	//
	// The cross terms will cancel out: one pair for each.

	newDeg := 2 * f.degree
	var newCoeffs *big.Int
	if f == p {
		// f can be p, so be careful not to overwrite the input.
		newCoeffs = new(big.Int)
	} else {
		// Reuse to avoid allocation.
		newCoeffs = &p.coeffs
		newCoeffs.SetInt64(0)
	}

	zero := new(big.Int)

	// Walk through each set bit and double the degree.
	coeffs := new(big.Int).Set(&f.coeffs)
	for coeffs.Cmp(zero) > 0 {
		deg := calcDegree(coeffs)
		if deg == 0 {
			newCoeffs.SetBit(newCoeffs, 0, 1)
		} else {
			newCoeffs.SetBit(newCoeffs, int(2*deg), 1)
		}
		// Eliminate.
		coeffs.SetBit(coeffs, int(deg), 0)
	}

	p.degree = newDeg
	p.coeffs = *newCoeffs
	return p
}

func (p *Polynomial) String() string {
	zero := new(big.Int)

	// As a special case, we need to handle the zero term.
	if p.degree == 0 && p.coeffs.Cmp(zero) == 0 {
		return "0"
	}

	terms := make([]string, 0)

	// We deal with sparse polynomials using big.Int's BitLen(), which
	// should be optimized for word-sized operations.  Thus, copy p and then
	// walk through the 1 bits, MSB first.
	coeffs := new(big.Int).Set(&p.coeffs)
	for coeffs.Cmp(zero) > 0 {
		deg := calcDegree(coeffs)
		if deg == 0 {
			terms = append(terms, "1")
		} else {
			terms = append(terms, fmt.Sprintf("x^%v", deg))
		}
		coeffs.SetBit(coeffs, int(deg), 0)
	}

	return strings.Join(terms, " + ")
}

// Returns (degree, coefficients).
// Coefficients holds the state of the variable bits (0, 1, ..., degree - 1).
// Note that the k-th degree of a k degree polynomial is fixed (ie., 1) in
// GF(2).
//
// The result is undefined if degree > 64.
func (p *Polynomial) Uint64() (uint, uint64) {
	if p.degree > 64 {
		panic(fmt.Sprint("degree is too large", p.degree))
	}
	tmp := new(big.Int).Set(&p.coeffs)
	if p.degree == 64 {
		// First, clear the MSB (bit), since we're interested in the
		// lower bits.
		tmp.SetBit(tmp, int(p.degree), 0)
	}

	mask := big.NewInt(int64(0xffffffff))

	half := new(big.Int)
	lower := uint32(half.And(tmp, mask).Int64())
	upper := uint32(half.Rsh(tmp, 32).And(half, mask).Int64())

	v := uint64(lower) | (uint64(upper) << 32)
	return p.degree, v
}
