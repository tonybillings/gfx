#!/bin/bash

# Run this script instead of "go test -v ." to avoid an import cycle issue.
# Tests are stored in the _test package to prevent access to non-exported
# objects, just as experienced by importers of the gfx package.  For some
# reason, explicitly importing the _test package from within the _test
# package is sometimes necessary, which can cause this import cycle issue
# if running "go test -v .".  For now, this is the recommended way to test.
echo "Skipping benchmark tests!"
for file in *_test.go; do
    if [ -f "$file" ]; then
        echo "################################################################################"
        echo "Searching for tests defined in $file"...
        echo "################################################################################"

        test_ran=0
        while read -r testname; do
            test_name=${testname%(*}

            # TODO: Invoking each test individually until an issue
            # relating to the use of a second GLFW window to render
            # objects (after destroying the first window) is resolved.
            echo ">>> Running $test_name* from $file..."
            go test -v "$file" -test.run="$test_name"
            if [ "$?" != 0 ]; then
                echo "Test failed: $test_name"
                exit 1
            fi

            test_ran=1
        done < <(grep -oP 'func \KTest\w+\(t \*testing\.T\)' "$file")

        if [ "$test_ran" == 0 ]; then
            echo "!!! No non-benchmarking tests found in this file, skipping..."
        fi
    fi
done

printf "\nAll tests passed!\n"
