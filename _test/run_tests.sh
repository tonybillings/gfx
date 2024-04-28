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
        echo "Running tests defined in $file"...
        echo "################################################################################"
        go test -v "$file"
        if [ "$?" != 0 ]; then
            exit 1
        fi
    fi
done

printf "\nAll tests passed!\n"
