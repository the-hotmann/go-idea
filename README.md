# go-idea

A pure-Go implementation of the **IDEA** (International Data Encryption Algorithm)
block cipher, with an optional SIMD-accelerated key schedule.

- 64-bit block, 128-bit key
- Implements the standard [`crypto/cipher.Block`](https://pkg.go.dev/crypto/cipher#Block) interface, so it composes with `cipher.NewCBCEncrypter`, `cipher.NewCTR`, and the rest of the standard modes
- **Zero allocations** on every operation — key setup, encrypt and decrypt
- **Constant-time**: no data-dependent branches in the cipher core or in modular inversion (enforced by a test that disassembles the package and audits its conditional jumps)
- Optional AVX2/AVX-512 key schedule via Go's `simd` experiment, with a pure-Go fallback that is always built when the experiment is off

```go
import "github.com/the-hotmann/go-idea"
```

## Install

```sh
go get github.com/the-hotmann/go-idea
```

Requires **Go 1.27+**.

## Usage

`NewCipher` returns a `cipher.Block`. Use it directly for a single block, or wrap
it in any standard cipher mode.

```go
package main

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	idea "github.com/the-hotmann/go-idea"
)

func main() {
	key := make([]byte, 16) // IDEA keys are exactly 16 bytes
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}

	block, err := idea.NewCipher(key)
	if err != nil {
		panic(err) // only returned for a key that is not 16 bytes
	}

	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		panic(err)
	}

	plaintext := []byte("8 bytes!") // must be a multiple of the 8-byte block
	ciphertext := make([]byte, len(plaintext))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, plaintext)

	fmt.Printf("%x\n", ciphertext)
}
```

A block returned by `NewCipher` is immutable and safe for concurrent use by
multiple goroutines.

An invalid key length yields a typed `KeySizeError`:

```go
_, err := idea.NewCipher(make([]byte, 8))
// err is idea.KeySizeError(8): "idea: invalid key size 8"
```

## SIMD acceleration

The key schedule inverts 20 multiplicative subkeys. On `amd64` that batch
inversion is vectorized with 16-wide `uint16` lanes. It is enabled when **both**:

- the `simd` experiment is on: `GOEXPERIMENT=simd`, and
- the target is AVX2 or newer: `GOAMD64=v3` (or `v4`).

```sh
GOEXPERIMENT=simd GOAMD64=v3 go build ./...
```

Everywhere else — including a plain `go build` with no experiment — the identical,
constant-time pure-Go path is compiled instead. The selection is a build tag, so
it costs nothing at runtime and the two paths are verified to produce identical
output.

## Benchmarks

Measured on an AMD EPYC (Genoa), Go 1.27, `GOEXPERIMENT=simd`, `-count=6`.
Every operation is **0 B/op, 0 allocs/op**.

### Top level

| Operation            | Generic (`v1`) | SIMD (`v3`, AVX2) |
|----------------------|---------------:|------------------:|
| `NewCipher` (key setup) | 209 ns       | **164 ns** (−21%) |
| `Encrypt` (one block)   | 86 ns        | 86 ns             |
| `Decrypt` (one block)   | 86 ns        | 86 ns             |

Block throughput is ~88 MiB/s per core. SIMD targets the **key schedule**, not
the block transform, so the win shows up in `NewCipher` (and any workload that
re-keys often), while per-block `Encrypt`/`Decrypt` are unchanged.

### Details

Full matrix across `GOAMD64` levels (`v1` generic SSE2 baseline, `v3`/`v4` take
the vectorized key schedule). `MulInvAll` is the batch inversion; `InvertKey` and
`NewCipher` are what it feeds:

```
                     │   generic    │             sse2 (v2)             │            avx2 (v3)             │           avx512 (v4)            │
                     │    sec/op    │    sec/op     vs base             │    sec/op     vs base            │    sec/op     vs base            │
NewCipher-20            209.1n ± 1%    202.8n ± 2%   -2.99%               164.2n ± 2%  -21.45%               164.2n ± 2%  -21.47%
Encrypt-20              86.21n ± 4%    86.90n ± 2%        ~               86.78n ± 2%        ~               87.31n ± 1%        ~
Decrypt-20              85.75n ± 3%    86.30n ± 2%        ~               85.77n ± 2%        ~               87.85n ± 3%        ~
Mul-20                  5.309n ± 3%    5.324n ± 1%        ~               5.239n ± 2%        ~               5.361n ± 1%        ~
MulInv-20               53.48n ± 2%    52.06n ± 2%   -2.66%               54.05n ± 2%        ~               53.84n ± 3%        ~
MulInvAll-20            150.4n ± 1%    154.4n ± 5%        ~               115.4n ± 2%  -23.30%               119.4n ± 3%  -20.61%
InvertKey-20            190.1n ± 5%    182.2n ± 2%   -4.10%               148.8n ± 3%  -21.73%               148.7n ± 3%  -21.76%
```

AVX2 and AVX-512 are equivalent here: both use the 16-wide `uint16` path, so `v4`
adds nothing over `v3`.

Reproduce (runs `v1 v2 v3 v4` and diffs them with `benchstat`):

```sh
make bench
```

## Security notes

IDEA is a 1991 cipher with a 64-bit block. Its patents have long expired, so it is
free to use, but the small block size makes it unsuitable as a default for new
systems — prefer AES. Use this package for interoperability with existing IDEA
data and for study.

This is an independent implementation and has not been externally audited. It is
verified against the NESSIE IDEA-128 test vectors and checked for data-dependent
branching, but that is not a substitute for an audit.

## License

MIT. See [LICENSE](LICENSE).
