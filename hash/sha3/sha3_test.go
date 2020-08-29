// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha3

// Tests include all the ShortMsgKATs provided by the Keccak team at
// https://github.com/gvanas/KeccakCodePackage
//
// They only include the zero-bit case of the bitwise testvectors
// published by NIST in the draft of FIPS-202.

import (
	"bytes"
	"compress/flate"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"strings"
	"testing"
)

const (
	testString  = "brekeccakkeccak koax koax"
	katFilename = "testdata/keccakKats.json.deflate"
)

// testDigests contains functions returning hash.Hash instances
// with output-length equal to the KAT length for SHA-3, Keccak
// and SHAKE instances.
var testDigests = map[string]func() hash.Hash{
	"SHA3-224": New224,
	"SHA3-256": New256,
	"SHA3-384": New384,
	"SHA3-512": New512,
}

// testShakes contains functions that return sha3.ShakeHash instances for
// with output-length equal to the KAT length.
var testShakes = map[string]struct {
	constructor  func(N []byte, S []byte) ShakeHash
	defAlgoName  string
	defCustomStr string
}{
	// NewCShake without customization produces same result as SHAKE
	"SHAKE128":  {NewCShake128, "", ""},
	"SHAKE256":  {NewCShake256, "", ""},
	"cSHAKE128": {NewCShake128, "CSHAKE128", "CustomStrign"},
	"cSHAKE256": {NewCShake256, "CSHAKE256", "CustomStrign"},
}

// decodeHex converts a hex-encoded string into a raw byte string.
func decodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// structs used to marshal JSON test-cases.
type KeccakKats struct {
	Kats map[string][]struct {
		Digest  string `json:"digest"`
		Length  int64  `json:"length"`
		Message string `json:"message"`

		// Defined only for cSHAKE
		N string `json:"N"`
		S string `json:"S"`
	}
}

func testUnalignedAndGeneric(t *testing.T, testf func(impl string)) {
	xorInOrig, copyOutOrig := xorIn, copyOut
	xorIn, copyOut = xorInGeneric, copyOutGeneric
	testf("generic")
	if xorImplementationUnaligned != "generic" {
		xorIn, copyOut = xorInUnaligned, copyOutUnaligned
		testf("unaligned")
	}
	xorIn, copyOut = xorInOrig, copyOutOrig
}

// TestKeccakKats tests the SHA-3 and Shake implementations against all the
// ShortMsgKATs from https://github.com/gvanas/KeccakCodePackage
// (The testvectors are stored in keccakKats.json.deflate due to their length.)
func TestKeccakKats(t *testing.T) {
	testUnalignedAndGeneric(t, func(impl string) {
		// Read the KATs.
		deflated, err := os.Open(katFilename)
		if err != nil {
			t.Errorf("error opening %s: %s", katFilename, err)
		}
		file := flate.NewReader(deflated)
		dec := json.NewDecoder(file)
		var katSet KeccakKats
		err = dec.Decode(&katSet)
		if err != nil {
			t.Errorf("error decoding KATs: %s", err)
		}

		for algo, function := range testDigests {
			d := function()
			for _, kat := range katSet.Kats[algo] {
				d.Reset()
				in, err := hex.DecodeString(kat.Message)
				if err != nil {
					t.Errorf("error decoding KAT: %s", err)
				}
				d.Write(in[:kat.Length/8])
				got := strings.ToUpper(hex.EncodeToString(d.Sum(nil)))
				if got != kat.Digest {
					t.Errorf("function=%s, implementation=%s, length=%d\nmessage:\n %s\ngot:\n %s\nwanted:\n %s",
						algo, impl, kat.Length, kat.Message, got, kat.Digest)
					t.Logf("wanted %+v", kat)
					t.FailNow()
				}
				continue
			}
		}
		for algo, v := range testShakes {
			for _, kat := range katSet.Kats[algo] {
				N, err := hex.DecodeString(kat.N)
				if err != nil {
					t.Errorf("error decoding KAT: %s", err)
				}

				S, err := hex.DecodeString(kat.S)
				if err != nil {
					t.Errorf("error decoding KAT: %s", err)
				}
				d := v.constructor(N, S)
				in, err := hex.DecodeString(kat.Message)
				if err != nil {
					t.Errorf("error decoding KAT: %s", err)
				}

				d.Write(in[:kat.Length/8])
				out := make([]byte, len(kat.Digest)/2)
				d.Read(out)
				got := strings.ToUpper(hex.EncodeToString(out))
				if got != kat.Digest {
					t.Errorf("function=%s, implementation=%s, length=%d N:%s\n S:%s\nmessage:\n %s \ngot:\n  %s\nwanted:\n %s",
						algo, impl, kat.Length, kat.N, kat.S, kat.Message, got, kat.Digest)
					t.Logf("wanted %+v", kat)
					t.FailNow()
				}
				continue
			}
		}
	})
}

