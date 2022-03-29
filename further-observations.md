# Further observations on cache implementations

The cache implementations seem to mainly be incremental improvements upon other inspirations. 
The reason why there are so many cache implementations (instead of one mega-implementation) are beyond me.

# Common implemmenation differentiators

## Lock scope 

There are multiple implementations to support concurrency and prevent the lock acquisition from being the main bottleneck. 
The primary implementation to avoid lock contention seems to be sub-partitioned "cache blocks" - effectively a specific lock that exists for a particular modulus of a key hash. 
In naive pseudocode, `locks[hash(key) % partitions]` would get the lock for the particular partition.

## Eviction policy and memory limitation

As caches are designed to be a memoization implementation, they are designed to sacrifice memory for speed.
Since memory is finite, the amount of calculations that can be cached is limited, and thus the algorithm used to determine what to keep (or similarly remove / evict) in the cache is important to the cache product.
The most common cache eviction policies is a Least Recently Used (LRU).
This will keep track of the usages of keys and will evict the oldest, unused key when memory needs to be freed.
There are many ways to implement the LRU policy, from naive, native maps to ring buffers.

## Cache value typing and garbage collection sweep avoidance

Caches can store any arbitrary value to prevent a long procedure from having to run again.
As Go is a language that uses a garbage collector, if there are objects in memory from the cache, the garbage collector will check those objects in memory for more references to objects that can be collected.
One way to avoid traversing a potentially complex object graph is to store simple data values (that should not incur a cost during a garbage collection), and a way to store complex objects in simple data values is to serialize the data.
There are certain guarantees that may be unserializable, especially depending on the format of the serialization.
For example, JSON serialization cannot natively serialize cyclical object graphs.

# Other notes

## Distributed caching

Note that distributed caching is outside the scope of this research.
This is specific to [`groupcache`](https://github.com/golang/groupcache), which the Hashicorp LRU cache states as their inspiration.
This has a unique miss-handling mechanism, where a specific cache instance is held responsible for tracking whether or not the specific key has a cache value or not, and cache instances not responsible are aware of how to ask the responsible cache instance for the truth of the cache value availability.


