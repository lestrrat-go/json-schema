#!/bin/bash

# Initialize JSON Schema Test Suite
# This script clones the official JSON Schema Test Suite if it doesn't already exist
# Usage: ./init-test-suite.sh [COMMIT_ID]

TEST_DIR="tests"
REPO_URL="https://github.com/json-schema-org/JSON-Schema-Test-Suite.git"
COMMIT_ID="${1:-}"

# Check if the directory already exists
if [ ! -d "$TEST_DIR" ]; then
    echo "Cloning JSON Schema Test Suite..."
    if git clone "$REPO_URL" "$TEST_DIR"; then
        echo "Successfully cloned test suite to $TEST_DIR"
        
        # Checkout specific commit if provided
        if [ -n "$COMMIT_ID" ]; then
            echo "Checking out commit $COMMIT_ID..."
            cd "$TEST_DIR" && git checkout "$COMMIT_ID"
            if [ $? -eq 0 ]; then
                echo "Successfully checked out commit $COMMIT_ID"
            else
                echo "Failed to checkout commit $COMMIT_ID" >&2
                exit 1
            fi
        fi
    else
        echo "Failed to clone test suite" >&2
        exit 1
    fi
else
    echo "Test suite directory already exists at $TEST_DIR"
    
    # If commit ID is provided, ensure we're on the right commit
    if [ -n "$COMMIT_ID" ]; then
        cd "$TEST_DIR"
        CURRENT_COMMIT=$(git rev-parse HEAD)
        if [ "$CURRENT_COMMIT" != "$COMMIT_ID" ]; then
            echo "Updating to commit $COMMIT_ID..."
            git fetch origin && git checkout "$COMMIT_ID"
            if [ $? -eq 0 ]; then
                echo "Successfully updated to commit $COMMIT_ID"
            else
                echo "Failed to checkout commit $COMMIT_ID" >&2
                exit 1
            fi
        else
            echo "Already on commit $COMMIT_ID"
        fi
    fi
fi
