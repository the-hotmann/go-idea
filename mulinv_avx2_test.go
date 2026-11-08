//go:build goexperiment.simd && amd64.v3

package idea

import (
	"fmt"
	"math/rand/v2"
	"simd/archsimd"
	"testing"
)

func TestMulVecMatchesModularMultiplicationInEveryLane(t *testing.T) {
	var (
		checked, failed int
		first           string
		xs, ys, got     [16]uint16
	)
	check := func() {
		mulVec(archsimd.LoadUint16x16Array(&xs), archsimd.LoadUint16x16Array(&ys)).StoreArray(&got)
		for i := range got {
			checked++
			if want := referenceMul(xs[i], ys[i]); got[i] != want {
				if failed == 0 {
					first = fmt.Sprintf("lane %d: mulVec(%#04x, %#04x) = %#04x, want %#04x", i, xs[i], ys[i], got[i], want)
				}
				failed++
			}
		}
	}
	for _, y := range []uint16{0, 1, 2, 0x7fff, 0x8000, 0x8001, 0xfffe, 0xffff} {
		for start := 0; start < 1<<16; start += len(xs) {
			for i := range xs {
				xs[i], ys[i] = uint16(start+i), y
			}
			check()
			xs, ys = ys, xs
			check()
		}
	}
	r := rand.New(rand.NewPCG(9, 10))
	for range 1 << 16 {
		for i := range xs {
			xs[i], ys[i] = uint16(r.Uint32()), uint16(r.Uint32())
		}
		check()
	}
	if failed > 0 {
		t.Errorf("mulVec is wrong in %d of %d lanes, first %s: the AVX2 key schedule would produce wrong decryption subkeys", failed, checked, first)
	}
}

func BenchmarkMulVec(b *testing.B) {
	b.ReportAllocs()
	r := rand.New(rand.NewPCG(1, 2))
	var ys [16]uint16
	for i := range ys {
		ys[i] = uint16(r.Uint32())
	}
	var (
		x = archsimd.BroadcastUint16x16(1)
		y = archsimd.LoadUint16x16Array(&ys)
	)
	for b.Loop() {
		x = mulVec(x, y)
	}
}
