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
4. Non-evicting native `map` 
5. Non-evicting standard library `sync.Map`
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

$ docker run --rm -it -v "$PWD":/w -w /w -e SINGLE_LOAD=y golang:1.15 go test -bench=. -benchmem -benchtime=5s .
goos: linux
goarch: amd64
pkg: github.com/viant/go-cache-benchmark
BenchmarkMapSet/1000000-4               13307522               476 ns/op             143 B/op          2 allocs/op
BenchmarkFreeCacheSet/1000000-4          9099657               614 ns/op              59 B/op          1 allocs/op
BenchmarkBigCacheSet/1000000-4           8694972               841 ns/op             469 B/op          2 allocs/op
BenchmarkConcurrentMapSet/1000000-4      5665552               985 ns/op             226 B/op          6 allocs/op
BenchmarkMapGet/1000000-4               15071960               427 ns/op              23 B/op          1 allocs/op
BenchmarkFreeCacheGet/1000000-4          8106286               788 ns/op             135 B/op          2 allocs/op
BenchmarkBigCacheGet/1000000-4           7949700               821 ns/op             151 B/op          3 allocs/op
BenchmarkConcurrentMapGet/1000000-4     11390961               494 ns/op              23 B/op          1 allocs/op
BenchmarkFreeCacheSetParallel/1000000-4                 15078682               459 ns/op              73 B/op          2 allocs/op
BenchmarkBigCacheSetParallel/1000000-4                  18148712              2091 ns/op             494 B/op          2 allocs/op
BenchmarkConcurrentMapSetParallel/1000000-4              1788368              5427 ns/op             386 B/op          9 allocs/op
BenchmarkFreeCacheGetParallel/1000000-4                 18953143               309 ns/op             135 B/op          2 allocs/op
BenchmarkBigCacheGetParallel/1000000-4                  34834102               187 ns/op             151 B/op          3 allocs/op
BenchmarkSCacheGetParallel/1000000-4                    37698138               186 ns/op              23 B/op          1 allocs/op
BenchmarkConcurrentMapGetParallel/1000000-4             29622260               225 ns/op              23 B/op          1 allocs/op
BenchmarkFreeCacheZipfParallel/1000000-4                15800661               378 ns/op            156670 misses            154 B/op          2 allocs/op
BenchmarkBigCacheZipfParallel/1000000-4                 24453490               217 ns/op            232891 misses            160 B/op          3 allocs/op
BenchmarkSCacheZipfParallel/1000000-4                   36476641               165 ns/op           1198419 misses             35 B/op          1 allocs/op
BenchmarkHashiCacheZipfParallel/1000000-4                7619289               748 ns/op            152205 misses             41 B/op          2 allocs/op
PASS
ok      github.com/viant/go-cache-benchmark     251.860s

```

# Running tests

## Using native Go

Requires at least Go 1.15

`go test -bench=. -benchmen -benchtime=4s .`

## Using Docker

`docker run --rm -it -v "$PWD":/w -w /w golang:1.15 go test -bench=. -benchmem -benchtime=4s .`

Note that you can use whatever version of Go (after 1.5 for best results).

## Options

Use the `-bench` options to filter benchmarks (e.g. `-bench=Zipf` to only run Zipf eviction tests).
Refer to [standard library documentation](https://pkg.go.dev/cmd/go/internal/test) for more `go test` options.

## Environment variables

*This section may change if there's a better way to control tests.*

* `TEST_SIZE_FACTOR` - defaults to `1`. Multiplies the number of elements stored.
* `SINGLE_LOAD` - set to non-empty string to only benchmark the run with 1,000,000 elements (multiplier still applied).

See [standard library `rand`'s `Zipf` type](https://pkg.go.dev/math/rand#NewZipf)

* `ZIPF_FACTOR` - defaults to `2`. Multiplies the maximum of range of the Zipf distribution, used to calculate `imax`.
* `ZIPF_S` - defaults to `1.01`. Sets curvature of Zipf probability (increases hit likelihood dramatically), set as `s`.
* `ZIPF_V` - defaults to `1`. Sets initial offset for Zipf probability, set as `v`.

* `MISS_PENALTY` - defaults to `0`. Sets milliseconds of wait in the case of a cache miss for benchmarks that test eviction.

### Golang provided environment variables

Some useful ones include

* `GODEBUG` with value `gctrace=1` to have Go print metrics per garbage collect.
* `GOGC` with value `off` to turn off Go's garbage collection.
