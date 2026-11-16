// Package hashkit provides the hashing algorithms supported by the checker.
package hashkit

import (
	"encoding/binary"
	"hash"
	"math/bits"
)

// sm3IV is the initial value defined by GB/T 32905-2016.
var sm3IV = [8]uint32{
	0x7380166f, 0x4914b2b9, 0x172442d7, 0xda8a0600,
	0xa96f30bc, 0x163138aa, 0xe38dee4d, 0xb0fb0e4e,
}

type sm3Digest struct {
	h   [8]uint32
	buf [64]byte
	nx  int
	len uint64
}

// NewSM3 returns a hash.Hash computing the Chinese SM3 digest.
func NewSM3() hash.Hash {
	d := &sm3Digest{}
	d.Reset()
	return d
}

func (d *sm3Digest) Size() int      { return 32 }
func (d *sm3Digest) BlockSize() int { return 64 }

func (d *sm3Digest) Reset() {
	d.h = sm3IV
	d.nx = 0
	d.len = 0
}

func (d *sm3Digest) Write(p []byte) (int, error) {
	n := len(p)
	d.len += uint64(n)

	if d.nx > 0 {
		k := copy(d.buf[d.nx:], p)
		d.nx += k
		if d.nx == 64 {
			sm3Block(&d.h, d.buf[:])
			d.nx = 0
		}
		p = p[k:]
	}
	for len(p) >= 64 {
		sm3Block(&d.h, p[:64])
		p = p[64:]
	}
	if len(p) > 0 {
		d.nx = copy(d.buf[:], p)
	}
	return n, nil
}

func (d *sm3Digest) Sum(in []byte) []byte {
	// Work on a copy so that Sum does not change the running state.
	c := *d
	bitLen := c.len << 3

	var pad [72]byte
	pad[0] = 0x80
	padLen := 56 - int(c.len%64)
	if padLen <= 0 {
		padLen += 64
	}
	binary.BigEndian.PutUint64(pad[padLen:], bitLen)
	_, _ = c.Write(pad[:padLen+8])

	var out [32]byte
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(out[i*4:], c.h[i])
	}
	return append(in, out[:]...)
}

func sm3Block(h *[8]uint32, block []byte) {
	var w [68]uint32
	var w1 [64]uint32

	for i := 0; i < 16; i++ {
		w[i] = binary.BigEndian.Uint32(block[i*4:])
	}
	for i := 16; i < 68; i++ {
		w[i] = sm3P1(w[i-16]^w[i-9]^bits.RotateLeft32(w[i-3], 15)) ^
			bits.RotateLeft32(w[i-13], 7) ^ w[i-6]
	}
	for i := 0; i < 64; i++ {
		w1[i] = w[i] ^ w[i+4]
	}

	a, b, c, dd := h[0], h[1], h[2], h[3]
	e, f, g, hh := h[4], h[5], h[6], h[7]

	for j := 0; j < 64; j++ {
		var t uint32 = 0x79cc4519
		if j >= 16 {
			t = 0x7a879d8a
		}
		ss1 := bits.RotateLeft32(
			bits.RotateLeft32(a, 12)+e+bits.RotateLeft32(t, j), 7)
		ss2 := ss1 ^ bits.RotateLeft32(a, 12)

		var ff, gg uint32
		if j < 16 {
			ff = a ^ b ^ c
			gg = e ^ f ^ g
		} else {
			ff = (a & b) | (a & c) | (b & c)
			gg = (e & f) | (^e & g)
		}

		tt1 := ff + dd + ss2 + w1[j]
		tt2 := gg + hh + ss1 + w[j]

		dd = c
		c = bits.RotateLeft32(b, 9)
		b = a
		a = tt1
		hh = g
		g = bits.RotateLeft32(f, 19)
		f = e
		e = sm3P0(tt2)
	}

	h[0] ^= a
	h[1] ^= b
	h[2] ^= c
	h[3] ^= dd
	h[4] ^= e
	h[5] ^= f
	h[6] ^= g
	h[7] ^= hh
}

func sm3P0(x uint32) uint32 {
	return x ^ bits.RotateLeft32(x, 9) ^ bits.RotateLeft32(x, 17)
}

func sm3P1(x uint32) uint32 {
	return x ^ bits.RotateLeft32(x, 15) ^ bits.RotateLeft32(x, 23)
}
