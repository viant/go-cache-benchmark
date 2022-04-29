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

The benchmark runs with a default of 10,000,000 elements stored, then 1,000,000, then lastly 100,000 elements.

1. Benchmark setting of values without eviction, in serial and parallel.
2. Prepopulate and benchmark getting values without misses, in serial and parallel.
3. Prepoulate and benchmark eviction policy using requests following a Zipf distribution, setting on cache miss, in parallel only.

More will be added, including:
* Test GC churn caused by application behavior while cache system is running in memory
* Comparison with naive eviction algorithms
* Comparisons with other caching libraries

# Observations

1. `SCache` is the only cache that does not allocate memory on read. This can be checked by using the `-cpuprofile` option and viewing the CPU trace resulting in no pathways that lead to an `alloc` (and subsequently a GC).
2. `Scache` seems to be generally the fastest performing cache. Standard operation seems to occur in 68% of `BigCache` and 45% of `FreeCache`.
3. `SCache` tends to have the highest miss / lowest hit ratio with the Zipf distribution test. The other caches seem to get about 1-2% miss rate whereas `SCache` seems to get about 5% miss rate. This could be problematic if cache misses are dramatically more expensive than cache hits.

[Further observations.](./further-observations.md)

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
go: downloading github.com/allegro/bigcache/v2 v2.1.3
go: downloading github.com/coocood/freecache v1.1.0
go: downloading github.com/pkg/errors v0.9.1
go: downloading github.com/cespare/xxhash v1.1.0
goos: linux
goarch: amd64
pkg: github.com/viant/go-cache-benchmark
BenchmarkFreeCacheSet/1000000-4                 10000000               639 ns/op                 7.00 gc
BenchmarkBigCacheSet/1000000-4                  10000000               821 ns/op                 7.00 gc
BenchmarkSCacheSet/1000000-4                    10000000               454 ns/op                 7.00 gc
BenchmarkFreeCacheGet/1000000-4                 10000000               756 ns/op                10.0 gc
BenchmarkBigCacheGet/1000000-4                  10000000               664 ns/op                12.0 gc
BenchmarkSCacheGet/1000000-4                    10000000               555 ns/op                 5.00 gc
BenchmarkFreeCacheSetParallel/1000000-4         10000000               438 ns/op                 8.00 gc
BenchmarkBigCacheSetParallel/1000000-4          10000000               388 ns/op                 8.00 gc
BenchmarkFreeCacheGetParallel/1000000-4         10000000               291 ns/op                12.0 gc
BenchmarkBigCacheGetParallel/1000000-4          10000000               181 ns/op                12.0 gc
BenchmarkSCacheGetParallel/1000000-4            10000000               174 ns/op                 6.00 gc
BenchmarkFreeCacheEvictZipfParallel/1000000-4           10000000               388 ns/op                 5.00 gc            102417 misses
BenchmarkBigCacheEvictZipfParallel/1000000-4            10000000               253 ns/op                 5.00 gc            102761 misses
BenchmarkSCacheEvictZipfParallel/1000000-4              10000000               218 ns/op                 1.00 gc            434173 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-4          10000000               892 ns/op                 1.00 gc            199071 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-4        10000000               371 ns/op                 5.00 gc            719011 misses
BenchmarkBigCacheEvictUniformParallel/1000000-4         10000000               346 ns/op                 5.00 gc            721349 misses
BenchmarkSCacheEvictUniformParallel/1000000-4           10000000               320 ns/op                 1.00 gc           1880019 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-4       10000000               998 ns/op                 3.00 gc           1256406 misses
PASS
ok      github.com/viant/go-cache-benchmark     106.775s
```

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
BenchmarkFreeCacheEvictZipfParallel/1000000-0000.9990-4                 10000000               400 ns/op                 5.00 gc                 0 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.0000-4                 10000000               383 ns/op                 4.00 gc                 0 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.0010-4                 10000000               396 ns/op                 4.00 gc               160 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.0100-4                 10000000               375 ns/op                 5.00 gc              1511 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.1000-4                 10000000               432 ns/op                 5.00 gc             14231 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0002.0000-4                 10000000               372 ns/op                 5.00 gc            102576 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0010.0000-4                 10000000               373 ns/op                 5.00 gc            312239 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0100.0000-4                 10000000               369 ns/op                 5.00 gc            559361 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-1000.0000-4                 10000000               371 ns/op                 5.00 gc            757529 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0000.9990-4                  10000000               256 ns/op                 5.00 gc                 0 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0001.0000-4                  10000000               242 ns/op                 6.00 gc                 2.00 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0001.0010-4                  10000000               252 ns/op                 5.00 gc               193 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0001.0100-4                  10000000               259 ns/op                 6.00 gc              1634 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0001.1000-4                  10000000               242 ns/op                 6.00 gc             15454 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0002.0000-4                  10000000               246 ns/op                 5.00 gc            103738 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0010.0000-4                  10000000               295 ns/op                 5.00 gc            318739 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0100.0000-4                  10000000               296 ns/op                 6.00 gc            591390 misses
BenchmarkBigCacheEvictZipfParallel/1000000-1000.0000-4                  10000000               305 ns/op                 5.00 gc            790669 misses
BenchmarkSCacheEvictZipfParallel/1000000-0000.9990-4                    10000000               216 ns/op                 1.00 gc                 0 misses
BenchmarkSCacheEvictZipfParallel/1000000-0001.0000-4                    10000000               197 ns/op                 0 gc            0 misses
BenchmarkSCacheEvictZipfParallel/1000000-0001.0010-4                    10000000               190 ns/op                 1.00 gc            346059 misses
BenchmarkSCacheEvictZipfParallel/1000000-0001.0100-4                    10000000               185 ns/op                 1.00 gc            348027 misses
BenchmarkSCacheEvictZipfParallel/1000000-0001.1000-4                    10000000               199 ns/op                 1.00 gc            348070 misses
BenchmarkSCacheEvictZipfParallel/1000000-0002.0000-4                    10000000               194 ns/op                 0 gc       462426 misses
BenchmarkSCacheEvictZipfParallel/1000000-0010.0000-4                    10000000               220 ns/op                 1.00 gc            720425 misses
BenchmarkSCacheEvictZipfParallel/1000000-0100.0000-4                    10000000               225 ns/op                 1.00 gc            989807 misses
BenchmarkSCacheEvictZipfParallel/1000000-1000.0000-4                    10000000               250 ns/op                 1.00 gc           1209708 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0000.9990-4                10000000               850 ns/op                 2.00 gc                 0 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.0000-4                10000000               829 ns/op                 1.00 gc                 0 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.0010-4                10000000               842 ns/op                 1.00 gc              1006 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.0100-4                10000000               843 ns/op                 2.00 gc              6172 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.1000-4                10000000               860 ns/op                 1.00 gc             41971 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0002.0000-4                10000000               885 ns/op                 1.00 gc            198142 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0010.0000-4                10000000               907 ns/op                 1.00 gc            481880 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0100.0000-4                10000000               968 ns/op                 1.00 gc            762560 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-1000.0000-4                10000000               977 ns/op                 2.00 gc            958817 misses
PASS
ok      github.com/viant/go-cache-benchmark     187.244s
```

