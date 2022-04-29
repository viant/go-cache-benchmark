// Copyright (c) 2022 Viant Inc.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/allegro/bigcache/v2"
	"github.com/coocood/freecache"
	lru "github.com/hashicorp/golang-lru"
	"github.com/viant/scache"
)

const maxEntrySize = 256
const defaultShards = 256

// Trivial tests

// serial Set

func BenchmarkFreeCacheSet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		cache := freecache.NewCache(testSize * maxEntrySize)
		for i := 0; i < b.N; i++ {
			cache.Set([]byte(key(i%testSize)), value(), 0)
		}
	})
}

func BenchmarkBigCacheSet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		cache := initBigCache(testSize)
		for i := 0; i < b.N; i++ {
			cache.Set(key(i%testSize), value())
		}
	})
}

func BenchmarkSCacheSet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		cache := initSCache(testSize)
		for i := 0; i < b.N; i++ {
			cache.Set(key(i%testSize), value())
		}
	})
}

// serial Get

func BenchmarkFreeCacheGet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()

		cache := freecache.NewCache(testSize * maxEntrySize)
		for i := 0; i < testSize; i++ {
			cache.Set([]byte(key(i)), value(), 0)
		}

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			cache.Get([]byte(key(i % testSize)))
		}
	})
}

func BenchmarkBigCacheGet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()
		cache := initBigCache(testSize)
		for i := 0; i < testSize; i++ {
			cache.Set(key(i), value())
		}

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(key(i % testSize))
		}
	})
}

func BenchmarkSCacheGet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()
		cache := initSCache(testSize)
		for i := 0; i < testSize; i++ {
			cache.Set(key(i), value())
		}

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(key(i % testSize))
		}
	})
}

// Parallel set

func BenchmarkFreeCacheSetParallel(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		cache := freecache.NewCache(testSize * maxEntrySize)
		rand.Seed(time.Now().Unix())

		b.RunParallel(func(pb *testing.PB) {
			id := rand.Intn(1000)
			counter := 0
			for pb.Next() {
				cache.Set([]byte(parallelKey(id, counter%testSize)), value(), 0)
				counter = counter + 1
			}
		})
	})
}

func BenchmarkBigCacheSetParallel(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		cache := initBigCache(testSize)
		rand.Seed(time.Now().Unix())

		b.RunParallel(func(pb *testing.PB) {
			id := rand.Intn(1000)
			counter := 0
			for pb.Next() {
				cache.Set(parallelKey(id, counter%testSize), value())
				counter = counter + 1
			}
		})
	})
}

// Parallel get

func BenchmarkFreeCacheGetParallel(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()
		cache := freecache.NewCache(testSize * maxEntrySize)
		for i := 0; i < testSize; i++ {
			cache.Set([]byte(key(i)), value(), 0)
		}

		b.StartTimer()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				cache.Get([]byte(key(counter % testSize)))
				counter = counter + 1
			}
		})
	})
}

func BenchmarkBigCacheGetParallel(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()
		cache := initBigCache(testSize)
		for i := 0; i < testSize; i++ {
			cache.Set(key(i), value())
		}

		b.StartTimer()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				cache.Get(key(counter % testSize))
				counter = counter + 1
			}
		})
	})
}

func BenchmarkSCacheGetParallel(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()
		cache := initSCache(testSize)
		for i := 0; i < testSize; i++ {
			cache.Set(key(i), value())
		}

		b.StartTimer()
		hitCount := 0
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				v, _ := cache.Get(key(counter % testSize))
				if v != nil {
					hitCount++
				}

				counter = counter + 1
			}
		})
	})
}

// refactor

type Benchmarked interface {
	Set(i int)
	Get(i int) bool
}

// freecache

type eFreecache struct {
	*freecache.Cache
}

func (c *eFreecache) Set(i int) {
	c.Cache.Set([]byte(key(i)), value(), 0)
}

func (c *eFreecache) Get(i int) bool {
	_, err := c.Cache.Get([]byte(key(i)))
	return err != nil
}

// bigcache

type eBigCache struct {
	*bigcache.BigCache
}

func (c *eBigCache) Set(i int) {
	c.BigCache.Set(key(i), value())
}

func (c *eBigCache) Get(i int) bool {
	_, err := c.BigCache.Get(key(i))
	return err != nil
}

// scache

