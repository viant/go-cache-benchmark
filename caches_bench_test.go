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
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
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

func BenchmarkMapSet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		m := makeMap(testSize)
		for i := 0; i < b.N; i++ {
			m[key(i%testSize)] = value()
		}
	})
}

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

func BenchmarkConcurrentMapSet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		var m sync.Map
		for i := 0; i < b.N; i++ {
			m.Store(key(i%testSize), value())
		}
	})
}

// serial Get

func BenchmarkMapGet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()

		m := makeMap(testSize)
		for i := 0; i < testSize; i++ {
			m[key(i)] = value()
		}

		var ignored int = 0
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			if m[key(i%testSize)] != nil {
				ignored++
			}
		}
	})
}

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
		cache := initBigCache(b.N)
		for i := 0; i < testSize; i++ {
			cache.Set(key(i), value())
		}

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(key(i % testSize))
		}
	})
}

func BenchmarkConcurrentMapGet(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()

		var m sync.Map
		for i := 0; i < testSize; i++ {
			m.Store(key(i), value())
		}

		ignored := 0
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			_, ok := m.Load(key(i % testSize))
			if ok {
				ignored++
			}
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

func BenchmarkConcurrentMapSetParallel(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		var m sync.Map

		b.RunParallel(func(pb *testing.PB) {
			id := rand.Intn(1000)
			counter := 0
			for pb.Next() {
				m.Store(parallelKey(id, counter%testSize), value())
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

func BenchmarkConcurrentMapGetParallel(b *testing.B) {
	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()

		var m sync.Map
		for i := 0; i < testSize; i++ {
			m.Store(key(i), value())
		}

		b.StartTimer()
		hitCount := 0
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				_, ok := m.Load(key(counter % testSize))
				if ok {
					hitCount++
				}

				counter = counter + 1
			}
		})
	})
}

// parallel Zipf + eviction

func BenchmarkFreeCacheZipfParallel(b *testing.B) {
	b.StopTimer()
	missPenalty := getMissPenalty()
	testWithSizes(b, func(b *testing.B, testSize int) {
		cache := freecache.NewCache(testSize * maxEntrySize)
		for i := 0; i < testSize; i++ {
			cache.Set([]byte(key(i)), value(), 0)
		}

		var misses uint64
		b.StartTimer()
		b.RunParallel(func(pb *testing.PB) {
			zipf := getZipf(testSize)

			var missed uint64
			for pb.Next() {
				k := []byte(key(int(zipf.Uint64())))
				v, _ := cache.Get(k)
				if v == nil {
					cache.Set(k, value(), 0)
					missed = missed + 1

					if missPenalty > 0 {
						time.Sleep(missPenalty)
					}
				}
			}

			atomic.AddUint64(&misses, missed)
		})

		b.StopTimer()
		b.ReportMetric(float64(misses), "misses")
	})

	runtime.GC()
}

func BenchmarkBigCacheZipfParallel(b *testing.B) {
	b.StopTimer()
	missPenalty := getMissPenalty()
	testWithSizes(b, func(b *testing.B, testSize int) {
		cache := initBigCache(testSize)
		for i := 0; i < testSize; i++ {
			cache.Set(key(i), value())
		}

		var misses uint64
		b.StartTimer()
		b.RunParallel(func(pb *testing.PB) {
			zipf := getZipf(testSize)

			var missed uint64
			for pb.Next() {
				k := key(int(zipf.Uint64()))
				_, e := cache.Get(k)
				if e != nil {
					cache.Set(k, value())
					missed++

					if missPenalty > 0 {
						time.Sleep(missPenalty)
					}
				}
			}

			atomic.AddUint64(&misses, missed)
		})

		b.StopTimer()
		b.ReportMetric(float64(misses), "misses")
	})

	runtime.GC()
}

