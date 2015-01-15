# Go implementation of Rabin fingerprinting with SSE2 optimizations

Number are based on an Intel Core 2 Duo 2.26ghz:

```
$ go test -bench=".*"
PASS
Benchmark_Rabin 20000000               105 ns/op
Benchmark_RabinGeneric  10000000               126 ns/op
Benchmark_MD5    2000000               720 ns/op
Benchmark_RabinLong            5         203355985 ns/op
Benchmark_RabinGenericLong             5         214017019 ns/op
Benchmark_Rabin64Long         10         180310324 ns/op
Benchmark_Rabin64GenericLong          10         191728738 ns/op
Benchmark_MD5Long              2         665961119 ns/op
...
Benchmark_Crc64Block        2000           1059137 ns/op
Benchmark_MD5Block          3000            592506 ns/op
Benchmark_Rabin32Block      3000            574193 ns/op
Benchmark_Rabin32GenericBlock       2000            812863 ns/op
Benchmark_Rabin64Block      5000            353489 ns/op
Benchmark_Rabin64GenericBlock       3000            586740 ns/op
```

In the above, "Generic" versions use native Go implementations.  `Rabin32` and `Rabin64` use SSE and SSE2, respectively,
if available.  The SSE/SSE2 optimizations are more of an exercise here, as the performance increase is marginal: 15-40%,
depending on input data size.  This is likely because the SSE/SSE2 implementation only optimizes XORs and 
consumes data at the same rate as the native versions (due to hash state).

Of interest are the "Block" benchmarks, which run various schemes over 256KB inputs.  `Benchmark_Crc64Block` uses
`hash/crc64` from Go's standard library and is a useful comparison with Rabin fingerprinting, since both are
fundamentally similar (i.e., operating with polynomials over GF(2)).