*Try extreme ranges for possible inputs for both Zipf and uniform distributions*

`SWEEP_DIST='[0.1,100,10000,100000000] go test -benchtime=10000000x -bench=Evict .`

```
$ docker run --rm -it -v "$PWD":/w -w /w -e SWEEP_DIST='[0.1,100,10000,100000000]' golang:1.15 go test -bench=Evict -benchtime=10000000x .
go: downloading github.com/viant/scache v0.5.0
go: downloading github.com/hashicorp/golang-lru v0.5.4
go: downloading github.com/allegro/bigcache/v2 v2.1.3
go: downloading github.com/coocood/freecache v1.1.0
go: downloading github.com/pkg/errors v0.9.1
go: downloading github.com/cespare/xxhash v1.1.0
goos: linux
goarch: amd64
pkg: github.com/viant/go-cache-benchmark
BenchmarkFreeCacheEvictZipfParallel/1000000-000000000.1-4               10000000               437 ns/op                 4.00 gc                 0 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-000000100.0-4               10000000               383 ns/op                 5.00 gc            561977 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-000010000.0-4               10000000               391 ns/op                 4.00 gc            931059 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-100000000.0-4               10000000               386 ns/op                 4.00 gc           1342408 misses
BenchmarkBigCacheEvictZipfParallel/1000000-000000000.1-4                10000000               249 ns/op                 6.00 gc                 0 misses
BenchmarkBigCacheEvictZipfParallel/1000000-000000100.0-4                10000000               316 ns/op                 5.00 gc            576990 misses
BenchmarkBigCacheEvictZipfParallel/1000000-000010000.0-4                10000000               339 ns/op                 5.00 gc            950480 misses
BenchmarkBigCacheEvictZipfParallel/1000000-100000000.0-4                10000000               382 ns/op                 5.00 gc           1370620 misses
BenchmarkSCacheEvictZipfParallel/1000000-000000000.1-4                  10000000               185 ns/op                 0 gc            0 misses
BenchmarkSCacheEvictZipfParallel/1000000-000000100.0-4                  10000000               224 ns/op                 1.00 gc            984134 misses
BenchmarkSCacheEvictZipfParallel/1000000-000010000.0-4                  10000000               262 ns/op                 1.00 gc           1345075 misses
BenchmarkSCacheEvictZipfParallel/1000000-100000000.0-4                  10000000               285 ns/op                 1.00 gc           1714319 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-000000000.1-4              10000000               782 ns/op                 1.00 gc                 0 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-000000100.0-4              10000000               979 ns/op                 2.00 gc            767605 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-000010000.0-4              10000000              1039 ns/op                 2.00 gc           1166627 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-100000000.0-4              10000000              1066 ns/op                 2.00 gc           1499113 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-000000000.1-4            10000000               378 ns/op                 5.00 gc                 0 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-000000100.0-4            10000000               408 ns/op                 4.00 gc           2460063 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-000010000.0-4            10000000               410 ns/op                 4.00 gc           2501610 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-100000000.0-4            10000000               410 ns/op                 5.00 gc           2500814 misses
BenchmarkBigCacheEvictUniformParallel/1000000-000000000.1-4             10000000               259 ns/op                 5.00 gc                 0 misses
BenchmarkBigCacheEvictUniformParallel/1000000-000000100.0-4             10000000               414 ns/op                 4.00 gc           2463603 misses
BenchmarkBigCacheEvictUniformParallel/1000000-000010000.0-4             10000000               412 ns/op                 4.00 gc           2505315 misses
BenchmarkBigCacheEvictUniformParallel/1000000-100000000.0-4             10000000               422 ns/op                 6.00 gc           2505043 misses
BenchmarkSCacheEvictUniformParallel/1000000-000000000.1-4               10000000               218 ns/op                 0 gc            0 misses
BenchmarkSCacheEvictUniformParallel/1000000-000000100.0-4               10000000               305 ns/op                 1.00 gc           2785412 misses
BenchmarkSCacheEvictUniformParallel/1000000-000010000.0-4               10000000               318 ns/op                 1.00 gc           2809346 misses
BenchmarkSCacheEvictUniformParallel/1000000-100000000.0-4               10000000               305 ns/op                 2.00 gc           2795954 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-000000000.1-4           10000000               800 ns/op                 2.00 gc                 0 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-000000100.0-4           10000000               880 ns/op                 3.00 gc           2481520 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-000010000.0-4           10000000               882 ns/op                 4.00 gc           2506585 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-100000000.0-4           10000000               945 ns/op                 3.00 gc           2501872 misses
PASS
ok      github.com/viant/go-cache-benchmark     187.620s
```


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

