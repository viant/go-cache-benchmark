# cache-bench

**This repository is currently a work-in-progress.**

Requires at least Go 1.15

Benchmarks for comparing cache Golang cache libraries.

Based off https://github.com/allegro/bigcache-bench.

The primary motivation of this repository is to create an even comparison of caching libraries.

Currently compares the following available libraries / implementations:

1. [FreeCache](https://github.com/coocood/freecache)
2. [BigCache](https://github.com/allegro/bigcache)
3. [SCache](https://github.com/viant/scache)
4. Non-evicting native `map` 
5. Non-evicting standard library `sync.Map`

# Descriptions of benchmarks

The benchmark runs with a default of 10,000,000 elements stored, then 1,000,000, then lastly 100,000 elements.

1. Benchmark setting of values without eviction, in serial and parallel.
2. Prepopulate and benchmark getting values without misses, in serial and parallel.
3. Prepoulate and benchmark eviction policy using requests following a Zipf distribution, setting on cache miss, in parallel only.

More will be added, including:
* Test GC churn caused by application behavior while cache system is running in memory
* Comparison with naive eviction algorithms
* Comparisons with other caching libraries

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

$ docker run --rm -it -v $HOME/go-workspace:/go -v "$PWD":/w -w /w -e TEST_SIZE_FACTOR=0.1 golang:1.15 go test -bench=.  -benchmem -benchtime=1s .
goos: linux
goarch: amd64
pkg: github.com/allegro/bigcache-bench
BenchmarkMapSet/1000000-4                2265510               466 ns/op             177 B/op          2 allocs/op
BenchmarkMapSet/100000-4                 2939761               420 ns/op             138 B/op          2 allocs/op
BenchmarkMapSet/10000-4                  3196062               348 ns/op             136 B/op          2 allocs/op
BenchmarkFreeCacheSet/1000000-4          1937762               605 ns/op             191 B/op          2 allocs/op
BenchmarkFreeCacheSet/100000-4           2023972               574 ns/op              45 B/op          1 allocs/op
BenchmarkFreeCacheSet/10000-4            2886192               423 ns/op              26 B/op          1 allocs/op
BenchmarkBigCacheSet/1000000-4           1603279               651 ns/op             204 B/op          2 allocs/op
BenchmarkBigCacheSet/100000-4            1623900               766 ns/op             516 B/op          2 allocs/op
BenchmarkBigCacheSet/10000-4             1920404               603 ns/op             367 B/op          1 allocs/op
BenchmarkConcurrentMapSet/1000000-4      1000000              1502 ns/op             347 B/op          8 allocs/op
BenchmarkConcurrentMapSet/100000-4       1366564               780 ns/op             207 B/op          6 allocs/op
BenchmarkConcurrentMapSet/10000-4        1997832               578 ns/op             200 B/op          5 allocs/op
BenchmarkMapGet/1000000-4                2739915               396 ns/op              23 B/op          1 allocs/op
BenchmarkMapGet/100000-4                 3503272               325 ns/op              23 B/op          1 allocs/op
BenchmarkMapGet/10000-4                  4931890               258 ns/op              23 B/op          1 allocs/op
BenchmarkFreeCacheGet/1000000-4          1666851               725 ns/op             135 B/op          2 allocs/op
BenchmarkFreeCacheGet/100000-4           1784475               661 ns/op             135 B/op          2 allocs/op
BenchmarkFreeCacheGet/10000-4            2167072               575 ns/op             135 B/op          2 allocs/op
BenchmarkBigCacheGet/1000000-4           1799558               675 ns/op             151 B/op          3 allocs/op
BenchmarkBigCacheGet/100000-4            1838730               698 ns/op             151 B/op          3 allocs/op
BenchmarkBigCacheGet/10000-4             2089288               593 ns/op             151 B/op          3 allocs/op
BenchmarkConcurrentMapGet/1000000-4      2295189               490 ns/op              23 B/op          1 allocs/op
BenchmarkConcurrentMapGet/100000-4       2979280               429 ns/op              23 B/op          1 allocs/op
BenchmarkConcurrentMapGet/10000-4        3969945               296 ns/op              23 B/op          1 allocs/op
BenchmarkFreeCacheSetParallel/1000000-4                  3627200               355 ns/op             155 B/op          3 allocs/op
BenchmarkFreeCacheSetParallel/100000-4                   3615232               340 ns/op              60 B/op          2 allocs/op
BenchmarkFreeCacheSetParallel/10000-4                    3792136               306 ns/op              48 B/op          2 allocs/op
BenchmarkBigCacheSetParallel/1000000-4                   3541106               362 ns/op             330 B/op          2 allocs/op
BenchmarkBigCacheSetParallel/100000-4                    2972002               472 ns/op             545 B/op          3 allocs/op
BenchmarkBigCacheSetParallel/10000-4                     3187263               361 ns/op             458 B/op          2 allocs/op
BenchmarkConcurrentMapSetParallel/1000000-4              1000000              1656 ns/op             369 B/op          8 allocs/op
BenchmarkConcurrentMapSetParallel/100000-4               1000000              1067 ns/op             265 B/op          7 allocs/op
BenchmarkConcurrentMapSetParallel/10000-4                1484668               756 ns/op             227 B/op          7 allocs/op
BenchmarkFreeCacheGetParallel/1000000-4                  4695729               284 ns/op             135 B/op          2 allocs/op
BenchmarkFreeCacheGetParallel/100000-4                   5024266               253 ns/op             135 B/op          2 allocs/op
BenchmarkFreeCacheGetParallel/10000-4                    4821295               251 ns/op             135 B/op          2 allocs/op
BenchmarkBigCacheGetParallel/1000000-4                   7000062               196 ns/op             151 B/op          3 allocs/op
BenchmarkBigCacheGetParallel/100000-4                    6524695               159 ns/op             151 B/op          3 allocs/op
BenchmarkBigCacheGetParallel/10000-4                     4979718               240 ns/op             151 B/op          3 allocs/op
BenchmarkSCacheGetParallel/1000000-4                     7034186               145 ns/op              23 B/op          1 allocs/op
BenchmarkSCacheGetParallel/100000-4                      8794845               119 ns/op              23 B/op          1 allocs/op
BenchmarkSCacheGetParallel/10000-4                       9553612               116 ns/op              23 B/op          1 allocs/op
BenchmarkConcurrentMapGetParallel/1000000-4              5244427               197 ns/op              23 B/op          1 allocs/op
BenchmarkConcurrentMapGetParallel/100000-4               8149916               198 ns/op              23 B/op          1 allocs/op
BenchmarkConcurrentMapGetParallel/10000-4                9975601               119 ns/op              23 B/op          1 allocs/op
BenchmarkFreeCacheZipfParallel/1000000-4                 3219261               321 ns/op             34462 misses            131 B/op          2 allocs/op
BenchmarkFreeCacheZipfParallel/100000-4                  4098817               305 ns/op             40798 misses            131 B/op          2 allocs/op
BenchmarkFreeCacheZipfParallel/10000-4                   3498136               343 ns/op             28719 misses            131 B/op          2 allocs/op
BenchmarkBigCacheZipfParallel/1000000-4                  6620007               226 ns/op             71553 misses            147 B/op          3 allocs/op
BenchmarkBigCacheZipfParallel/100000-4                   5801552               179 ns/op             52313 misses            147 B/op          3 allocs/op
BenchmarkBigCacheZipfParallel/10000-4                    5297527               236 ns/op              9999 misses            148 B/op          3 allocs/op
BenchmarkSCacheZipfParallel/1000000-4                    6855507               154 ns/op            329306 misses             25 B/op          1 allocs/op
BenchmarkSCacheZipfParallel/100000-4                     8994006               140 ns/op            310938 misses             22 B/op          1 allocs/op
BenchmarkSCacheZipfParallel/10000-4                      7559128               227 ns/op            931651 misses             33 B/op          1 allocs/op
PASS
ok      github.com/allegro/bigcache-bench       142.277s
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


See [standard library `rand`'s `Zipf` type](https://pkg.go.dev/math/rand#NewZipf)

* `ZIPF_FACTOR` - defaults to `2`. Multiplies the maximum of range of the Zipf distribution, used to calculate `imax`.
* `ZIPF_S` - defaults to `1.01`. Sets curvature of Zipf probability (increases hit likelihood dramatically), set as `s`.
* `ZIPF_V` - defaults to `1`. Sets initial offset for Zipf probability, set as `v`.


* `MISS_PENALTY` - defaults to `0`. Sets milliseconds of wait in the case of a cache miss for benchmarks that test eviction.

