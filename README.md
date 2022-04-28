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
go: downloading github.com/coocood/freecache v1.1.0
go: downloading github.com/hashicorp/golang-lru v0.5.4
go: downloading github.com/viant/scache v0.5.0
go: downloading github.com/allegro/bigcache/v2 v2.1.3
go: downloading github.com/cespare/xxhash v1.1.0
go: downloading github.com/pkg/errors v0.9.1
goos: linux
goarch: amd64
pkg: github.com/viant/go-cache-benchmark
BenchmarkFreeCacheSet/1000000-4                 10000000               655 ns/op
BenchmarkBigCacheSet/1000000-4                  10000000               818 ns/op
BenchmarkFreeCacheGet/1000000-4                 10000000               748 ns/op
BenchmarkBigCacheGet/1000000-4                  10000000               747 ns/op
BenchmarkFreeCacheSetParallel/1000000-4         10000000               458 ns/op
BenchmarkBigCacheSetParallel/1000000-4          10000000               369 ns/op
BenchmarkFreeCacheGetParallel/1000000-4         10000000               276 ns/op
BenchmarkBigCacheGetParallel/1000000-4          10000000               174 ns/op
BenchmarkSCacheGetParallel/1000000-4            10000000               154 ns/op
BenchmarkFreeCacheEvictZipfParallel/1000000-4           10000000               389 ns/op            102738 misses
BenchmarkBigCacheEvictZipfParallel/1000000-4            10000000               268 ns/op            106150 misses
BenchmarkSCacheEvictZipfParallel/1000000-4              10000000               196 ns/op            437107 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-4          10000000               778 ns/op            196808 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-4        10000000               401 ns/op            723942 misses
BenchmarkBigCacheEvictUniformParallel/1000000-4         10000000               344 ns/op            721281 misses
BenchmarkSCacheEvictUniformParallel/1000000-4           10000000               317 ns/op           1872326 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-4       10000000               987 ns/op           1254594 misses
PASS
ok      github.com/viant/go-cache-benchmark     97.257s
```

# Running tests

## Using native Go

Requires at least Go 1.15

Generally recommended to use `-benchtime=Xx` instead of `-benchtime=Xs`, especially to see effects on hit rate.

*Run all benchmarks*

`go test -benchmem -benchtime=10000000x -bench=. .`

*Run eviction strategy benchmarks*

`go test -benchmem -benchtime=10000000x -bench=Evict .`

*Extend range of possible inputs for Zipf*

`ZIPF_FACTOR=8 go test -benchmem -benchtime=10000000x -bench=Zipf .`

*Try different ranges for possible inputs for Zipf*

`SWEEP_DIST='[0.999,1,1.001,1.01,1.1,2,10,100,1000]' go test -benchmem -benchtime=10000000x -bench=Zipf`

```
goos: linux
goarch: amd64
pkg: github.com/viant/go-cache-benchmark
BenchmarkFreeCacheEvictZipfParallel/1000000-0000.999-4          10000000               377 ns/op                 0 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.000-4          10000000               365 ns/op                 1.00 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.001-4          10000000               396 ns/op               136 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.010-4          10000000               367 ns/op              1517 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0001.100-4          10000000               399 ns/op             14432 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0002.000-4          10000000               369 ns/op            102228 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0010.000-4          10000000               377 ns/op            319885 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-0100.000-4          10000000               368 ns/op            566235 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-1000.000-4          10000000               382 ns/op            762181 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0000.999-4           10000000               270 ns/op                 0 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0001.000-4           10000000               240 ns/op                 1.00 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0001.001-4           10000000               247 ns/op               204 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0001.010-4           10000000               267 ns/op              2014 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0001.100-4           10000000               245 ns/op             15643 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0002.000-4           10000000               262 ns/op            104307 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0010.000-4           10000000               277 ns/op            313208 misses
BenchmarkBigCacheEvictZipfParallel/1000000-0100.000-4           10000000               288 ns/op            583161 misses
BenchmarkBigCacheEvictZipfParallel/1000000-1000.000-4           10000000               297 ns/op            758926 misses
BenchmarkSCacheEvictZipfParallel/1000000-0000.999-4             10000000               210 ns/op                 0 misses
BenchmarkSCacheEvictZipfParallel/1000000-0001.000-4             10000000               209 ns/op                 0 misses
BenchmarkSCacheEvictZipfParallel/1000000-0001.001-4             10000000               190 ns/op            356031 misses
BenchmarkSCacheEvictZipfParallel/1000000-0001.010-4             10000000               186 ns/op            355829 misses
BenchmarkSCacheEvictZipfParallel/1000000-0001.100-4             10000000               194 ns/op            358652 misses
BenchmarkSCacheEvictZipfParallel/1000000-0002.000-4             10000000               216 ns/op            466240 misses
BenchmarkSCacheEvictZipfParallel/1000000-0010.000-4             10000000               219 ns/op            700887 misses
BenchmarkSCacheEvictZipfParallel/1000000-0100.000-4             10000000               233 ns/op            946551 misses
BenchmarkSCacheEvictZipfParallel/1000000-1000.000-4             10000000               249 ns/op           1175783 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0000.999-4         10000000               696 ns/op                 0 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.000-4         10000000               697 ns/op                 0 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.001-4         10000000               701 ns/op               952 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.010-4         10000000               726 ns/op              6467 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0001.100-4         10000000               737 ns/op             40577 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0002.000-4         10000000               761 ns/op            199027 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0010.000-4         10000000               813 ns/op            543037 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-0100.000-4         10000000               835 ns/op            779987 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-1000.000-4         10000000               862 ns/op            999072 misses
PASS
ok      github.com/viant/go-cache-benchmark     190.162s
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
BenchmarkFreeCacheEvictZipfParallel/1000000-000000000.1-4               10000000               402 ns/op                 0 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-000000100.0-4               10000000               398 ns/op            566003 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-000010000.0-4               10000000               369 ns/op            914335 misses
BenchmarkFreeCacheEvictZipfParallel/1000000-100000000.0-4               10000000               460 ns/op           1344099 misses
BenchmarkBigCacheEvictZipfParallel/1000000-000000000.1-4                10000000               229 ns/op                 0 misses
BenchmarkBigCacheEvictZipfParallel/1000000-000000100.0-4                10000000               325 ns/op            580607 misses
BenchmarkBigCacheEvictZipfParallel/1000000-000010000.0-4                10000000               354 ns/op            920043 misses
BenchmarkBigCacheEvictZipfParallel/1000000-100000000.0-4                10000000               355 ns/op           1373385 misses
BenchmarkSCacheEvictZipfParallel/1000000-000000000.1-4                  10000000               194 ns/op                 0 misses
BenchmarkSCacheEvictZipfParallel/1000000-000000100.0-4                  10000000               248 ns/op            983033 misses
BenchmarkSCacheEvictZipfParallel/1000000-000010000.0-4                  10000000               248 ns/op           1375111 misses
BenchmarkSCacheEvictZipfParallel/1000000-100000000.0-4                  10000000               274 ns/op           1722078 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-000000000.1-4              10000000               607 ns/op                 0 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-000000100.0-4              10000000               832 ns/op            749243 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-000010000.0-4              10000000               878 ns/op           1137320 misses
BenchmarkHashiCacheEvictZipfParallel/1000000-100000000.0-4              10000000               927 ns/op           1502707 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-000000000.1-4            10000000               365 ns/op                 0 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-000000100.0-4            10000000               412 ns/op           2459946 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-000010000.0-4            10000000               412 ns/op           2500175 misses
BenchmarkFreeCacheEvictUniformParallel/1000000-100000000.0-4            10000000               420 ns/op           2501747 misses
BenchmarkBigCacheEvictUniformParallel/1000000-000000000.1-4             10000000               269 ns/op                 0 misses
BenchmarkBigCacheEvictUniformParallel/1000000-000000100.0-4             10000000               418 ns/op           2463058 misses
BenchmarkBigCacheEvictUniformParallel/1000000-000010000.0-4             10000000               417 ns/op           2506048 misses
BenchmarkBigCacheEvictUniformParallel/1000000-100000000.0-4             10000000               415 ns/op           2505713 misses
BenchmarkSCacheEvictUniformParallel/1000000-000000000.1-4               10000000               235 ns/op                 0 misses
BenchmarkSCacheEvictUniformParallel/1000000-000000100.0-4               10000000               299 ns/op           2784388 misses
BenchmarkSCacheEvictUniformParallel/1000000-000010000.0-4               10000000               300 ns/op           2805622 misses
BenchmarkSCacheEvictUniformParallel/1000000-100000000.0-4               10000000               316 ns/op           2827530 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-000000000.1-4           10000000               797 ns/op                 0 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-000000100.0-4           10000000               897 ns/op           2495135 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-000010000.0-4           10000000               862 ns/op           2506445 misses
BenchmarkHashiCacheEvictUniformParallel/1000000-100000000.0-4           10000000               908 ns/op           2512219 misses
PASS
ok      github.com/viant/go-cache-benchmark     186.709s
```


## Using Docker

`docker run --rm -it -v "$PWD":/w -w /w golang:1.15 go test -bench=. -benchtime=10000000x .`

Note that you can use whatever version of Go (after 1.5 for best results).

## Options

Use the `-bench` options to filter benchmarks (e.g. `-bench=Zipf` to only run Zipf eviction tests).
Refer to [standard library documentation](https://pkg.go.dev/cmd/go/internal/test) for more `go test` options.

## Environment variables

*This section may change if there's a better way to control tests.*

* `TEST_SIZE_FACTOR` - defaults to `1`. Multiplies the number of elements stored.

* `MULTI_SIZES` - set to non-empty string to only benchmark the run with 10,000,000 elements stored, then 1,000,000, then lastly 100,000 elements (`TEST_SIZE_FACTOR` still applies).

See [standard library `rand`'s `Zipf` type](https://pkg.go.dev/math/rand#NewZipf)

* `ZIPF_FACTOR` - defaults to `2`. Multiplies the maximum of range of the Zipf distribution, used to calculate `imax`. Only applies if `SWEEP_DIST` is not provided.
* `ZIPF_S` - defaults to `1.01`. Sets curvature of Zipf probability (increases hit likelihood dramatically), set as `s`.
* `ZIPF_V` - defaults to `1`. Sets initial offset for Zipf probability, set as `v`.

* `SCACHE_ENTRIES_DIV` - defaults to `2`. Sets the number of entries that are used in `scache` configuration initialization, since [`scache` allocates twice the amount of memory than expected](https://github.com/viant/scache/blob/master/config.go#L33). Set to `1` to use twice the amount of memory than other caches. The resulting number of entries supported by the cache will be `expectedEntries / SCACHE_ENTRIES_DIV`. To get a specific "extended" size, divide 2 by the desired additional size. For example, to allocate 10% more memory for `scache`, use `SCACHE_ENTRIES_DIV` of `2 / 1.1` or `1.8182`.

* `MISS_PENALTY` - defaults to `0`. Sets milliseconds of wait in the case of a cache miss for benchmarks that test eviction.

### Golang provided environment variables

Some useful ones include

* `GODEBUG` with value `gctrace=1` to have Go print metrics per garbage collect.
* `GOGC` with value `off` to turn off Go's garbage collection.
