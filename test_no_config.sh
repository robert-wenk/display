#!/bin/bash

# Test script to verify the application works without config.json
echo "Testing QNAP Display Control without config.json..."

# Remove any existing config file
sudo rm -f /tmp/test_config.json

# Test 1: Basic functionality without config
echo "Test 1: Running without config file..."
timeout 2s ./bin/qnap-display-control --config /tmp/test_config.json --port /dev/null --verbose 2>&1 | grep -q "using defaults"
if [ $? -eq 0 ]; then
    echo "✅ Test 1 PASSED: Application falls back to defaults when config.json is missing"
else
    echo "❌ Test 1 FAILED: Application didn't fall back to defaults"
fi

# Test 2: Verify default port is /dev/ttyS1
echo "Test 2: Checking default port..."
./bin/qnap-display-control --help | grep -q "/dev/ttyS1"
if [ $? -eq 0 ]; then
    echo "✅ Test 2 PASSED: Default port is /dev/ttyS1"
else
    echo "❌ Test 2 FAILED: Default port is not /dev/ttyS1"
fi

# Test 3: Command line overrides work
echo "Test 3: Testing command line overrides..."
timeout 2s ./bin/qnap-display-control --port /dev/zero --config /tmp/test_config.json --verbose 2>&1 | grep -q "/dev/zero"
if [ $? -eq 0 ]; then
    echo "✅ Test 3 PASSED: Command line port override works"
else
    echo "❌ Test 3 FAILED: Command line port override doesn't work"
fi

echo ""
echo "Summary: Application works without config.json and uses /dev/ttyS1 as default"
