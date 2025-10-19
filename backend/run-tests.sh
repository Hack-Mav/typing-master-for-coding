#!/bin/bash
# Test runner script for Typing Master Backend
# This script runs all unit tests with coverage reporting

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default values
PACKAGE="./..."
VERBOSE=""
COVERAGE=""
RACE=""
SHORT=""
RUN=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -c|--coverage)
            COVERAGE="true"
            shift
            ;;
        -r|--race)
            RACE="-race"
            shift
            ;;
        -s|--short)
            SHORT="-short"
            shift
            ;;
        --run)
            RUN="-run $2"
            shift 2
            ;;
        -p|--package)
            PACKAGE="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${CYAN}========================================"
echo -e "Typing Master Backend - Test Suite"
echo -e "========================================${NC}"
echo ""

# Build test command
TEST_CMD="go test $PACKAGE $VERBOSE $RACE $SHORT $RUN"

if [ "$COVERAGE" = "true" ]; then
    TEST_CMD="$TEST_CMD -coverprofile=coverage.out -covermode=atomic"
fi

# Run tests
echo -e "${YELLOW}Running tests...${NC}"
echo -e "${CYAN}Command: $TEST_CMD${NC}"
echo ""

export GO_ENV=test

if eval $TEST_CMD; then
    echo ""
    echo -e "${GREEN}✓ All tests passed!${NC}"
    
    # Generate coverage report if requested
    if [ "$COVERAGE" = "true" ] && [ -f "coverage.out" ]; then
        echo ""
        echo -e "${YELLOW}Generating coverage report...${NC}"
        
        # Display coverage summary
        go tool cover -func=coverage.out
        
        # Generate HTML coverage report
        go tool cover -html=coverage.out -o coverage.html
        echo ""
        echo -e "${GREEN}✓ Coverage report generated: coverage.html${NC}"
        
        # Calculate total coverage
        COVERAGE_TOTAL=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}')
        echo ""
        echo -e "${CYAN}Coverage Summary:${NC}"
        echo -e "${NC}Total Coverage: $COVERAGE_TOTAL${NC}"
    fi
    
    echo ""
    echo -e "${CYAN}========================================"
    echo -e "Test run completed"
    echo -e "========================================${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Tests failed!${NC}"
    echo ""
    echo -e "${CYAN}========================================"
    echo -e "Test run completed with errors"
    echo -e "========================================${NC}"
    exit 1
fi