type eSCache struct {
	*scache.Cache
}

func (c *eSCache) Set(i int) {
	c.Cache.Set(key(i), value())
}

func (c *eSCache) Get(i int) bool {
	_, err := c.Cache.Get(key(i))
	return err != nil
}

// lru (HashiCache)

type eHashiCache struct {
	*lru.Cache
}

func (c *eHashiCache) Set(i int) {
	c.Cache.Add(key(i), value())
}

func (c *eHashiCache) Get(i int) bool {
	_, ok := c.Cache.Get(key(i))
	return !ok
}

// parallel Zipf + eviction

func BenchmarkFreeCacheEvictZipfParallel(b *testing.B) {
	testSweepZipf(b, func(b *testing.B, s int, mp time.Duration, g func() distMaker) {
		fci := freecache.NewCache(s * maxEntrySize)
		c := eFreecache{fci}
		benchCacheEvict(b, s, mp, g, &c)
	})
}

/* this is to compare the interfaced version with the direct calls, difference seems to be noise
// n=10000000, cpu=4
// new: 401, 407, 364, 375, 371, 393, 373, 369, 357, 367
// old: 373, 371, 381, 364, 378, 372, 364, 380, 354, 388
func BenchmarkOldFreeCacheEvictZipfParallel(b *testing.B) {
	testSweepZipf(b, func(b *testing.B, testSize int, missPenalty time.Duration, getDist func() distMaker) {
		cache := freecache.NewCache(testSize * maxEntrySize)
		for i := 0; i < testSize; i++ {
			cache.Set([]byte(key(i)), value(), 0)
		}

		runEvictParallel(b, func(pb *testing.PB, em *evictMeta) {
			dist := getDist()

			var misses, inCache, exCache uint64
			for pb.Next() {
				uv := dist.Uint64()
				if uv > uint64(testSize) {
					exCache++
				} else {
					inCache++
				}

				k := []byte(key(int(uv)))
				v, _ := cache.Get(k)
				if v == nil {
					cache.Set(k, value(), 0)
					misses++

					if missPenalty > 0 {
						time.Sleep(missPenalty)
					}
				}
			}

			em.misses = misses
			em.inCache = inCache
			em.exCache = exCache
		})
	})
}
//*/

func BenchmarkBigCacheEvictZipfParallel(b *testing.B) {
	testSweepZipf(b, func(b *testing.B, testSize int, missPenalty time.Duration, getDist func() distMaker) {
		cache := initBigCache(testSize)
		c := eBigCache{cache}
		benchCacheEvict(b, testSize, missPenalty, getDist, &c)
	})
}

func BenchmarkSCacheEvictZipfParallel(b *testing.B) {
	testSweepZipf(b, func(b *testing.B, testSize int, missPenalty time.Duration, getDist func() distMaker) {
		cache := initSCache(testSize)
		c := eSCache{cache}
		benchCacheEvict(b, testSize, missPenalty, getDist, &c)
	})
}

func BenchmarkHashiCacheEvictZipfParallel(b *testing.B) {
	testSweepZipf(b, func(b *testing.B, testSize int, missPenalty time.Duration, getDist func() distMaker) {
		cache, err := lru.New(testSize)
		if err != nil {
			b.Errorf("%s", err)
		}

		c := eHashiCache{cache}
		benchCacheEvict(b, testSize, missPenalty, getDist, &c)
	})
}

// uniform (unrealistic) distribution

func BenchmarkFreeCacheEvictUniformParallel(b *testing.B) {
	testSweepUniform(b, func(b *testing.B, testSize int, missPenalty time.Duration, getDist func() distMaker) {
		fci := freecache.NewCache(testSize * maxEntrySize)
		c := eFreecache{fci}
		benchCacheEvict(b, testSize, missPenalty, getDist, &c)
	})
}

func BenchmarkBigCacheEvictUniformParallel(b *testing.B) {
	testSweepUniform(b, func(b *testing.B, testSize int, missPenalty time.Duration, getDist func() distMaker) {
		cache := initBigCache(testSize)
		c := eBigCache{cache}
		benchCacheEvict(b, testSize, missPenalty, getDist, &c)
	})
}

