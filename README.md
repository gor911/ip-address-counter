# IP Uniqueness Counter
Count hundreds of gigabytes of IPv4 addresses in under a minute with just 512 MiB of RAM.
The program prints the total number of unique IPs and exits.

# Task
This tool solves the [IP Addr Counter](https://github.com/Ecwid/new-job/blob/master/IP-Addr-Counter-GO.md) GO challenge.

## Features
- **Fixed-memory bitset**: 2³² bits packed into 1<<26 `uint64` words (512 MiB total)
- **Chunk-based reading**: Processes the input file in constant 10 MiB slices
- **High-throughput workers**: Parses and sets bits in parallel using an atomic CAS loop
- **Zero-per-line allocations**: In-place parsing of ASCII IPs, no temporary strings or maps
- **Fast popcount**: Uses hardware-assisted `math/bits.OnesCount64` to count set bits efficiently

## Performance
- Takes about the same time as a pure chunk-read approach.
- Processes a 114GB file under 50 seconds on an M4 Mac mini.

## Installation & Usage
```bash
# Build the executable
go build -o ipcounter

# Run against a large file
./ipcounter resources/ip_addresses
