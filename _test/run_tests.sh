#!/bin/bash

# doing this instead of "go test -v ." to avoid import cycle issue
echo "Skipping benchmark tests!"
for file in *_test.go; do
    if [ -f "$file" ]; then
        echo "Running tests for $file"
        go test -v "$file"
    fi
done
