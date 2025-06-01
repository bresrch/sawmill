#!/bin/bash

# Logging Library Performance Benchmark Runner
# Compares sawmill against popular Go logging libraries

set -e

echo "🚀 Running Logging Library Performance Benchmarks"
echo "=================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# Ensure we're in the right directory
if [ ! -f "benchmark_test.go" ]; then
    echo -e "${RED}Error: benchmark_test.go not found. Run this script from the benchmark directory.${NC}"
    exit 1
fi

echo -e "${BLUE}Building and preparing benchmarks...${NC}"
go mod tidy

# Build the main sawmill package from parent directory
cd .. && go build . && cd benchmark || {
    echo -e "${RED}Error: Failed to build sawmill package${NC}"
    exit 1
}

echo -e "${GREEN}✓ Build successful${NC}"
echo ""

# Function to run a specific benchmark and format output
run_benchmark() {
    local bench_name="$1"
    local description="$2"
    
    echo -e "${PURPLE}📊 Running: $description${NC}"
    echo -e "${CYAN}Benchmark: $bench_name${NC}"
    echo ""
    
    go test -bench="^$bench_name$" -benchmem -count=3 -cpu=1,2,4 \
        -benchtime=1s -timeout=10m \
        | grep -E "(^Benchmark|^PASS|^FAIL)" \
        | sed 's/^Benchmark/  Benchmark/' \
        || echo -e "${RED}Failed to run $bench_name${NC}"
    
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Function to run allocation-focused benchmarks
run_alloc_benchmark() {
    local bench_name="$1"
    local description="$2"
    
    echo -e "${PURPLE}🔍 Running: $description${NC}"
    echo -e "${CYAN}Benchmark: $bench_name (Memory Focus)${NC}"
    echo ""
    
    go test -bench="^$bench_name$" -benchmem -count=5 \
        -benchtime=100000x -timeout=10m \
        | grep -E "(^Benchmark|^PASS|^FAIL)" \
        | sed 's/^Benchmark/  Benchmark/' \
        || echo -e "${RED}Failed to run $bench_name${NC}"
    
    echo ""
    echo "----------------------------------------"
    echo ""
}

echo -e "${YELLOW}🎯 Performance Benchmark Suite${NC}"
echo ""

# Simple message logging
run_benchmark "BenchmarkSimpleMessage" "Simple Message Logging"

# Structured logging with multiple fields
run_benchmark "BenchmarkStructuredLogging" "Structured Logging (5 fields)"

# Complex struct logging
run_benchmark "BenchmarkComplexStructLogging" "Complex Struct Logging"

# High-frequency logging
run_benchmark "BenchmarkHighFrequency" "High-Frequency Debug Logging"

# Disabled level performance
run_benchmark "BenchmarkDisabledLevel" "Disabled Log Level Performance"

# Concurrent logging
run_benchmark "BenchmarkConcurrent" "Concurrent Logging"

echo -e "${YELLOW}💾 Memory Allocation Analysis${NC}"
echo ""

# Memory allocation analysis
run_alloc_benchmark "BenchmarkAllocations" "Memory Allocation Analysis"

echo -e "${GREEN}✅ All benchmarks completed!${NC}"
echo ""

echo -e "${BLUE}📈 Benchmark Summary${NC}"
echo "===================="
echo ""
echo "The benchmarks compare sawmill against:"
echo "• Standard library log package"
echo "• Go 1.21+ slog (structured logging)"
echo "• Logrus (structured logging with hooks)"
echo "• Zap (high-performance logging)"
echo "• Zap Sugar (convenience wrapper)"
echo ""
echo "Key metrics:"
echo "• ns/op: Nanoseconds per operation (lower is better)"
echo "• B/op: Bytes allocated per operation (lower is better)"  
echo "• allocs/op: Number of allocations per operation (lower is better)"
echo ""
echo "For detailed analysis, run individual benchmarks:"
echo "  go test -bench=BenchmarkSimpleMessage -benchmem -v"
echo "  go test -bench=BenchmarkStructuredLogging -benchmem -v"
echo "  go test -bench=BenchmarkComplexStructLogging -benchmem -v"
echo ""
echo -e "${CYAN}Pro tip: Use -cpuprofile and -memprofile for detailed profiling${NC}"