func BenchmarkSCacheZipfParallel(b *testing.B) {
	b.StopTimer()
	missPenalty := getMissPenalty()

	testWithSizes(b, func(b *testing.B, testSize int) {
		cache := initSCache(testSize)
		for i := 0; i < testSize; i++ {
			cache.Set(key(i), value())
		}

		var totalMisses uint64
		b.StartTimer()
		b.RunParallel(func(pb *testing.PB) {
			zipf := getZipf(testSize)

			var misses uint64
			for pb.Next() {
				k := key(int(zipf.Uint64()))
				_, e := cache.Get(k)
				if e != nil {
					cache.Set(k, value())
					misses = misses + 1
					if missPenalty > 0 {
						time.Sleep(missPenalty)
					}
				}
			}

			atomic.AddUint64(&totalMisses, misses)
		})

		b.StopTimer()
		b.ReportMetric(float64(totalMisses), "misses")
	})

	runtime.GC()
}

func BenchmarkHashiCacheZipfParallel(b *testing.B) {
	b.StopTimer()
	missPenalty := getMissPenalty()

	testWithSizes(b, func(b *testing.B, testSize int) {
		b.StopTimer()
		cache, err := lru.New(testSize)
		if err != nil {
			b.Errorf("%s", err)
		}

		for i := 0; i < testSize; i++ {
			cache.Add(key(i), value())
		}

		var totalMisses uint64
		b.StartTimer()
		b.RunParallel(func(pb *testing.PB) {
			zipf := getZipf(testSize)

			var misses uint64
			for pb.Next() {
				k := key(int(zipf.Uint64()))
				v, _ := cache.Get(k)
				if v == nil {
					cache.Add(k, value())
					misses = misses + 1
					if missPenalty > 0 {
						time.Sleep(missPenalty)
					}
				}
			}

			atomic.AddUint64(&totalMisses, misses)
		})

		b.StopTimer()
		b.ReportMetric(float64(totalMisses), "misses")
	})

	runtime.GC()
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

func testWithSizes(b *testing.B, f func(b *testing.B, testSize int)) {
	multFactor := getEnvFloat64("TEST_SIZE_FACTOR", 1.0)

	singleOnly := os.Getenv("SINGLE_LOAD")
	var testSizes []int
	if singleOnly == "" {
		testSizes = []int{int(10000000 * multFactor), int(1000000 * multFactor), int(100000 * multFactor)}
	} else {
		testSizes = []int{int(1000000 * multFactor)}
	}

	for _, testSize := range testSizes {
		b.Run(fmt.Sprintf("%d", testSize), func(b *testing.B) {
			f(b, testSize)
		})
	}
}

func key(i int) string {
	// generates a 16 byte key
	return fmt.Sprintf("key-%012d", i)
}

func parallelKey(threadID int, counter int) string {
	// generates a 17 byte key with parallel support, used avoid collision
	return fmt.Sprintf("key-%04d-%08d", threadID, counter)
}

func value() []byte {
	// allocates emptye byte space
	return make([]byte, 100)
}

func getMissPenalty() time.Duration {
	v := getEnvInt64("MISS_PENALTY", 0)
	return time.Duration(v) * time.Millisecond
}

// rand helpers

func getZipf(testSize int) *rand.Zipf {
	src := rand.NewSource(time.Now().Unix())
	randObj := rand.New(src)

	zipfS := getEnvFloat64("ZIPF_S", 1.01)
	zipfV := getEnvFloat64("ZIPF_V", 1.0)
	zipfFactor := getEnvFloat64("ZIPF_FACTOR", 2.0)

	return rand.NewZipf(randObj, zipfS, zipfV, uint64(math.Round(float64(testSize)*zipfFactor)))
}

// cache helpers

func makeMap(size int) map[string][]byte {
	return make(map[string][]byte, size)
}

func initSCache(entries int) *scache.Cache {
	// since SCache allocates 2x buffer, divide max entries by 2
	// https://github.com/viant/scache/blob/master/config.go#L33
	cache, _ := scache.New(&scache.Config{
		Shards:     defaultShards,
		MaxEntries: entries / 2,
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
