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

$ docker run --rm -it -v "$PWD":/w -w /w golang:1.15 go test -bench=. -benchmem -benchtime=10000000x ./
go: downloading github.com/viant/scache v0.5.0
go: downloading github.com/coocood/freecache v1.1.0
go: downloading github.com/allegro/bigcache/v2 v2.1.3
go: downloading github.com/hashicorp/golang-lru v0.5.4
go: downloading github.com/cespare/xxhash v1.1.0
go: downloading github.com/pkg/errors v0.9.1
goos: linux
goarch: amd64
pkg: github.com/viant/go-cache-benchmark
BenchmarkFreeCacheSet/1000000-4                 10000000               639 ns/op              56 B/op          1 allocs/op
BenchmarkBigCacheSet/1000000-4                  10000000               818 ns/op              66 B/op          2 allocs/op
BenchmarkFreeCacheGet/1000000-4                 10000000               728 ns/op             135 B/op          2 allocs/op
BenchmarkBigCacheGet/1000000-4                  10000000               738 ns/op             151 B/op          3 allocs/op
BenchmarkFreeCacheSetParallel/1000000-4         10000000               458 ns/op              87 B/op          2 allocs/op
BenchmarkBigCacheSetParallel/1000000-4          10000000               376 ns/op              95 B/op          3 allocs/op
BenchmarkFreeCacheGetParallel/1000000-4         10000000               298 ns/op             135 B/op          2 allocs/op
BenchmarkBigCacheGetParallel/1000000-4          10000000               180 ns/op             151 B/op          3 allocs/op
BenchmarkSCacheGetParallel/1000000-4            10000000               155 ns/op              23 B/op          1 allocs/op
BenchmarkFreeCacheEvictZipfParallel/1000000-4           10000000               397 ns/op            102669 misses            166 B/op          2 allocs/op
BenchmarkBigCacheEvictZipfParallel/1000000-4            10000000               260 ns/op            104361 misses            178 B/op          3 allocs/op
BenchmarkSCacheEvictZipfParallel/1000000-4              10000000               197 ns/op            437547 misses             67 B/op          1 allocs/op
BenchmarkHashiCacheEvictZipfParallel/1000000-4          10000000               786 ns/op            201540 misses             41 B/op          2 allocs/op
PASS
ok      github.com/viant/go-cache-benchmark     72.104s
```

# Running tests

## Using native Go

Requires at least Go 1.15

Generally recommended to use `-benchtime=Xx` instead of `-benchtime=Xs`, especially to see effects on hit rate.

*Run all benchmarks*

`go test -bench=. -benchmem -benchtime=10000000x .`

*Run eviction strategy benchmarks*

`go test -bench=Evict -benchmem -benchtime=10000000x .`

*Extend range of possible inputs*

`ZIPF_FACTOR=8 go test -bench=Evict -benchmem -benchtime=10000000x .`

```
$ docker run --rm -it -v "$PWD":/w -w /w -e ZIPF_FACTOR=8 golang:1.15 go test -bench=Evict -benchmem -benchtime=10000000x ./
go: downloading github.com/viant/scache v0.5.0
go: downloading github.com/hashicorp/golang-lru v0.5.4
go: downloading github.com/coocood/freecache v1.1.0
go: downloading github.com/allegro/bigcache/v2 v2.1.3
go: downloading github.com/pkg/errors v0.9.1
go: downloading github.com/cespare/xxhash v1.1.0
goos: linux
goarch: amd64
pkg: github.com/viant/go-cache-benchmark
BenchmarkFreeCacheEvictZipfParallel/1000000-4           10000000               416 ns/op            291498 misses            168 B/op          2 allocs/op
BenchmarkBigCacheEvictZipfParallel/1000000-4            10000000               267 ns/op            296929 misses            176 B/op          3 allocs/op
BenchmarkSCacheEvictZipfParallel/1000000-4              10000000               204 ns/op            691164 misses             71 B/op          1 allocs/op
BenchmarkHashiCacheEvictZipfParallel/1000000-4          10000000               783 ns/op            450316 misses             47 B/op          2 allocs/op
PASS
ok      github.com/viant/go-cache-benchmark     21.911s
```


## Using Docker

`docker run --rm -it -v "$PWD":/w -w /w golang:1.15 go test -bench=. -benchmem -benchtime=10000000x .`

Note that you can use whatever version of Go (after 1.5 for best results).

## Options

Use the `-bench` options to filter benchmarks (e.g. `-bench=Zipf` to only run Zipf eviction tests).
Refer to [standard library documentation](https://pkg.go.dev/cmd/go/internal/test) for more `go test` options.

## Environment variables

*This section may change if there's a better way to control tests.*

* `TEST_SIZE_FACTOR` - defaults to `1`. Multiplies the number of elements stored.
* `MULTI_SIZES` - set to non-empty string to only benchmark the run with 10,000,000 elements stored, then 1,000,000, then lastly 100,000 elements (multiplier still applied).

See [standard library `rand`'s `Zipf` type](https://pkg.go.dev/math/rand#NewZipf)

* `ZIPF_FACTOR` - defaults to `2`. Multiplies the maximum of range of the Zipf distribution, used to calculate `imax`.
* `ZIPF_S` - defaults to `1.01`. Sets curvature of Zipf probability (increases hit likelihood dramatically), set as `s`.
* `ZIPF_V` - defaults to `1`. Sets initial offset for Zipf probability, set as `v`.

* `SCACHE_ENTRIES_DIV` - defaults to `2`. Sets the number of entries that are used in `scache` configuration initialization, since [`scache` allocates twice the amount of memory than expected](https://github.com/viant/scache/blob/master/config.go#L33). Set to `1` to use twice the amount of memory than other caches, the resulting number of entries supported by the cache will be `expectedEntries / SCACHE_ENTRIES_DIV`. To get a specific "extended" size, divide 2 by the desired additional size. For example, to allocate 10% more memory for `scache`, use `SCACHE_ENTRIES_DIV` of `2 / 1.1` or `1.8182`.

* `MISS_PENALTY` - defaults to `0`. Sets milliseconds of wait in the case of a cache miss for benchmarks that test eviction.

### Golang provided environment variables

Some useful ones include

* `GODEBUG` with value `gctrace=1` to have Go print metrics per garbage collect.
* `GOGC` with value `off` to turn off Go's garbage collection.
