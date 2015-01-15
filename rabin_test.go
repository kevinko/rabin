// Copyright 2012, Kevin Ko <kevin@faveset.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rabin

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"hash/crc64"
	"math/rand"
	"testing"
)

// 26 + 26 + 10 + 11 = 73 total
// [a-zA-Z0-9\.]
var urlChars [73]rune

func init() {
	cIndex := 0
	for ii := 0; ii < 26; ii++ {
		urlChars[cIndex] = 'a' + rune(ii)
		cIndex++
		urlChars[cIndex] = 'A' + rune(ii)
		cIndex++
	}
	for ii := 0; ii < 10; ii++ {
		urlChars[cIndex] = '0' + rune(ii)
		cIndex++
	}
	urlChars[cIndex] = '.'
}

func makeSequence(length int) []byte {
	buff := make([]byte, 0)

	// Avoid a vacuous byte.  (Adding 0 in GF(2) does nothing.)
	char := byte(1)
	for ii := 0; ii < length; ii++ {
		buff = append(buff, char)
		char++
	}
	return buff
}

func Test_Rabin(t *testing.T) {
	for ii := 0; ii < 256; ii++ {
		buff := makeSequence(ii)
		hash := New()
		hash.Write(buff)
		sum := hash.Sum64()

		cmp := RabinFingerprintFixed(buff)
		if sum != cmp {
			t.Error(fmt.Sprintf("mismatch %d: (%q)", ii, buff))
		}

		hash = New64()
		hash.Write(buff)
		sum = hash.Sum64()
		if sum != cmp {
			t.Error(fmt.Sprintf("mismatch %d: (%q)", ii, buff))
		}
	}
}

func Test_RabinRandom(t *testing.T) {
	testData := makeTestData()

	hash := New()

	// Verify correctness.
	count := 1000
	if count > len(testData) {
		count = len(testData)
	}
	for jj := 0; jj < count; jj++ {
		buff := []byte(testData[jj])
		hash.Write(buff)
		fp := hash.Sum64()
		cmp := RabinFingerprintFixed(buff)
		if fp != cmp {
			t.Error(fmt.Sprintf("mismatch %d: (%q)", jj, buff))
		}
		hash.Reset()
	}
}

func Test_Rabin64Random(t *testing.T) {
	testData := makeTestData()

	hash := New64()

	// Verify correctness.
	count := 1000
	if count > len(testData) {
		count = len(testData)
	}
	for jj := 0; jj < count; jj++ {
		buff := []byte(testData[jj])
		hash.Write(buff)
		fp := hash.Sum64()
		cmp := RabinFingerprintFixed(buff)
		if fp != cmp {
			t.Error(fmt.Sprintf("mismatch %d: (%q)", jj, buff))
		}
		hash.Reset()
	}
}

func Benchmark_Rabin(b *testing.B) {
	b.StopTimer()
	buff := makeSequence(32)
	hash := New()

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		hash.Write(buff)
		hash.Sum64()
		hash.Reset()
	}
}

// Native go implementation.
func Benchmark_RabinGeneric(b *testing.B) {
	b.StopTimer()
	buff := makeSequence(32)
	hash := new(digest)

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		hash.writeGeneric(buff)
		hash.Sum64()
		hash.Reset()
	}
}

func Benchmark_MD5(b *testing.B) {
	b.StopTimer()
	buff := makeSequence(32)
	hash := md5.New()
	sum := make([]byte, md5.Size)

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		hash.Write(buff)
		hash.Sum(sum)
		hash.Reset()
	}
}

func makeRandomUrl(r *rand.Rand) string {
	urlLen := 20
	buff := new(bytes.Buffer)

	for ii := 0; ii < urlLen; ii++ {
		index := r.Intn(len(urlChars))
		buff.WriteRune(urlChars[index])
	}
	return buff.String()
}

// Generates 1 million unique random "URLs" (20 char strings with URL chars)
func makeTestData() []string {
	r := rand.New(rand.NewSource(0))

	seen := make(map[string]bool)

	testData := make([]string, 1000000)
	for ii := 0; ii < len(testData); {
		for {
			url := makeRandomUrl(r)
			if seen[url] {
				continue
			}
			seen[url] = true

			testData[ii] = url
			ii++
			break
		}
	}
	return testData
}

func Benchmark_RabinLong(b *testing.B) {
	b.StopTimer()
	testData := makeTestData()

	hash := New()

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		for jj := 0; jj < len(testData); jj++ {
			hash.Write([]byte(testData[jj]))
			hash.Sum64()
			hash.Reset()
		}
	}
}

func Benchmark_RabinGenericLong(b *testing.B) {
	b.StopTimer()
	testData := makeTestData()

	hash := new(digest)

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		for jj := 0; jj < len(testData); jj++ {
			hash.writeGeneric([]byte(testData[jj]))
			hash.Sum64()
			hash.Reset()
		}
	}
}

