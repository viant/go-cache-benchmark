# cache-bench

**This repository is currently a work-in-progress.**

Requires at least Go 1.13

Benchmarks for comparing cache Golang cache libraries.

Based off https://github.com/allegro/bigcache-bench.

The primary motivation of this repository is to create an even comparison of caching libraries.

Currently compares the following available libraries / implementations:

1. [FreeCache](https://github.com/coocood/freecache)
2. [BigCache](https://github.com/allegro/bigcache)
3. [SCache](https://github.com/viant/scache)
6. [`golang-lru`](https://github.com/hashicorp/golang-lru) 

# Descriptions of benchmarks

The benchmarks run, for each cache implementation, by creating a cache with a set maximum size which, by default, is 10M, will:

1. Benchmark setting of values without eviction, in serial and parallel.
2. Prepopulate and benchmark getting values without misses, in serial and parallel.
3. Prepopulate and benchmark eviction policy using requests following a Zipf and uniform distribution, setting on cache miss, in parallel only.

# Observations

1. `scache` seems to allocate the least memory on read or write. This is shown by looking at the GC count (`gc`) metric. Using default GC configuration, `scache` seems to have about 1/10th the number of GCs. This is more evident the longer the benchmarks are run.
2. `scache` seems to be generally the fastest performing cache. Standard operation seems to occur in 68% of `BigCache` and 45% of `freecache`.
3. `scache` tends to have the highest miss rate with the eviction tests. 
    - For the Zipf distribution, `freecache` and `bigcache` seem to get about 3% miss rates, `golang-lru` gets about 7%, but `scache` seems to get about a 11% miss rate. 
    - For the uniform distribution, `freecache` and `bigcache` get about 22%, `golang-lru` gets about 50%, and `scache` gets about 65% miss rate.
    - This could be problematic if cache misses are dramatically more expensive than cache hits.
4. `scache` seems to start dramatically drop hit rate as soon as the cache is too small, whereas the other caches slowly begin to drop their hit rates.
5. `scache` seems to have consistent performance when an eviction is required, whereas other cache implementations can be twice as slow. 
6. `scache` requires at least 2x as much memory to get a similar hit rate as other caches.
7. For difficult to cache data usage distributions, hit rates eventually converge for all caches.

Do note that within the scope of usage, the cache overhead is most likely not the biggest cost. 
Although "twice as slow" may sound scary, if the application using the cache takes about 1ms to respond to a request, the cache overhead on miss would at most add 0.001ms or about 0.1% additional time to the request. 
Also, as one of the primary purposes of a cache is to memoize long-running computations, usually the cache overhead from a miss is insigificant to the computation required to populate the cache entry.

# Sample results

Run on VM

```
$ uname -a
Linux alpine-1 5.10.45-0-virt #1-Alpine SMP Mon, 21 Jun 2021 07:19:03 +0000 x86_64 Linux 

$ free -m
total        used        free      shared  buff/cache   available
Mem:           3942         141        3455           0         344        3740
Swap:          3998         207        3791

$ cat /proc/cpuinfo | egrep 'processor|cpu (MHz|cores)|cache' 
processor       : 0
cpu MHz         : 2593.994
cache size      : 6144 KB
cpu cores       : 4
cache_alignment : 64
processor       : 1
cpu MHz         : 2593.994
cache size      : 6144 KB
cpu cores       : 4
cache_alignment : 64
processor       : 2
cpu MHz         : 2593.994
cache size      : 6144 KB
cpu cores       : 4
cache_alignment : 64
processor       : 3
cpu MHz         : 2593.994
cache size      : 6144 KB
cpu cores       : 4
cache_alignment : 64

$ docker run --rm -it -v "$PWD":/w -w /w golang:1.15 go test -bench=. -benchtime=10000000x .
go: downloading github.com/viant/scache v0.5.0
go: downloading github.com/hashicorp/golang-lru v0.5.4
go: downloading github.com/coocood/freecache v1.1.0
go: downloading github.com/allegro/bigcache/v2 v2.1.3
go: downloading github.com/cespare/xxhash v1.1.0
go: downloading github.com/pkg/errors v0.9.1
goos: linux
goarch: amd64
pkg: github.com/viant/go-cache-benchmark
BenchmarkFreeCacheSet/1000000-4                 10000000               757 ns/op                10.0 gc
BenchmarkBigCacheSet/1000000-4                  10000000               884 ns/op                 9.00 gc
BenchmarkSCacheSet/1000000-4                    10000000               514 ns/op                 8.00 gc
BenchmarkFreeCacheGet/1000000-4                 10000000               862 ns/op                15.0 gc
BenchmarkBigCacheGet/1000000-4                  10000000               750 ns/op                14.0 gc
BenchmarkSCacheGet/1000000-4                    10000000               632 ns/op                 7.00 gc
BenchmarkFreeCacheSetParallel/1000000-4         10000000               434 ns/op                 8.00 gc
BenchmarkBigCacheSetParallel/1000000-4          10000000               402 ns/op                 9.00 gc
BenchmarkFreeCacheGetParallel/1000000-4         10000000               313 ns/op                13.0 gc
BenchmarkBigCacheGetParallel/1000000-4          10000000               225 ns/op                15.0 gc
BenchmarkSCacheGetParallel/1000000-4            10000000               168 ns/op                 7.00 gc
BenchmarkFreeCacheEvictZipfParallel/1000000-4           10000000               432 ns/op            427452 expc          7.00 gc            345323 miss
BenchmarkBigCacheEvictZipfParallel/1000000-4            10000000               353 ns/op            426961 expc          8.00 gc            345126 miss
BenchmarkSCacheEvictZipfParallel/1000000-4              10000000               271 ns/op            427927 expc          2.00 gc           1134548 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-4          10000000              1137 ns/op            427359 expc          3.00 gc            683287 miss
BenchmarkFreeCacheEvictUniformParallel/1000000-4        10000000               560 ns/op           4999961 expc          7.00 gc           2232494 miss
BenchmarkBigCacheEvictUniformParallel/1000000-4         10000000               554 ns/op           5000092 expc          7.00 gc           2154031 miss
BenchmarkSCacheEvictUniformParallel/1000000-4           10000000               468 ns/op           5001840 expc          4.00 gc           6530118 miss
BenchmarkHashiCacheEvictUniformParallel/1000000-4       10000000              2054 ns/op           4998118 expc          6.00 gc           5001616 miss
PASS
ok      github.com/viant/go-cache-benchmark     133.718s

```

# Outputs and reported metrics

Aside from the standard benchmark metrics, there are custom metrics reported:

* `gc` - the difference of garbage collections run since the start of the benchmark until the end of the benchmark. Refer to [`runtime.ReadMemStats`](https://pkg.go.dev/runtime#MemStats).
* `expc` - Shows on benchmarks with eviction distributions. Represents "ex-pre-cache", or values that were not within the initial precaching of the cache.
* `miss` - Shows on benchmarks with eviction distributions. Counts the number of times the cache responded to a "get" with a miss.

# Running tests

## Using native Go

Requires at least Go 1.15

Generally recommended to use `-benchtime=Xx` instead of `-benchtime=Xs`, especially to see effects on hit rate.

*Run all benchmarks*

`go test -benchtime=10000000x -bench=. .`

*Run eviction strategy benchmarks*

`go test -benchtime=10000000x -bench=Evict .`

*Extend range of possible inputs for Zipf*

`ZIPF_FACTOR=8 go test -benchtime=10000000x -bench=Zipf .`

*Try different ranges for possible inputs for Zipf*

`SWEEP_DIST='[0.999,1,1.001,1.01,1.1,2,10,100,1000]' go test -benchtime=10000000x -bench=Zipf`

```
goos: linux
goarch: amd64
pkg: github.com/viant/go-cache-benchmark
BenchmarkFreeCacheEvictZipfParallel/1000000-0000.9990-4                 10000000               428 ns/op                 0 expc          7.00 gc                0 miss
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.0000-4                 10000000               405 ns/op                 0 expc          8.00 gc                0 miss
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.0010-4                 10000000               418 ns/op               663 expc          8.00 gc              498 miss
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.0100-4                 10000000               409 ns/op              6469 expc          8.00 gc             4768 miss
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.1000-4                 10000000               423 ns/op             61016 expc          8.00 gc            45701 miss
BenchmarkFreeCacheEvictZipfParallel/1000000-0002.0000-4                 10000000               409 ns/op            427799 expc          7.00 gc           345453 miss
BenchmarkFreeCacheEvictZipfParallel/1000000-0010.0000-4                 10000000               469 ns/op           1281717 expc          7.00 gc          1347459 miss
BenchmarkFreeCacheEvictZipfParallel/1000000-0100.0000-4                 10000000               467 ns/op           2252014 expc          8.00 gc          2553736 miss
BenchmarkFreeCacheEvictZipfParallel/1000000-1000.0000-4                 10000000               503 ns/op           3012731 expc          7.00 gc          3412157 miss
BenchmarkBigCacheEvictZipfParallel/1000000-0000.9990-4                  10000000               272 ns/op                 0 expc          8.00 gc                0 miss
BenchmarkBigCacheEvictZipfParallel/1000000-0001.0000-4                  10000000               299 ns/op                 0 expc          8.00 gc                0 miss
BenchmarkBigCacheEvictZipfParallel/1000000-0001.0010-4                  10000000               273 ns/op               638 expc          8.00 gc              469 miss
BenchmarkBigCacheEvictZipfParallel/1000000-0001.0100-4                  10000000               273 ns/op              6446 expc          7.00 gc             4722 miss
BenchmarkBigCacheEvictZipfParallel/1000000-0001.1000-4                  10000000               312 ns/op             61491 expc          8.00 gc            45950 miss
BenchmarkBigCacheEvictZipfParallel/1000000-0002.0000-4                  10000000               314 ns/op            426544 expc          8.00 gc           344749 miss
BenchmarkBigCacheEvictZipfParallel/1000000-0010.0000-4                  10000000               422 ns/op           1282637 expc          7.00 gc          1485386 miss
BenchmarkBigCacheEvictZipfParallel/1000000-0100.0000-4                  10000000               510 ns/op           2251553 expc          7.00 gc          2726767 miss
BenchmarkBigCacheEvictZipfParallel/1000000-1000.0000-4                  10000000               568 ns/op           3015042 expc          6.00 gc          3608655 miss
BenchmarkSCacheEvictZipfParallel/1000000-0000.9990-4                    10000000               232 ns/op                 0 expc          2.00 gc                0 miss
BenchmarkSCacheEvictZipfParallel/1000000-0001.0000-4                    10000000               232 ns/op                 0 expc          1.00 gc                0 miss
BenchmarkSCacheEvictZipfParallel/1000000-0001.0010-4                    10000000               255 ns/op               683 expc          2.00 gc           637281 miss
BenchmarkSCacheEvictZipfParallel/1000000-0001.0100-4                    10000000               232 ns/op              6527 expc          2.00 gc           644363 miss
BenchmarkSCacheEvictZipfParallel/1000000-0001.1000-4                    10000000               243 ns/op             61330 expc          2.00 gc           705193 miss
BenchmarkSCacheEvictZipfParallel/1000000-0002.0000-4                    10000000               266 ns/op            426436 expc          2.00 gc          1133277 miss
BenchmarkSCacheEvictZipfParallel/1000000-0010.0000-4                    10000000               292 ns/op           1281681 expc          2.00 gc          2131414 miss
BenchmarkSCacheEvictZipfParallel/1000000-0100.0000-4                    10000000               317 ns/op           2252555 expc          3.00 gc          3196230 miss
BenchmarkSCacheEvictZipfParallel/1000000-1000.0000-4                    10000000               339 ns/op           3014547 expc          3.00 gc          3979280 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-0000.9990-4                10000000               966 ns/op                 0 expc          2.00 gc                0 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.0000-4                10000000               960 ns/op                 1.00 expc               2.00 gc         7.00 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.0010-4                10000000               991 ns/op               665 expc          2.00 gc             3216 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.0100-4                10000000              1027 ns/op              6459 expc          2.00 gc            22058 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.1000-4                10000000              1045 ns/op             61510 expc          2.00 gc           140479 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-0002.0000-4                10000000              1121 ns/op            428473 expc          3.00 gc           684840 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-0010.0000-4                10000000              1286 ns/op           1281995 expc          4.00 gc          1799002 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-0100.0000-4                10000000              1431 ns/op           2253585 expc          5.00 gc          2944195 miss
BenchmarkHashiCacheEvictZipfParallel/1000000-1000.0000-4                10000000              1517 ns/op           3011711 expc          4.00 gc          3769726 miss
PASS
ok      github.com/viant/go-cache-benchmark     232.748s
```

*Try extreme ranges for possible inputs for both Zipf and uniform distributions*

`SWEEP_DIST='[0.1,100,10000,100000000] go test -benchtime=10000000x -bench=Evict .`

## Using Docker

`docker run --rm -it -v "$PWD":/w -w /w golang:1.15 go test -bench=. -benchtime=10000000x .`

Note that you can use whatever version of Go (after 1.5 for best results). 
Also, some of the examples above use Docker.

## Options

Use the `-bench` options to filter benchmarks (e.g. `-bench=Zipf` to only run Zipf eviction tests).
Refer to [standard library documentation](https://pkg.go.dev/cmd/go/internal/test) for more `go test` options.

## Environment variables

*This section may change if there's a better way to control tests.*

* `TEST_SIZE_FACTOR` - defaults to `1`. Multiplies the number of elements stored.

* `MULTI_SIZES` - set to non-empty string to only benchmark the run with 100M, elements stored, then 10M, then lastly 100k elements (`TEST_SIZE_FACTOR` still applies).

* `PRECACHE_FACTOR` - defaults to `1`. Multiple of the cache size, determines how much of the cache is pre-populated before the eviction benchmarks are run.

### Zipf & distribution modifiers

* `MISS_PENALTY` - defaults to `0`. Sets milliseconds of wait in the case of a cache miss for benchmarks that test eviction.

See [standard library `rand`'s `Zipf` type](https://pkg.go.dev/math/rand#NewZipf) for more information about Zipf

* `ZIPF_FACTOR` - defaults to `2`. Multiplies the maximum of range of the input distribution, used to calculate `imax`. Only applies if `SWEEP_DIST` is not provided. Note that this also applies to the uniform distribution test.
* `ZIPF_S` - defaults to `1.01`. Sets curvature of Zipf probability (increases hit likelihood dramatically), set as `s`.
* `ZIPF_V` - defaults to `1`. Sets initial offset for Zipf probability, set as `v`.

* `SWEEP_DIST` - defaults to effectively `[ZIPF_FACTOR]`. Should be a JSON string containing a series of `float64` values to multiply the maximum input possible, in relation to the expected cache size (related to `TEST_SIZE_FACTOR`). For example, if `SWEEP_DIST='[2,5]'` (the single quotes are for escaping shell interpretation of `[]`), then the tests will use 2x and 5x for maximum input possible, relative to the cache size (which is 1M by default) - so this means that for the Zipf distribution, the maximum output value will be 2M for the 2x and 5M for the 5x. Increasing the number increases the likelihood of a cache miss.
    - Regarding the benchmark description for sweeps, the benchmark name is followed by 3 numbers - the first number is the cache size, the second number is the maximum distribution value multiplier, and the last number is the parallelization.

### `scache` modifiers

* `SCACHE_ENTRIES_BUFFER` - default to `1`. Multiplies the amount of entries provided to the `scache` configuration. `SCACHE_ENTRIES_DIV` is still applied, but this is a simpler interface for seeing how much additional memory `scache` requires to for achieving comparable hit rates with other caches.
* `SCACHE_ENTRIES_DIV` - defaults to `2`. Sets the number of entries that are used in `scache` configuration initialization, since [`scache` allocates twice the amount of memory than expected](https://github.com/viant/scache/blob/master/config.go#L33). Set to `1` to use twice the amount of memory than other caches. The resulting number of entries supported by the cache will be `expectedEntries / SCACHE_ENTRIES_DIV`. To get a specific "extended" size, divide 2 by the desired additional size. For example, to allocate 10% more memory for `scache`, use `SCACHE_ENTRIES_DIV` of `2 / 1.1` or `1.8182`.

### Golang provided environment variables

Some useful ones include:

* `GODEBUG` with value `gctrace=1` to have Go print metrics per garbage collect.
* `GOGC` with value `off` to turn off Go's garbage collection.
