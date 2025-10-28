# Performance Testing Guide

This guide explains how to measure and compare the performance of sequential vs parallel scraping in the go-scraper project.

## Quick Start

The scraper now automatically tracks and displays performance metrics every time you run it.

### Basic Usage

**Sequential scraping:**
```bash
./scraper -states=IL,IN
```

**Parallel scraping:**
```bash
./scraper -states=IL,IN -concurrent
```

## Performance Metrics Output

After each run, you'll see detailed metrics like this:

```
=== Performance Metrics ===
Execution mode: Sequential
Total duration: 45.234s
States scraped: 2
Parks collected: 150
Average time per state: 22.617s
Average time per park: 301.56ms

Per-state timings:
  IL: 23.456s (75 parks)
  IN: 21.778s (75 parks)
```

## Metrics Explained

- **Total duration**: Complete end-to-end time for scraping
- **States scraped**: Number of states processed
- **Parks collected**: Total parks successfully scraped
- **Average time per state**: Total duration / number of states
- **Average time per park**: Total duration / number of parks
- **Per-state timings**: Individual timing for each state

## Comparing Sequential vs Parallel

### Method 1: Manual Comparison

Run both modes and compare the output:

```bash
# Sequential
./scraper -states=IL,IN > sequential.log

# Parallel
./scraper -states=IL,IN -concurrent > parallel.log
```

### Method 2: Using Go Benchmarks

Run formal benchmarks with statistical analysis:

```bash
# Benchmark sequential (2 states)
go test -bench=BenchmarkSequential -benchtime=1x

# Benchmark parallel (2 states)
go test -bench=BenchmarkParallel -benchtime=1x

# Run both and compare
go test -bench=. -benchtime=1x
```

**Note:** Benchmarks will make real network requests. Use `-benchtime=1x` to run only once.

### Method 3: Memory Profiling

Compare memory usage between modes:

```bash
# Sequential with memory stats
go test -bench=BenchmarkSequential -benchtime=1x -benchmem

# Parallel with memory stats
go test -bench=BenchmarkParallel -benchtime=1x -benchmem
```

## Advanced: CPU Profiling with pprof

To find performance bottlenecks:

1. Add profiling to your code (in `scraper.go`):
```go
import (
    "runtime/pprof"
    "os"
)

// Before scraping
f, _ := os.Create("cpu.pprof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
```

2. Run the scraper:
```bash
./scraper -states=IL,IN
```

3. Analyze the profile:
```bash
go tool pprof cpu.pprof
```

4. Useful pprof commands:
   - `top` - Show top CPU consumers
   - `list <function>` - Show line-by-line CPU usage
   - `web` - Generate visual graph (requires graphviz)

## Expected Performance Characteristics

### Sequential Scraping
- **Pros:** Simpler, more predictable, easier to debug
- **Cons:** Slower total time (states processed one-by-one)
- **Best for:** Single state, debugging, respecting rate limits

### Parallel Scraping
- **Pros:** Faster total time (states processed simultaneously)
- **Cons:** Higher memory usage, more complex error handling
- **Best for:** Multiple states, production runs

## Tips for Accurate Measurements

1. **Run multiple times** - Network variability affects results
2. **Use consistent test data** - Same states for comparison
3. **Consider network conditions** - Run tests during similar times
4. **Account for rate limiting** - Check your `config.json` requestDelay
5. **Monitor system resources** - Use `htop` or Activity Monitor
6. **Close other network apps** - Reduce interference

## Example Comparison Workflow

```bash
# 1. Test with 2 states
echo "Sequential (2 states):"
./scraper -states=IL,IN | grep "Total duration"

echo "Parallel (2 states):"
./scraper -states=IL,IN -concurrent | grep "Total duration"

# 2. Test with all states
echo "Sequential (all states):"
./scraper | grep "Total duration"

echo "Parallel (all states):"
./scraper -concurrent | grep "Total duration"

# 3. Run benchmarks
go test -bench=. -benchtime=1x -benchmem
```

## Logging Levels

Control log verbosity in `config.json`:

```json
{
  "logLevel": "info",  // Options: "debug", "info", "warn", "error"
  "requestDelay": 2
}
```

- `debug` - Detailed timing for every request
- `info` - High-level state completion messages
- `error` - Only errors

Set to `error` for cleaner performance testing output.
