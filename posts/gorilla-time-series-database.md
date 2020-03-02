+++
slug = "gorilla-time-series-database"
date = 2019-11-21
visibility = "draft"
+++

# Gorilla Time Series Database

Gorilla is an in-memory, time series database from Facebook optimized for writes, reading data in a few milliseconds, and high availability. At its core, Gorilla is a 26-hour write-through cache backed by durable storage in HBase. Gorilla optimizes for four attributes:

1.  High data insertion rate. The primary requirement is that Gorilla should always be available to take writes. The expected insertion rate is 10M timestamps with an associated 64 bit float value pairs per second.

2.  Real-time monitoring to show new data within tens of seconds.

3.  Reads in under one millisecond and fast scans over all in-memory data in tens of milliseconds.

4.  Reliability requirements. Gorilla always serves reads even if a server crashes or when an entire region fails.

Gorilla compromises on the following attributes:

- Flexibility. The only supported data type is a named stream with 64 bit floating point values. Higher level abstractions can be built on top of Gorilla.

- Duration. Gorilla only stores the last 26 hours of data.

- Granularity. The minimum granularity is 15 seconds.

- Durability. A server crash can cause data loss of up to 64kB which is 1-2 seconds of data. During prolonged outages, Gorilla preserves the most recent 1 minute of data and discards the rest of the data.

- Consistency. Data is streamed between datacenter regions without attempting to guarantee consistency.

- Query flexibility - Gorilla serves raw compressed blocks to clients. There’s no query engine in Gorilla so clients are expected to process the compressed blocks.

Gorilla’s contributions include a novel, streaming time stamp compression scheme.

## Time stamp compression

Gorilla introduces a novel lossless compression scheme for streaming timestamps. Gorilla’s timestamp encoding is based on the observation that the vast majority of timestamps arrive at a fixed interval. Using sampled production data, the Gorilla team found 96% of timestamps compress to a single bit. Compressing to a single bit implies that 96% of metrics arrive on a fixed schedule. This is surprising because I’d expect most metrics to be request driven and not adhere to a fixed schedule.

Each stream is divided into blocks aligned at two-hour intervals. The block header contains a 64 bit timestamp of the beginning of the block, e.g. `2019-09-05T02:00`. The first timestamp is the delta from the header timestamp stored in 14 bits. Using 14 bits allows one second granularity within the two hour window. Subsequent timestamps are encoded with a delta of deltas scheme.

```bash
# Block timestamp aligned to 8:00:00.
timestamps =   [8:00:30, 8:01:30, 8:02:30, 8:03:28]
deltas =       [     30,      60,      60,      57]
delta_deltas = [       ,      30,       0,      -4]
```

The `delta_deltas` are encoded using a variable sized integer encoding.

- If the delta is zero, store a single `0` bit.
- If the delta is in `[-63, 64]` store `0b10` followed by the signed value in 7 bits.
- If the delta is in `[-255, 256]` store `0b110` followed by the signed value in 9 bits.
- If the delta is in `[-2047, 2048]` store `0b1110` followed by the signed value in 12 bits.
- Otherwise, store `0b1111` followed by the delta in 32 bits.

The example above can be represented as:

```bash
Block header: Timestamp at 08:00:30
14 bits: 1st Timestamp delta: 30
9 bits: 0b10 + binary(30)
1 bit: 0
9 bits: 0b10 + binary(-4)
```

### Late arriving time stamps

Gorilla allows out of order timestamps by supporting signed integers. I couldn’t figure out what happens to severely out-of-order timestamps like if a timestamp is 2 hours late.

### Alternative timestamp schemes

Since Gorilla is willing to drop data, there’s interesting optimization opportunities if Gorilla required sorted timestamps. Sorted timestamps are equivalent to a postings list, also known as a reverse index, which has 50 years of research of fast compression strategies.

The delta-of-delta scheme is appropriate when the data arrives at a fixed interval. For instance, a service might log a single point every 60 seconds. For other uses cases, like user-generated timestamps, it’s not clear that the delta-of-delta compression is superior to a single delta approach.

Recent research takes advantage of SIMD and avoids the branchy code of variable width timestamps.

TODO: add blurb on fast postings list compression.

## Time series value compression

The compression scheme for time stamp values takes advantage of the fact that most timestamp values don’t change significantly compared to neighboring values. 59% of values are identical to the previous value and compress to a single bit. If values are close, `xor` compression will drop the sign, exponent, and first few bits of the mantissa.

Since the encoding is variable length, the entire two hour block must be decoded to access values. This isn’t a problem for time series databases because the value of the data is in aggregation, not in single points.

## Sharding

A Paxos-based system called _ShardManager_ assigns shards to nodes. I think each time series is contained by a single shard. It’s unclear if Gorilla mitigates hotspots that might occur for frequent metrics like response time per Facebook page request.

## Data structures

The in-memory organization of Gorilla is a two-level map:
1\. _TSMap_ is the first level map from a shard ID to a time series map.
2\. The _TSMap_ maps a string name to a `TimeSeries` data structure.

The `TimeSeries` data structure is a collection of closed blocks containing historical data and a single open block containing the previous two hours of data. Upon receiving a query:
1\. The Gorilla node checks the Shard ID map to get the _TSMap_. If the value is null, this node doesn’t own the shard.
2\. Next, the _TSMap_ is read-locked and the node copies the pointer to the `timeSeries`.
3\. The node spin-locks the `TimeSeries` to copy the data and returns the raw blocks to the client.

Write path:
1\. Presumably, the Gorilla write client gets the correct node from the Shard Manager and gets a new node if the node dies.
2\. The Gorilla node looks up time series name in key-list file to get the Shard ID for the time series name.
3\. The Gorilla node streams the new value into the open block in the `TimeSeries` data structure using the compression described above.
4\. The Gorilla node writes the compressed value to the append-only log file. The file buffer is flushed every 64kB so it’s not a write-ahead log. Since a shard has many time series, each time stamp-value pair is tagged with the 32-bit index.
5\. After two hours, the Gorilla node closes all open blocks and flushes each one to disk with a corresponding checkpoint file. After all `TimeSeries` for a shard are flushed, the Gorilla node deletes the append-only log for that shard.

The Shar
