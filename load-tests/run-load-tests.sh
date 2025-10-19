#!/bin/bash

# Load Testing Runner Script
# Runs various load test scenarios using K6

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
RESULTS_DIR="./results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create results directory
mkdir -p "$RESULTS_DIR"

echo "Starting load tests against $BASE_URL"
echo "Results will be saved to $RESULTS_DIR"
echo ""

# Function to run a test
run_test() {
  local test_name=$1
  local test_file=$2
  local output_file="$RESULTS_DIR/${test_name}_${TIMESTAMP}"
  
  echo "Running $test_name..."
  k6 run \
    --out json="$output_file.json" \
    --out influxdb=http://localhost:8086/k6 \
    --summary-export="$output_file.summary.json" \
    "$test_file"
  
  echo "$test_name completed. Results saved to $output_file"
  echo ""
}

# 1. Smoke Test - Quick validation
echo "=== Smoke Test ==="
run_test "smoke" "k6-load-test.js" --vus 1 --duration 30s

# 2. Load Test - Normal load (5k concurrent users)
echo "=== Load Test (5k users) ==="
run_test "load" "k6-load-test.js"

# 3. Stress Test - Beyond normal capacity
echo "=== Stress Test (15k users) ==="
run_test "stress" "k6-stress-test.js"

# 4. Spike Test - Sudden traffic spike
echo "=== Spike Test ==="
run_test "spike" "k6-spike-test.js"

# 5. Soak Test - Extended duration
echo "=== Soak Test (1 hour) ==="
run_test "soak" "k6-soak-test.js"

# Generate HTML report
echo "Generating HTML report..."
k6-reporter "$RESULTS_DIR/load_${TIMESTAMP}.json" \
  --output "$RESULTS_DIR/report_${TIMESTAMP}.html"

echo ""
echo "All tests completed!"
echo "View results at: $RESULTS_DIR/report_${TIMESTAMP}.html"
