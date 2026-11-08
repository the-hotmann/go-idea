//go:build goexperiment.simd && amd64.v3

package idea

import "simd/archsimd"

func mulVec(x, y archsimd.Uint16x16) archsimd.Uint16x16 {

	var (
		zero   archsimd.Uint16x16
		lo     = x.Mul(y)
		hi     = x.MulHigh(y)
		borrow = lo.Less(hi).ToInt16x16().ToBits()
		isZero = x.Equal(zero).Or(y.Equal(zero))
	)

	{
		x, y := archsimd.BroadcastUint16x16(1).Sub(x).Sub(y), lo.Sub(hi).Sub(borrow)
		return x.IfElse(isZero, y)
	}

}

func mulInvAll(xs [mulSubkeys]uint16) [mulSubkeys]uint16 {

	var buf [32]uint16
	copy(buf[:], xs[:])

	var (
		ya = archsimd.LoadUint16x16Array((*[16]uint16)(buf[:16]))
		yb = archsimd.LoadUint16x16Array((*[16]uint16)(buf[16:]))
	)

	for _, k := range [...]int{1, 2, 4, 8} {
		ta, tb := ya, yb
		for range k {
			ta, tb = mulVec(ta, ta), mulVec(tb, tb)
		}
		ya, yb = mulVec(ta, ya), mulVec(tb, yb)
	}

	ya.StoreArray((*[16]uint16)(buf[:16]))
	yb.StoreArray((*[16]uint16)(buf[16:]))

	return [mulSubkeys]uint16(buf[:mulSubkeys])

}