func BenchmarkSCacheEvictUniformParallel(b *testing.B) {
	testSweepUniform(b, func(b *testing.B, testSize int, missPenalty time.Duration, getDist func() distMaker) {
		cache := initSCache(testSize)
		c := eSCache{cache}
		benchCacheEvict(b, testSize, missPenalty, getDist, &c)
	})
}

func BenchmarkHashiCacheEvictUniformParallel(b *testing.B) {
	testSweepUniform(b, func(b *testing.B, testSize int, missPenalty time.Duration, getDist func() distMaker) {
		cache, err := lru.New(testSize)
		if err != nil {
			b.Errorf("%s", err)
		}

		c := eHashiCache{cache}
		benchCacheEvict(b, testSize, missPenalty, getDist, &c)
	})
}

// util functions

type evictMeta struct {
	misses uint64
	// the distribution value is outside initial cache size
	exCache uint64
}

func benchCacheEvict(b *testing.B, testSize int, missPenalty time.Duration, getDist func() distMaker, cache Benchmarked) {
	for i := 0; i < testSize; i++ {
		cache.Set(i)
	}

	uts := uint64(testSize)

	runEvictParallel(b, func(pb *testing.PB, em *evictMeta) {
		dist := getDist()

		var misses, exCache uint64
		for pb.Next() {
			uv := dist.Uint64()
			if uv >= uts {
				exCache++
			}

			iv := int(uv)
			missed := cache.Get(iv)
			if missed {
				cache.Set(iv)
				misses++

				if missPenalty > 0 {
					time.Sleep(missPenalty)
				}
			}
		}

		em.misses = misses
		em.exCache = exCache
	})
}

func runEvictParallel(b *testing.B, rpf func(pb *testing.PB, em *evictMeta)) {
	var startMem runtime.MemStats
	runtime.ReadMemStats(&startMem)

	var totalMeta evictMeta

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		var em evictMeta
		rpf(pb, &em)

		atomic.AddUint64(&totalMeta.misses, em.misses)
		atomic.AddUint64(&totalMeta.exCache, em.exCache)
	})
	b.StopTimer()

	var endMem runtime.MemStats
	runtime.ReadMemStats(&endMem)
	b.ReportMetric(float64(endMem.NumGC-startMem.NumGC), "gc")

	b.ReportMetric(float64(totalMeta.misses), "miss")
	b.ReportMetric(float64(totalMeta.exCache), "expc")
}

func getEnvInt64(varName string, defaultVal int64) int64 {
	vs := os.Getenv(varName)
	v, e := strconv.ParseInt(vs, 10, 64)
	if e != nil {
		v = defaultVal
	}
	return v
}

func getEnvFloat64(varName string, defaultVal float64) float64 {
	vs := os.Getenv(varName)
	v, err := strconv.ParseFloat(vs, 64)
	if err != nil {
		v = defaultVal
	}
	return v
}

func getEnvCacheSizes() []int {
	multFactor := getEnvFloat64("TEST_SIZE_FACTOR", 1.0)

	multiSize := os.Getenv("MULTI_SIZES")
	var testSizes []int
	if multiSize != "" {
		testSizes = []int{int(10000000 * multFactor), int(1000000 * multFactor), int(100000 * multFactor)}
	} else {
		testSizes = []int{int(1000000 * multFactor)}
	}

	return testSizes
}

func testWithSizes(b *testing.B, f func(b *testing.B, testSize int)) {
	testSizes := getEnvCacheSizes()
	for _, testSize := range testSizes {
		b.Run(fmt.Sprintf("%d", testSize), func(b *testing.B) {
			var startMem runtime.MemStats
			runtime.ReadMemStats(&startMem)

			f(b, testSize)

			var endMem runtime.MemStats
			runtime.ReadMemStats(&endMem)

			b.ReportMetric(float64(endMem.NumGC-startMem.NumGC), "gc")
		})
	}
}

type distMaker interface {
	Uint64() uint64
}

type distFactory func(cs int, sf float64) distMaker

type sweepTest func(b *testing.B, cs int, mp time.Duration, df func() distMaker)