* `MULTI_SIZES` - set to non-empty string to only benchmark the run with 10,000,000 elements stored, then 1,000,000, then lastly 100,000 elements (`TEST_SIZE_FACTOR` still applies).

See [standard library `rand`'s `Zipf` type](https://pkg.go.dev/math/rand#NewZipf)

* `ZIPF_FACTOR` - defaults to `2`. Multiplies the maximum of range of the input distribution, used to calculate `imax`. Only applies if `SWEEP_DIST` is not provided. Note that this also applies to the uniform distribution test.
* `ZIPF_S` - defaults to `1.01`. Sets curvature of Zipf probability (increases hit likelihood dramatically), set as `s`.
* `ZIPF_V` - defaults to `1`. Sets initial offset for Zipf probability, set as `v`.

* `SWEEP_DIST` - defaults to effectively `[ZIPF_FACTOR]`. Should be a JSON string containing a series of `float64` values to multiply the maximum input possible, in relation to the expected cache size (related to `TEST_SIZE_FACTOR`). For example, if `SWEEP_DIST='[2,5]'` (the single quotes are for escaping shell interpretation of `[]`), then the tests will use 2x and 5x for maximum input possible, relative to the cache size, which is 1,000,000 by default - so this means that for the Zipf distribution, the maximum output value will be 2,000,000 for the 2x and 5,000,00 for the 5x. Increasing the number increases the likelihood of a cache miss.

* `SCACHE_ENTRIES_DIV` - defaults to `2`. Sets the number of entries that are used in `scache` configuration initialization, since [`scache` allocates twice the amount of memory than expected](https://github.com/viant/scache/blob/master/config.go#L33). Set to `1` to use twice the amount of memory than other caches. The resulting number of entries supported by the cache will be `expectedEntries / SCACHE_ENTRIES_DIV`. To get a specific "extended" size, divide 2 by the desired additional size. For example, to allocate 10% more memory for `scache`, use `SCACHE_ENTRIES_DIV` of `2 / 1.1` or `1.8182`.

* `MISS_PENALTY` - defaults to `0`. Sets milliseconds of wait in the case of a cache miss for benchmarks that test eviction.

### Golang provided environment variables

Some useful ones include

* `GODEBUG` with value `gctrace=1` to have Go print metrics per garbage collect.
* `GOGC` with value `off` to turn off Go's garbage collection.
