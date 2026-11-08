//go:build !(goexperiment.simd && amd64.v3)

package idea

func mulInvAll(xs [mulSubkeys]uint16) [mulSubkeys]uint16 {

	var (
		prefix [mulSubkeys]uint16
		a, b   = uint16(1), uint16(1)
	)

	for i := 0; i < mulSubkeys; i += 2 {
		prefix[i], prefix[i+1] = a, b
		a, b = mul(a, xs[i]), mul(b, xs[i+1])
	}

	inv := mulInv(mul(a, b))
	a, b = mul(inv, b), mul(inv, a)
	for i := mulSubkeys - 2; i >= 0; i -= 2 {
		xs[i], a = mul(a, prefix[i]), mul(a, xs[i])
		xs[i+1], b = mul(b, prefix[i+1]), mul(b, xs[i+1])
	}

	return xs

}