func Benchmark_Rabin64Long(b *testing.B) {
	b.StopTimer()
	testData := makeTestData()

	hash := New64()

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		for jj := 0; jj < len(testData); jj++ {
			hash.Write([]byte(testData[jj]))
			hash.Sum64()
			hash.Reset()
		}
	}
}

func Benchmark_Rabin64GenericLong(b *testing.B) {
	b.StopTimer()
	testData := makeTestData()

	hash := new(digest64)

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		for jj := 0; jj < len(testData); jj++ {
			hash.writeGeneric([]byte(testData[jj]))
			hash.Sum64()
			hash.Reset()
		}
	}
}

func Benchmark_MD5Long(b *testing.B) {
	b.StopTimer()
	testData := makeTestData()

	hash := md5.New()

	sum := make([]byte, md5.Size)
	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		for jj := 0; jj < len(testData); jj++ {
			hash.Write([]byte(testData[jj]))
			hash.Sum(sum)
			hash.Reset()
		}
	}
}

func Benchmark_RabinLongCollisions(b *testing.B) {
	b.StopTimer()
	fmt.Println("generating test data")
	testData := makeTestData()

	hash := New()

	fmt.Println("hashing")
	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		// Maintains a count of collisions
		seen := make(map[uint64]int)
		for jj := 0; jj < len(testData); jj++ {
			hash.Write([]byte(testData[jj]))
			h := hash.Sum64()

			count, _ := seen[h]
			seen[h] = count + 1
			if count > 0 {
				b.Error(fmt.Sprintf("collision found (%s) 0x%x", testData[jj], h))
			}

			hash.Reset()
		}
	}
}

func makeBlock(size int) []byte {
	b := make([]byte, size)
	for ii := 0; ii < size; ii++ {
		// Offset by one to avoid a vacuous first element.
		b[ii] = uint8(ii + 1)
	}
	return b
}

func Benchmark_Crc64Block(b *testing.B) {
	b.StopTimer()

	buff := makeBlock(256 * 1024)
	hash := crc64.New(crc64.MakeTable(crc64.ECMA))

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		hash.Write(buff)
		hash.Sum64()
		hash.Reset()
	}
}

func Benchmark_MD5Block(b *testing.B) {
	b.StopTimer()

	buff := makeBlock(256 * 1024)
	hash := md5.New()
	sum := make([]byte, md5.Size)

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		hash.Write(buff)
		hash.Sum(sum)
		hash.Reset()
	}
}

func Benchmark_Rabin32Block(b *testing.B) {
	b.StopTimer()

	buff := makeBlock(256 * 1024)
	hash := New()

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		hash.Write(buff)
		hash.Sum64()
		hash.Reset()
	}
}

func Benchmark_Rabin32GenericBlock(b *testing.B) {
	b.StopTimer()

	buff := makeBlock(256 * 1024)
	hash := new(digest)

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		hash.writeGeneric(buff)
		hash.Sum64()
		hash.Reset()
	}
}

func Benchmark_Rabin64Block(b *testing.B) {
	b.StopTimer()

	buff := makeBlock(256 * 1024)
	hash := New64()

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		hash.Write(buff)
		hash.Sum64()
		hash.Reset()
	}
}

func Benchmark_Rabin64GenericBlock(b *testing.B) {
	b.StopTimer()

	buff := makeBlock(256 * 1024)
	hash := new(digest64)

	b.StartTimer()
	for ii := 0; ii < b.N; ii++ {
		hash.writeGeneric(buff)
		hash.Sum64()
		hash.Reset()
	}
}

func Test_Roll(t *testing.T) {
	hash := NewRolling(128)

	// Load the hash.
	buff := makeSequence(4096)
	hash.Write(buff[:128])
	sum := hash.Sum64()
	cmp := RabinFingerprintFixed(buff[:128])
	if sum != cmp {
		t.Error(fmt.Sprintf("mismatch 0x%x != 0x%x", sum, cmp))
	}

	// Returns true on success.
	checkRoll := func(oldData, newData, cmpData []byte) bool {
		length, err := hash.Roll(oldData, newData)
		if err != nil || length != len(oldData) {
			t.Error("Roll")
		}
		sum = hash.Sum64()
		cmp = RabinFingerprintFixed(cmpData)
		return sum == cmp
	}

	// Move the window 1 byte.  This is vacuous since the first byte
	// is 0.
	if !checkRoll(buff[:1], buff[128:129], buff[1:129]) {
		t.Error("mismatch")
	}

	// Move the window 1 byte.
	if !checkRoll(buff[1:2], buff[129:130], buff[2:130]) {
		t.Error("mismatch")
	}

	// Move the window 2 bytes.
	if !checkRoll(buff[2:4], buff[130:132], buff[4:132]) {
		t.Error("mismatch")
	}

	// Move the window 3 bytes.
	if !checkRoll(buff[4:7], buff[132:135], buff[7:135]) {
		t.Error("mismatch")
	}

	// Move the window 4 bytes.
	if !checkRoll(buff[7:11], buff[135:139], buff[11:139]) {
		t.Error("mismatch")
	}
}