// TestUnalignedWrite tests that writing data in an arbitrary pattern with
// small input buffers.
func TestUnalignedWrite(t *testing.T) {
	testUnalignedAndGeneric(t, func(impl string) {
		buf := generateData(0x10000)
		for alg, df := range testDigests {
			d := df()
			d.Reset()
			d.Write(buf)
			want := d.Sum(nil)
			d.Reset()
			for i := 0; i < len(buf); {
				// Cycle through offsets which make a 137 byte sequence.
				// Because 137 is prime this sequence should exercise all corner cases.
				offsets := [17]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 1}
				for _, j := range offsets {
					if v := len(buf) - i; v < j {
						j = v
					}
					d.Write(buf[i : i+j])
					i += j
				}
			}
			got := d.Sum(nil)
			if !bytes.Equal(got, want) {
				t.Errorf("Unaligned writes, implementation=%s, alg=%s\ngot %q, want %q", impl, alg, got, want)
			}
		}

		// Same for SHAKE
		for alg, df := range testShakes {
			want := make([]byte, 16)
			got := make([]byte, 16)
			d := df.constructor([]byte(df.defAlgoName), []byte(df.defCustomStr))

			d.Reset()
			d.Write(buf)
			d.Read(want)
			d.Reset()
			for i := 0; i < len(buf); {
				// Cycle through offsets which make a 137 byte sequence.
				// Because 137 is prime this sequence should exercise all corner cases.
				offsets := [17]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 1}
				for _, j := range offsets {
					if v := len(buf) - i; v < j {
						j = v
					}
					d.Write(buf[i : i+j])
					i += j
				}
			}
			d.Read(got)
			if !bytes.Equal(got, want) {
				t.Errorf("Unaligned writes, implementation=%s, alg=%s\ngot %q, want %q", impl, alg, got, want)
			}
		}
	})
}

// TestAppend checks that appending works when reallocation is necessary.
func TestAppend(t *testing.T) {
	testUnalignedAndGeneric(t, func(impl string) {
		d := New224()

		for capacity := 2; capacity <= 66; capacity += 64 {
			// The first time around the loop, Sum will have to reallocate.
			// The second time, it will not.
			buf := make([]byte, 2, capacity)
			d.Reset()
			d.Write([]byte{0xcc})
			buf = d.Sum(buf)
			expected := "0000DF70ADC49B2E76EEE3A6931B93FA41841C3AF2CDF5B32A18B5478C39"
			if got := strings.ToUpper(hex.EncodeToString(buf)); got != expected {
				t.Errorf("got %s, want %s", got, expected)
			}
		}
	})
}

// TestAppendNoRealloc tests that appending works when no reallocation is necessary.
func TestAppendNoRealloc(t *testing.T) {
	testUnalignedAndGeneric(t, func(impl string) {
		buf := make([]byte, 1, 200)
		d := New224()
		d.Write([]byte{0xcc})
		buf = d.Sum(buf)
		expected := "00DF70ADC49B2E76EEE3A6931B93FA41841C3AF2CDF5B32A18B5478C39"
		if got := strings.ToUpper(hex.EncodeToString(buf)); got != expected {
			t.Errorf("%s: got %s, want %s", impl, got, expected)
		}
	})
}

// TestSqueezing checks that squeezing the full output a single time produces
// the same output as repeatedly squeezing the instance.
func TestSqueezing(t *testing.T) {
	testUnalignedAndGeneric(t, func(impl string) {
		for algo, v := range testShakes {
			d0 := v.constructor([]byte(v.defAlgoName), []byte(v.defCustomStr))
			d0.Write([]byte(testString))
			ref := make([]byte, 32)
			d0.Read(ref)

			d1 := v.constructor([]byte(v.defAlgoName), []byte(v.defCustomStr))
			d1.Write([]byte(testString))
			var multiple []byte
			for range ref {
				one := make([]byte, 1)
				d1.Read(one)
				multiple = append(multiple, one...)
			}
			if !bytes.Equal(ref, multiple) {
				t.Errorf("%s (%s): squeezing %d bytes one at a time failed", algo, impl, len(ref))
			}
		}
	})
}

