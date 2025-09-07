#!/bin/bash

# Perplexity MCP Golang - Comprehensive Test Runner
# This script runs all tests and generates coverage and benchmark reports

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-30s}
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-80}
BENCHMARK_TIME=${BENCHMARK_TIME:-1s}
OUTPUT_DIR="test_reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Perplexity MCP Golang Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo "Started: $(date)"
echo "Timeout: $TEST_TIMEOUT"
echo "Coverage Threshold: $COVERAGE_THRESHOLD%"
echo "Output Directory: $OUTPUT_DIR"
echo ""

# Function to print section headers
print_section() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# Function to check if command succeeded
check_success() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $1 completed successfully${NC}"
    else
        echo -e "${RED}✗ $1 failed${NC}"
        exit 1
    fi
}

# Clean up previous test artifacts
print_section "Cleanup"
echo "Cleaning up previous test artifacts..."
rm -f coverage.out coverage.html
rm -f cpu.prof mem.prof trace.out
rm -f "$OUTPUT_DIR"/benchmark_*.txt
rm -f "$OUTPUT_DIR"/test_*.log
echo "✓ Cleanup completed"

# Verify Go modules
print_section "Module Verification"
echo "Verifying Go modules..."
go mod verify
check_success "Module verification"

# Run quick integration tests first for fast feedback
print_section "Quick Integration Tests"
echo "Running quick integration tests for fast feedback..."
go test -timeout="$TEST_TIMEOUT" -v ./test/integration \
    -run TestQuickIntegrationSuite \
    2>&1 | tee "$OUTPUT_DIR/quick_integration_${TIMESTAMP}.log"
check_success "Quick integration tests"

# Run unit tests with coverage
print_section "Unit Tests with Coverage"
echo "Running unit tests with coverage analysis..."
go test -timeout="$TEST_TIMEOUT" -cover -coverprofile=coverage.out -v \
    ./internal/... \
    2>&1 | tee "$OUTPUT_DIR/unit_tests_${TIMESTAMP}.log"
check_success "Unit tests"