func testSweep(b *testing.B, fsg distFactory, f sweepTest) {
	missPenalty := getMissPenalty()

	cacheSizes := getEnvCacheSizes()

	sweepDistStr := os.Getenv("SWEEP_DIST")
	sweepDist := sweepDistStr != ""
	var distFactors []float64
	if sweepDist {
		err := json.Unmarshal([]byte(sweepDistStr), &distFactors)
		if err != nil {
			distFactors = []float64{0.99, 1.0, 1.01, 1.05, 1.1, 1.5, 2.0}
		}
	} else {
		zipfFactor := getEnvFloat64("ZIPF_FACTOR", 2.0)
		distFactors = []float64{zipfFactor}
	}

	var wholes, fp int
	for _, distFactor := range distFactors {
		rounded := math.Round(distFactor)
		wholeDigits := len(fmt.Sprintf("%0.0f", rounded))
		if wholeDigits > wholes {
			wholes = wholeDigits
		}

		maxPrinted := fmt.Sprintf("%0.15f", distFactor-rounded)
		for i := len(maxPrinted) - 1; i >= 0; i-- {
			if maxPrinted[i] != '0' {
				if fp < i-1 {
					fp = i - 1
				}

				break
			}
		}
	}

	floatFormat := "%d-%0" + fmt.Sprintf("%d", wholes+fp+1) + "." + fmt.Sprintf("%d", fp) + "f"

	for _, cacheSize := range cacheSizes {
		for _, distFactor := range distFactors {
			var benchName string
			if sweepDist {
				benchName = fmt.Sprintf(floatFormat, cacheSize, distFactor)
			} else {
				benchName = fmt.Sprintf("%d", cacheSize)
			}

			b.Run(benchName, func(b *testing.B) {
				getDistMaker := func() distMaker {
					return fsg(cacheSize, distFactor)
				}

				f(b, cacheSize, missPenalty, getDistMaker)
			})
		}
	}
}

type uniformDist struct {
	r   *rand.Rand
	max int64
}

func (u *uniformDist) Uint64() uint64 {
	return uint64(u.r.Int63n(u.max))
}

func testSweepUniform(b *testing.B, sweepTester sweepTest) {
	b.StopTimer()

	g := func(testSize int, f float64) distMaker {
		src := rand.NewSource(time.Now().UnixNano())
		randObj := rand.New(src)
		return &uniformDist{
			r:   randObj,
			max: int64(math.Round(float64(testSize) * f)),
		}
	}

	testSweep(b, g, sweepTester)

	runtime.GC()
}

func testSweepZipf(b *testing.B, sweepTester sweepTest) {
	b.StopTimer()

	zipfS := getEnvFloat64("ZIPF_S", 1.01)
	zipfV := getEnvFloat64("ZIPF_V", 1.0)

	zipfGen := func(testSize int, zipfFactor float64) distMaker {
		src := rand.NewSource(time.Now().UnixNano())
		randObj := rand.New(src)

		return rand.NewZipf(randObj, zipfS, zipfV, uint64(math.Round(float64(testSize)*zipfFactor)))
	}

	testSweep(b, zipfGen, sweepTester)

	runtime.GC()
}

func key(i int) string {
	// generates a 36 (4+32) byte key
	return fmt.Sprintf("key-%032d", i)
}

func parallelKey(threadID int, counter int) string {
	// generates a 17 (4+4+1+8) byte key with parallel support, used avoid collision
	return fmt.Sprintf("key-%04d-%08d", threadID, counter)
}

func value() []byte {
	return make([]byte, 100)
}

func getMissPenalty() time.Duration {
	v := getEnvInt64("MISS_PENALTY", 0)
	return time.Duration(v) * time.Millisecond
}

// cache helpers

func initSCache(entries int) *scache.Cache {
	// since SCache allocates 2x buffer, divide max entries by 2
	// https://github.com/viant/scache/blob/master/config.go#L33
	entriesDiv := getEnvFloat64("SCACHE_ENTRIES_DIV", 2)
	cache, _ := scache.New(&scache.Config{
		Shards:     defaultShards,
		MaxEntries: int(math.Round(float64(entries) / entriesDiv)),
		EntrySize:  maxEntrySize,
	})

	return cache
}

func initBigCache(entriesInWindow int) *bigcache.BigCache {
	cache, _ := bigcache.NewBigCache(bigcache.Config{
		Shards:             defaultShards,
		LifeWindow:         10 * time.Minute,
		MaxEntriesInWindow: entriesInWindow,
		MaxEntrySize:       maxEntrySize,
		HardMaxCacheSize:   maxEntrySize * entriesInWindow / 1024 / 1024,
		Verbose:            false,
	})

	return cache
}