func doSum(h hash.Hash, data []byte) (digest []byte) {
	half := int(len(data) / 2)
	h.Write(data[:half])
	h.Write(data[half:])
	digest = h.Sum(data[:0])
	return
}

// generateData produces a buffer of size consecutive bytes 0x00, 0x01, ..., used for testing.
func generateData(size int) []byte {
	result := make([]byte, size)
	for i := range result {
		result[i] = byte(i)
	}
	return result
}

func TestReset(t *testing.T) {
	out1 := make([]byte, 32)
	out2 := make([]byte, 32)

	for _, v := range testShakes {
		// Calculate hash for the first time
		c := v.constructor([]byte(v.defAlgoName), []byte(v.defCustomStr))
		c.Write(generateData(0x100))
		c.Read(out1)

		// Calculate hash again
		c.Reset()
		c.Write(generateData(0x100))
		c.Read(out2)

		if !bytes.Equal(out1, out2) {
			t.Error("\nExpected:\n", out1, "\ngot:\n", out2)
		}
	}
}

func TestClone(t *testing.T) {
	out1 := make([]byte, 16)
	out2 := make([]byte, 16)
	in := generateData(0x100)

	for _, v := range testShakes {
		h1 := v.constructor([]byte(v.defAlgoName), []byte(v.defCustomStr))
		h1.Write([]byte{0x01})

		h2 := h1.Clone()

		h1.Write(in)
		h1.Read(out1)

		h2.Write(in)
		h2.Read(out2)

		if !bytes.Equal(out1, out2) {
			t.Error("\nExpected:\n", hex.EncodeToString(out1), "\ngot:\n", hex.EncodeToString(out2))
		}
	}
}

// BenchmarkPermutationFunction measures the speed of the permutation function
// with no input data.
func BenchmarkPermutationFunction(b *testing.B) {
	b.SetBytes(int64(200))
	var lanes [25]uint64
	for i := 0; i < b.N; i++ {
		keccakF1600(&lanes)
	}
}

// benchmarkHash tests the speed to hash num buffers of buflen each.
// This function uses heap
func benchmarkHashChunked(b *testing.B, h hash.Hash, size, num int) {
	b.StopTimer()
	data := generateData(size)
	digestBuf := make([]byte, h.Size())
	b.SetBytes(int64(size * num))
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		h.Reset()
		for j := 0; j < num; j++ {
			h.Write(data)
		}
		digestBuf = h.Sum(digestBuf[:])
		// needed to avoid alocations
		digestBuf = digestBuf[:0]
	}
	b.StopTimer()
	h.Reset()
}

// benchmarkShake is specialized to the Shake instances, which don't
// require a copy on reading output.
func benchmarkShake(b *testing.B, h ShakeHash, size, num int) {
	b.StopTimer()
	out := make([]byte, 32)
	data := generateData(size)

	b.SetBytes(int64(size * num))
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		h.Reset()
		for j := 0; j < num; j++ {
			h.Write(data)
		}
		h.Read(out[:])
	}
}

var domainString = []byte("SHAKE")
var customString = []byte("CustomString")

// benchmarkShake is specialized to the Shake instances, which don't
// require a copy on reading output.
func benchmarkCShake(b *testing.B, f func(N, S []byte) ShakeHash, size, num int) {
	b.StopTimer()
	h := f(domainString, customString)
	out := make([]byte, 32)
	data := generateData(size)

	b.SetBytes(int64(size * num))
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		h.Reset()
		for j := 0; j < num; j++ {
			h.Write(data)
		}
		h.Read(out[:])
	}
}

func BenchmarkSha3Chunk_x01(b *testing.B) {
	b.Run("SHA-3/224", func(b *testing.B) { benchmarkHashChunked(b, New224(), 2047, 1) })
	b.Run("SHA-3/256", func(b *testing.B) { benchmarkHashChunked(b, New256(), 2047, 1) })
	b.Run("SHA-3/384", func(b *testing.B) { benchmarkHashChunked(b, New384(), 2047, 1) })
	b.Run("SHA-3/512", func(b *testing.B) { benchmarkHashChunked(b, New512(), 2047, 1) })
}