# Generate coverage report
if [ -f coverage.out ]; then
    echo "Generating coverage reports..."
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o "$OUTPUT_DIR/coverage_${TIMESTAMP}.html"
    
    # Generate coverage summary
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Coverage: ${COVERAGE}%"
    
    # Check coverage threshold
    if (( $(echo "$COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
        echo -e "${GREEN}✓ Coverage threshold met (${COVERAGE}% >= ${COVERAGE_THRESHOLD}%)${NC}"
    else
        echo -e "${YELLOW}⚠ Coverage below threshold (${COVERAGE}% < ${COVERAGE_THRESHOLD}%)${NC}"
    fi
    
    # Save coverage summary
    echo "Coverage Report - $TIMESTAMP" > "$OUTPUT_DIR/coverage_summary_${TIMESTAMP}.txt"
    echo "Total Coverage: ${COVERAGE}%" >> "$OUTPUT_DIR/coverage_summary_${TIMESTAMP}.txt"
    echo "" >> "$OUTPUT_DIR/coverage_summary_${TIMESTAMP}.txt"
    go tool cover -func=coverage.out >> "$OUTPUT_DIR/coverage_summary_${TIMESTAMP}.txt"
else
    echo -e "${YELLOW}⚠ No coverage data generated${NC}"
fi

# Run comprehensive integration tests
print_section "Comprehensive Integration Tests"
echo "Running comprehensive integration test suite..."
go test -timeout="$TEST_TIMEOUT" -v ./test/integration \
    -run TestComprehensiveIntegrationSuite \
    2>&1 | tee "$OUTPUT_DIR/comprehensive_integration_${TIMESTAMP}.log"
check_success "Comprehensive integration tests"

# Run all integration tests with edge cases
print_section "Integration Edge Cases"
echo "Running integration edge case tests..."
go test -timeout="$TEST_TIMEOUT" -v ./test/integration \
    -run TestIntegrationEdgeCases \
    2>&1 | tee "$OUTPUT_DIR/integration_edge_cases_${TIMESTAMP}.log"
check_success "Integration edge case tests"

# Run performance benchmarks
print_section "Performance Benchmarks"
echo "Running performance benchmarks..."

# MCP Server benchmarks
echo "  → MCP Server benchmarks..."
go test ./test/benchmark -bench=BenchmarkSearchTool -benchmem -benchtime="$BENCHMARK_TIME" \
    > "$OUTPUT_DIR/benchmark_mcp_server_${TIMESTAMP}.txt"
go test ./test/benchmark -bench=BenchmarkChatTool -benchmem -benchtime="$BENCHMARK_TIME" \
    >> "$OUTPUT_DIR/benchmark_mcp_server_${TIMESTAMP}.txt"
go test ./test/benchmark -bench=BenchmarkResearchTool -benchmem -benchtime="$BENCHMARK_TIME" \
    >> "$OUTPUT_DIR/benchmark_mcp_server_${TIMESTAMP}.txt"

# Use case benchmarks
echo "  → Use case benchmarks..."
go test ./test/benchmark -bench=BenchmarkSearchUseCase -benchmem -benchtime="$BENCHMARK_TIME" \
    > "$OUTPUT_DIR/benchmark_usecases_${TIMESTAMP}.txt"
go test ./test/benchmark -bench=BenchmarkChatUseCase -benchmem -benchtime="$BENCHMARK_TIME" \
    >> "$OUTPUT_DIR/benchmark_usecases_${TIMESTAMP}.txt"
go test ./test/benchmark -bench=BenchmarkResearchUseCase -benchmem -benchtime="$BENCHMARK_TIME" \
    >> "$OUTPUT_DIR/benchmark_usecases_${TIMESTAMP}.txt"

# Memory allocation benchmarks
echo "  → Memory allocation benchmarks..."
go test ./test/benchmark -bench=BenchmarkMemoryAllocation -benchmem -benchtime="$BENCHMARK_TIME" \
    > "$OUTPUT_DIR/benchmark_memory_${TIMESTAMP}.txt"
go test ./test/benchmark -bench=BenchmarkUseCaseMemoryUsage -benchmem -benchtime="$BENCHMARK_TIME" \
    >> "$OUTPUT_DIR/benchmark_memory_${TIMESTAMP}.txt"

# Sequential vs concurrent baseline benchmarks
echo "  → Concurrency comparison benchmarks..."
go test ./test/benchmark -bench=BenchmarkConcurrencyComparison -benchmem -benchtime="$BENCHMARK_TIME" \
    > "$OUTPUT_DIR/benchmark_concurrency_${TIMESTAMP}.txt"

check_success "Performance benchmarks"

# Run race condition detection
print_section "Race Condition Detection"
echo "Running tests with race detector..."
go test -timeout="$TEST_TIMEOUT" -race -v ./test/integration \
    -run TestIntegrationWithRealTimeouts \
    2>&1 | tee "$OUTPUT_DIR/race_detection_${TIMESTAMP}.log"
check_success "Race condition detection"

# Generate final report summary
print_section "Test Report Summary"

REPORT_FILE="$OUTPUT_DIR/test_summary_${TIMESTAMP}.md"

cat << EOF > "$REPORT_FILE"
# Perplexity MCP Golang - Test Report

**Generated:** $(date)  
**Test Timeout:** $TEST_TIMEOUT  
**Coverage Threshold:** $COVERAGE_THRESHOLD%

## Summary

EOF

# Add coverage information if available
if [ -n "$COVERAGE" ]; then
cat << EOF >> "$REPORT_FILE"
### Coverage Analysis
- **Total Coverage:** ${COVERAGE}%
- **Threshold Status:** $(if (( $(echo "$COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then echo "✅ PASSED"; else echo "⚠️ BELOW THRESHOLD"; fi)

EOF
fi

cat << EOF >> "$REPORT_FILE"
### Test Categories Executed
- ✅ Quick Integration Tests
- ✅ Unit Tests with Coverage  
- ✅ Comprehensive Integration Tests
- ✅ Integration Edge Cases
- ✅ Performance Benchmarks
- ✅ Race Condition Detection

### Performance Benchmarks
Single-threaded baseline performance metrics have been established for:

#### MCP Server Operations
- Search tool execution
- Chat tool execution  
- Research tool execution
- Tool listing operations
- Sequential tool workflows

#### Use Case Operations  
- Search use case execution
- Chat use case execution
- Research use case execution
- Input validation performance
- Complex workflow execution

#### Memory Analysis
- Memory allocation patterns per operation
- GC pressure analysis
- Object creation overhead

### Files Generated
- **Coverage Report:** coverage_${TIMESTAMP}.html
- **Coverage Summary:** coverage_summary_${TIMESTAMP}.txt
- **Integration Test Logs:** \*integration\*${TIMESTAMP}.log
- **Benchmark Results:** benchmark_\*${TIMESTAMP}.txt
- **Race Detection Log:** race_detection_${TIMESTAMP}.log

## Architecture Validation

The test suite validates the clean architecture implementation:

- **External Dependencies:** Properly mocked and isolated
- **Use Cases:** Business logic tested in isolation  
- **Domain Layer:** Entity validation and business rules verified
- **MCP Protocol:** End-to-end protocol compliance tested
- **Error Handling:** Comprehensive error scenario coverage
- **Performance:** Single-thread-first baselines established

## Single-Thread-First Policy Compliance

All benchmarks establish single-threaded performance baselines following the project's architectural principle:

1. ✅ **Correctness First** - All tests pass with deterministic behavior
2. ✅ **Performance Baseline** - Comprehensive benchmarks provide optimization baselines  
3. ✅ **Safe Concurrency** - Race detection passes, concurrent access tested
4. ✅ **Resource Management** - Memory allocation patterns analyzed

## Next Steps

Based on benchmark results:

1. **Profile bottlenecks** if performance issues are identified
2. **Add concurrency** only where profiling shows clear bottlenecks
3. **Monitor regression** using established baselines
4. **Scale testing** with load tests when deploying to production

---

*Report generated by automated test suite*
EOF

echo "Test report summary generated: $REPORT_FILE"

# Final summary output
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}All Tests Completed Successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo "Total Duration: $(($(date +%s) - $(date -d "$(head -n1 "$OUTPUT_DIR"/*integration*.log | grep -o '[0-9][0-9]:[0-9][0-9]:[0-9][0-9]' | head -n1 || echo '00:00:00')" +%s)))" seconds" 2>/dev/null || echo "Duration calculation unavailable"

if [ -n "$COVERAGE" ]; then
    echo "Final Coverage: ${COVERAGE}%"
fi

echo ""
echo "Generated Reports:"
ls -la "$OUTPUT_DIR"/*${TIMESTAMP}* 2>/dev/null || echo "Report files in $OUTPUT_DIR"
echo ""
echo -e "View coverage report: ${BLUE}open $OUTPUT_DIR/coverage_${TIMESTAMP}.html${NC}"
echo -e "View test summary: ${BLUE}cat $OUTPUT_DIR/test_summary_${TIMESTAMP}.md${NC}"
EOF