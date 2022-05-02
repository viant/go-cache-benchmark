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

func BenchmarkSCacheSetParallel(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		cache := initSCache(testSize)
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

// Complex tests

// parallel Zipf + eviction

func BenchmarkFreeCacheEvictZipfParallel(b *testing.B) {
	testSweepZipf(b, func(b *testing.B, stp sweepTestParam) {
		fci := freecache.NewCache(stp.cacheSize * maxEntrySize)
		c := &eFreecache{fci}

		benchCacheEvict(b, benchEvictParam{
			sweepTestParam: stp,
			cache:          c,
		})
	})
}

func BenchmarkBigCacheEvictZipfParallel(b *testing.B) {
	testSweepZipf(b, func(b *testing.B, stp sweepTestParam) {
		cache := initBigCache(stp.cacheSize)
		c := &eBigCache{cache}
		benchCacheEvict(b, benchEvictParam{
			sweepTestParam: stp,
			cache:          c,
		})
	})
}

func BenchmarkSCacheEvictZipfParallel(b *testing.B) {
	testSweepZipf(b, func(b *testing.B, stp sweepTestParam) {
		cache := initSCache(stp.cacheSize)
		c := &eSCache{cache}
		benchCacheEvict(b, benchEvictParam{
			sweepTestParam: stp,
			cache:          c,
		})
	})
}

func BenchmarkHashiCacheEvictZipfParallel(b *testing.B) {
	testSweepZipf(b, func(b *testing.B, stp sweepTestParam) {
		cache, err := lru.New(stp.cacheSize)
		if err != nil {
			b.Errorf("%s", err)
		}

		c := &eHashiCache{cache}
		benchCacheEvict(b, benchEvictParam{
			sweepTestParam: stp,
			cache:          c,
		})
	})
}

// uniform (unrealistic) distribution

func BenchmarkFreeCacheEvictUniformParallel(b *testing.B) {
	testSweepUniform(b, func(b *testing.B, stp sweepTestParam) {
		fci := freecache.NewCache(stp.cacheSize * maxEntrySize)
		c := &eFreecache{fci}
		benchCacheEvict(b, benchEvictParam{
			sweepTestParam: stp,
			cache:          c,
		})
	})
}

func BenchmarkBigCacheEvictUniformParallel(b *testing.B) {
	testSweepUniform(b, func(b *testing.B, stp sweepTestParam) {
		cache := initBigCache(stp.cacheSize)
		c := &eBigCache{cache}
		benchCacheEvict(b, benchEvictParam{
			sweepTestParam: stp,
			cache:          c,
		})
	})
}

func BenchmarkSCacheEvictUniformParallel(b *testing.B) {
	testSweepUniform(b, func(b *testing.B, stp sweepTestParam) {
		cache := initSCache(stp.cacheSize)
		c := &eSCache{cache}
		benchCacheEvict(b, benchEvictParam{
			sweepTestParam: stp,
			cache:          c,
		})
	})
}

func BenchmarkHashiCacheEvictUniformParallel(b *testing.B) {
	testSweepUniform(b, func(b *testing.B, stp sweepTestParam) {
		cache, err := lru.New(stp.cacheSize)
		if err != nil {
			b.Errorf("%s", err)
		}

		c := &eHashiCache{cache}
		benchCacheEvict(b, benchEvictParam{
			sweepTestParam: stp,
			cache:          c,
		})
	})
}

// trivial test helpers

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

// Complex test helpers

type sweepTest func(b *testing.B, stp sweepTestParam)

type sweepTestParam struct {
	cacheSize     int
	precacheRatio float64
	missPenalty   time.Duration
	getDist       dmm
	writer        chan cacheResult
}

// distMakerMaker
type dmm func() distMaker

type distMaker interface {
	Uint64() uint64
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

type uniformDist struct {
	r   *rand.Rand
	max int64
}

// implements distMaker
func (u *uniformDist) Uint64() uint64 {
	return uint64(u.r.Int63n(u.max))
}

func testSweep(b *testing.B, fsg func(cs int, sf float64) distMaker, f sweepTest) {
	missPenalty := getMissPenalty()

	cacheSizes := getEnvCacheSizes()

	precacheRatio := getEnvFloat64("PRECACHE_RATIO", 1.0)

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

	// TODO use format resolver
	floatFormat := "%d-%0" + fmt.Sprintf("%d", wholes+fp+1) + "." + fmt.Sprintf("%d", fp) + "f"

	logDistPrefix := os.Getenv("LOG_DIST_PREFIX")
	for _, cacheSize := range cacheSizes {
		for _, distFactor := range distFactors {
			var benchName string
			if sweepDist {
				benchName = fmt.Sprintf(floatFormat, cacheSize, distFactor)
			} else {
				benchName = fmt.Sprintf("%d", cacheSize)
			}

			var writer chan cacheResult
			var terminator chan bool
			var logFile os.File
			if logDistPrefix != "" {
				logName := fmt.Sprintf("%s%s.%s.log", logDistPrefix, b.Name(), benchName)
				logFile, err := os.Create(logName)
				if err != nil {
					fmt.Println(err)
					b.Fail()
				}

				writer = make(chan cacheResult)
				terminator = make(chan bool)

				go func() {
					for {
						select {
						case cr := <-writer:
							logFile.Write([]byte(fmt.Sprintf("%d,%t\n", cr.v, cr.m)))
						case q := <-terminator:
							if q {
								close(writer)
								return
							}
						}
					}
				}()
			}

			b.Run(benchName, func(b *testing.B) {
				getDistMaker := func() distMaker {
					return fsg(cacheSize, distFactor)
				}

				f(b, sweepTestParam{
					cacheSize:     cacheSize,
					precacheRatio: precacheRatio,
					missPenalty:   missPenalty,
					getDist:       getDistMaker,
					writer:        writer,
				})
			})

			if terminator != nil {
				terminator <- true
				logFile.Close()
			}
		}
	}

}

type benchEvictParam struct {
	sweepTestParam

	cache Benchmarked
}

type cacheResult struct {
	v int
	m bool
}

func benchCacheEvict(b *testing.B, bep benchEvictParam) {
	testSize := bep.cacheSize
	cache := bep.cache
	precacheRatio := bep.precacheRatio

	precacheSize := int(math.Round(float64(testSize) * precacheRatio))
	for i := 0; i < precacheSize; i++ {
		cache.Set(i)
	}

	uts := uint64(precacheSize)

	missPenalty := bep.missPenalty
	runEvictParallel(b, func(pb *testing.PB, em *evictMeta) {
		dist := bep.getDist()

		var misses, exCache uint64
		for pb.Next() {
			uv := dist.Uint64()
			if uv >= uts {
				exCache++
			}

			iv := int(uv)

			missed := cache.Get(iv)

			if bep.writer != nil {
				bep.writer <- cacheResult{iv, missed}
			}

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

// metrics

type evictMeta struct {
	misses uint64
	// the distribution's random value is outside initial cache size
	exCache uint64
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

// generalized interface for cache

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

// util functions

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

func key(i int) string {
	// generates a 36 (4+32) byte key
	return fmt.Sprintf("key-%032d", i)
}

func parallelKey(threadID int, counter int) string {
	// generates a 36 (4+4+1+27) byte key with parallel support, used avoid collision
	return fmt.Sprintf("key-%04d-%027d", threadID, counter)
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
	entriesBuffer := getEnvFloat64("SCACHE_ENTRIES_BUFFER", 1)

	cache, _ := scache.New(&scache.Config{
		Shards:     defaultShards,
		MaxEntries: int(math.Round(float64(entries) * entriesBuffer / entriesDiv)),
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