func BenchmarkSha3Chunk_x16(b *testing.B) {
	b.Run("SHA-3/224", func(b *testing.B) { benchmarkHashChunked(b, New224(), 16, 1024) })
	b.Run("SHA-3/256", func(b *testing.B) { benchmarkHashChunked(b, New256(), 16, 1024) })
	b.Run("SHA-3/384", func(b *testing.B) { benchmarkHashChunked(b, New384(), 16, 1024) })
	b.Run("SHA-3/512", func(b *testing.B) { benchmarkHashChunked(b, New512(), 16, 1024) })
}

func BenchmarkShake_x01(b *testing.B) {
	b.Run("SHAKE-128", func(b *testing.B) { benchmarkShake(b, NewShake128(), 1350, 1) })
	b.Run("SHAKE-256", func(b *testing.B) { benchmarkShake(b, NewShake256(), 1350, 1) })
}

func BenchmarkShake_x16(b *testing.B) {
	b.Run("SHAKE-128", func(b *testing.B) { benchmarkShake(b, NewShake128(), 16, 1024) })
	b.Run("SHAKE-256", func(b *testing.B) { benchmarkShake(b, NewShake256(), 16, 1024) })
}

func BenchmarkCShake(b *testing.B) {
	b.Run("cSHAKE-128", func(b *testing.B) { benchmarkCShake(b, NewCShake128, 2047, 1) })
	b.Run("cSHAKE-256", func(b *testing.B) { benchmarkCShake(b, NewCShake256, 2047, 1) })
}

func Example_sum() {
	buf := []byte("some data to hash")
	// A hash needs to be 64 bytes long to have 256-bit collision resistance.
	h := make([]byte, 64)
	// Compute a 64-byte hash of buf and put it in h.
	ShakeSum256(h, buf)
	fmt.Printf("%x\n", h)
	// Output: 0f65fe41fc353e52c55667bb9e2b27bfcc8476f2c413e9437d272ee3194a4e3146d05ec04a25d16b8f577c19b82d16b1424c3e022e783d2b4da98de3658d363d
}

func Example_mac() {
	k := []byte("this is a secret key; you should generate a strong random key that's at least 32 bytes long")
	buf := []byte("and this is some data to authenticate")
	// A MAC with 32 bytes of output has 256-bit security strength -- if you use at least a 32-byte-long key.
	h := make([]byte, 32)
	d := NewShake256()
	// Write the key into the hash.
	d.Write(k)
	// Now write the data.
	d.Write(buf)
	// Read 32 bytes of output from the hash into h.
	d.Read(h)
	fmt.Printf("%x\n", h)
	// Output: 78de2974bd2711d5549ffd32b753ef0f5fa80a0db2556db60f0987eb8a9218ff
}

func ExampleCShake256() {
	out := make([]byte, 32)
	msg := []byte("The quick brown fox jumps over the lazy dog")

	// Example 1: Simple cshake
	c1 := NewCShake256([]byte("NAME"), []byte("Partition1"))
	c1.Write(msg)
	c1.Read(out)
	fmt.Println(hex.EncodeToString(out))

	// Example 2: Different customization string produces different digest
	c1 = NewCShake256([]byte("NAME"), []byte("Partition2"))
	c1.Write(msg)
	c1.Read(out)
	fmt.Println(hex.EncodeToString(out))

	// Example 3: Different output length produces different digest
	out = make([]byte, 64)
	c1 = NewCShake256([]byte("NAME"), []byte("Partition1"))
	c1.Write(msg)
	c1.Read(out)
	fmt.Println(hex.EncodeToString(out))

	// Example 4: Next read produces different result
	c1.Read(out)
	fmt.Println(hex.EncodeToString(out))

	// Output:
	//a90a4c6ca9af2156eba43dc8398279e6b60dcd56fb21837afe6c308fd4ceb05b
	//a8db03e71f3e4da5c4eee9d28333cdd355f51cef3c567e59be5beb4ecdbb28f0
	//a90a4c6ca9af2156eba43dc8398279e6b60dcd56fb21837afe6c308fd4ceb05b9dd98c6ee866ca7dc5a39d53e960f400bcd5a19c8a2d6ec6459f63696543a0d8
	//85e73a72228d08b46515553ca3a29d47df3047e5d84b12d6c2c63e579f4fd1105716b7838e92e981863907f434bfd4443c9e56ea09da998d2f9b47db71988109
}

func ExampleSum256() {
	d := generateData(32)
	var data [32]byte
	h := New256()
	h.Write(d)
	s1 := h.Sum(data[:0])
	fmt.Printf("%X\n", s1)
	//Output:
	// 050A48733BD5C2756BA95C5828CC83EE16FABCD3C086885B7744F84A0F9E0D94
}
